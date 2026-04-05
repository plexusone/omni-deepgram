# v0.3.0 — Batch Transcription Support

This release adds batch transcription support for pre-recorded audio files and URLs, complementing the existing real-time streaming transcription.

## Highlights

- Batch transcription support for pre-recorded audio files and URLs

## Added

- `Transcribe` method for batch transcription of audio bytes ([`d64c4ea`](https://github.com/agentplexus/omnivoice-deepgram/commit/d64c4ea))
- `TranscribeFile` method for batch transcription of local audio files ([`d64c4ea`](https://github.com/agentplexus/omnivoice-deepgram/commit/d64c4ea))
- `TranscribeURL` method for batch transcription of remote audio URLs ([`d64c4ea`](https://github.com/agentplexus/omnivoice-deepgram/commit/d64c4ea))
- Pre-recorded transcription conversion helpers for Deepgram responses ([`3fc08dd`](https://github.com/agentplexus/omnivoice-deepgram/commit/3fc08dd))

## Tests

- `TranscribeFile` and `TranscribeURL` conformance tests with Ferris Bueller test audio ([`35c593e`](https://github.com/agentplexus/omnivoice-deepgram/commit/35c593e))

## Dependencies

- Bump `github.com/agentplexus/omnivoice` from v0.3.0 to v0.4.1 ([`35c593e`](https://github.com/agentplexus/omnivoice-deepgram/commit/35c593e))

## Documentation

- README updated with batch transcription examples ([`fe09c20`](https://github.com/agentplexus/omnivoice-deepgram/commit/fe09c20))

## Quick Start

```go
import (
    deepgramstt "github.com/agentplexus/omnivoice-deepgram/omnivoice/stt"
    "github.com/agentplexus/omnivoice/stt"
)

// Create STT provider
provider, err := deepgramstt.New(deepgramstt.WithAPIKey("your-api-key"))
if err != nil {
    log.Fatal(err)
}

// Transcribe a local file
result, err := provider.TranscribeFile(ctx, "audio.mp3", stt.TranscriptionConfig{
    Language: "en",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Text)

// Transcribe from URL
result, err = provider.TranscribeURL(ctx, "https://example.com/audio.mp3", stt.TranscriptionConfig{
    Language: "en",
})
```

See the [README](https://github.com/agentplexus/omnivoice-deepgram#readme) for more examples.
