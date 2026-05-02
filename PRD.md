# OmniVoice Deepgram Provider - Product Requirements Document

This document defines the feature scope for the omni-deepgram provider, mapping capabilities across OmniVoice interfaces, the Deepgram SDK, and the current implementation status.

## Feature Matrix

### Legend

- [x] Supported/Implemented
- [ ] Not supported/Not implemented
- N/A - Not applicable

---

## Speech-to-Text (STT)

### Core STT Methods

| Feature | OmniVoice Interface | Deepgram SDK | omni-deepgram |
|---------|:-------------------:|:------------:|:------------------:|
| `Transcribe` (batch from bytes) | [x] | [x] | [ ] |
| `TranscribeFile` (from file path) | [x] | [x] | [ ] |
| `TranscribeURL` (from URL) | [x] | [x] | [ ] |
| `TranscribeStream` (real-time streaming) | [x] | [x] | [x] |

### STT Streaming Features

| Feature | OmniVoice Interface | Deepgram SDK | omni-deepgram |
|---------|:-------------------:|:------------:|:------------------:|
| Interim/partial results | [x] | [x] | [x] |
| Final results | [x] | [x] | [x] |
| Speech start detection (`EventSpeechStart`) | [x] | [x] | [x] |
| Speech end detection (`EventSpeechEnd`) | [x] | [x] | [x] |
| Utterance end detection | [x] | [x] | [x] |

### STT Configuration Options

| Feature | OmniVoice Interface | Deepgram SDK | omni-deepgram |
|---------|:-------------------:|:------------:|:------------------:|
| Language selection | [x] | [x] | [x] |
| Model selection | [x] | [x] | [x] |
| Sample rate | [x] | [x] | [x] |
| Channels | [x] | [x] | [x] |
| Encoding format | [x] | [x] | [x] |
| Punctuation | [x] | [x] | [x] |
| Word-level timestamps | [x] | [x] | [x] |
| Confidence scores | [x] | [x] | [x] |
| Speaker diarization | [x] | [x] | [x] |
| Max speakers | [x] | [x] | [x] |
| Keyword boosting | [x] | [x] | [x] |
| Custom vocabulary | [x] | [x] | [ ] |

### STT Advanced Features (Deepgram-specific)

| Feature | OmniVoice Interface | Deepgram SDK | omni-deepgram |
|---------|:-------------------:|:------------:|:------------------:|
| Smart formatting | N/A | [x] | [ ] |
| Profanity filtering | N/A | [x] | [ ] |
| Redaction | N/A | [x] | [ ] |
| Numerals formatting | N/A | [x] | [ ] |
| Search terms | N/A | [x] | [ ] |
| Replace terms | N/A | [x] | [ ] |
| Interim results rate control | N/A | [x] | [ ] |
| Endpointing | N/A | [x] | [ ] |
| VAD events | N/A | [x] | [ ] |
| Multichannel | N/A | [x] | [ ] |

---

## Text-to-Speech (TTS)

### Core TTS Methods

| Feature | OmniVoice Interface | Deepgram SDK | omni-deepgram |
|---------|:-------------------:|:------------:|:------------------:|
| `Synthesize` (batch, returns bytes) | [x] | [x] | [x] |
| `SynthesizeStream` (streaming output) | [x] | [x] | [x] |
| `SynthesizeFromReader` (streaming input) | [x] | [x] | [x] |
| `ListVoices` | [x] | [ ] | [x] |
| `GetVoice` | [x] | [ ] | [x] |

### TTS Deepgram SDK Methods

| Feature | OmniVoice Interface | Deepgram SDK | omni-deepgram |
|---------|:-------------------:|:------------:|:------------------:|
| `ToStream` (REST, to buffer) | N/A | [x] | [x] |
| `ToFile` (REST, to io.Writer) | N/A | [x] | [ ] |
| `ToSave` (REST, to file path) | N/A | [x] | [ ] |
| WebSocket streaming (callback) | N/A | [x] | [x] |
| WebSocket streaming (channel) | N/A | [x] | [ ] |
| `SpeakWithText` (WS send text) | N/A | [x] | [x] |
| `Flush` (WS flush buffer) | N/A | [x] | [x] |
| `Reset` (WS clear buffer) | N/A | [x] | [ ] |

### TTS Configuration Options

| Feature | OmniVoice Interface | Deepgram SDK | omni-deepgram |
|---------|:-------------------:|:------------:|:------------------:|
| Voice/Model selection | [x] | [x] | [x] |
| Output format | [x] | [x] | [x] |
| Sample rate | [x] | [x] | [x] |
| Bit rate | N/A | [x] | [ ] |
| Speed control | [x] | [ ] | [ ] |
| Pitch control | [x] | [ ] | [ ] |
| Stability | [x] | [ ] | [ ] |
| Similarity boost | [x] | [ ] | [ ] |

### TTS WebSocket Events

| Feature | OmniVoice Interface | Deepgram SDK | omni-deepgram |
|---------|:-------------------:|:------------:|:------------------:|
| Open event | N/A | [x] | [x] |
| Metadata event | N/A | [x] | [x] |
| Binary audio chunks | [x] (StreamChunk) | [x] | [x] |
| Flushed event | N/A | [x] | [x] |
| Cleared event | N/A | [x] | [x] |
| Warning event | N/A | [x] | [x] |
| Error event | [x] (StreamChunk.Error) | [x] | [x] |
| Close event | N/A | [x] | [x] |

---

## Audio Formats

### STT Input Formats

| Format | OmniVoice Interface | Deepgram SDK | omni-deepgram |
|--------|:-------------------:|:------------:|:------------------:|
| mu-law (`mulaw`) | [x] | [x] | [x] |
| A-law (`alaw`) | [x] | [x] | [x] |
| Linear PCM (`linear16`) | [x] | [x] | [x] |
| MP3 (`mp3`) | [x] | [x] | [x] |
| WAV (`wav`) | [x] | [x] | [x] |
| Opus (`opus`) | [x] | [x] | [x] |
| FLAC (`flac`) | [x] | [x] | [x] |
| WebM (`webm`) | N/A | [x] | [ ] |
| OGG (`ogg`) | N/A | [x] | [ ] |

### TTS Output Formats

| Format | OmniVoice Interface | Deepgram SDK | omni-deepgram |
|--------|:-------------------:|:------------:|:------------------:|
| Linear PCM (`linear16`) | [x] | [x] | [x] |
| MP3 (`mp3`) | [x] | [x] | [x] |
| WAV (`wav`) | [x] | [x] | [x] |
| Opus (`opus`) | [x] | [x] | [x] |
| FLAC (`flac`) | [x] | [x] | [x] |
| mu-law (`mulaw`) | [x] | [x] | [x] |
| A-law (`alaw`) | [x] | [x] | [x] |

---

## Deepgram Models

### STT Models

| Model | Description | omni-deepgram |
|-------|-------------|:------------------:|
| `nova-2` | Latest general model (default) | [x] |
| `nova-2-phonecall` | Optimized for phone calls | [x] |
| `nova-2-meeting` | Optimized for meetings | [x] |
| `nova-2-voicemail` | Optimized for voicemail | [x] |
| `nova-2-finance` | Financial domain | [x] |
| `nova-2-medical` | Medical domain | [x] |
| `nova` | Previous generation | [x] |
| `enhanced` | Enhanced accuracy | [x] |
| `base` | Base model | [x] |

### TTS Models/Voices

| Model | Description | omni-deepgram |
|-------|-------------|:------------------:|
| `aura-asteria-en` | Default English voice | [x] |
| `aura-2-thalia-en` | Aura 2 English voice | [x] |
| Additional Aura voices | Various voices | [x] |

---

## API Constraints

| Constraint | Value | Notes |
|------------|-------|-------|
| TTS max characters per request | 2000 | Deepgram limit |
| STT max audio duration (streaming) | Unlimited | Keep-alive required |
| STT max audio duration (batch) | Provider dependent | Check Deepgram docs |
| Rate limiting | Project-specific | 429 error when exceeded |

---

## Implementation Priority

### Phase 1: Complete STT (Done)

- [x] Streaming transcription
- [ ] Batch transcription (`Transcribe`)
- [ ] File transcription (`TranscribeFile`)
- [ ] URL transcription (`TranscribeURL`)

### Phase 2: TTS Non-Streaming (Done)

- [x] `Synthesize` (batch) via REST API
- [x] Voice listing (`ListVoices`)
- [x] Voice retrieval (`GetVoice`)

### Phase 3: TTS Streaming (Done)

- [x] `SynthesizeStream` via WebSocket
- [x] `SynthesizeFromReader` for LLM integration

### Phase 4: Advanced Features

- [ ] Custom vocabulary support
- [ ] Smart formatting
- [ ] Redaction/profanity filtering
- [ ] Advanced TTS controls

---

## References

- [OmniVoice Repository](https://github.com/plexusone/omnivoice)
- [Deepgram Go SDK](https://github.com/deepgram/deepgram-go-sdk)
- [Deepgram STT API Docs](https://developers.deepgram.com/docs/getting-started-with-live-streaming-audio)
- [Deepgram TTS API Docs](https://developers.deepgram.com/docs/text-to-speech)
- [Deepgram TTS REST Reference](https://developers.deepgram.com/reference/text-to-speech-api)
- [Deepgram TTS WebSocket Reference](https://developers.deepgram.com/reference/transform-text-to-speech-websocket)
