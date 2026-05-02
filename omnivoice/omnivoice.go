// Package omnivoice provides OmniVoice provider implementations using Deepgram.
//
// This package adapts the official Deepgram Go SDK to OmniVoice interfaces,
// enabling Deepgram's speech-to-text capabilities within the OmniVoice framework.
//
// # Usage
//
// For STT (Speech-to-Text):
//
//	import (
//	    "github.com/plexusone/omni-deepgram/omnivoice/stt"
//	    "github.com/plexusone/omnivoice-core/stt"
//	)
//
//	// Create provider with API key
//	provider, err := deepgramstt.New(deepgramstt.WithAPIKey("your-api-key"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Use with OmniVoice STT client
//	client := stt.NewClient(provider)
//
// # Streaming Transcription
//
// The STT provider supports real-time streaming transcription, ideal for
// voice agents and telephony applications:
//
//	config := stt.TranscriptionConfig{
//	    Model:      "nova-2",
//	    Language:   "en-US",
//	    Encoding:   "mulaw",    // Telephony format
//	    SampleRate: 8000,       // Telephony sample rate
//	}
//
//	stream, err := provider.TranscribeStream(ctx, audioReader, config)
//	for event := range stream {
//	    if event.IsFinal {
//	        fmt.Println("Final:", event.Transcript)
//	    }
//	}
package omnivoice

import (
	"sync"

	client "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/listen"
)

// ProviderName is the identifier for the Deepgram provider.
const ProviderName = "deepgram"

// Version is the version of this OmniVoice adapter.
const Version = "0.1.0"

// sdkInitOnce ensures the Deepgram SDK is initialized only once across all providers.
var sdkInitOnce sync.Once

// InitSDK initializes the Deepgram SDK. Safe to call multiple times.
func InitSDK() {
	sdkInitOnce.Do(func() {
		client.Init(client.InitLib{
			LogLevel: client.LogLevelDefault,
		})
	})
}
