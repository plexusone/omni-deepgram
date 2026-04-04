package omnivoice

import (
	"errors"
	"testing"

	"github.com/plexusone/omnivoice-core/resilience"
)

func TestDeepgramClassifier_Classify(t *testing.T) {
	classifier := &DeepgramClassifier{}

	tests := []struct {
		name         string
		err          error
		wantCategory resilience.ErrorCategory
		wantCode     string
		wantRetry    bool
	}{
		{
			name:         "nil error",
			err:          nil,
			wantCategory: resilience.CategoryUnknown,
			wantCode:     "",
			wantRetry:    false,
		},
		{
			name:         "rate limit error",
			err:          errors.New("rate limit exceeded"),
			wantCategory: resilience.CategoryRateLimit,
			wantCode:     "RATE_LIMITED",
			wantRetry:    true,
		},
		{
			name:         "too many requests",
			err:          errors.New("too many requests"),
			wantCategory: resilience.CategoryRateLimit,
			wantCode:     "RATE_LIMITED",
			wantRetry:    true,
		},
		{
			name:         "HTTP 429",
			err:          errors.New("status 429"),
			wantCategory: resilience.CategoryRateLimit,
			wantCode:     "RATE_LIMITED",
			wantRetry:    true,
		},
		{
			name:         "unauthorized",
			err:          errors.New("unauthorized access"),
			wantCategory: resilience.CategoryAuth,
			wantCode:     "UNAUTHORIZED",
			wantRetry:    false,
		},
		{
			name:         "invalid api key",
			err:          errors.New("invalid api key"),
			wantCategory: resilience.CategoryAuth,
			wantCode:     "UNAUTHORIZED",
			wantRetry:    false,
		},
		{
			name:         "HTTP 401",
			err:          errors.New("status 401"),
			wantCategory: resilience.CategoryAuth,
			wantCode:     "UNAUTHORIZED",
			wantRetry:    false,
		},
		{
			name:         "forbidden",
			err:          errors.New("forbidden"),
			wantCategory: resilience.CategoryAuth,
			wantCode:     "FORBIDDEN",
			wantRetry:    false,
		},
		{
			name:         "HTTP 403",
			err:          errors.New("status 403"),
			wantCategory: resilience.CategoryAuth,
			wantCode:     "FORBIDDEN",
			wantRetry:    false,
		},
		{
			name:         "not found",
			err:          errors.New("resource not found"),
			wantCategory: resilience.CategoryNotFound,
			wantCode:     "NOT_FOUND",
			wantRetry:    false,
		},
		{
			name:         "HTTP 404",
			err:          errors.New("status 404"),
			wantCategory: resilience.CategoryNotFound,
			wantCode:     "NOT_FOUND",
			wantRetry:    false,
		},
		{
			name:         "invalid request",
			err:          errors.New("invalid parameter"),
			wantCategory: resilience.CategoryValidation,
			wantCode:     "VALIDATION_ERROR",
			wantRetry:    false,
		},
		{
			name:         "unsupported format",
			err:          errors.New("unsupported format"),
			wantCategory: resilience.CategoryValidation,
			wantCode:     "VALIDATION_ERROR",
			wantRetry:    false,
		},
		{
			name:         "bad request",
			err:          errors.New("bad request"),
			wantCategory: resilience.CategoryValidation,
			wantCode:     "VALIDATION_ERROR",
			wantRetry:    false,
		},
		{
			name:         "HTTP 400",
			err:          errors.New("status 400"),
			wantCategory: resilience.CategoryValidation,
			wantCode:     "VALIDATION_ERROR",
			wantRetry:    false,
		},
		{
			name:         "quota exceeded",
			err:          errors.New("quota exceeded"),
			wantCategory: resilience.CategoryQuota,
			wantCode:     "QUOTA_EXCEEDED",
			wantRetry:    false,
		},
		{
			name:         "insufficient credits",
			err:          errors.New("insufficient credits"),
			wantCategory: resilience.CategoryQuota,
			wantCode:     "QUOTA_EXCEEDED",
			wantRetry:    false,
		},
		{
			name:         "server error",
			err:          errors.New("internal server error"),
			wantCategory: resilience.CategoryServer,
			wantCode:     "SERVER_ERROR",
			wantRetry:    true,
		},
		{
			name:         "HTTP 500",
			err:          errors.New("status 500"),
			wantCategory: resilience.CategoryServer,
			wantCode:     "SERVER_ERROR",
			wantRetry:    true,
		},
		{
			name:         "HTTP 502",
			err:          errors.New("status 502"),
			wantCategory: resilience.CategoryServer,
			wantCode:     "SERVER_ERROR",
			wantRetry:    true,
		},
		{
			name:         "HTTP 503",
			err:          errors.New("status 503"),
			wantCategory: resilience.CategoryServer,
			wantCode:     "SERVER_ERROR",
			wantRetry:    true,
		},
		{
			name:         "connection error",
			err:          errors.New("connection refused"),
			wantCategory: resilience.CategoryTransient,
			wantCode:     "CONNECTION_ERROR",
			wantRetry:    true,
		},
		{
			name:         "websocket error",
			err:          errors.New("websocket disconnected"),
			wantCategory: resilience.CategoryTransient,
			wantCode:     "CONNECTION_ERROR",
			wantRetry:    true,
		},
		{
			name:         "failed to connect",
			err:          errors.New("failed to connect to server"),
			wantCategory: resilience.CategoryTransient,
			wantCode:     "CONNECTION_ERROR",
			wantRetry:    true,
		},
		{
			name:         "unknown error",
			err:          errors.New("something unexpected happened"),
			wantCategory: resilience.CategoryUnknown,
			wantCode:     "UNKNOWN",
			wantRetry:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := classifier.Classify(tt.err)

			if info.Category != tt.wantCategory {
				t.Errorf("Category = %v, want %v", info.Category, tt.wantCategory)
			}

			if info.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", info.Code, tt.wantCode)
			}

			if info.Retryable != tt.wantRetry {
				t.Errorf("Retryable = %v, want %v", info.Retryable, tt.wantRetry)
			}
		})
	}
}

func TestClassifyError(t *testing.T) {
	err := errors.New("rate limit exceeded")
	pe := ClassifyError("Synthesize", err)

	if pe.Provider != ProviderName {
		t.Errorf("Provider = %v, want %v", pe.Provider, ProviderName)
	}

	if pe.Op != "Synthesize" {
		t.Errorf("Op = %v, want Synthesize", pe.Op)
	}

	if pe.Info.Category != resilience.CategoryRateLimit {
		t.Errorf("Category = %v, want %v", pe.Info.Category, resilience.CategoryRateLimit)
	}

	if !pe.IsRetryable() {
		t.Error("Expected error to be retryable")
	}
}

func TestExtractValidationMessage(t *testing.T) {
	tests := []struct {
		msg  string
		want string
	}{
		{"invalid voice id", "Invalid voice ID"},
		{"invalid model specified", "Invalid model specified"},
		{"unsupported encoding format", "Invalid audio encoding or format"},
		{"invalid sample rate", "Invalid sample rate"},
		{"unsupported language", "Invalid or unsupported language"},
		{"some other validation error", "Invalid request parameters"},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			got := extractValidationMessage(tt.msg)
			if got != tt.want {
				t.Errorf("extractValidationMessage(%q) = %q, want %q", tt.msg, got, tt.want)
			}
		})
	}
}
