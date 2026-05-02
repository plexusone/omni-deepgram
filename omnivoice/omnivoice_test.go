package omnivoice_test

import (
	"testing"

	"github.com/plexusone/omni-deepgram/omnivoice"
	"github.com/plexusone/omni-deepgram/omnivoice/stt"
	"github.com/plexusone/omni-deepgram/omnivoice/tts"
)

// TestDualProviderInstantiation verifies that both STT and TTS providers
// can be instantiated together without klog flag redefinition panic.
// This is a regression test for the shared SDK init fix.
func TestDualProviderInstantiation(t *testing.T) {
	apiKey := "test-api-key"

	// Create STT provider first
	sttProvider, err := stt.New(stt.WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("stt.New() error = %v", err)
	}
	if sttProvider == nil {
		t.Fatal("stt.New() returned nil provider")
	}

	// Create TTS provider second - this should NOT panic
	ttsProvider, err := tts.New(tts.WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("tts.New() error = %v", err)
	}
	if ttsProvider == nil {
		t.Fatal("tts.New() returned nil provider")
	}

	// Verify both providers are functional
	if sttProvider.Name() != omnivoice.ProviderName {
		t.Errorf("sttProvider.Name() = %q, want %q", sttProvider.Name(), omnivoice.ProviderName)
	}
	if ttsProvider.Name() != omnivoice.ProviderName {
		t.Errorf("ttsProvider.Name() = %q, want %q", ttsProvider.Name(), omnivoice.ProviderName)
	}
}

// TestDualProviderInstantiationReverse verifies the same but with TTS first.
func TestDualProviderInstantiationReverse(t *testing.T) {
	apiKey := "test-api-key"

	// Create TTS provider first
	ttsProvider, err := tts.New(tts.WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("tts.New() error = %v", err)
	}

	// Create STT provider second - this should NOT panic
	sttProvider, err := stt.New(stt.WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("stt.New() error = %v", err)
	}

	// Verify both providers are functional
	if sttProvider.Name() != omnivoice.ProviderName {
		t.Errorf("sttProvider.Name() = %q, want %q", sttProvider.Name(), omnivoice.ProviderName)
	}
	if ttsProvider.Name() != omnivoice.ProviderName {
		t.Errorf("ttsProvider.Name() = %q, want %q", ttsProvider.Name(), omnivoice.ProviderName)
	}
}

func TestInitSDK(t *testing.T) {
	// InitSDK should be safe to call multiple times
	omnivoice.InitSDK()
	omnivoice.InitSDK()
	omnivoice.InitSDK()
	// If we get here without panic, the test passes
}
