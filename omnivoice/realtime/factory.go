package realtime

import (
	"github.com/plexusone/omnivoice-core/gateway"
	corereal "github.com/plexusone/omnivoice-core/realtime"
)

// Factory creates Deepgram Voice Agent providers from gateway configuration.
// It implements [gateway.RealtimeProviderFactory].
type Factory struct{}

// NewFactory creates a new Deepgram realtime provider factory.
func NewFactory() *Factory {
	return &Factory{}
}

// Ensure Factory implements gateway.RealtimeProviderFactory.
var _ gateway.RealtimeProviderFactory = (*Factory)(nil)

// Create creates a Deepgram RealtimeProvider from the given configuration.
func (f *Factory) Create(config *gateway.RealtimeConfig) (corereal.Provider, error) {
	if config.APIKey == "" {
		return nil, ErrAPIKeyRequired
	}

	opts := []Option{
		WithAPIKey(config.APIKey),
		WithExperimental(true),      // Enable latency metrics
		WithInputSampleRate(48000),  // LiveKit sends 48kHz audio
		WithOutputSampleRate(24000), // Deepgram outputs 24kHz
	}

	if config.Model != "" {
		opts = append(opts, WithModel(config.Model))
	}
	if config.Voice != "" {
		opts = append(opts, WithModel(config.Voice)) // Voice is the model in Deepgram
	}

	// Set a default greeting so the agent speaks first
	// This triggers Deepgram Voice Agent to initiate the conversation
	opts = append(opts, WithGreeting("Hello! How can I help you today?"))

	return New(opts...)
}

// Name returns the provider name.
func (f *Factory) Name() string {
	return "deepgram"
}

// ProviderName is the name used to identify Deepgram realtime provider.
const ProviderName = "deepgram"
