// Command deepgram-voice-test is a CLI tool for testing the Deepgram Voice Agent provider.
//
// Usage:
//
//	export DEEPGRAM_API_KEY="your-api-key"
//	go run ./cmd/deepgram-voice-test --greeting "Hello! How can I help?"
//
// This tool connects to Deepgram Voice Agent and prints events to stdout.
// It can be used for debugging the realtime provider implementation.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/plexusone/omni-deepgram/omnivoice/realtime"
	corereal "github.com/plexusone/omnivoice-core/realtime"
	"github.com/spf13/cobra"
)

var (
	flagGreeting     string
	flagVoice        string
	flagLanguage     string
	flagInstructions string
	flagDuration     time.Duration
	flagExperimental bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "deepgram-voice-test",
		Short: "Test the Deepgram Voice Agent provider",
		Long: `A CLI tool for testing the Deepgram Voice Agent provider.

This tool connects to Deepgram's Voice Agent API and prints events to stdout.
It sends silence to trigger the greeting and monitors for responses.

Environment:
  DEEPGRAM_API_KEY  Required. Your Deepgram API key.

Examples:
  # Basic test with greeting
  deepgram-voice-test --greeting "Hello! How can I help you?"

  # Test with custom voice and instructions
  deepgram-voice-test --voice aura-2-thalia-en --instructions "You are a helpful assistant."

  # Run for a specific duration
  deepgram-voice-test --duration 30s`,
		RunE: runTest,
	}

	rootCmd.Flags().StringVar(&flagGreeting, "greeting", "", "Greeting message for the agent to speak")
	rootCmd.Flags().StringVar(&flagVoice, "voice", "", "Voice model (e.g., aura-2-thalia-en)")
	rootCmd.Flags().StringVar(&flagLanguage, "language", "en", "Language code")
	rootCmd.Flags().StringVar(&flagInstructions, "instructions", "You are a helpful assistant.", "System instructions")
	rootCmd.Flags().DurationVar(&flagDuration, "duration", 30*time.Second, "How long to run the test")
	rootCmd.Flags().BoolVar(&flagExperimental, "experimental", true, "Enable experimental features (latency metrics)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runTest(cmd *cobra.Command, args []string) error {
	apiKey := os.Getenv("DEEPGRAM_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("DEEPGRAM_API_KEY environment variable is required")
	}

	fmt.Println("=== Deepgram Voice Agent Test ===")
	fmt.Printf("Language: %s\n", flagLanguage)
	fmt.Printf("Voice: %s\n", flagVoice)
	fmt.Printf("Greeting: %s\n", flagGreeting)
	fmt.Printf("Duration: %s\n", flagDuration)
	fmt.Println()

	// Build provider options
	opts := []realtime.Option{
		realtime.WithAPIKey(apiKey),
		realtime.WithLanguage(flagLanguage),
		realtime.WithExperimental(flagExperimental),
	}

	if flagVoice != "" {
		opts = append(opts, realtime.WithModel(flagVoice))
	}
	if flagGreeting != "" {
		opts = append(opts, realtime.WithGreeting(flagGreeting))
	}

	// Create provider
	fmt.Println("Creating provider...")
	provider, err := realtime.New(opts...)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}
	defer provider.Close()

	// Setup context with cancellation
	ctx, cancel := context.WithTimeout(context.Background(), flagDuration)
	defer cancel()

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nReceived interrupt, shutting down...")
		cancel()
	}()

	// Create audio input channel
	audioIn := make(chan []byte, 100)

	// Start the session
	fmt.Println("Starting audio stream...")
	audioCh, transcriptCh, err := provider.ProcessAudioStream(ctx, audioIn, corereal.ProcessConfig{
		Instructions: flagInstructions,
		Voice:        flagVoice,
	})
	if err != nil {
		return fmt.Errorf("failed to start audio stream: %w", err)
	}

	fmt.Println("Connected to Deepgram Voice Agent")
	fmt.Println()

	// Send silence to keep connection alive and trigger greeting
	go func() {
		silence := make([]byte, 320) // 10ms of silence at 16kHz mono
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				select {
				case audioIn <- silence:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Track statistics
	var totalAudioBytes int
	var audioChunks int
	var transcriptCount int
	startTime := time.Now()

	// Process events
	fmt.Println("Listening for events...")
	fmt.Println("----------------------------------------")

	for {
		select {
		case <-ctx.Done():
			goto done

		case chunk, ok := <-audioCh:
			if !ok {
				fmt.Println("[Audio channel closed]")
				goto done
			}
			if len(chunk.Audio) > 0 {
				audioChunks++
				totalAudioBytes += len(chunk.Audio)
				elapsed := time.Since(startTime).Truncate(time.Millisecond)
				fmt.Printf("[%s] AUDIO: %d bytes (chunk #%d, total: %d bytes)\n",
					elapsed, len(chunk.Audio), audioChunks, totalAudioBytes)
			}
			if chunk.IsFinal {
				elapsed := time.Since(startTime).Truncate(time.Millisecond)
				fmt.Printf("[%s] AUDIO_DONE: Agent finished speaking\n", elapsed)
			}

		case transcript, ok := <-transcriptCh:
			if !ok {
				fmt.Println("[Transcript channel closed]")
				goto done
			}
			if transcript.Text != "" {
				transcriptCount++
				elapsed := time.Since(startTime).Truncate(time.Millisecond)
				role := "AGENT"
				if transcript.IsInput {
					role = "USER"
				}
				fmt.Printf("[%s] %s: %q (final=%v)\n",
					elapsed, role, transcript.Text, transcript.IsFinal)
			}
		}
	}

done:
	close(audioIn)

	fmt.Println("----------------------------------------")
	fmt.Println()
	fmt.Println("=== Test Summary ===")
	fmt.Printf("Duration: %s\n", time.Since(startTime).Truncate(time.Second))
	fmt.Printf("Audio chunks received: %d\n", audioChunks)
	fmt.Printf("Total audio bytes: %d\n", totalAudioBytes)
	fmt.Printf("Transcripts received: %d\n", transcriptCount)

	return nil
}
