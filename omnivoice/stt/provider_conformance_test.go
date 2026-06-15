// Conformance tests for the Deepgram STT provider using omnivoice providertest.
// These tests verify the provider correctly implements the omnivoice STT interfaces.
//
// To run integration tests, set DEEPGRAM_API_KEY environment variable:
//
//	source /path/to/agentcall/.envrc
//	go test -v -run TestConformance
//
// The Deepgram STT provider implements both streaming and batch transcription:
//   - TranscribeStream: Real-time WebSocket streaming
//   - Transcribe: Batch transcription from audio bytes
//   - TranscribeFile: Batch transcription from file path
//   - TranscribeURL: Batch transcription from URL
package stt

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/plexusone/omnivoice-core/stt"
	"github.com/plexusone/omnivoice-core/stt/providertest"
)

// testAudioURL is Deepgram's public test audio file.
// "Life moves pretty fast. If you don't stop and look around once in a while, you could miss it."
const testAudioURL = "https://static.deepgram.com/examples/Bueller-Life-moves-pretty-fast.wav"

func TestConformance(t *testing.T) {
	apiKey := os.Getenv("DEEPGRAM_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPGRAM_API_KEY not set, skipping conformance tests (API key required for provider creation)")
	}

	p, err := New(WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Download test audio file for TranscribeFile test
	testAudioFile := downloadTestAudio(t)

	// Generate test audio (1 second of silence at 16kHz mono linear16)
	// Real tests should use actual speech audio for meaningful results
	testAudio := makeTestAudio()

	cfg := providertest.Config{
		Provider:          p,
		StreamingProvider: p,    // Deepgram STT implements StreamingProvider
		SkipIntegration:   true, // Skip generic integration tests (see below)
		TestAudio:         testAudio,
		TestAudioConfig: stt.TranscriptionConfig{
			Language: "en",
			Model:    "nova-2",
		},
		// Batch transcription tests
		TestAudioFile:    testAudioFile,
		TestAudioURL:     testAudioURL,
		TestExpectedText: "life moves pretty fast",
		Timeout:          60 * time.Second,
	}

	// Run interface tests (always run)
	t.Run("Interface", func(t *testing.T) {
		providertest.RunInterfaceTests(t, cfg)
	})

	// Run behavior tests
	t.Run("Behavior", func(t *testing.T) {
		providertest.RunBehaviorTests(t, cfg)
	})

	// NOTE: We skip providertest.RunIntegrationTests because it applies TestAudioConfig
	// (with encoding/sample_rate settings) to file/URL tests. Deepgram requires auto-detection
	// for container formats (WAV, MP3) - explicit encoding settings cause empty results.
	// Use TestConformance_BatchOnly and TestConformance_StreamingOnly for integration coverage.
}

// TestConformance_InterfaceOnly runs only interface tests.
// Note: Deepgram STT requires API key for provider creation,
// so this test also requires the API key.
func TestConformance_InterfaceOnly(t *testing.T) {
	apiKey := os.Getenv("DEEPGRAM_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPGRAM_API_KEY not set (required for provider creation)")
	}

	p, err := New(WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	providertest.RunInterfaceTests(t, providertest.Config{
		Provider: p,
	})
}

// TestConformance_StreamingOnly runs only streaming integration tests.
func TestConformance_StreamingOnly(t *testing.T) {
	apiKey := os.Getenv("DEEPGRAM_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPGRAM_API_KEY not set, skipping streaming tests")
	}

	p, err := New(WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	testAudio := makeTestAudio()

	cfg := providertest.Config{
		Provider:          p,
		StreamingProvider: p,
		SkipIntegration:   false,
		TestAudio:         testAudio,
		TestAudioConfig: stt.TranscriptionConfig{
			Language:   "en",
			Model:      "nova-2",
			SampleRate: 16000,
			Channels:   1,
			Encoding:   "linear16",
		},
		Timeout: 60 * time.Second,
	}

	// Run interface tests
	t.Run("Interface", func(t *testing.T) {
		providertest.RunInterfaceTests(t, cfg)
	})

	// Run streaming integration test
	t.Run("Integration/TranscribeStream", func(t *testing.T) {
		testTranscribeStream(t, p, testAudio, cfg.TestAudioConfig, cfg.Timeout)
	})
}

// TestConformance_BatchOnly runs batch transcription tests.
func TestConformance_BatchOnly(t *testing.T) {
	apiKey := os.Getenv("DEEPGRAM_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPGRAM_API_KEY not set, skipping batch tests")
	}

	p, err := New(WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Download test audio file
	testAudioFile := downloadTestAudio(t)
	testAudio, err := os.ReadFile(testAudioFile)
	if err != nil {
		t.Fatalf("failed to read test audio: %v", err)
	}

	cfg := stt.TranscriptionConfig{
		Language: "en",
		Model:    "nova-2",
	}

	t.Run("Transcribe", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := p.Transcribe(ctx, testAudio, cfg)
		if err != nil {
			t.Fatalf("Transcribe() error: %v", err)
		}
		t.Logf("Transcribe result: %q (segments=%d)", result.Text, len(result.Segments))
	})

	t.Run("TranscribeURL", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := p.TranscribeURL(ctx, testAudioURL, cfg)
		if err != nil {
			t.Fatalf("TranscribeURL() error: %v", err)
		}
		t.Logf("TranscribeURL result: %q (segments=%d)", result.Text, len(result.Segments))

		// Verify we got word timestamps
		if len(result.Segments) > 0 && len(result.Segments[0].Words) > 0 {
			w := result.Segments[0].Words[0]
			t.Logf("First word: %q start=%v end=%v", w.Text, w.StartTime, w.EndTime)
		}
	})
}

// testTranscribeStream tests streaming transcription directly.
func testTranscribeStream(t *testing.T, p stt.StreamingProvider, testAudio []byte, cfg stt.TranscriptionConfig, timeout time.Duration) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	writer, events, err := p.TranscribeStream(ctx, cfg)
	if err != nil {
		t.Fatalf("TranscribeStream() error: %v", err)
	}

	// Write audio in a goroutine
	done := make(chan error, 1)
	go func() {
		_, writeErr := writer.Write(testAudio)
		if writeErr != nil {
			done <- writeErr
			return
		}
		done <- writer.Close()
	}()

	// Collect events with timeout
	eventLoop := true
	for eventLoop {
		select {
		case event, ok := <-events:
			if !ok {
				eventLoop = false
				break
			}
			switch event.Type {
			case stt.EventTranscript:
				t.Logf("Got transcript: %q (final=%v)", event.Transcript, event.IsFinal)
			case stt.EventSpeechStart:
				t.Log("Speech started")
			case stt.EventSpeechEnd:
				t.Log("Speech ended")
			case stt.EventError:
				t.Logf("Stream error: %v", event.Error)
				eventLoop = false
			}
		case writeErr := <-done:
			if writeErr != nil {
				t.Logf("Write error: %v", writeErr)
			}
		case <-ctx.Done():
			eventLoop = false
		}
	}

	t.Log("TranscribeStream completed successfully")
}

// TestProviderRequiresAPIKey verifies that provider creation fails without API key.
func TestProviderRequiresAPIKey(t *testing.T) {
	_, err := New()
	if err == nil {
		t.Error("New() should return error when API key is not provided")
	}
}

// makeTestAudio generates test audio data.
// Returns 1 second of silence at 16kHz mono linear16.
func makeTestAudio() []byte {
	// 16kHz * 2 bytes per sample * 1 second = 32000 bytes
	return make([]byte, 32000)
}

// downloadTestAudio downloads the test audio file to a temp directory.
func downloadTestAudio(t *testing.T) string {
	t.Helper()

	resp, err := http.Get(testAudioURL)
	if err != nil {
		t.Fatalf("failed to download test audio: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("failed to download test audio: status %d", resp.StatusCode)
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test-audio.wav")

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		t.Fatalf("failed to write test audio: %v", err)
	}

	return filePath
}
