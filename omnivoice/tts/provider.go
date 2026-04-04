// Package tts provides an OmniVoice TTS provider implementation using Deepgram.
package tts

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode"

	speakapi "github.com/deepgram/deepgram-go-sdk/v3/pkg/api/speak/v1/rest"
	wsinterfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/api/speak/v1/websocket/interfaces"
	interfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/interfaces"
	speak "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/speak"
	"github.com/plexusone/omnivoice-core/tts"
	"github.com/plexusone/omnivoice-deepgram/omnivoice"
)

// Verify interface compliance at compile time.
var (
	_ tts.Provider          = (*Provider)(nil)
	_ tts.StreamingProvider = (*Provider)(nil)
)

// Provider implements tts.Provider using the Deepgram API.
type Provider struct {
	apiKey string
	client *speakapi.Client

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

// New creates a new Deepgram TTS provider.
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

	// Create REST client with empty options (not nil)
	restClient := speak.NewREST(cfg.apiKey, &interfaces.ClientOptions{})
	client := speakapi.New(restClient)

	return &Provider{
		apiKey: cfg.apiKey,
		client: client,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return omnivoice.ProviderName
}

// Synthesize converts text to speech and returns audio data.
func (p *Provider) Synthesize(ctx context.Context, text string, config tts.SynthesisConfig) (*tts.SynthesisResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Convert config to Deepgram options
	opts := omnivoice.ConfigToSpeakOptions(config)

	// Get audio into buffer
	var buffer interfaces.RawResponse
	resp, err := p.client.ToStream(ctx, text, opts, &buffer)
	if err != nil {
		return nil, omnivoice.ClassifyError("Synthesize", err)
	}

	// Determine output format
	outputFormat := config.OutputFormat
	if outputFormat == "" {
		outputFormat = "linear16"
	}

	// Determine sample rate
	sampleRate := config.SampleRate
	if sampleRate == 0 {
		sampleRate = 24000 // Deepgram default
	}

	return &tts.SynthesisResult{
		Audio:          buffer.Bytes(),
		Format:         outputFormat,
		SampleRate:     sampleRate,
		CharacterCount: resp.Characters,
	}, nil
}

// SynthesizeStream converts text to speech with streaming output.
func (p *Provider) SynthesizeStream(ctx context.Context, text string, config tts.SynthesisConfig) (<-chan tts.StreamChunk, error) {
	// Convert config to Deepgram WebSocket options
	opts := omnivoice.ConfigToWSSpeakOptions(config)

	chunkCh := make(chan tts.StreamChunk, 100)

	// Create callback handler
	handler := &ttsCallbackHandler{
		chunkCh: chunkCh,
		ctx:     ctx,
	}

	// Create WebSocket client with callback and API key
	wsClient, err := speak.NewWSUsingCallback(ctx, p.apiKey, &interfaces.ClientOptions{}, opts, handler)
	if err != nil {
		close(chunkCh)
		return nil, omnivoice.ClassifyError("SynthesizeStream", err)
	}

	// Connect to Deepgram
	if !wsClient.Connect() {
		close(chunkCh)
		return nil, omnivoice.ClassifyError("SynthesizeStream", fmt.Errorf("failed to connect to Deepgram TTS"))
	}

	// Send text and manage connection in goroutine
	go func() {
		defer func() {
			wsClient.Finish()
			handler.mu.Lock()
			if !handler.closed {
				handler.closed = true
				close(chunkCh)
			}
			handler.mu.Unlock()
		}()

		// Send text
		if err := wsClient.SpeakWithText(text); err != nil {
			handler.sendChunk(tts.StreamChunk{Error: fmt.Errorf("failed to send text: %w", err)})
			return
		}

		// Flush to signal end of input
		if err := wsClient.Flush(); err != nil {
			handler.sendChunk(tts.StreamChunk{Error: fmt.Errorf("failed to flush: %w", err)})
			return
		}

		// Wait for context cancellation or flush completion
		<-ctx.Done()
	}()

	return chunkCh, nil
}

// ListVoices returns available voices from this provider.
func (p *Provider) ListVoices(ctx context.Context) ([]tts.Voice, error) {
	voices := make([]tts.Voice, len(omnivoice.DeepgramVoices))
	for i, v := range omnivoice.DeepgramVoices {
		voices[i] = omnivoice.VoiceToOmniVoice(v)
	}
	return voices, nil
}

// GetVoice returns a specific voice by ID.
func (p *Provider) GetVoice(ctx context.Context, voiceID string) (*tts.Voice, error) {
	for _, v := range omnivoice.DeepgramVoices {
		if v.ID == voiceID {
			voice := omnivoice.VoiceToOmniVoice(v)
			return &voice, nil
		}
	}
	return nil, tts.ErrVoiceNotFound
}

// SynthesizeFromReader reads text from a reader and streams audio output.
// This is useful for streaming LLM output directly to TTS.
// Text is buffered and split into sentences for natural speech synthesis.
func (p *Provider) SynthesizeFromReader(ctx context.Context, reader io.Reader, config tts.SynthesisConfig) (<-chan tts.StreamChunk, error) {
	// Convert config to Deepgram WebSocket options
	opts := omnivoice.ConfigToWSSpeakOptions(config)

	chunkCh := make(chan tts.StreamChunk, 100)

	// Create callback handler
	handler := &ttsCallbackHandler{
		chunkCh: chunkCh,
		ctx:     ctx,
	}

	// Create WebSocket client with callback and API key
	wsClient, err := speak.NewWSUsingCallback(ctx, p.apiKey, &interfaces.ClientOptions{}, opts, handler)
	if err != nil {
		close(chunkCh)
		return nil, omnivoice.ClassifyError("SynthesizeStream", err)
	}

	// Connect to Deepgram
	if !wsClient.Connect() {
		close(chunkCh)
		return nil, omnivoice.ClassifyError("SynthesizeStream", fmt.Errorf("failed to connect to Deepgram TTS"))
	}

	// Process text from reader in goroutine
	go func() {
		defer func() {
			wsClient.Finish()
			handler.mu.Lock()
			if !handler.closed {
				handler.closed = true
				close(chunkCh)
			}
			handler.mu.Unlock()
		}()

		// Create a buffered reader for efficient reading
		bufReader := bufio.NewReader(reader)
		var textBuffer strings.Builder

		for {
			select {
			case <-ctx.Done():
				// Flush any remaining text before exit
				if textBuffer.Len() > 0 {
					text := strings.TrimSpace(textBuffer.String())
					if text != "" {
						if err := wsClient.SpeakWithText(text); err != nil {
							handler.sendChunk(tts.StreamChunk{Error: fmt.Errorf("failed to send text: %w", err)})
						}
					}
				}
				if err := wsClient.Flush(); err != nil {
					handler.sendChunk(tts.StreamChunk{Error: fmt.Errorf("failed to flush: %w", err)})
				}
				return
			default:
				// Read a chunk of text
				chunk, err := bufReader.ReadString('\n')
				if err != nil && err != io.EOF {
					handler.sendChunk(tts.StreamChunk{Error: fmt.Errorf("failed to read text: %w", err)})
					return
				}

				if len(chunk) > 0 {
					textBuffer.WriteString(chunk)

					// Check if we have complete sentences to send
					sentences := splitIntoSentences(textBuffer.String())
					if len(sentences) > 1 {
						// Send all complete sentences except the last (potentially incomplete) one
						for _, sentence := range sentences[:len(sentences)-1] {
							sentence = strings.TrimSpace(sentence)
							if sentence != "" {
								if err := wsClient.SpeakWithText(sentence); err != nil {
									handler.sendChunk(tts.StreamChunk{Error: fmt.Errorf("failed to send text: %w", err)})
									return
								}
							}
						}
						// Keep the last (potentially incomplete) sentence in the buffer
						textBuffer.Reset()
						textBuffer.WriteString(sentences[len(sentences)-1])
					}
				}

				if err == io.EOF {
					// End of input - flush remaining text
					remaining := strings.TrimSpace(textBuffer.String())
					if remaining != "" {
						if err := wsClient.SpeakWithText(remaining); err != nil {
							handler.sendChunk(tts.StreamChunk{Error: fmt.Errorf("failed to send text: %w", err)})
							return
						}
					}
					if err := wsClient.Flush(); err != nil {
						handler.sendChunk(tts.StreamChunk{Error: fmt.Errorf("failed to flush: %w", err)})
					}
					// Wait for flush callback to signal completion
					<-ctx.Done()
					return
				}
			}
		}
	}()

	return chunkCh, nil
}

// splitIntoSentences splits text into sentences based on common delimiters.
// Returns a slice where the last element may be an incomplete sentence.
func splitIntoSentences(text string) []string {
	var sentences []string
	var current strings.Builder

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		current.WriteRune(r)

		// Check for sentence-ending punctuation
		if r == '.' || r == '!' || r == '?' {
			// Look ahead to see if this is really the end of a sentence
			// (not an abbreviation like "Dr." or decimal like "3.14")
			isEndOfSentence := true

			if r == '.' && i > 0 {
				// Check if it's likely an abbreviation (single uppercase letter before dot)
				prevRune := runes[i-1]
				if unicode.IsUpper(prevRune) && (i < 2 || !unicode.IsLetter(runes[i-2])) {
					isEndOfSentence = false
				}
				// Check if it's a number (decimal point)
				if unicode.IsDigit(prevRune) && i+1 < len(runes) && unicode.IsDigit(runes[i+1]) {
					isEndOfSentence = false
				}
			}

			// Check if followed by space or end of text
			if isEndOfSentence && (i+1 >= len(runes) || unicode.IsSpace(runes[i+1])) {
				sentences = append(sentences, current.String())
				current.Reset()
			}
		}
	}

	// Add any remaining text as the last element
	if current.Len() > 0 {
		sentences = append(sentences, current.String())
	}

	return sentences
}

// ttsCallbackHandler implements the Deepgram TTS callback interface.
type ttsCallbackHandler struct {
	chunkCh chan tts.StreamChunk
	ctx     context.Context
	closed  bool
	mu      sync.Mutex
}

// sendChunk safely sends a chunk to the channel.
func (h *ttsCallbackHandler) sendChunk(chunk tts.StreamChunk) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return
	}

	select {
	case h.chunkCh <- chunk:
	case <-h.ctx.Done():
	default:
		// Channel full, drop chunk
	}
}

// Open is called when the connection is established.
func (h *ttsCallbackHandler) Open(or *wsinterfaces.OpenResponse) error {
	return nil
}

// Metadata is called when metadata is received.
func (h *ttsCallbackHandler) Metadata(md *wsinterfaces.MetadataResponse) error {
	return nil
}

// Flush is called when a flush response is received.
func (h *ttsCallbackHandler) Flush(fr *wsinterfaces.FlushedResponse) error {
	// Mark final chunk after flush
	h.sendChunk(tts.StreamChunk{IsFinal: true})
	return nil
}

// Clear is called when a clear response is received.
func (h *ttsCallbackHandler) Clear(cr *wsinterfaces.ClearedResponse) error {
	return nil
}

// Close is called when the connection is closed.
func (h *ttsCallbackHandler) Close(cr *wsinterfaces.CloseResponse) error {
	return nil
}

// Warning is called when a warning is received.
func (h *ttsCallbackHandler) Warning(wr *wsinterfaces.WarningResponse) error {
	return nil
}

// Error is called when an error occurs.
func (h *ttsCallbackHandler) Error(er *wsinterfaces.ErrorResponse) error {
	if er == nil {
		return nil
	}

	err := fmt.Errorf("deepgram TTS error: %s", er.Description)
	h.sendChunk(tts.StreamChunk{
		Error: omnivoice.ClassifyError("SynthesizeStream", err),
	})
	return nil
}

// UnhandledEvent is called for unhandled events.
func (h *ttsCallbackHandler) UnhandledEvent(raw []byte) error {
	return nil
}

// Binary is called when audio data is received.
func (h *ttsCallbackHandler) Binary(data []byte) error {
	// Copy data to avoid reference issues
	audio := make([]byte, len(data))
	copy(audio, data)

	h.sendChunk(tts.StreamChunk{Audio: audio})
	return nil
}
