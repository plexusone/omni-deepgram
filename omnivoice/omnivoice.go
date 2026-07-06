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
	"os"
	"sync"

	agentclient "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/agent"
	client "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/listen"
)

// ProviderName is the identifier for the Deepgram provider.
const ProviderName = "deepgram"

// Version is the version of this OmniVoice adapter.
const Version = "0.1.0"

// sdkInitOnce ensures the Deepgram SDK is initialized only once across all providers.
var sdkInitOnce sync.Once

// sdkAgentInitOnce ensures the Deepgram Agent SDK is initialized only once.
var sdkAgentInitOnce sync.Once

// InitSDK initializes the Deepgram SDK. Safe to call multiple times.
func InitSDK() {
	sdkInitOnce.Do(func() {
		client.Init(client.InitLib{
			LogLevel: client.LogLevelDefault,
		})
	})
}

// InitAgentSDK initializes the Deepgram Agent SDK. Safe to call multiple times.
// Note: The SDK calls flag.Parse() internally, so we temporarily clear os.Args
// to prevent conflicts with CLI frameworks like Cobra that have already parsed flags.
func InitAgentSDK() {
	sdkAgentInitOnce.Do(func() {
		// Save original args and clear them to prevent SDK from parsing CLI flags
		origArgs := os.Args
		os.Args = []string{origArgs[0]}

		agentclient.Init(agentclient.InitLib{
			LogLevel: agentclient.LogLevelDefault,
		})

		// Restore original args
		os.Args = origArgs
	})
}
