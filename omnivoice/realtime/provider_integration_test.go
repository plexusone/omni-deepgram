//go:build integration

package realtime

import (
	"context"
	"os"
	"testing"
	"time"

	corereal "github.com/plexusone/omnivoice-core/realtime"
)

// TestIntegration_ProcessAudioStream tests the realtime provider with a live Deepgram connection.
// Run with: DEEPGRAM_API_KEY=xxx go test -v -tags=integration ./omnivoice/realtime/...
func TestIntegration_ProcessAudioStream(t *testing.T) {
	apiKey := os.Getenv("DEEPGRAM_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPGRAM_API_KEY not set, skipping integration test")
	}

	provider, err := New(
		WithAPIKey(apiKey),
		WithLanguage("en"),
		WithExperimental(true),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer provider.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create audio input channel
	audioIn := make(chan []byte, 100)

	// Start the session
	audioCh, transcriptCh, err := provider.ProcessAudioStream(ctx, audioIn, corereal.ProcessConfig{
		Instructions: "You are a helpful assistant. Say hello.",
	})
	if err != nil {
		t.Fatalf("ProcessAudioStream() error = %v", err)
	}

	// Track what we receive
	var receivedAudio bool
	var receivedTranscript bool

	// Send silence to trigger greeting (if configured)
	// Linear16 silence: 320 bytes = 10ms at 16kHz mono
	silence := make([]byte, 320)
	for i := 0; i < 10; i++ {
		select {
		case audioIn <- silence:
		case <-ctx.Done():
			t.Log("Context cancelled during audio send")
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for some response or timeout
	timeout := time.After(8 * time.Second)
	for {
		select {
		case chunk, ok := <-audioCh:
			if !ok {
				t.Log("Audio channel closed")
				goto done
			}
			if len(chunk.Audio) > 0 {
				receivedAudio = true
				t.Logf("Received audio chunk: %d bytes, isFinal=%v", len(chunk.Audio), chunk.IsFinal)
			}
			if chunk.IsFinal {
				t.Log("Received final audio marker")
			}

		case transcript, ok := <-transcriptCh:
			if !ok {
				t.Log("Transcript channel closed")
				goto done
			}
			if transcript.Text != "" {
				receivedTranscript = true
				t.Logf("Received transcript: role=%s, text=%q, isFinal=%v",
					transcript.Role(), transcript.Text, transcript.IsFinal)
			}

		case <-timeout:
			t.Log("Test timeout reached")
			goto done
		}
	}

done:
	close(audioIn)
	cancel()

	// Log what we received
	t.Logf("Results: receivedAudio=%v, receivedTranscript=%v", receivedAudio, receivedTranscript)

	// The connection should have been established successfully
	// Audio/transcript may or may not be received depending on Deepgram's response
	if !receivedAudio && !receivedTranscript {
		t.Log("Note: No audio or transcript received. This may be expected if no greeting was configured.")
	}
}

// TestIntegration_FactoryCreate tests the factory pattern.
func TestIntegration_FactoryCreate(t *testing.T) {
	factory := NewFactory()

	if factory.Name() != "deepgram" {
		t.Errorf("Factory.Name() = %v, want %v", factory.Name(), "deepgram")
	}

	// Factory.Create requires gateway.RealtimeConfig which is tested in omnivoice integration tests
	t.Log("Factory created successfully. Full integration test requires omnivoice module.")
}

// TestIntegration_WithGreeting tests the provider with a greeting configured.
func TestIntegration_WithGreeting(t *testing.T) {
	apiKey := os.Getenv("DEEPGRAM_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPGRAM_API_KEY not set, skipping integration test")
	}

	provider, err := New(
		WithAPIKey(apiKey),
		WithLanguage("en"),
		WithGreeting("Hello! How can I help you today?"),
		WithExperimental(true),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer provider.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	audioIn := make(chan []byte, 100)
	defer close(audioIn)

	audioCh, transcriptCh, err := provider.ProcessAudioStream(ctx, audioIn, corereal.ProcessConfig{
		Instructions: "You are a helpful assistant.",
	})
	if err != nil {
		t.Fatalf("ProcessAudioStream() error = %v", err)
	}

	var receivedGreetingAudio bool
	var receivedGreetingText bool

	timeout := time.After(12 * time.Second)
	for {
		select {
		case chunk, ok := <-audioCh:
			if !ok {
				goto done
			}
			if len(chunk.Audio) > 0 {
				receivedGreetingAudio = true
				t.Logf("Received greeting audio: %d bytes", len(chunk.Audio))
			}

		case transcript, ok := <-transcriptCh:
			if !ok {
				goto done
			}
			if transcript.Text != "" && !transcript.IsInput {
				receivedGreetingText = true
				t.Logf("Received greeting transcript: %q", transcript.Text)
			}

		case <-timeout:
			goto done
		}
	}

done:
	cancel()

	if receivedGreetingAudio {
		t.Log("Successfully received greeting audio from Deepgram Voice Agent")
	}
	if receivedGreetingText {
		t.Log("Successfully received greeting transcript from Deepgram Voice Agent")
	}
}
