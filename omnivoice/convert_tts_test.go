package omnivoice

import (
	"testing"

	"github.com/plexusone/omnivoice-core/tts"
)

func TestConfigToSpeakOptions(t *testing.T) {
	tests := []struct {
		name           string
		config         tts.SynthesisConfig
		wantModel      string
		wantEncoding   string
		wantSampleRate int
	}{
		{
			name:           "empty config uses defaults",
			config:         tts.SynthesisConfig{},
			wantModel:      DefaultTTSModel,
			wantEncoding:   "linear16",
			wantSampleRate: 0,
		},
		{
			name: "model specified",
			config: tts.SynthesisConfig{
				Model: "aura-luna-en",
			},
			wantModel:      "aura-luna-en",
			wantEncoding:   "linear16",
			wantSampleRate: 0,
		},
		{
			name: "voiceID used when model empty",
			config: tts.SynthesisConfig{
				VoiceID: "aura-orion-en",
			},
			wantModel:      "aura-orion-en",
			wantEncoding:   "linear16",
			wantSampleRate: 0,
		},
		{
			name: "model takes precedence over voiceID",
			config: tts.SynthesisConfig{
				Model:   "aura-luna-en",
				VoiceID: "aura-orion-en",
			},
			wantModel:      "aura-luna-en",
			wantEncoding:   "linear16",
			wantSampleRate: 0,
		},
		{
			name: "mp3 encoding",
			config: tts.SynthesisConfig{
				OutputFormat: "mp3",
			},
			wantModel:      DefaultTTSModel,
			wantEncoding:   "mp3",
			wantSampleRate: 0,
		},
		{
			name: "wav maps to linear16",
			config: tts.SynthesisConfig{
				OutputFormat: "wav",
			},
			wantModel:      DefaultTTSModel,
			wantEncoding:   "linear16",
			wantSampleRate: 0,
		},
		{
			name: "sample rate specified",
			config: tts.SynthesisConfig{
				SampleRate: 22050,
			},
			wantModel:      DefaultTTSModel,
			wantEncoding:   "linear16",
			wantSampleRate: 22050,
		},
		{
			name: "full config",
			config: tts.SynthesisConfig{
				Model:        "aura-2-thalia-en",
				OutputFormat: "opus",
				SampleRate:   48000,
			},
			wantModel:      "aura-2-thalia-en",
			wantEncoding:   "opus",
			wantSampleRate: 48000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := ConfigToSpeakOptions(tt.config)

			if opts.Model != tt.wantModel {
				t.Errorf("Model = %q, want %q", opts.Model, tt.wantModel)
			}
			if opts.Encoding != tt.wantEncoding {
				t.Errorf("Encoding = %q, want %q", opts.Encoding, tt.wantEncoding)
			}
			if opts.SampleRate != tt.wantSampleRate {
				t.Errorf("SampleRate = %d, want %d", opts.SampleRate, tt.wantSampleRate)
			}
		})
	}
}

func TestConfigToWSSpeakOptions(t *testing.T) {
	tests := []struct {
		name           string
		config         tts.SynthesisConfig
		wantModel      string
		wantEncoding   string
		wantSampleRate int
	}{
		{
			name:           "empty config uses defaults",
			config:         tts.SynthesisConfig{},
			wantModel:      DefaultTTSModel,
			wantEncoding:   "linear16",
			wantSampleRate: 0,
		},
		{
			name: "full config",
			config: tts.SynthesisConfig{
				Model:        "aura-luna-en",
				OutputFormat: "mp3",
				SampleRate:   24000,
			},
			wantModel:      "aura-luna-en",
			wantEncoding:   "mp3",
			wantSampleRate: 24000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := ConfigToWSSpeakOptions(tt.config)

			if opts.Model != tt.wantModel {
				t.Errorf("Model = %q, want %q", opts.Model, tt.wantModel)
			}
			if opts.Encoding != tt.wantEncoding {
				t.Errorf("Encoding = %q, want %q", opts.Encoding, tt.wantEncoding)
			}
			if opts.SampleRate != tt.wantSampleRate {
				t.Errorf("SampleRate = %d, want %d", opts.SampleRate, tt.wantSampleRate)
			}
		})
	}
}

func TestNormalizeEncoding(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"mp3", "mp3"},
		{"linear16", "linear16"},
		{"pcm", "linear16"},
		{"pcm_s16le", "linear16"},
		{"wav", "linear16"},
		{"mulaw", "mulaw"},
		{"ulaw", "mulaw"},
		{"g711u", "mulaw"},
		{"pcm_mulaw", "mulaw"},
		{"alaw", "alaw"},
		{"g711a", "alaw"},
		{"pcm_alaw", "alaw"},
		{"opus", "opus"},
		{"flac", "flac"},
		{"aac", "aac"},
		{"speex", "speex"},
		{"webm", "webm"},
		{"", "linear16"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeEncoding(tt.input)
			if got != tt.want {
				t.Errorf("normalizeEncoding(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestVoiceToOmniVoice(t *testing.T) {
	v := Voice{
		ID:       "aura-asteria-en",
		Name:     "Asteria",
		Language: "en-US",
		Gender:   "female",
	}

	got := VoiceToOmniVoice(v)

	if got.ID != v.ID {
		t.Errorf("ID = %q, want %q", got.ID, v.ID)
	}
	if got.Name != v.Name {
		t.Errorf("Name = %q, want %q", got.Name, v.Name)
	}
	if got.Language != v.Language {
		t.Errorf("Language = %q, want %q", got.Language, v.Language)
	}
	if got.Gender != v.Gender {
		t.Errorf("Gender = %q, want %q", got.Gender, v.Gender)
	}
	if got.Provider != ProviderName {
		t.Errorf("Provider = %q, want %q", got.Provider, ProviderName)
	}
}

func TestDeepgramVoices(t *testing.T) {
	if len(DeepgramVoices) == 0 {
		t.Error("DeepgramVoices should not be empty")
	}

	// Verify all voices have required fields
	for _, v := range DeepgramVoices {
		if v.ID == "" {
			t.Errorf("Voice ID is empty for voice %s", v.Name)
		}
		if v.Name == "" {
			t.Errorf("Voice Name is empty for voice %s", v.ID)
		}
		if v.Language == "" {
			t.Errorf("Voice Language is empty for voice %s", v.ID)
		}
		if v.Gender == "" {
			t.Errorf("Voice Gender is empty for voice %s", v.ID)
		}
	}

	// Verify default voice exists
	found := false
	for _, v := range DeepgramVoices {
		if v.ID == DefaultTTSModel {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Default TTS model %q not found in DeepgramVoices", DefaultTTSModel)
	}
}
