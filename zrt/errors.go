package zrt

import "errors"

// Sentinel errors returned by the SDK. Match them with errors.Is to branch on
// failure modes (for example to drive reconnect or re-auth logic):
//
//	if errors.Is(err, zrt.ErrNoCredentials) {
//		// prompt the user for an auth token
//	}
//
// Errors returned by the SDK wrap these with additional context, so always use
// errors.Is rather than == for comparison.
var (
	// ErrNoCredentials means no usable Zero Runtime credentials were found:
	// set ZRT_AUTH_TOKEN, or ZRT_API_KEY + ZRT_SECRET_KEY, or pass an auth
	// token in RoomOptions.
	ErrNoCredentials = errors.New("zrt: no credentials available")

	// ErrAuthFailed means credentials were present but authentication failed
	// (for example minting a JWT from the API key and secret failed).
	ErrAuthFailed = errors.New("zrt: authentication failed")

	// ErrConnection means the SDK could not reach the runtime.
	ErrConnection = errors.New("zrt: connection failed")

	// ErrSessionRejected means the runtime rejected the session request
	// (for example invalid configuration or quota exceeded).
	ErrSessionRejected = errors.New("zrt: session rejected")

	// ErrSessionNotStarted means an operation was attempted on a session that
	// has not been started yet. Call AgentSession.Start first.
	ErrSessionNotStarted = errors.New("zrt: session not started")

	// ErrNotImplemented means the requested capability is not available.
	ErrNotImplemented = errors.New("zrt: not implemented")
)
