# Technical Design: TTS Implementation

This document describes the technical implementation plan for adding Text-to-Speech (TTS) capabilities to omni-deepgram, supporting both REST API (non-streaming) and WebSocket (streaming) modes.

## Goals

1. Implement `tts.Provider` interface from OmniVoice
2. Support non-streaming synthesis via Deepgram REST API
3. Support streaming synthesis via Deepgram WebSocket API
4. Maintain consistency with existing STT implementation patterns

## Architecture Overview

```
omni-deepgram/
├── omnivoice/
│   ├── omnivoice.go          # Shared constants, provider name
│   ├── convert.go            # STT config conversions (existing)
│   ├── convert_tts.go        # TTS config conversions (new)
│   ├── stt/
│   │   └── provider.go       # STT provider (existing)
│   └── tts/
│       └── provider.go       # TTS provider (new)
```

## Interface Implementation

### OmniVoice TTS Provider Interface

```go
type Provider interface {
    Name() string
    Synthesize(ctx context.Context, text string, config SynthesisConfig) (*SynthesisResult, error)
    SynthesizeStream(ctx context.Context, text string, config SynthesisConfig) (<-chan StreamChunk, error)
    ListVoices(ctx context.Context) ([]Voice, error)
    GetVoice(ctx context.Context, voiceID string) (*Voice, error)
}
```

### Optional StreamingProvider Interface

```go
type StreamingProvider interface {
    Provider
    SynthesizeFromReader(ctx context.Context, reader io.Reader, config SynthesisConfig) (<-chan StreamChunk, error)
}
```

## Implementation Plan

### Phase 1: Non-Streaming TTS (REST API)

#### 1.1 Create `omnivoice/convert_tts.go`

Maps OmniVoice `SynthesisConfig` to Deepgram `SpeakOptions`:

```go
// OmniVoice SynthesisConfig → Deepgram SpeakOptions
func ConfigToSpeakOptions(config tts.SynthesisConfig) *interfaces.SpeakOptions {
    opts := &interfaces.SpeakOptions{
        Model:      config.Model,      // e.g., "aura-2-thalia-en"
        Encoding:   config.OutputFormat, // e.g., "linear16", "mp3"
        SampleRate: config.SampleRate,
    }

    // If VoiceID provided but Model not, use VoiceID as model
    if opts.Model == "" && config.VoiceID != "" {
        opts.Model = config.VoiceID
    }

    // Default model if none specified
    if opts.Model == "" {
        opts.Model = "aura-asteria-en"
    }

    return opts
}
```

#### 1.2 Create `omnivoice/tts/provider.go`

```go
package tts

import (
    "context"
    "github.com/plexusone/omni-deepgram/omnivoice"
    "github.com/plexusone/omnivoice/tts"
    speak "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/speak"
    speakapi "github.com/deepgram/deepgram-go-sdk/v3/pkg/api/speak/v1/rest"
    interfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/interfaces"
)

var _ tts.Provider = (*Provider)(nil)

type Provider struct {
    apiKey string
    client *speak.RESTClient
}

func New(opts ...Option) (*Provider, error) {
    // Initialize with API key
    // Create REST client
}

func (p *Provider) Name() string {
    return omnivoice.ProviderName
}

func (p *Provider) Synthesize(ctx context.Context, text string, config tts.SynthesisConfig) (*tts.SynthesisResult, error) {
    // 1. Convert config to Deepgram options
    // 2. Call ToStream() to get audio bytes
    // 3. Convert response to SynthesisResult
}

func (p *Provider) ListVoices(ctx context.Context) ([]tts.Voice, error) {
    // Return static list of known Deepgram voices
    // Deepgram doesn't have a voices API endpoint
}

func (p *Provider) GetVoice(ctx context.Context, voiceID string) (*tts.Voice, error) {
    // Lookup from static voice list
}
```

#### 1.3 Synthesize Implementation Details

```go
func (p *Provider) Synthesize(ctx context.Context, text string, config tts.SynthesisConfig) (*tts.SynthesisResult, error) {
    opts := omnivoice.ConfigToSpeakOptions(config)

    // Create API wrapper
    api := speakapi.New(p.client)

    // Get audio into buffer
    var buffer interfaces.RawResponse
    resp, err := api.ToStream(ctx, text, opts, &buffer)
    if err != nil {
        return nil, fmt.Errorf("deepgram TTS failed: %w", err)
    }

    return &tts.SynthesisResult{
        Audio:          buffer.Bytes(),
        Format:         config.OutputFormat,
        SampleRate:     config.SampleRate,
        CharacterCount: resp.Characters,
    }, nil
}
```

### Phase 2: Streaming TTS (WebSocket API)

#### 2.1 SynthesizeStream Implementation

Uses Deepgram WebSocket API with callback pattern:

```go
func (p *Provider) SynthesizeStream(ctx context.Context, text string, config tts.SynthesisConfig) (<-chan tts.StreamChunk, error) {
    opts := omnivoice.ConfigToWSSpeakOptions(config)

    chunkCh := make(chan tts.StreamChunk, 100)

    // Create callback handler
    handler := &ttsCallbackHandler{
        chunkCh: chunkCh,
        ctx:     ctx,
    }

    // Create WebSocket client
    wsClient, err := speak.NewWSUsingCallback(ctx, p.apiKey, nil, opts, handler)
    if err != nil {
        close(chunkCh)
        return nil, err
    }

    // Connect
    if !wsClient.Connect() {
        close(chunkCh)
        return nil, fmt.Errorf("failed to connect to Deepgram TTS")
    }

    // Send text and flush
    go func() {
        defer func() {
            wsClient.Stop()
            close(chunkCh)
        }()

        if err := wsClient.SpeakWithText(text); err != nil {
            chunkCh <- tts.StreamChunk{Error: err}
            return
        }

        if err := wsClient.Flush(); err != nil {
            chunkCh <- tts.StreamChunk{Error: err}
            return
        }

        // Wait for completion or context cancellation
        <-ctx.Done()
    }()

    return chunkCh, nil
}
```

#### 2.2 Callback Handler

```go
type ttsCallbackHandler struct {
    chunkCh chan tts.StreamChunk
    ctx     context.Context
    done    bool
    mu      sync.Mutex
}

func (h *ttsCallbackHandler) Binary(data []byte) error {
    h.mu.Lock()
    defer h.mu.Unlock()

    if h.done {
        return nil
    }

    // Copy data to avoid reference issues
    audio := make([]byte, len(data))
    copy(audio, data)

    select {
    case h.chunkCh <- tts.StreamChunk{Audio: audio}:
    case <-h.ctx.Done():
        return h.ctx.Err()
    default:
        // Channel full, drop chunk
    }
    return nil
}

func (h *ttsCallbackHandler) Flush(fr *wsinterfaces.FlushedResponse) error {
    // Mark final chunk after flush
    h.mu.Lock()
    defer h.mu.Unlock()

    h.done = true
    select {
    case h.chunkCh <- tts.StreamChunk{IsFinal: true}:
    case <-h.ctx.Done():
    default:
    }
    return nil
}

func (h *ttsCallbackHandler) Error(er *wsinterfaces.ErrorResponse) error {
    select {
    case h.chunkCh <- tts.StreamChunk{Error: fmt.Errorf("deepgram error: %s", er.Description)}:
    case <-h.ctx.Done():
    default:
    }
    return nil
}

// Implement remaining callback methods: Open, Metadata, Close, Warning, Clear, UnhandledEvent
```

### Phase 3: Streaming Input (SynthesizeFromReader)

For streaming LLM output directly to TTS:

```go
func (p *Provider) SynthesizeFromReader(ctx context.Context, reader io.Reader, config tts.SynthesisConfig) (<-chan tts.StreamChunk, error) {
    opts := omnivoice.ConfigToWSSpeakOptions(config)

    chunkCh := make(chan tts.StreamChunk, 100)
    handler := &ttsCallbackHandler{chunkCh: chunkCh, ctx: ctx}

    wsClient, err := speak.NewWSUsingCallback(ctx, p.apiKey, nil, opts, handler)
    if err != nil {
        close(chunkCh)
        return nil, err
    }

    if !wsClient.Connect() {
        close(chunkCh)
        return nil, fmt.Errorf("failed to connect")
    }

    go func() {
        defer func() {
            wsClient.Stop()
            close(chunkCh)
        }()

        scanner := bufio.NewScanner(reader)
        scanner.Split(bufio.ScanLines) // Or custom split for sentences

        for scanner.Scan() {
            select {
            case <-ctx.Done():
                return
            default:
                text := scanner.Text()
                if text == "" {
                    continue
                }
                if err := wsClient.SpeakWithText(text); err != nil {
                    chunkCh <- tts.StreamChunk{Error: err}
                    return
                }
            }
        }

        // Flush remaining audio
        wsClient.Flush()

        // Wait for all audio to be received
        time.Sleep(500 * time.Millisecond) // Or use done signal from Flush callback
    }()

    return chunkCh, nil
}
```

## Voice Management

Deepgram doesn't provide a voices API. Implement with static voice list:

```go
var deepgramVoices = []tts.Voice{
    {ID: "aura-asteria-en", Name: "Asteria", Language: "en-US", Gender: "female", Provider: "deepgram"},
    {ID: "aura-luna-en", Name: "Luna", Language: "en-US", Gender: "female", Provider: "deepgram"},
    {ID: "aura-stella-en", Name: "Stella", Language: "en-US", Gender: "female", Provider: "deepgram"},
    {ID: "aura-athena-en", Name: "Athena", Language: "en-US", Gender: "female", Provider: "deepgram"},
    {ID: "aura-hera-en", Name: "Hera", Language: "en-US", Gender: "female", Provider: "deepgram"},
    {ID: "aura-orion-en", Name: "Orion", Language: "en-US", Gender: "male", Provider: "deepgram"},
    {ID: "aura-arcas-en", Name: "Arcas", Language: "en-US", Gender: "male", Provider: "deepgram"},
    {ID: "aura-perseus-en", Name: "Perseus", Language: "en-US", Gender: "male", Provider: "deepgram"},
    {ID: "aura-angus-en", Name: "Angus", Language: "en-IE", Gender: "male", Provider: "deepgram"},
    {ID: "aura-orpheus-en", Name: "Orpheus", Language: "en-US", Gender: "male", Provider: "deepgram"},
    {ID: "aura-helios-en", Name: "Helios", Language: "en-GB", Gender: "male", Provider: "deepgram"},
    {ID: "aura-zeus-en", Name: "Zeus", Language: "en-US", Gender: "male", Provider: "deepgram"},
    // Aura 2 voices
    {ID: "aura-2-thalia-en", Name: "Thalia (Aura 2)", Language: "en-US", Gender: "female", Provider: "deepgram"},
    // Add more as Deepgram releases them
}

func (p *Provider) ListVoices(ctx context.Context) ([]tts.Voice, error) {
    return deepgramVoices, nil
}

func (p *Provider) GetVoice(ctx context.Context, voiceID string) (*tts.Voice, error) {
    for _, v := range deepgramVoices {
        if v.ID == voiceID {
            return &v, nil
        }
    }
    return nil, tts.ErrVoiceNotFound
}
```

## Type Mappings

### OmniVoice → Deepgram

| OmniVoice SynthesisConfig | Deepgram SpeakOptions | Notes |
|---------------------------|----------------------|-------|
| `VoiceID` | `Model` | Deepgram uses model name as voice ID |
| `Model` | `Model` | Takes precedence over VoiceID |
| `OutputFormat` | `Encoding` | "mp3", "linear16", "opus", etc. |
| `SampleRate` | `SampleRate` | Hz value |
| `Speed` | — | Not supported by Deepgram |
| `Pitch` | — | Not supported by Deepgram |
| `Stability` | — | Not supported by Deepgram |
| `SimilarityBoost` | — | Not supported by Deepgram |

### Deepgram → OmniVoice

| Deepgram SpeakResponse | OmniVoice SynthesisResult | Notes |
|------------------------|---------------------------|-------|
| `buffer.Bytes()` | `Audio` | Raw audio bytes |
| `resp.Characters` | `CharacterCount` | Characters processed |
| (from config) | `Format` | Pass through from config |
| (from config) | `SampleRate` | Pass through from config |
| — | `DurationMs` | Calculate from audio size/format |

## Testing Strategy

### Unit Tests

```go
// provider_test.go

func TestSynthesize(t *testing.T) {
    // Mock Deepgram REST client
    // Verify correct options mapping
    // Verify result conversion
}

func TestSynthesizeStream(t *testing.T) {
    // Mock Deepgram WebSocket client
    // Verify chunks received in order
    // Verify IsFinal on last chunk
    // Verify error handling
}

func TestListVoices(t *testing.T) {
    // Verify all voices returned
    // Verify voice structure
}

func TestGetVoice(t *testing.T) {
    // Verify found voice
    // Verify ErrVoiceNotFound for invalid ID
}
```

### Integration Tests

```go
// provider_integration_test.go (requires DEEPGRAM_API_KEY)

func TestSynthesizeIntegration(t *testing.T) {
    if os.Getenv("DEEPGRAM_API_KEY") == "" {
        t.Skip("DEEPGRAM_API_KEY not set")
    }
    // Test actual API call
    // Verify audio bytes returned
}
```

## Implementation Checklist

### Phase 1: Non-Streaming (REST)

- [ ] Create `omnivoice/convert_tts.go` with config conversion
- [ ] Create `omnivoice/tts/provider.go` with Provider struct
- [ ] Implement `New()` constructor with options
- [ ] Implement `Name()` method
- [ ] Implement `Synthesize()` method
- [ ] Implement `ListVoices()` with static voice list
- [ ] Implement `GetVoice()` with lookup
- [ ] Add unit tests
- [ ] Add integration tests
- [ ] Update README.md feature table

### Phase 2: Streaming (WebSocket)

- [ ] Add WebSocket config conversion to `convert_tts.go`
- [ ] Implement `ttsCallbackHandler` struct
- [ ] Implement callback methods (Binary, Flush, Error, etc.)
- [ ] Implement `SynthesizeStream()` method
- [ ] Add streaming unit tests
- [ ] Add streaming integration tests
- [ ] Update README.md feature table

### Phase 3: Streaming Input

- [ ] Implement `SynthesizeFromReader()` method
- [ ] Add sentence/phrase splitting logic
- [ ] Add unit tests
- [ ] Add integration tests
- [ ] Update README.md feature table

### Documentation

- [ ] Update README.md with TTS usage examples
- [ ] Update PRD.md checkboxes
- [ ] Add CHANGELOG entry

## References

- [Deepgram TTS REST API](https://developers.deepgram.com/reference/text-to-speech-api)
- [Deepgram TTS WebSocket API](https://developers.deepgram.com/reference/transform-text-to-speech-websocket)
- [Deepgram Go SDK Examples](https://github.com/deepgram/deepgram-go-sdk/tree/main/examples/text-to-speech)
- [OmniVoice TTS Interface](https://github.com/plexusone/omnivoice/blob/main/tts/tts.go)
