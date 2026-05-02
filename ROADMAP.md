# Roadmap

This document tracks planned features and improvements for omni-deepgram.

## Planned

### STT Batch Transcription

- **Status**: Not implemented
- **Methods to implement**:
  - `Transcribe(ctx, audio []byte, config)` - Transcribe audio bytes
  - `TranscribeFile(ctx, filePath string, config)` - Transcribe from file path
  - `TranscribeURL(ctx, url string, config)` - Transcribe from URL
- **Deepgram API**: Use prerecorded/batch transcription endpoint
- **Priority**: Medium (streaming transcription is the primary use case)

## Completed

### Core Implementation

- [x] TTS Provider with batch synthesis (`Synthesize`)
- [x] TTS Provider with streaming synthesis (`SynthesizeStream`, `SynthesizeFromReader`)
- [x] STT Provider with streaming transcription (`TranscribeStream`)
- [x] Voice listing and retrieval
- [x] Audio format conversion utilities

### Conformance Tests

- [x] TTS interface conformance tests
- [x] TTS behavior conformance tests
- [x] TTS integration tests (batch synthesis)
- [x] TTS integration tests (streaming synthesis)
- [x] STT interface conformance tests
- [x] STT streaming conformance tests
