// Package realtime provides an OmniVoice realtime provider implementation using Deepgram Voice Agent.
//
// This package adapts the Deepgram Voice Agent API to the omnivoice-core realtime.Provider
// interface, enabling native voice-to-voice conversations with ~100-300ms latency.
//
// # Usage
//
//	import (
//	    "github.com/plexusone/omni-deepgram/omnivoice/realtime"
//	    corereal "github.com/plexusone/omnivoice-core/realtime"
//	)
//
//	// Create provider with API key
//	provider, err := realtime.New(realtime.WithAPIKey("your-api-key"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer provider.Close()
//
//	// Start a voice session
//	audioCh, transcriptCh, err := provider.ProcessAudioStream(ctx, audioIn, corereal.ProcessConfig{
//	    Instructions: "You are a helpful assistant.",
//	    Voice:        "aura-2-thalia-en",
//	})
//
//	// Process output
//	for {
//	    select {
//	    case chunk, ok := <-audioCh:
//	        if !ok {
//	            return
//	        }
//	        // Play audio to user
//	    case transcript, ok := <-transcriptCh:
//	        if !ok {
//	            return
//	        }
//	        // Display transcript
//	    }
//	}
//
// # Features
//
// The Deepgram Voice Agent provides:
//   - Native voice-to-voice with ~100-300ms latency
//   - Real-time transcription for both user and agent
//   - Function calling support
//   - Multiple voice options
//   - Configurable LLM providers
package realtime
