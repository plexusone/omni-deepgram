package omnivoice

import (
	"errors"
	"net"
	"strings"
	"syscall"

	"github.com/plexusone/omnivoice-core/resilience"
)

// DeepgramClassifier provides Deepgram-specific error classification.
// It analyzes errors from the Deepgram SDK and maps them to resilience categories.
type DeepgramClassifier struct{}

// Classify categorizes a Deepgram error and returns actionable metadata.
func (c *DeepgramClassifier) Classify(err error) resilience.ErrorInfo {
	if err == nil {
		return resilience.ErrorInfo{Category: resilience.CategoryUnknown}
	}

	// Check for wrapped ProviderError first
	if pe, ok := resilience.IsProviderError(err); ok {
		return pe.Info
	}

	// Check for network errors (transient)
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return resilience.ErrorInfo{
				Category:   resilience.CategoryTransient,
				Retryable:  true,
				Code:       "TIMEOUT",
				Message:    "Request timed out",
				Suggestion: "Retry with exponential backoff",
			}
		}
		return resilience.ErrorInfo{
			Category:   resilience.CategoryTransient,
			Retryable:  true,
			Code:       "NETWORK_ERROR",
			Message:    "Network error occurred",
			Suggestion: "Check network connectivity and retry",
		}
	}

	// Check for connection refused
	if errors.Is(err, syscall.ECONNREFUSED) {
		return resilience.ErrorInfo{
			Category:   resilience.CategoryServer,
			Retryable:  true,
			Code:       "CONNECTION_REFUSED",
			Message:    "Connection refused",
			Suggestion: "Deepgram server may be unavailable, retry later",
		}
	}

	// Analyze error message for Deepgram-specific patterns
	msg := strings.ToLower(err.Error())

	// Rate limiting
	if strings.Contains(msg, "rate limit") || strings.Contains(msg, "too many requests") ||
		strings.Contains(msg, "429") {
		return resilience.ErrorInfo{
			Category:   resilience.CategoryRateLimit,
			Retryable:  true,
			Code:       "RATE_LIMITED",
			Message:    "Rate limit exceeded",
			Suggestion: "Wait and retry with exponential backoff",
		}
	}

	// Authentication errors
	if strings.Contains(msg, "unauthorized") || strings.Contains(msg, "invalid api key") ||
		strings.Contains(msg, "401") || strings.Contains(msg, "authentication") {
		return resilience.ErrorInfo{
			Category:   resilience.CategoryAuth,
			Retryable:  false,
			Code:       "UNAUTHORIZED",
			Message:    "Invalid or missing API key",
			Suggestion: "Check your Deepgram API key",
		}
	}

	// Permission errors
	if strings.Contains(msg, "forbidden") || strings.Contains(msg, "403") ||
		strings.Contains(msg, "permission") {
		return resilience.ErrorInfo{
			Category:   resilience.CategoryAuth,
			Retryable:  false,
			Code:       "FORBIDDEN",
			Message:    "Permission denied",
			Suggestion: "Check API key permissions for this feature",
		}
	}

	// Resource not found
	if strings.Contains(msg, "not found") || strings.Contains(msg, "404") {
		return resilience.ErrorInfo{
			Category:   resilience.CategoryNotFound,
			Retryable:  false,
			Code:       "NOT_FOUND",
			Message:    "Resource not found",
			Suggestion: "Verify the requested resource exists",
		}
	}

	// Validation errors
	if strings.Contains(msg, "invalid") || strings.Contains(msg, "validation") ||
		strings.Contains(msg, "unsupported") || strings.Contains(msg, "bad request") ||
		strings.Contains(msg, "400") {
		return resilience.ErrorInfo{
			Category:   resilience.CategoryValidation,
			Retryable:  false,
			Code:       "VALIDATION_ERROR",
			Message:    extractValidationMessage(msg),
			Suggestion: "Check request parameters (voice, model, encoding, etc.)",
		}
	}

	// Quota errors
	if strings.Contains(msg, "quota") || strings.Contains(msg, "limit exceeded") ||
		strings.Contains(msg, "insufficient") {
		return resilience.ErrorInfo{
			Category:   resilience.CategoryQuota,
			Retryable:  false,
			Code:       "QUOTA_EXCEEDED",
			Message:    "Quota or credits exceeded",
			Suggestion: "Check your Deepgram account balance or upgrade plan",
		}
	}

	// Server errors
	if strings.Contains(msg, "internal") || strings.Contains(msg, "server error") ||
		strings.Contains(msg, "500") || strings.Contains(msg, "502") ||
		strings.Contains(msg, "503") || strings.Contains(msg, "504") {
		return resilience.ErrorInfo{
			Category:   resilience.CategoryServer,
			Retryable:  true,
			Code:       "SERVER_ERROR",
			Message:    "Deepgram server error",
			Suggestion: "Retry with exponential backoff",
		}
	}

	// Connection/WebSocket errors (transient)
	if strings.Contains(msg, "connection") || strings.Contains(msg, "websocket") ||
		strings.Contains(msg, "failed to connect") || strings.Contains(msg, "eof") {
		return resilience.ErrorInfo{
			Category:   resilience.CategoryTransient,
			Retryable:  true,
			Code:       "CONNECTION_ERROR",
			Message:    "Connection error",
			Suggestion: "Retry the connection",
		}
	}

	// Unknown error
	return resilience.ErrorInfo{
		Category:   resilience.CategoryUnknown,
		Retryable:  false,
		Code:       "UNKNOWN",
		Message:    err.Error(),
		Suggestion: "Check the error details",
	}
}

// extractValidationMessage tries to extract a meaningful validation message.
func extractValidationMessage(msg string) string {
	// Common validation patterns
	if strings.Contains(msg, "voice") {
		return "Invalid voice ID"
	}
	if strings.Contains(msg, "model") {
		return "Invalid model specified"
	}
	if strings.Contains(msg, "encoding") || strings.Contains(msg, "format") {
		return "Invalid audio encoding or format"
	}
	if strings.Contains(msg, "sample rate") || strings.Contains(msg, "sample_rate") {
		return "Invalid sample rate"
	}
	if strings.Contains(msg, "language") {
		return "Invalid or unsupported language"
	}
	return "Invalid request parameters"
}

// ClassifyError is a convenience function that creates a ProviderError with classification.
func ClassifyError(op string, err error) *resilience.ProviderError {
	classifier := &DeepgramClassifier{}
	info := classifier.Classify(err)

	return resilience.NewProviderError(ProviderName, op, err, info)
}

// Verify interface compliance.
var _ resilience.ErrorClassifier = (*DeepgramClassifier)(nil)
