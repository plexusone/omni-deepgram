# OmniVoice Deepgram Provider

OmniVoice provider implementation for [Deepgram](https://deepgram.com/) speech-to-text and text-to-speech services.

This package adapts the official [Deepgram Go SDK](https://github.com/deepgram/deepgram-go-sdk) to the [OmniVoice](https://github.com/plexusone/omnivoice-core) interfaces, enabling Deepgram's STT and TTS capabilities within the OmniVoice framework.

## Features

### Speech-to-Text (STT)

- Real-time streaming transcription via WebSocket
- Batch transcription from files, URLs, or bytes
- Support for telephony audio formats (mu-law, a-law)
- Interim and final transcription results
- Speech start/end detection for natural turn-taking
- Speaker diarization support
- Keyword boosting

### Text-to-Speech (TTS)

- Non-streaming synthesis via REST API
- Real-time streaming synthesis via WebSocket
- Streaming input support (pipe LLM output directly to TTS)
- Automatic sentence splitting for natural speech
- Multiple Aura voices (male/female, US/UK/IE accents)
- Multiple output formats (mp3, linear16, mulaw, opus, etc.)

### Agent Experience (AX)

- Error classification with 8 categories (transient, rate_limit, validation, auth, not_found, server, quota, unknown)
- Automatic retry decisions based on error type
- Actionable error suggestions for recovery
- Integration with `resilience.ProviderError` from omnivoice-core

## Installation

```bash
go get github.com/plexusone/omni-deepgram
```

## Quick Start

### Speech-to-Text

```go
import (
    deepgramstt "github.com/plexusone/omni-deepgram/omnivoice/stt"
    "github.com/plexusone/omnivoice-core/stt"
)

provider, err := deepgramstt.New(deepgramstt.WithAPIKey("your-api-key"))
if err != nil {
    log.Fatal(err)
}

result, err := provider.TranscribeURL(ctx, "https://example.com/audio.mp3", stt.TranscriptionConfig{
    Model:    "nova-2",
    Language: "en-US",
})
```

### Text-to-Speech

```go
import (
    deepgramtts "github.com/plexusone/omni-deepgram/omnivoice/tts"
    "github.com/plexusone/omnivoice-core/tts"
)

provider, err := deepgramtts.New(deepgramtts.WithAPIKey("your-api-key"))
if err != nil {
    log.Fatal(err)
}

result, err := provider.Synthesize(ctx, "Hello, world!", tts.SynthesisConfig{
    VoiceID:      "aura-asteria-en",
    OutputFormat: "mp3",
})
```

## Requirements

- Go 1.21 or later
- Deepgram API key ([get one here](https://console.deepgram.com/))

## Related Projects

- [omnivoice-core](https://github.com/plexusone/omnivoice-core) - Voice agent framework interfaces
- [elevenlabs-go](https://github.com/plexusone/elevenlabs-go) - ElevenLabs TTS provider
- [omnivoice-twilio](https://github.com/plexusone/omnivoice-twilio) - Twilio Media Streams transport
