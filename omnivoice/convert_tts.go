package omnivoice

import (
	interfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/interfaces"
	"github.com/plexusone/omnivoice-core/audio/format"
	"github.com/plexusone/omnivoice-core/tts"
)

// ConfigToSpeakOptions converts OmniVoice SynthesisConfig to Deepgram SpeakOptions.
func ConfigToSpeakOptions(config tts.SynthesisConfig) *interfaces.SpeakOptions {
	encoding := normalizeEncoding(config.OutputFormat)
	opts := &interfaces.SpeakOptions{
		Model:    config.Model,
		Encoding: encoding,
	}

	// For raw audio formats: set container=none for telephony use.
	if format.IsRawEncoding(encoding) {
		opts.Container = "none"
	}

	// MP3 doesn't support sample_rate or container options.
	// Other formats (including raw and container) can use sample_rate.
	if format.Encoding(encoding).Normalize() != format.MP3 {
		opts.SampleRate = config.SampleRate
	}

	// If VoiceID provided but Model not set, use VoiceID as model
	// Deepgram uses the model name as the voice identifier
	if opts.Model == "" && config.VoiceID != "" {
		opts.Model = config.VoiceID
	}

	// Default model if none specified
	if opts.Model == "" {
		opts.Model = DefaultTTSModel
	}

	return opts
}

// ConfigToWSSpeakOptions converts OmniVoice SynthesisConfig to Deepgram WSSpeakOptions.
func ConfigToWSSpeakOptions(config tts.SynthesisConfig) *interfaces.WSSpeakOptions {
	opts := &interfaces.WSSpeakOptions{
		Model:      config.Model,
		Encoding:   normalizeEncoding(config.OutputFormat),
		SampleRate: config.SampleRate,
	}

	// If VoiceID provided but Model not set, use VoiceID as model
	if opts.Model == "" && config.VoiceID != "" {
		opts.Model = config.VoiceID
	}

	// Default model if none specified
	if opts.Model == "" {
		opts.Model = DefaultTTSModel
	}

	return opts
}

// DefaultTTSModel is the default TTS model to use.
const DefaultTTSModel = "aura-asteria-en"

// Voice represents a Deepgram TTS voice.
type Voice struct {
	ID       string
	Name     string
	Language string
	Gender   string
}

// DeepgramVoices contains the list of available Deepgram TTS voices.
// Deepgram doesn't have a voices API, so we maintain a static list.
var DeepgramVoices = []Voice{
	// Aura 1 voices - Female
	{ID: "aura-asteria-en", Name: "Asteria", Language: "en-US", Gender: "female"},
	{ID: "aura-luna-en", Name: "Luna", Language: "en-US", Gender: "female"},
	{ID: "aura-stella-en", Name: "Stella", Language: "en-US", Gender: "female"},
	{ID: "aura-athena-en", Name: "Athena", Language: "en-US", Gender: "female"},
	{ID: "aura-hera-en", Name: "Hera", Language: "en-US", Gender: "female"},

	// Aura 1 voices - Male
	{ID: "aura-orion-en", Name: "Orion", Language: "en-US", Gender: "male"},
	{ID: "aura-arcas-en", Name: "Arcas", Language: "en-US", Gender: "male"},
	{ID: "aura-perseus-en", Name: "Perseus", Language: "en-US", Gender: "male"},
	{ID: "aura-angus-en", Name: "Angus", Language: "en-IE", Gender: "male"},
	{ID: "aura-orpheus-en", Name: "Orpheus", Language: "en-US", Gender: "male"},
	{ID: "aura-helios-en", Name: "Helios", Language: "en-GB", Gender: "male"},
	{ID: "aura-zeus-en", Name: "Zeus", Language: "en-US", Gender: "male"},

	// Aura 2 voices
	{ID: "aura-2-thalia-en", Name: "Thalia (Aura 2)", Language: "en-US", Gender: "female"},
	{ID: "aura-2-andromeda-en", Name: "Andromeda (Aura 2)", Language: "en-US", Gender: "female"},
	{ID: "aura-2-helena-en", Name: "Helena (Aura 2)", Language: "en-US", Gender: "female"},
	{ID: "aura-2-apollo-en", Name: "Apollo (Aura 2)", Language: "en-US", Gender: "male"},
	{ID: "aura-2-aries-en", Name: "Aries (Aura 2)", Language: "en-US", Gender: "male"},
}

// VoiceToOmniVoice converts an internal Voice to an OmniVoice tts.Voice.
func VoiceToOmniVoice(v Voice) tts.Voice {
	return tts.Voice{
		ID:       v.ID,
		Name:     v.Name,
		Language: v.Language,
		Gender:   v.Gender,
		Provider: ProviderName,
	}
}
