# v0.2.0 — Text-to-Speech (TTS) Provider

This release adds a full TTS provider with streaming support and LLM integration, complementing the existing STT provider.

## Highlights

- Text-to-Speech (TTS) provider with streaming support and LLM integration

## Added

- TTS provider implementing `tts.Provider` and `tts.StreamingProvider` interfaces ([`0f554ae`](https://github.com/agentplexus/omnivoice-deepgram/commit/0f554ae))
- `Synthesize` method for non-streaming TTS via Deepgram REST API ([`0f554ae`](https://github.com/agentplexus/omnivoice-deepgram/commit/0f554ae))
- `SynthesizeStream` method for streaming TTS via Deepgram WebSocket API ([`0f554ae`](https://github.com/agentplexus/omnivoice-deepgram/commit/0f554ae))
- `SynthesizeFromReader` method for streaming LLM output directly to TTS ([`0f554ae`](https://github.com/agentplexus/omnivoice-deepgram/commit/0f554ae))
- `ListVoices` method returning available Deepgram Aura voices ([`0f554ae`](https://github.com/agentplexus/omnivoice-deepgram/commit/0f554ae))
- `GetVoice` method for voice lookup by ID ([`0f554ae`](https://github.com/agentplexus/omnivoice-deepgram/commit/0f554ae))
- Automatic sentence splitting for natural speech synthesis ([`0f554ae`](https://github.com/agentplexus/omnivoice-deepgram/commit/0f554ae))
- Support for TTS output formats: mp3, linear16, mulaw, alaw, opus, flac ([`0f554ae`](https://github.com/agentplexus/omnivoice-deepgram/commit/0f554ae))
- Configuration converter from OmniVoice `SynthesisConfig` to Deepgram options ([`0f554ae`](https://github.com/agentplexus/omnivoice-deepgram/commit/0f554ae))
- Static voice list with 17 Aura voices (Aura 1 and Aura 2) ([`0f554ae`](https://github.com/agentplexus/omnivoice-deepgram/commit/0f554ae))
- TTS conformance tests using omnivoice `providertest` package ([`854e14d`](https://github.com/agentplexus/omnivoice-deepgram/commit/854e14d))
- STT conformance tests using omnivoice `providertest` package ([`854e14d`](https://github.com/agentplexus/omnivoice-deepgram/commit/854e14d))

## Fixed

- STT provider initialization uses `sync.Once` to prevent flag redefinition panic ([`8a584f7`](https://github.com/agentplexus/omnivoice-deepgram/commit/8a584f7))

## Dependencies

- Bump `github.com/agentplexus/omnivoice` from v0.2.0 to v0.3.0 ([`0f554ae`](https://github.com/agentplexus/omnivoice-deepgram/commit/0f554ae))

## Quick Start

```go
import (
    deepgramtts "github.com/agentplexus/omnivoice-deepgram/omnivoice/tts"
    "github.com/agentplexus/omnivoice/tts"
)

// Create TTS provider
provider, err := deepgramtts.New(deepgramtts.WithAPIKey("your-api-key"))
if err != nil {
    log.Fatal(err)
}

// Synthesize text to speech
result, err := provider.Synthesize(ctx, "Hello, world!", tts.SynthesisConfig{
    VoiceID:      "aura-asteria-en",
    OutputFormat: "mp3",
    SampleRate:   24000,
})
```

See the [README](https://github.com/agentplexus/omnivoice-deepgram#readme) for streaming and LLM integration examples.
