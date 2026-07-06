package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	agentinterfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/api/agent/v1/websocket/interfaces"
	agent "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/agent"
	agentws "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/agent/v1/websocket"
	interfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/interfaces"
	"github.com/plexusone/omni-deepgram/omnivoice"
	corereal "github.com/plexusone/omnivoice-core/realtime"
)

// Verify interface compliance at compile time.
var _ corereal.Provider = (*Provider)(nil)

// Provider implements corereal.Provider using the Deepgram Voice Agent API.
type Provider struct {
	config Config
	mu     sync.Mutex
}

// New creates a new Deepgram Voice Agent provider.
func New(opts ...Option) (*Provider, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.APIKey == "" {
		return nil, ErrAPIKeyRequired
	}

	// Initialize the Deepgram agent client library
	omnivoice.InitAgentSDK()

	return &Provider{
		config: cfg,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return omnivoice.ProviderName + "-agent"
}

// Close releases any resources held by the provider.
func (p *Provider) Close() error {
	return nil
}

// ProcessAudioStream starts a real-time voice session.
func (p *Provider) ProcessAudioStream(
	ctx context.Context,
	audioIn <-chan []byte,
	config corereal.ProcessConfig,
) (<-chan corereal.AudioChunk, <-chan corereal.Transcript, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	fmt.Println("[DEEPGRAM] ProcessAudioStream starting...")

	// Create channel handler for receiving events
	handler := newChanHandler()

	// Build SettingsOptions from config
	settingsOpts := ConfigToSettingsOptions(p.config, config)
	fmt.Printf("[DEEPGRAM] Greeting: %q\n", p.config.Greeting)
	fmt.Printf("[DEEPGRAM] Language: %s, InputRate: %d, OutputRate: %d\n",
		p.config.Language, p.config.InputSampleRate, p.config.OutputSampleRate)

	// Create WebSocket client
	wsClient, err := agent.NewWSUsingChan(
		ctx,
		p.config.APIKey,
		&interfaces.ClientOptions{
			EnableKeepAlive: true,
		},
		settingsOpts,
		handler,
	)
	if err != nil {
		fmt.Printf("[DEEPGRAM] Failed to create WebSocket client: %v\n", err)
		return nil, nil, omnivoice.ClassifyError("ProcessAudioStream", err)
	}

	// Connect to Deepgram
	fmt.Println("[DEEPGRAM] Connecting to Deepgram Voice Agent...")
	if !wsClient.Connect() {
		fmt.Println("[DEEPGRAM] Connection failed!")
		return nil, nil, omnivoice.ClassifyError("ProcessAudioStream", ErrConnectionFailed)
	}
	fmt.Println("[DEEPGRAM] Connected successfully!")

	// Create output channels
	audioCh := make(chan corereal.AudioChunk, 100)
	transcriptCh := make(chan corereal.Transcript, 100)

	// Start session to begin processing
	session := &voiceSession{
		wsClient:     wsClient,
		handler:      handler,
		config:       config,
		audioCh:      audioCh,
		transcriptCh: transcriptCh,
		done:         make(chan struct{}),
	}

	// Forward audio input to Deepgram
	go session.forwardAudioInput(ctx, audioIn)

	// Process events from Deepgram
	go session.processEvents(ctx)

	return audioCh, transcriptCh, nil
}

// voiceSession manages a single voice session.
type voiceSession struct {
	wsClient     *agentws.WSChannel
	handler      *chanHandler
	config       corereal.ProcessConfig
	audioCh      chan corereal.AudioChunk
	transcriptCh chan corereal.Transcript
	done         chan struct{}
	closeOnce    sync.Once

	// turnAudioDone tracks when agent audio for current turn is complete.
	// This prevents out-of-order audio chunks from being processed after IsFinal.
	turnAudioDone bool
	turnMu        sync.Mutex
}

// forwardAudioInput reads from audioIn and sends to Deepgram.
func (s *voiceSession) forwardAudioInput(ctx context.Context, audioIn <-chan []byte) {
	defer s.cleanup()

	audioChunks := 0
	totalBytes := 0

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[DEEPGRAM] forwardAudioInput: context cancelled after %d chunks, %d bytes\n", audioChunks, totalBytes)
			return
		case <-s.done:
			fmt.Printf("[DEEPGRAM] forwardAudioInput: done signal after %d chunks, %d bytes\n", audioChunks, totalBytes)
			return
		case audio, ok := <-audioIn:
			if !ok {
				fmt.Printf("[DEEPGRAM] forwardAudioInput: input channel closed after %d chunks, %d bytes\n", audioChunks, totalBytes)
				return
			}
			if len(audio) > 0 {
				audioChunks++
				totalBytes += len(audio)
				if audioChunks == 1 {
					fmt.Printf("[DEEPGRAM] Sending audio to Deepgram! First chunk: %d bytes\n", len(audio))
				}
				if audioChunks%100 == 0 {
					fmt.Printf("[DEEPGRAM] Sent %d audio chunks to Deepgram (%d bytes)\n", audioChunks, totalBytes)
				}
				_, err := s.wsClient.Write(audio)
				if err != nil {
					fmt.Printf("[DEEPGRAM] forwardAudioInput: write error: %v\n", err)
					return
				}
			}
		}
	}
}

// processEvents reads events from Deepgram and sends to output channels.
func (s *voiceSession) processEvents(ctx context.Context) {
	defer s.cleanup()
	fmt.Println("[DEEPGRAM] processEvents started, waiting for events...")

	audioChunks := 0
	totalAudioBytes := 0

	for {
		select {
		case <-ctx.Done():
			fmt.Println("[DEEPGRAM] processEvents: context cancelled")
			return
		case <-s.done:
			fmt.Println("[DEEPGRAM] processEvents: done signal received")
			return

		// Binary audio data from agent
		case audio := <-s.handler.binaryChan:
			// Skip audio chunks that arrive after turn is complete (out-of-order from select)
			s.turnMu.Lock()
			audioDone := s.turnAudioDone
			s.turnMu.Unlock()
			if audioDone {
				continue
			}

			if audio != nil && len(*audio) > 0 {
				audioChunks++
				totalAudioBytes += len(*audio)
				if audioChunks == 1 {
					fmt.Printf("[DEEPGRAM] Receiving audio! First chunk: %d bytes\n", len(*audio))
				}
				if audioChunks%50 == 0 {
					fmt.Printf("[DEEPGRAM] Audio progress: %d chunks, %d total bytes\n", audioChunks, totalAudioBytes)
				}

				// Copy audio data
				data := make([]byte, len(*audio))
				copy(data, *audio)

				select {
				case s.audioCh <- corereal.AudioChunk{Audio: data}:
				case <-ctx.Done():
					return
				case <-s.done:
					return
				default:
					// Channel full, drop
					fmt.Println("[DEEPGRAM] Warning: audio channel full, dropping chunk")
				}
			}

		// Conversation text (transcripts)
		case text := <-s.handler.conversationTextChan:
			if text != nil && text.Content != "" {
				transcript := corereal.Transcript{
					Text:    text.Content,
					IsFinal: true, // Deepgram sends final transcripts
					IsInput: text.Role == "user",
				}

				select {
				case s.transcriptCh <- transcript:
				case <-ctx.Done():
					return
				case <-s.done:
					return
				default:
					// Channel full, drop
				}
			}

		// Agent audio done - mark as final
		case <-s.handler.agentAudioDoneChan:
			// Mark turn audio as done to prevent out-of-order chunks
			s.turnMu.Lock()
			s.turnAudioDone = true
			s.turnMu.Unlock()

			select {
			case s.audioCh <- corereal.AudioChunk{IsFinal: true}:
			case <-ctx.Done():
				return
			case <-s.done:
				return
			default:
				// Channel full, drop
			}

		// Function call request
		case funcCall := <-s.handler.functionCallRequestChan:
			if funcCall != nil {
				s.handleFunctionCall(ctx, funcCall)
			}

		// Connection closed
		case <-s.handler.closeChan:
			return

		// Error
		case errResp := <-s.handler.errorChan:
			if errResp != nil {
				// Log or handle error
				_ = errResp
			}

		// Log important events
		case <-s.handler.openChan:
			fmt.Println("[DEEPGRAM] Event: connection opened")
		case <-s.handler.welcomeChan:
			fmt.Println("[DEEPGRAM] Event: welcome received - session ready")
		case <-s.handler.userStartedSpeakingChan:
			fmt.Println("[DEEPGRAM] Event: user started speaking")
		case <-s.handler.agentThinkingChan:
			fmt.Println("[DEEPGRAM] Event: agent thinking")
		case <-s.handler.agentStartedSpeakingChan:
			// Reset turn audio done flag for new turn
			s.turnMu.Lock()
			s.turnAudioDone = false
			s.turnMu.Unlock()
			fmt.Println("[DEEPGRAM] Event: agent started speaking")
		case <-s.handler.settingsAppliedChan:
			fmt.Println("[DEEPGRAM] Event: settings applied")
		case <-s.handler.injectionRefusedChan:
			fmt.Println("[DEEPGRAM] Event: injection refused")
		case <-s.handler.keepAliveChan:
			// Don't log keepalives, too noisy
		case unhandled := <-s.handler.unhandledChan:
			if unhandled != nil {
				// Log first 200 chars of unhandled event to see what it is
				content := string(*unhandled)
				if len(content) > 200 {
					content = content[:200] + "..."
				}
				fmt.Printf("[DEEPGRAM] Unhandled event: %s\n", content)
			}
		}
	}
}

// handleFunctionCall processes a function call request from the agent.
func (s *voiceSession) handleFunctionCall(_ context.Context, req *agentinterfaces.FunctionCallRequestResponse) {
	if s.config.OnFunctionCall == nil {
		// No handler configured, send empty response
		s.sendFunctionCallResponse(req.FunctionCallID, "", nil)
		return
	}

	// Convert input to JSON string
	argsJSON, err := json.Marshal(req.Input)
	if err != nil {
		s.sendFunctionCallResponse(req.FunctionCallID, "", fmt.Errorf("failed to marshal args: %w", err))
		return
	}

	// Call the handler
	result, err := s.config.OnFunctionCall(req.FunctionCallID, req.FunctionName, string(argsJSON))
	s.sendFunctionCallResponse(req.FunctionCallID, result, err)
}

// sendFunctionCallResponse sends a function call response to Deepgram.
func (s *voiceSession) sendFunctionCallResponse(callID string, result any, err error) {
	var output string
	if err != nil {
		output = fmt.Sprintf(`{"error": %q}`, err.Error())
	} else {
		data, marshalErr := json.Marshal(result)
		if marshalErr != nil {
			output = fmt.Sprintf(`{"error": "failed to marshal result: %v"}`, marshalErr)
		} else {
			output = string(data)
		}
	}

	response := map[string]any{
		"type":             "FunctionCallResponse",
		"function_call_id": callID,
		"output":           output,
	}

	_ = s.wsClient.WriteJSON(response)
}

// cleanup closes channels and releases resources.
func (s *voiceSession) cleanup() {
	s.closeOnce.Do(func() {
		close(s.done)
		close(s.audioCh)
		close(s.transcriptCh)
		s.wsClient.Finish()
	})
}
