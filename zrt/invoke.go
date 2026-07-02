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

// Room is the per-invocation room/meeting config. All fields are optional; leave
// RoomID empty to auto-create a playground room (requires ZRT auth via AuthToken /
// ZRT_AUTH_TOKEN / ZRT_API_KEY+ZRT_SECRET_KEY).
type Room struct {
	// RoomID is the target room/meeting id. Empty auto-creates a playground room.
	RoomID string
	// AuthToken is the ZRT auth token; empty falls back to ZRT_AUTH_TOKEN or ZRT_API_KEY+ZRT_SECRET_KEY.
	AuthToken string
	// AgentName is the display name for the agent participant. Defaults to the agentID.
	AgentName string
	// Playground (defaults true) governs whether Invoke returns a clickable
	// PlaygroundURL. It is NOT sent on the wire — the runtime reserves this field.
	Playground *bool
	// Vision enables camera/video input for the session.
	Vision bool
	// Recording enables session recording.
	Recording bool
	// BackgroundAudio enables background audio playback for the session.
	BackgroundAudio bool
	// AudioListenerEnabled enables the raw audio listener stream.
	AudioListenerEnabled bool
	AutoEndSession       *bool // defaults true
	// SessionTimeoutSeconds is the max session lifetime in seconds; nil = runtime default.
	SessionTimeoutSeconds *int
	// NoParticipantTimeoutSeconds ends the session after this many seconds with no participant; nil = runtime default.
	NoParticipantTimeoutSeconds *int
	// AgentParticipantID is an optional fixed participant id for the agent.
	AgentParticipantID string
	// SignalingBaseURL overrides the signaling server base URL used for room creation.
	SignalingBaseURL    string
	SendLogsToDashboard *bool  // defaults true
	DashboardLogLevel   string // defaults "INFO"
}

// Sip carries SIP connection details for a telephony session, mapped onto the
// dispatch metadata (first-class keys win over Extra).
type Sip struct {
	// CallTo is the destination phone number/SIP URI (sipCallTo).
	CallTo string
	// CallFrom is the originating phone number/SIP URI (sipCallFrom).
	CallFrom string
	// CallType is the call type (e.g. "sip"); maps to callType.
	CallType string
	// CallID is the telephony call id (callId).
	CallID string
	// WebhookURL is the callback URL for call events (webhook_url).
	WebhookURL string
	// Extra holds additional dispatch metadata keys; first-class fields above win on conflict.
	Extra map[string]string
}

// InvokeOptions configures Invoke.
type InvokeOptions struct {
	// Room is the room/meeting config (defaults to an auto-created playground room).
	Room *Room
	// Sip carries SIP connection details for telephony sessions.
	Sip *Sip
	// Labels constrain the invocation to registered workers whose labels match.
	Labels map[string]string
	// Metadata is arbitrary key/value dispatch metadata (stringified).
	Metadata map[string]string
	// Recording, when set, overrides the recording config for this session and
	// forces recording on (equivalent to Room.Recording=true plus custom output
	// settings). It maps onto the dispatch RecordingOverride.
	Recording *RecordingConfig
	// SessionID is an optional caller-chosen session id (runtime mints one if empty).
	SessionID string
	// RuntimeAddress is the runtime host:port (defaults to $ZRT_RUNTIME_ADDRESS,
	// then "us1.rt.zeroruntime.ai:443").
	RuntimeAddress string
	// Timeout is the RPC timeout (defaults to 30s).
	Timeout time.Duration
}

// InvokeResult is returned when an invocation is accepted.
type InvokeResult struct {
	// SessionID is the id of the started session.
	SessionID string
	// WorkerID is the id of the worker that accepted the invocation.
	WorkerID string
	// RoomID is the room the session joined (resolved or auto-created).
	RoomID string
	// PlaygroundURL is a clickable link to join the session, present when
	// Room.Playground is set (the default) and auth is available.
	PlaygroundURL string
}

// Invoke starts a session for the agent registered under agentID (the AgentID
// passed to Serve) and returns {SessionID, WorkerID, RoomID} on acceptance (plus
// PlaygroundURL when Room.Playground is set and auth is available). The room is
// auto-created when Room.RoomID is empty and auth is available. This is a one-shot
// client call.
//
// It returns an error if agentID is empty, if a room is required but cannot be
// created, or if the runtime rejects the invocation.
func Invoke(agentID string, opts InvokeOptions) (*InvokeResult, error) {
	if agentID == "" {
		return nil, fmt.Errorf("zrt.Invoke: agentID is required (the AgentID you passed to Serve())")
	}

	room := opts.Room
	if room == nil {
		room = &Room{}
	}
	addr := cmp.Or(opts.RuntimeAddress, os.Getenv("ZRT_RUNTIME_ADDRESS"), "us1.rt.zeroruntime.ai:443")
	timeout := cmp.Or(opts.Timeout, 30*time.Second)

	// Resolve the room: explicit > env > auto-create.
	roomID := cmp.Or(room.RoomID, strings.TrimSpace(os.Getenv("ZRT_ROOM_ID")))

	token, tokErr := ResolveAuthToken(room.AuthToken)
	if tokErr != nil {
		if roomID == "" {
			return nil, fmt.Errorf("zrt.Invoke needs Room.RoomID, or ZRT auth "+
				"(Room.AuthToken / ZRT_AUTH_TOKEN / ZRT_API_KEY + ZRT_SECRET_KEY) to auto-create one: %w", tokErr)
		}
		token = ""
	}

	if roomID == "" {
		created, err := createRoomStatic(token, room.SignalingBaseURL)
		if err != nil {
			return nil, fmt.Errorf("zrt.Invoke: auto-create room failed: %w", err)
		}
		roomID = created
	}

	// dispatch metadata: Metadata + Sip.Extra + first-class SIP keys (first-class wins).
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
		Vision:                 room.Vision,
		RecordingEnabled:       room.Recording || opts.Recording != nil,
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
		AgentId:          agentID,
		Room:             roomPB,
		DispatchMetadata: dispatchMetadata,
		LabelSelector:    copyAnyMap(opts.Labels),
		SessionId:        opts.SessionID,
	}
	if opts.Recording != nil {
		req.RecordingOverride = buildRecordingConfig(opts.Recording)
	}

	conn, err := openGRPCChannel(addr, token)
	if err != nil {
		return nil, fmt.Errorf("zrt.Invoke: dial runtime %s: %w", addr, err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := pb.NewAgentRuntimeClient(conn).Dispatch(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("zrt.Invoke: Dispatch RPC failed: %w", err)
	}

	if acc := resp.GetAccepted(); acc != nil {
		result := &InvokeResult{SessionID: acc.GetSessionId(), WorkerID: acc.GetWorkerId(), RoomID: roomID}
		if BoolOr(room.Playground, true) && roomID != "" && token != "" {
			base := strings.TrimRight(cmp.Or(os.Getenv("ZRT_PLAYGROUND_URL"), "https://playground.zeroruntime.ai/cli"), "/")
			result.PlaygroundURL = fmt.Sprintf("%s?token=%s&meetingId=%s", base, token, roomID)
		}
		return result, nil
	}

	rej := resp.GetRejected()
	if rej == nil {
		return nil, fmt.Errorf("zrt.Invoke: invocation returned no result")
	}
	return nil, fmt.Errorf("zrt.Invoke: invocation rejected (%s): %s [registered_agents=%d, available_agents=%d]",
		rej.GetReason(), rej.GetMessage(), rej.GetRegisteredAgents(), rej.GetAvailableAgents())
}
