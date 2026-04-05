# Release Notes: v0.3.1

**Release Date:** 2026-02-22

## Summary

Bug fix release that resolves a klog flag conflict when using both STT and TTS providers together.

## Bug Fixes

### Shared SDK Initialization

Fixed a panic that occurred when instantiating both STT and TTS providers in the same application. The Deepgram SDK's `Init()` function registers klog flags with Go's flag package, and calling it twice caused a "flag redefined" panic.

**Solution:** Moved SDK initialization to a shared `InitSDK()` function in the `omnivoice` package. Both STT and TTS providers now call this shared function, which uses `sync.Once` to ensure initialization happens only once.

**Before (panic):**
```go
sttProvider, _ := stt.New(stt.WithAPIKey(key))
ttsProvider, _ := tts.New(tts.WithAPIKey(key)) // PANIC: flag redefined
```

**After (works):**
```go
sttProvider, _ := stt.New(stt.WithAPIKey(key))
ttsProvider, _ := tts.New(tts.WithAPIKey(key)) // OK
```

## Upgrade Guide

This is a drop-in replacement for v0.3.0. No code changes required.

```bash
go get github.com/agentplexus/omnivoice-deepgram@v0.3.1
```
