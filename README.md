# OmniVoice Deepgram Provider

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/omni-deepgram/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/omni-deepgram/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/omni-deepgram/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/omni-deepgram/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/omni-deepgram/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/omni-deepgram/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/plexusone/omni-deepgram
 [goreport-url]: https://goreportcard.com/report/github.com/plexusone/omni-deepgram
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/omni-deepgram
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/omni-deepgram
 [viz-svg]: https://img.shields.io/badge/visualizaton-Go-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fomni-deepgram
 [loc-svg]: https://tokei.rs/b1/github/plexusone/omni-deepgram
 [repo-url]: https://github.com/plexusone/omni-deepgram
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/omni-deepgram/blob/main/LICENSE

OmniVoice provider implementation for [Deepgram](https://deepgram.com/) speech-to-text and text-to-speech services.

This package adapts the official [Deepgram Go SDK](https://github.com/deepgram/deepgram-go-sdk) to the [OmniVoice](https://github.com/plexusone/omnivoice) interfaces, enabling Deepgram's STT and TTS capabilities within the OmniVoice framework.

## OmniVoice Feature Support

This table shows which [OmniVoice](https://github.com/plexusone/omnivoice) abstracted capabilities are supported by this provider.

### Core Voice Capabilities

| Capability | Supported | Notes |
|------------|:---------:|-------|
| **STT (Speech-to-Text)** | ✅ | Full capability |
| STT Streaming | ✅ | Real-time via WebSocket |
| STT Batch | ✅ | From audio bytes via REST |
| STT File | ✅ | From file path via REST |
| STT URL | ✅ | From URL via REST |
| **TTS (Text-to-Speech)** | ✅ | Aura voices via REST and WebSocket |
| TTS Synthesize | ✅ | Non-streaming via REST API |
| TTS Streaming | ✅ | Real-time via WebSocket |
| TTS Voice List | ✅ | Static list of Aura voices |
| **Voice Agent** | — | N/A (use with agent orchestration) |

### STT Features

| Feature | Supported | Notes |
|---------|:---------:|-------|
| Interim results | ✅ | Real-time partial transcripts |
| Final results | ✅ | Complete utterance transcripts |
| Speech start detection | ✅ | `EventSpeechStart` events |
| Speech end detection | ✅ | `EventSpeechEnd` / utterance end |
| Speaker diarization | ✅ | Multi-speaker identification |
| Keyword boosting | ✅ | Boost specific terms |
| Punctuation | ✅ | Optional auto-punctuation |
| Word-level timestamps | ✅ | Per-word timing data |
| Confidence scores | ✅ | Per-word and per-utterance |

### TTS Features

| Feature | Supported | Notes |
|---------|:---------:|-------|
| Non-streaming synthesis | ✅ | REST API returns full audio |
| Streaming synthesis | ✅ | WebSocket streams audio chunks |
| Streaming input | ✅ | Pipe LLM output directly to TTS |
| Sentence splitting | ✅ | Automatic splitting for natural speech |
| Voice selection | ✅ | Aura 1 and Aura 2 voices |
| Output formats | ✅ | mp3, linear16, mulaw, alaw, opus, flac |
| Sample rate control | ✅ | Configurable output sample rate |

### Transport Layer

| Transport | Supported | Notes |
|-----------|:---------:|-------|
| WebSocket | ✅ | Native streaming transport |
| HTTP | ✅ | Batch/pre-recorded API |
| WebRTC | — | Use with transport provider |
| SIP | — | Use with transport provider |
| PSTN | — | Use with transport provider |

### Call System Integration

| Call System | Supported | Notes |
|-------------|:---------:|-------|
| Twilio | — | Use with [omnivoice-twilio](https://github.com/plexusone/omnivoice-twilio) |
| RingCentral | — | Use with call system provider |
| Zoom | — | Use with call system provider |
| LiveKit | — | Use with call system provider |
| Daily | — | Use with call system provider |

**Legend:** ✅ Supported | ❌ Not implemented | — Not applicable (use with other providers)

## Features

### Speech-to-Text (STT)

- Real-time streaming transcription via WebSocket
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
- Configurable sample rate

## Installation

```bash
go get github.com/plexusone/omni-deepgram
```

## Usage

### Batch Transcription (File/URL)

```go
import (
    deepgramstt "github.com/plexusone/omni-deepgram/omnivoice/stt"
    "github.com/plexusone/omnivoice/stt"
)

// Create provider with API key
provider, err := deepgramstt.New(deepgramstt.WithAPIKey("your-api-key"))
if err != nil {
    log.Fatal(err)
}

config := stt.TranscriptionConfig{
    Model:    "nova-2",
    Language: "en-US",
}

// Transcribe from URL
result, err := provider.TranscribeURL(ctx, "https://example.com/audio.mp3", config)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Transcript: %s\n", result.Text)
fmt.Printf("Duration: %v\n", result.Duration)

// Access word-level timestamps
for _, segment := range result.Segments {
    for _, word := range segment.Words {
        fmt.Printf("%s: %v - %v\n", word.Text, word.StartTime, word.EndTime)
    }
}

// Transcribe from file
result, err = provider.TranscribeFile(ctx, "/path/to/audio.mp3", config)

// Transcribe from bytes
audioData, _ := os.ReadFile("/path/to/audio.mp3")
result, err = provider.Transcribe(ctx, audioData, config)
```

### Streaming Transcription (Real-time)

```go
import (
    deepgramstt "github.com/plexusone/omni-deepgram/omnivoice/stt"
    "github.com/plexusone/omnivoice/stt"
)

// Create provider with API key
provider, err := deepgramstt.New(deepgramstt.WithAPIKey("your-api-key"))
if err != nil {
    log.Fatal(err)
}

// Configure for telephony audio
config := stt.TranscriptionConfig{
    Model:      "nova-2",
    Language:   "en-US",
    Encoding:   "mulaw",
    SampleRate: 8000,
}

// Start streaming transcription
writer, events, err := provider.TranscribeStream(ctx, config)
if err != nil {
    log.Fatal(err)
}

// Send audio data
go func() {
    defer writer.Close()
    io.Copy(writer, audioSource)
}()

// Receive transcription events
for event := range events {
    switch event.Type {
    case stt.EventTranscript:
        if event.IsFinal {
            fmt.Println("Final:", event.Transcript)
        }
    case stt.EventSpeechStart:
        fmt.Println("Speech started")
    case stt.EventSpeechEnd:
        fmt.Println("Speech ended")
    case stt.EventError:
        log.Printf("Error: %v", event.Error)
    }
}
```

### Basic Text-to-Speech

```go
import (
    deepgramtts "github.com/plexusone/omni-deepgram/omnivoice/tts"
    "github.com/plexusone/omnivoice/tts"
)

// Create TTS provider with API key
provider, err := deepgramtts.New(deepgramtts.WithAPIKey("your-api-key"))
if err != nil {
    log.Fatal(err)
}

// Configure synthesis
config := tts.SynthesisConfig{
    VoiceID:      "aura-asteria-en",  // Female US voice
    OutputFormat: "mp3",
    SampleRate:   24000,
}

// Synthesize text to speech
result, err := provider.Synthesize(ctx, "Hello, world!", config)
if err != nil {
    log.Fatal(err)
}

// result.Audio contains the synthesized audio bytes
fmt.Printf("Generated %d bytes of audio\n", len(result.Audio))
```

### Streaming Text-to-Speech

```go
// Start streaming synthesis
chunkCh, err := provider.SynthesizeStream(ctx, "Hello, this is streaming TTS.", config)
if err != nil {
    log.Fatal(err)
}

// Receive audio chunks as they're generated
for chunk := range chunkCh {
    if chunk.Error != nil {
        log.Printf("Error: %v", chunk.Error)
        break
    }
    if len(chunk.Audio) > 0 {
        // Process or play audio chunk
        audioPlayer.Write(chunk.Audio)
    }
    if chunk.IsFinal {
        fmt.Println("Synthesis complete")
    }
}
```

### List Available Voices

```go
voices, err := provider.ListVoices(ctx)
if err != nil {
    log.Fatal(err)
}

for _, voice := range voices {
    fmt.Printf("%s: %s (%s, %s)\n", voice.ID, voice.Name, voice.Language, voice.Gender)
}
```

### Streaming Input from LLM

Stream text from an LLM directly to TTS for low-latency voice responses:

```go
// Create a pipe to connect LLM output to TTS input
pr, pw := io.Pipe()

// Start streaming synthesis from the reader
chunkCh, err := provider.SynthesizeFromReader(ctx, pr, config)
if err != nil {
    log.Fatal(err)
}

// Simulate streaming LLM output in a goroutine
go func() {
    defer pw.Close()

    // Write text chunks as they arrive from LLM
    pw.Write([]byte("Hello! "))
    pw.Write([]byte("This is streaming from an LLM. "))
    pw.Write([]byte("Each sentence is synthesized as it arrives."))
}()

// Receive audio chunks as they're generated
for chunk := range chunkCh {
    if chunk.Error != nil {
        log.Printf("Error: %v", chunk.Error)
        break
    }
    if len(chunk.Audio) > 0 {
        audioPlayer.Write(chunk.Audio)
    }
}
```

### With OmniVoice Pipeline

For a complete voice agent example using Deepgram STT and TTS with Twilio Media Streams, see the [omnivoice-examples](https://github.com/plexusone/omnivoice-examples) repository.

## Supported Audio Formats

| Format | Encoding Value | Typical Use |
|--------|---------------|-------------|
| mu-law | `mulaw` | Twilio, telephony |
| A-law | `alaw` | European telephony |
| Linear PCM | `linear16` | General audio |
| FLAC | `flac` | Compressed lossless |
| Opus | `opus` | WebRTC |
| MP3 | `mp3` | Compressed lossy |

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `Model` | Deepgram model | `nova-2` |
| `Language` | Language code | `en-US` |
| `SampleRate` | Audio sample rate | `8000` |
| `Channels` | Audio channels | `1` |
| `EnablePunctuation` | Add punctuation | `false` |
| `EnableSpeakerDiarization` | Identify speakers | `false` |
| `Keywords` | Words to boost | `[]` |

## Requirements

- Go 1.21 or later
- Deepgram API key ([get one here](https://console.deepgram.com/))

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related Projects

- [omnivoice](https://github.com/plexusone/omnivoice) - Voice agent framework interfaces
- [go-elevenlabs](https://github.com/plexusone/go-elevenlabs) - ElevenLabs TTS provider
- [omnivoice-twilio](https://github.com/plexusone/omnivoice-twilio) - Twilio Media Streams transport
- [omnivoice-examples](https://github.com/plexusone/omnivoice-examples) - Complete voice agent examples
