package zrt

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	pb "github.com/ZeroRuntimeAI/zrt-golang-sdk/internal/pb"
)

// Room is the room/meeting config for a dispatched session — the runtime joins
// using this. It lives on the dispatch (per call), not on Serve: one serving
// worker handles many rooms over its lifetime.
//
// All fields are optional. Leave RoomID empty to auto-create a playground room
// (requires ZRT auth via AuthToken / ZRT_AUTH_TOKEN / ZRT_API_KEY+ZRT_SECRET_KEY).
type Room struct {
	RoomID                      string
	AuthToken                   string
	AgentName                   string
	Playground                  *bool // defaults true
	Vision                      bool
	Recording                   bool
	BackgroundAudio             bool
	AudioListenerEnabled        bool
	AutoEndSession              *bool // defaults true
	SessionTimeoutSeconds       *int
	NoParticipantTimeoutSeconds *int
	AgentParticipantID          string
	SignalingBaseURL            string
	SendLogsToDashboard         *bool  // defaults true
	DashboardLogLevel           string // defaults "INFO"
}

// Sip carries SIP connection details for a telephony session. They are mapped
// onto the dispatch metadata (first-class keys win over Extra).
type Sip struct {
	CallTo     string
	CallFrom   string
	CallType   string
	CallID     string
	WebhookURL string
	Extra      map[string]string
}

// DispatchOptions configures DispatchAgent.
type DispatchOptions struct {
	// Room is the room/meeting config (defaults to an auto-created playground room).
	Room *Room
	// Sip carries SIP connection details for telephony sessions.
	Sip *Sip
	// Labels constrain dispatch to registered workers whose labels match.
	Labels map[string]string
	// Metadata is arbitrary key/value dispatch metadata (stringified).
	Metadata map[string]string
	// SessionID is an optional caller-chosen session id (runtime mints one if empty).
	SessionID string
	// RuntimeAddress is the runtime host:port (defaults to $ZRT_RUNTIME_ADDRESS,
	// then "localhost:50051").
	RuntimeAddress string
	// Timeout is the RPC timeout (defaults to 30s).
	Timeout time.Duration
}

// DispatchResult is returned when a dispatch is accepted.
type DispatchResult struct {
	SessionID string
	AgentID   string
}

// DispatchAgent dispatches a session to a registered agent and returns its ids.
//
// It targets the agent registered under agentID (the AgentID passed to Serve),
// hands it a room (auto-created if Room.RoomID is empty and auth is available)
// plus any SIP / per-call config, and returns {SessionID, AgentID} on
// acceptance. This is a one-shot client call — run it from anywhere (a script,
// a CLI, a web handler).
//
// It returns an error if agentID is empty, if a room is required but cannot be
// created, or if the runtime rejects the dispatch (e.g. no agent available).
func DispatchAgent(agentID string, opts DispatchOptions) (*DispatchResult, error) {
	if agentID == "" {
		return nil, fmt.Errorf("zrt.DispatchAgent: agentID is required (the AgentID you passed to Serve())")
	}

	room := opts.Room
	if room == nil {
		room = &Room{}
	}
	addr := cmp.Or(opts.RuntimeAddress, os.Getenv("ZRT_RUNTIME_ADDRESS"), "localhost:50051")
	timeout := cmp.Or(opts.Timeout, 30*time.Second)

	// Resolve the room: explicit > env > auto-create.
	roomID := cmp.Or(room.RoomID, strings.TrimSpace(os.Getenv("ZRT_ROOM_ID")))

	token, tokErr := ResolveAuthToken(room.AuthToken)
	if tokErr != nil {
		if roomID == "" {
			return nil, fmt.Errorf("zrt.DispatchAgent needs Room.RoomID, or ZRT auth "+
				"(Room.AuthToken / ZRT_AUTH_TOKEN / ZRT_API_KEY + ZRT_SECRET_KEY) to auto-create one: %w", tokErr)
		}
		token = ""
	}

	if roomID == "" {
		created, err := createRoomStatic(token, room.SignalingBaseURL)
		if err != nil {
			return nil, fmt.Errorf("zrt.DispatchAgent: auto-create room failed: %w", err)
		}
		roomID = created
	}

	// dispatch metadata: arbitrary metadata + sip.Extra + first-class SIP (first-class wins).
	dispatchMetadata := map[string]string{}
	for k, v := range opts.Metadata {
		dispatchMetadata[k] = v
	}
	if opts.Sip != nil {
		for k, v := range opts.Sip.Extra {
			dispatchMetadata[k] = v
		}
		for _, kv := range []struct{ key, val string }{
			{"sipCallTo", opts.Sip.CallTo},
			{"sipCallFrom", opts.Sip.CallFrom},
			{"callType", opts.Sip.CallType},
			{"callId", opts.Sip.CallID},
			{"webhook_url", opts.Sip.WebhookURL},
		} {
			if kv.val != "" {
				dispatchMetadata[kv.key] = kv.val
			}
		}
	}

	roomPB := &pb.RoomConfig{
		RoomId:                 roomID,
		AuthToken:              token,
		AgentName:              cmp.Or(room.AgentName, agentID),
		Playground:             BoolOr(room.Playground, true),
		Vision:                 room.Vision,
		RecordingEnabled:       room.Recording,
		BackgroundAudioEnabled: room.BackgroundAudio,
		AudioListenerEnabled:   room.AudioListenerEnabled,
		AutoEndSession:         BoolOr(room.AutoEndSession, true),
		SendLogsToDashboard:    BoolOr(room.SendLogsToDashboard, true),
		DashboardLogLevel:      cmp.Or(room.DashboardLogLevel, "INFO"),
		AgentParticipantId:     room.AgentParticipantID,
	}
	if room.SessionTimeoutSeconds != nil {
		roomPB.SessionTimeoutSeconds = uint32(*room.SessionTimeoutSeconds)
	}
	if room.NoParticipantTimeoutSeconds != nil {
		roomPB.NoParticipantTimeoutSeconds = uint32(*room.NoParticipantTimeoutSeconds)
	}
	if room.SignalingBaseURL != "" {
		roomPB.SignalingBaseUrl = room.SignalingBaseURL
	}

	req := &pb.DispatchRequest{
		AgentKind:        agentID,
		Room:             roomPB,
		DispatchMetadata: dispatchMetadata,
		LabelSelector:    copyAnyMap(opts.Labels),
		SessionId:        opts.SessionID,
	}

	// openGRPCChannel honors ZRT_RUNTIME_INSECURE (insecure) vs TLS and injects
	// auth metadata on every call.
	conn, err := openGRPCChannel(addr, token)
	if err != nil {
		return nil, fmt.Errorf("zrt.DispatchAgent: dial runtime %s: %w", addr, err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := pb.NewAgentRuntimeClient(conn).Dispatch(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("zrt.DispatchAgent: Dispatch RPC failed: %w", err)
	}

	if acc := resp.GetAccepted(); acc != nil {
		return &DispatchResult{SessionID: acc.GetSessionId(), AgentID: acc.GetAgentId()}, nil
	}

	rej := resp.GetRejected()
	if rej == nil {
		return nil, fmt.Errorf("zrt.DispatchAgent: dispatch returned no result")
	}
	return nil, fmt.Errorf("zrt.DispatchAgent: dispatch rejected (%s): %s [registered_agents=%d, available_agents=%d]",
		rej.GetReason(), rej.GetMessage(), rej.GetRegisteredAgents(), rej.GetAvailableAgents())
}
