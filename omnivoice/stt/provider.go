// Package stt provides an OmniVoice STT provider implementation using Deepgram.
package stt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	restapi "github.com/deepgram/deepgram-go-sdk/v3/pkg/api/listen/v1/rest"
	wsinterfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/api/listen/v1/websocket/interfaces"
	interfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/interfaces"
	client "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/listen"
	"github.com/plexusone/omni-deepgram/omnivoice"
	"github.com/plexusone/omnivoice-core/stt"
)

// Verify interface compliance at compile time.
var _ stt.StreamingProvider = (*Provider)(nil)

// Provider implements stt.StreamingProvider using the Deepgram API.
type Provider struct {
	apiKey string

	mu sync.Mutex
}

// Option configures the Provider.
type Option func(*options)

type options struct {
	apiKey string
}

// WithAPIKey sets the Deepgram API key.
func WithAPIKey(apiKey string) Option {
	return func(o *options) {
		o.apiKey = apiKey
	}
}

// New creates a new Deepgram STT provider.
func New(opts ...Option) (*Provider, error) {
	cfg := &options{}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Initialize the Deepgram client library (shared across STT/TTS)
	omnivoice.InitSDK()

	return &Provider{
		apiKey: cfg.apiKey,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return omnivoice.ProviderName
}

// Transcribe converts audio to text (batch mode).
func (p *Provider) Transcribe(ctx context.Context, audio []byte, config stt.TranscriptionConfig) (*stt.TranscriptionResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Create REST client
	c := client.NewREST(p.apiKey, &interfaces.ClientOptions{})
	dg := restapi.New(c)

	// Convert config to Deepgram options
	opts := omnivoice.ConfigToPreRecordedOptions(config)

	// Transcribe from stream (bytes)
	resp, err := dg.FromStream(ctx, bytes.NewReader(audio), opts)
	if err != nil {
		return nil, omnivoice.ClassifyError("Transcribe", err)
	}

	// Convert response to OmniVoice result
	return omnivoice.PreRecordedResponseToResult(resp), nil
}

// TranscribeFile transcribes audio from a file path.
func (p *Provider) TranscribeFile(ctx context.Context, filePath string, config stt.TranscriptionConfig) (*stt.TranscriptionResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Create REST client
	c := client.NewREST(p.apiKey, &interfaces.ClientOptions{})
	dg := restapi.New(c)

	// Convert config to Deepgram options
	opts := omnivoice.ConfigToPreRecordedOptions(config)

	// Transcribe from file
	resp, err := dg.FromFile(ctx, filePath, opts)
	if err != nil {
		return nil, omnivoice.ClassifyError("TranscribeFile", err)
	}

	// Convert response to OmniVoice result
	return omnivoice.PreRecordedResponseToResult(resp), nil
}

// TranscribeURL transcribes audio from a URL.
func (p *Provider) TranscribeURL(ctx context.Context, url string, config stt.TranscriptionConfig) (*stt.TranscriptionResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Create REST client
	c := client.NewREST(p.apiKey, &interfaces.ClientOptions{})
	dg := restapi.New(c)

	// Convert config to Deepgram options
	opts := omnivoice.ConfigToPreRecordedOptions(config)

	// Transcribe from URL
	resp, err := dg.FromURL(ctx, url, opts)
	if err != nil {
		return nil, omnivoice.ClassifyError("TranscribeURL", err)
	}

	// Convert response to OmniVoice result
	return omnivoice.PreRecordedResponseToResult(resp), nil
}

// TranscribeStream starts a streaming transcription session.
// Returns a writer for sending audio and a channel for receiving events.
func (p *Provider) TranscribeStream(ctx context.Context, config stt.TranscriptionConfig) (io.WriteCloser, <-chan stt.StreamEvent, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Convert config to Deepgram options
	dgOptions := omnivoice.ConfigToLiveTranscriptionOptions(config)

	// Create the callback handler
	eventCh := make(chan stt.StreamEvent, 100)
	handler := &callbackHandler{
		eventCh: eventCh,
		ctx:     ctx,
	}

	// Create WebSocket client with callback
	dgClient, err := client.NewWSUsingCallbackWithDefaults(ctx, dgOptions, handler)
	if err != nil {
		close(eventCh)
		return nil, nil, omnivoice.ClassifyError("TranscribeStream", err)
	}

	// Connect to Deepgram
	if !dgClient.Connect() {
		close(eventCh)
		return nil, nil, omnivoice.ClassifyError("TranscribeStream", fmt.Errorf("failed to connect to Deepgram"))
	}

	// Create the audio writer
	writer := &streamWriter{
		client:  dgClient,
		eventCh: eventCh,
		ctx:     ctx,
		done:    make(chan struct{}),
	}

	// Handle context cancellation
	go func() {
		select {
		case <-ctx.Done():
			_ = writer.Close()
		case <-writer.done:
		}
	}()

	return writer, eventCh, nil
}

// streamWriter implements io.WriteCloser for sending audio to Deepgram.
type streamWriter struct {
	client  DeepgramClient
	eventCh chan stt.StreamEvent
	ctx     context.Context
	done    chan struct{}
	closed  bool
	mu      sync.Mutex
}

// DeepgramClient interface for the Deepgram WebSocket client.
type DeepgramClient interface {
	Write(p []byte) (n int, err error)
	Stop()
}

func (w *streamWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return 0, io.ErrClosedPipe
	}
	w.mu.Unlock()

	return w.client.Write(p)
}

func (w *streamWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}
	w.closed = true

	// Stop the Deepgram client
	w.client.Stop()

	// Close channels
	close(w.done)
	close(w.eventCh)

	return nil
}

// callbackHandler implements the Deepgram callback interface.
type callbackHandler struct {
	eventCh chan stt.StreamEvent
	ctx     context.Context
}

// Open is called when the connection is established.
func (h *callbackHandler) Open(or *wsinterfaces.OpenResponse) error {
	return nil
}

// Message is called when a transcription message is received.
func (h *callbackHandler) Message(mr *wsinterfaces.MessageResponse) error {
	if mr == nil {
		return nil
	}

	// Convert to our internal type
	result := &omnivoice.MessageResponse{
		IsFinal:  mr.IsFinal,
		Duration: mr.Duration,
		Start:    mr.Start,
	}

	// Copy channel data
	if len(mr.Channel.Alternatives) > 0 {
		result.Channel.Alternatives = make([]omnivoice.Alternative, len(mr.Channel.Alternatives))
		for i, alt := range mr.Channel.Alternatives {
			result.Channel.Alternatives[i] = omnivoice.Alternative{
				Transcript: alt.Transcript,
				Confidence: alt.Confidence,
			}
			if len(alt.Words) > 0 {
				result.Channel.Alternatives[i].Words = make([]omnivoice.Word, len(alt.Words))
				for j, w := range alt.Words {
					word := omnivoice.Word{
						Word:       w.Word,
						Start:      w.Start,
						End:        w.End,
						Confidence: w.Confidence,
					}
					if w.Speaker != nil {
						word.Speaker = w.Speaker
					}
					result.Channel.Alternatives[i].Words[j] = word
				}
			}
		}
	}

	// Convert to OmniVoice event
	event := omnivoice.MessageResponseToStreamEvent(result)

	select {
	case h.eventCh <- event:
	case <-h.ctx.Done():
		return h.ctx.Err()
	default:
		// Channel full, drop event
	}

	return nil
}

// Metadata is called when metadata is received.
func (h *callbackHandler) Metadata(md *wsinterfaces.MetadataResponse) error {
	return nil
}

// SpeechStarted is called when speech is detected.
func (h *callbackHandler) SpeechStarted(ssr *wsinterfaces.SpeechStartedResponse) error {
	event := stt.StreamEvent{
		Type:          stt.EventSpeechStart,
		SpeechStarted: true,
	}

	select {
	case h.eventCh <- event:
	case <-h.ctx.Done():
		return h.ctx.Err()
	default:
	}

	return nil
}

// UtteranceEnd is called when an utterance ends.
func (h *callbackHandler) UtteranceEnd(ur *wsinterfaces.UtteranceEndResponse) error {
	event := stt.StreamEvent{
		Type:        stt.EventSpeechEnd,
		SpeechEnded: true,
	}

	select {
	case h.eventCh <- event:
	case <-h.ctx.Done():
		return h.ctx.Err()
	default:
	}

	return nil
}

// Close is called when the connection is closed.
func (h *callbackHandler) Close(cr *wsinterfaces.CloseResponse) error {
	return nil
}

// Error is called when an error occurs.
func (h *callbackHandler) Error(er *wsinterfaces.ErrorResponse) error {
	if er == nil {
		return nil
	}

	err := fmt.Errorf("deepgram error: %s", er.Description)
	event := stt.StreamEvent{
		Type:  stt.EventError,
		Error: omnivoice.ClassifyError("TranscribeStream", err),
	}

	select {
	case h.eventCh <- event:
	case <-h.ctx.Done():
		return h.ctx.Err()
	default:
	}

	return nil
}

// UnhandledEvent is called for unhandled events.
func (h *callbackHandler) UnhandledEvent(raw []byte) error {
	// Try to parse as a generic message for debugging
	var msg map[string]interface{}
	if err := json.Unmarshal(raw, &msg); err == nil {
		// Log unhandled event type if present
		if msgType, ok := msg["type"].(string); ok {
			_ = msgType // Could log this for debugging
		}
	}
	return nil
}
