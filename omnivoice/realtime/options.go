package realtime

// Config holds configuration for the Deepgram Voice Agent provider.
type Config struct {
	// APIKey is the Deepgram API key.
	APIKey string

	// Model is the voice model to use (e.g., "aura-2-thalia-en").
	// If empty, defaults to the Deepgram default.
	Model string

	// Language is the language code (e.g., "en-US", "es").
	// Default: "en"
	Language string

	// InputEncoding is the audio encoding for input audio.
	// Default: "linear16"
	InputEncoding string

	// InputSampleRate is the sample rate for input audio.
	// Default: 16000
	InputSampleRate int

	// OutputEncoding is the audio encoding for output audio.
	// Default: "linear16"
	OutputEncoding string

	// OutputSampleRate is the sample rate for output audio.
	// Default: 24000
	OutputSampleRate int

	// Greeting is an optional greeting message the agent will speak first.
	Greeting string

	// Experimental enables experimental features like latency metrics.
	Experimental bool

	// ThinkProvider configures the LLM provider for the agent's "think" step.
	// Keys depend on the provider (e.g., "type": "open_ai", "model": "gpt-4o").
	ThinkProvider map[string]any

	// ListenProvider configures the STT provider for the agent's "listen" step.
	// If nil, uses Deepgram's default STT.
	ListenProvider map[string]any

	// SpeakProvider configures the TTS provider for the agent's "speak" step.
	// If nil, uses Deepgram's default TTS.
	SpeakProvider map[string]any
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Language:         "en",
		InputEncoding:    "linear16",
		InputSampleRate:  16000,
		OutputEncoding:   "linear16",
		OutputSampleRate: 24000,
	}
}

// Option configures the Provider.
type Option func(*Config)

// WithAPIKey sets the Deepgram API key.
func WithAPIKey(apiKey string) Option {
	return func(c *Config) {
		c.APIKey = apiKey
	}
}

// WithModel sets the voice model.
func WithModel(model string) Option {
	return func(c *Config) {
		c.Model = model
	}
}

// WithLanguage sets the language code.
func WithLanguage(language string) Option {
	return func(c *Config) {
		c.Language = language
	}
}

// WithInputEncoding sets the input audio encoding.
func WithInputEncoding(encoding string) Option {
	return func(c *Config) {
		c.InputEncoding = encoding
	}
}

// WithInputSampleRate sets the input audio sample rate.
func WithInputSampleRate(sampleRate int) Option {
	return func(c *Config) {
		c.InputSampleRate = sampleRate
	}
}

// WithOutputEncoding sets the output audio encoding.
func WithOutputEncoding(encoding string) Option {
	return func(c *Config) {
		c.OutputEncoding = encoding
	}
}

// WithOutputSampleRate sets the output audio sample rate.
func WithOutputSampleRate(sampleRate int) Option {
	return func(c *Config) {
		c.OutputSampleRate = sampleRate
	}
}

// WithGreeting sets an optional greeting message.
func WithGreeting(greeting string) Option {
	return func(c *Config) {
		c.Greeting = greeting
	}
}

// WithExperimental enables experimental features.
func WithExperimental(enabled bool) Option {
	return func(c *Config) {
		c.Experimental = enabled
	}
}

// WithThinkProvider sets the LLM provider configuration.
func WithThinkProvider(provider map[string]any) Option {
	return func(c *Config) {
		c.ThinkProvider = provider
	}
}

// WithListenProvider sets the STT provider configuration.
func WithListenProvider(provider map[string]any) Option {
	return func(c *Config) {
		c.ListenProvider = provider
	}
}

// WithSpeakProvider sets the TTS provider configuration.
func WithSpeakProvider(provider map[string]any) Option {
	return func(c *Config) {
		c.SpeakProvider = provider
	}
}
