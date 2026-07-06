package realtime

import (
	"encoding/json"

	agent "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/agent"
	interfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/interfaces"
	corereal "github.com/plexusone/omnivoice-core/realtime"
)

// ConfigToSettingsOptions converts the provider Config and ProcessConfig to Deepgram SettingsOptions.
func ConfigToSettingsOptions(cfg Config, processConfig corereal.ProcessConfig) *interfaces.SettingsOptions {
	// Use SDK's helper to get pre-initialized options with defaults
	opts := agent.NewSettingsConfigurationOptions()

	// Override experimental setting
	opts.Experimental = cfg.Experimental

	// Audio configuration
	opts.Audio.Input.Encoding = cfg.InputEncoding
	opts.Audio.Input.SampleRate = cfg.InputSampleRate
	opts.Audio.Output.Encoding = cfg.OutputEncoding
	opts.Audio.Output.SampleRate = cfg.OutputSampleRate

	// Agent configuration
	opts.Agent.Language = cfg.Language

	// Set greeting if provided
	if cfg.Greeting != "" {
		opts.Agent.Greeting = cfg.Greeting
	}

	// Think configuration (LLM) - use map index pattern like the SDK example
	if cfg.ThinkProvider != nil {
		for k, v := range cfg.ThinkProvider {
			opts.Agent.Think.Provider[k] = v
		}
	} else {
		opts.Agent.Think.Provider["type"] = "open_ai"
		opts.Agent.Think.Provider["model"] = "gpt-4o-mini"
	}
	opts.Agent.Think.Prompt = processConfig.Instructions

	// Convert functions if provided
	if len(processConfig.Functions) > 0 {
		funcs := make([]any, len(processConfig.Functions))
		for i, f := range processConfig.Functions {
			funcs[i] = functionDeclarationToSDKMap(f)
		}
		opts.Agent.Think.Provider["functions"] = funcs
	}

	// Listen configuration (STT) - use map index pattern
	if cfg.ListenProvider != nil {
		for k, v := range cfg.ListenProvider {
			opts.Agent.Listen.Provider[k] = v
		}
	} else {
		opts.Agent.Listen.Provider["type"] = "deepgram"
		opts.Agent.Listen.Provider["model"] = "nova-3"
	}

	// Speak configuration (TTS) - use map index pattern
	if cfg.SpeakProvider != nil {
		for k, v := range cfg.SpeakProvider {
			opts.Agent.Speak.Provider[k] = v
		}
	} else {
		opts.Agent.Speak.Provider["type"] = "deepgram"
	}

	// Set model from ProcessConfig, cfg.Model, or default
	if processConfig.Voice != "" {
		opts.Agent.Speak.Provider["model"] = processConfig.Voice
	} else if cfg.Model != "" {
		opts.Agent.Speak.Provider["model"] = cfg.Model
	} else if opts.Agent.Speak.Provider["model"] == nil {
		opts.Agent.Speak.Provider["model"] = "aura-2-thalia-en"
	}

	return opts
}

// functionDeclarationToSDKMap converts an omnivoice FunctionDeclaration to a map for Deepgram.
func functionDeclarationToSDKMap(f corereal.FunctionDeclaration) map[string]any {
	fn := map[string]any{
		"name":        f.Name,
		"description": f.Description,
	}

	// Parse JSON Schema parameters if provided
	if len(f.Parameters) > 0 {
		var schema map[string]any
		if err := json.Unmarshal(f.Parameters, &schema); err == nil {
			fn["parameters"] = schema
		}
	}

	return fn
}
