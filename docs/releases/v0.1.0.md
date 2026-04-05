# Release Notes - v0.1.0

**Release Date:** 2026-01-19

## Overview

Initial release of omnivoice-deepgram, an OmniVoice STT provider for Deepgram speech-to-text services.

This package adapts the official Deepgram Go SDK to the OmniVoice interfaces, enabling real-time streaming transcription for voice agents and telephony applications.

## Highlights

- OmniVoice STT provider for Deepgram with real-time streaming transcription support

## What's Included

### Streaming Transcription

- Full implementation of `stt.StreamingProvider` interface
- Real-time transcription via Deepgram WebSocket API
- Interim results for responsive UI feedback
- Final results with word-level timing information

### Audio Format Support

- mu-law (Twilio, telephony)
- A-law (European telephony)
- Linear PCM (general audio)
- FLAC, Opus, MP3

### Voice Agent Features

- Speech start/end detection for natural turn-taking
- Speaker diarization for multi-speaker scenarios
- Keyword boosting for domain-specific vocabulary

### Configuration

- OmniVoice `TranscriptionConfig` to Deepgram options converter
- Sensible defaults for telephony (8kHz, mono, nova-2 model)

## Installation

```bash
go get github.com/agentplexus/omnivoice-deepgram
```

## Quick Start

```go
import (
    deepgramstt "github.com/agentplexus/omnivoice-deepgram/omnivoice/stt"
    "github.com/agentplexus/omnivoice/stt"
)

// Create provider
provider, err := deepgramstt.New(deepgramstt.WithAPIKey("your-api-key"))
if err != nil {
    log.Fatal(err)
}

// Start streaming
writer, events, err := provider.TranscribeStream(ctx, stt.TranscriptionConfig{
    Model:      "nova-2",
    Language:   "en-US",
    Encoding:   "mulaw",
    SampleRate: 8000,
})

// Send audio, receive transcripts
go func() {
    defer writer.Close()
    io.Copy(writer, audioSource)
}()

for event := range events {
    if event.IsFinal {
        fmt.Println(event.Transcript)
    }
}
```

## Dependencies

- `github.com/agentplexus/omnivoice` v0.2.0 - OmniVoice interfaces
- `github.com/deepgram/deepgram-go-sdk/v3` v3.0.0 - Official Deepgram SDK

## Related

- [omnivoice-examples](https://github.com/agentplexus/omnivoice-examples) - Complete voice agent examples combining Deepgram STT with ElevenLabs TTS
