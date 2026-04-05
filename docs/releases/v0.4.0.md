# Release Notes: v0.4.0

**Release Date:** 2026-02-28

## Summary

Organization rename from agentplexus to plexusone with updated module path.

## Breaking Changes

| Component | Before | After |
|-----------|--------|-------|
| Go module | `github.com/agentplexus/omnivoice-deepgram` | `github.com/plexusone/omnivoice-deepgram` |

## Migration Guide

Update your import paths:

```go
// Before
import "github.com/agentplexus/omnivoice-deepgram/omnivoice/stt"
import "github.com/agentplexus/omnivoice-deepgram/omnivoice/tts"

// After
import "github.com/plexusone/omnivoice-deepgram/omnivoice/stt"
import "github.com/plexusone/omnivoice-deepgram/omnivoice/tts"
```

Update your `go.mod`:

```bash
go mod edit -droprequire github.com/agentplexus/omnivoice-deepgram
go get github.com/plexusone/omnivoice-deepgram@v0.4.0
go mod tidy
```

## Dependencies

- Updated to `github.com/plexusone/omnivoice-core` v0.5.0 (was `github.com/agentplexus/omnivoice`)
