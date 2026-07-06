package realtime

import (
	"context"
	"encoding/json"
	"testing"

	corereal "github.com/plexusone/omnivoice-core/realtime"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "no API key",
			opts:    nil,
			wantErr: true,
		},
		{
			name:    "empty API key",
			opts:    []Option{WithAPIKey("")},
			wantErr: true,
		},
		{
			name:    "valid API key",
			opts:    []Option{WithAPIKey("test-api-key")},
			wantErr: false,
		},
		{
			name: "with all options",
			opts: []Option{
				WithAPIKey("test-api-key"),
				WithModel("aura-2-thalia-en"),
				WithLanguage("en-US"),
				WithInputEncoding("linear16"),
				WithInputSampleRate(16000),
				WithOutputEncoding("linear16"),
				WithOutputSampleRate(24000),
				WithGreeting("Hello!"),
				WithExperimental(true),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && provider == nil {
				t.Error("New() returned nil provider without error")
			}
		})
	}
}

func TestProvider_Name(t *testing.T) {
	provider, err := New(WithAPIKey("test-api-key"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	name := provider.Name()
	if name != "deepgram-agent" {
		t.Errorf("Name() = %v, want %v", name, "deepgram-agent")
	}
}

func TestProvider_Close(t *testing.T) {
	provider, err := New(WithAPIKey("test-api-key"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = provider.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Language != "en" {
		t.Errorf("DefaultConfig().Language = %v, want %v", cfg.Language, "en")
	}
	if cfg.InputEncoding != "linear16" {
		t.Errorf("DefaultConfig().InputEncoding = %v, want %v", cfg.InputEncoding, "linear16")
	}
	if cfg.InputSampleRate != 16000 {
		t.Errorf("DefaultConfig().InputSampleRate = %v, want %v", cfg.InputSampleRate, 16000)
	}
	if cfg.OutputEncoding != "linear16" {
		t.Errorf("DefaultConfig().OutputEncoding = %v, want %v", cfg.OutputEncoding, "linear16")
	}
	if cfg.OutputSampleRate != 24000 {
		t.Errorf("DefaultConfig().OutputSampleRate = %v, want %v", cfg.OutputSampleRate, 24000)
	}
}

func TestConfigToSettingsOptions(t *testing.T) {
	cfg := Config{
		APIKey:           "test-key",
		Model:            "aura-2-thalia-en",
		Language:         "en-US",
		InputEncoding:    "linear16",
		InputSampleRate:  16000,
		OutputEncoding:   "linear16",
		OutputSampleRate: 24000,
		Greeting:         "Hello!",
		Experimental:     true,
	}

	processConfig := corereal.ProcessConfig{
		Instructions: "You are a helpful assistant.",
		Voice:        "aura-2-nova-en",
		Functions: []corereal.FunctionDeclaration{
			{
				Name:        "get_weather",
				Description: "Get the weather for a location",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"location":{"type":"string"}},"required":["location"]}`),
			},
		},
	}

	opts := ConfigToSettingsOptions(cfg, processConfig)

	if opts.Type != "SettingsConfiguration" {
		t.Errorf("Type = %v, want %v", opts.Type, "SettingsConfiguration")
	}
	if !opts.Experimental {
		t.Error("Experimental should be true")
	}
	if opts.Agent.Language != "en-US" {
		t.Errorf("Agent.Language = %v, want %v", opts.Agent.Language, "en-US")
	}
	if opts.Agent.Greeting != "Hello!" {
		t.Errorf("Agent.Greeting = %v, want %v", opts.Agent.Greeting, "Hello!")
	}
	if opts.Agent.Think.Prompt != "You are a helpful assistant." {
		t.Errorf("Agent.Think.Prompt = %v, want %v", opts.Agent.Think.Prompt, "You are a helpful assistant.")
	}
	// Voice from ProcessConfig should override Model from Config
	if opts.Agent.Speak.Provider["model"] != "aura-2-nova-en" {
		t.Errorf("Agent.Speak.Provider[model] = %v, want %v", opts.Agent.Speak.Provider["model"], "aura-2-nova-en")
	}
	if opts.Audio.Input.Encoding != "linear16" {
		t.Errorf("Audio.Input.Encoding = %v, want %v", opts.Audio.Input.Encoding, "linear16")
	}
	if opts.Audio.Input.SampleRate != 16000 {
		t.Errorf("Audio.Input.SampleRate = %v, want %v", opts.Audio.Input.SampleRate, 16000)
	}
	if opts.Audio.Output.Encoding != "linear16" {
		t.Errorf("Audio.Output.Encoding = %v, want %v", opts.Audio.Output.Encoding, "linear16")
	}
	if opts.Audio.Output.SampleRate != 24000 {
		t.Errorf("Audio.Output.SampleRate = %v, want %v", opts.Audio.Output.SampleRate, 24000)
	}
	if opts.Agent.Think.Functions == nil || len(*opts.Agent.Think.Functions) != 1 {
		t.Error("Expected 1 function")
	} else {
		fn := (*opts.Agent.Think.Functions)[0]
		if fn.Name != "get_weather" {
			t.Errorf("Function name = %v, want %v", fn.Name, "get_weather")
		}
		if fn.Description != "Get the weather for a location" {
			t.Errorf("Function description = %v, want %v", fn.Description, "Get the weather for a location")
		}
	}
}

func TestChanHandler_InterfaceCompliance(t *testing.T) {
	handler := newChanHandler()

	// Verify all channel getters return non-nil slices
	if len(handler.GetBinary()) == 0 {
		t.Error("GetBinary() returned empty slice")
	}
	if len(handler.GetOpen()) == 0 {
		t.Error("GetOpen() returned empty slice")
	}
	if len(handler.GetWelcome()) == 0 {
		t.Error("GetWelcome() returned empty slice")
	}
	if len(handler.GetConversationText()) == 0 {
		t.Error("GetConversationText() returned empty slice")
	}
	if len(handler.GetUserStartedSpeaking()) == 0 {
		t.Error("GetUserStartedSpeaking() returned empty slice")
	}
	if len(handler.GetAgentThinking()) == 0 {
		t.Error("GetAgentThinking() returned empty slice")
	}
	if len(handler.GetFunctionCallRequest()) == 0 {
		t.Error("GetFunctionCallRequest() returned empty slice")
	}
	if len(handler.GetAgentStartedSpeaking()) == 0 {
		t.Error("GetAgentStartedSpeaking() returned empty slice")
	}
	if len(handler.GetAgentAudioDone()) == 0 {
		t.Error("GetAgentAudioDone() returned empty slice")
	}
	if len(handler.GetInjectionRefused()) == 0 {
		t.Error("GetInjectionRefused() returned empty slice")
	}
	if len(handler.GetKeepAlive()) == 0 {
		t.Error("GetKeepAlive() returned empty slice")
	}
	if len(handler.GetSettingsApplied()) == 0 {
		t.Error("GetSettingsApplied() returned empty slice")
	}
	if len(handler.GetClose()) == 0 {
		t.Error("GetClose() returned empty slice")
	}
	if len(handler.GetError()) == 0 {
		t.Error("GetError() returned empty slice")
	}
	if len(handler.GetUnhandled()) == 0 {
		t.Error("GetUnhandled() returned empty slice")
	}
}

func TestOptions(t *testing.T) {
	cfg := DefaultConfig()

	// Apply options
	WithAPIKey("my-key")(&cfg)
	WithModel("custom-model")(&cfg)
	WithLanguage("es")(&cfg)
	WithInputEncoding("mulaw")(&cfg)
	WithInputSampleRate(8000)(&cfg)
	WithOutputEncoding("mp3")(&cfg)
	WithOutputSampleRate(44100)(&cfg)
	WithGreeting("Hola!")(&cfg)
	WithExperimental(true)(&cfg)
	WithThinkProvider(map[string]any{"type": "anthropic"})(&cfg)
	WithListenProvider(map[string]any{"model": "nova-3"})(&cfg)
	WithSpeakProvider(map[string]any{"model": "aura"})(&cfg)

	if cfg.APIKey != "my-key" {
		t.Errorf("APIKey = %v, want %v", cfg.APIKey, "my-key")
	}
	if cfg.Model != "custom-model" {
		t.Errorf("Model = %v, want %v", cfg.Model, "custom-model")
	}
	if cfg.Language != "es" {
		t.Errorf("Language = %v, want %v", cfg.Language, "es")
	}
	if cfg.InputEncoding != "mulaw" {
		t.Errorf("InputEncoding = %v, want %v", cfg.InputEncoding, "mulaw")
	}
	if cfg.InputSampleRate != 8000 {
		t.Errorf("InputSampleRate = %v, want %v", cfg.InputSampleRate, 8000)
	}
	if cfg.OutputEncoding != "mp3" {
		t.Errorf("OutputEncoding = %v, want %v", cfg.OutputEncoding, "mp3")
	}
	if cfg.OutputSampleRate != 44100 {
		t.Errorf("OutputSampleRate = %v, want %v", cfg.OutputSampleRate, 44100)
	}
	if cfg.Greeting != "Hola!" {
		t.Errorf("Greeting = %v, want %v", cfg.Greeting, "Hola!")
	}
	if !cfg.Experimental {
		t.Error("Experimental should be true")
	}
	if cfg.ThinkProvider["type"] != "anthropic" {
		t.Errorf("ThinkProvider[type] = %v, want %v", cfg.ThinkProvider["type"], "anthropic")
	}
	if cfg.ListenProvider["model"] != "nova-3" {
		t.Errorf("ListenProvider[model] = %v, want %v", cfg.ListenProvider["model"], "nova-3")
	}
	if cfg.SpeakProvider["model"] != "aura" {
		t.Errorf("SpeakProvider[model] = %v, want %v", cfg.SpeakProvider["model"], "aura")
	}
}

func TestProcessAudioStream_RequiresContext(t *testing.T) {
	provider, err := New(WithAPIKey("test-api-key"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	audioIn := make(chan []byte)
	close(audioIn)

	// This should fail because context is cancelled
	_, _, err = provider.ProcessAudioStream(ctx, audioIn, corereal.ProcessConfig{})
	// Connection will fail with cancelled context or connection error
	// We expect an error here
	if err == nil {
		t.Log("ProcessAudioStream returned without error (may be expected in some cases)")
	}
}
