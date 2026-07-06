package realtime

import "errors"

// Package-level errors for the realtime provider.
var (
	// ErrAPIKeyRequired is returned when no API key is provided.
	ErrAPIKeyRequired = errors.New("API key is required")

	// ErrConnectionFailed is returned when the WebSocket connection fails.
	ErrConnectionFailed = errors.New("failed to connect to Deepgram Voice Agent")

	// ErrSessionClosed is returned when operations are attempted on a closed session.
	ErrSessionClosed = errors.New("voice session is closed")

	// ErrFunctionCallFailed is returned when a function call handler returns an error.
	ErrFunctionCallFailed = errors.New("function call failed")
)
