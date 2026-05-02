package tts

import (
	"context"
	"testing"

	"github.com/plexusone/omni-deepgram/omnivoice"
	"github.com/plexusone/omnivoice-core/tts"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "no API key returns error",
			opts:    nil,
			wantErr: true,
		},
		{
			name:    "empty API key returns error",
			opts:    []Option{WithAPIKey("")},
			wantErr: true,
		},
		{
			name:    "valid API key succeeds",
			opts:    []Option{WithAPIKey("test-api-key")},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && p == nil {
				t.Error("New() returned nil provider with no error")
			}
		})
	}
}

func TestProvider_Name(t *testing.T) {
	p, err := New(WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got := p.Name(); got != omnivoice.ProviderName {
		t.Errorf("Name() = %q, want %q", got, omnivoice.ProviderName)
	}
}

func TestProvider_ListVoices(t *testing.T) {
	p, err := New(WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx := context.Background()
	voices, err := p.ListVoices(ctx)
	if err != nil {
		t.Errorf("ListVoices() error = %v", err)
	}

	if len(voices) == 0 {
		t.Error("ListVoices() returned empty list")
	}

	// Verify all voices have required fields
	for _, v := range voices {
		if v.ID == "" {
			t.Errorf("Voice ID is empty for voice %s", v.Name)
		}
		if v.Name == "" {
			t.Errorf("Voice Name is empty for voice %s", v.ID)
		}
		if v.Provider != omnivoice.ProviderName {
			t.Errorf("Voice Provider = %q, want %q", v.Provider, omnivoice.ProviderName)
		}
	}
}

func TestProvider_GetVoice(t *testing.T) {
	p, err := New(WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		voiceID string
		wantErr error
	}{
		{
			name:    "existing voice",
			voiceID: "aura-asteria-en",
			wantErr: nil,
		},
		{
			name:    "non-existent voice",
			voiceID: "non-existent-voice",
			wantErr: tts.ErrVoiceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			voice, err := p.GetVoice(ctx, tt.voiceID)
			if err != tt.wantErr {
				t.Errorf("GetVoice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == nil {
				if voice == nil {
					t.Error("GetVoice() returned nil voice with no error")
					return
				}
				if voice.ID != tt.voiceID {
					t.Errorf("GetVoice() voice.ID = %q, want %q", voice.ID, tt.voiceID)
				}
				if voice.Provider != omnivoice.ProviderName {
					t.Errorf("GetVoice() voice.Provider = %q, want %q", voice.Provider, omnivoice.ProviderName)
				}
			}
		})
	}
}

func TestProvider_ImplementsInterface(t *testing.T) {
	p, err := New(WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Verify Provider implements tts.Provider
	var _ tts.Provider = p

	// Verify Provider implements tts.StreamingProvider
	var _ tts.StreamingProvider = p
}

func TestSplitIntoSentences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single sentence with period",
			input:    "Hello world.",
			expected: []string{"Hello world."},
		},
		{
			name:     "two sentences",
			input:    "Hello world. How are you?",
			expected: []string{"Hello world.", " How are you?"},
		},
		{
			name:     "sentence with exclamation",
			input:    "Hello! How are you?",
			expected: []string{"Hello!", " How are you?"},
		},
		{
			name:     "incomplete sentence",
			input:    "Hello world",
			expected: []string{"Hello world"},
		},
		{
			name:     "sentence with abbreviation",
			input:    "Dr. Smith is here.",
			expected: []string{"Dr.", " Smith is here."}, // Abbreviations may split; acceptable for TTS
		},
		{
			name:     "sentence with decimal number",
			input:    "The price is 3.14 dollars.",
			expected: []string{"The price is 3.14 dollars."},
		},
		{
			name:     "multiple complete sentences",
			input:    "First sentence. Second sentence. Third sentence.",
			expected: []string{"First sentence.", " Second sentence.", " Third sentence."},
		},
		{
			name:     "sentence with trailing incomplete",
			input:    "Complete sentence. Incomplete",
			expected: []string{"Complete sentence.", " Incomplete"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: []string{"   "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitIntoSentences(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("splitIntoSentences(%q) returned %d sentences, want %d\nGot: %v\nWant: %v",
					tt.input, len(got), len(tt.expected), got, tt.expected)
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("splitIntoSentences(%q)[%d] = %q, want %q",
						tt.input, i, got[i], tt.expected[i])
				}
			}
		})
	}
}
