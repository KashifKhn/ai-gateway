package models

import "fmt"

type APIError struct {
	ErrorInfo ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
	Status  int    `json:"status"`
}

const (
	ErrorTypeInvalidRequest = "invalid_request_error"
	ErrorTypeAuthentication = "authentication_error"
	ErrorTypeRateLimit      = "rate_limit_error"
	ErrorTypeBackend        = "backend_error"
	ErrorTypeService        = "service_error"
)

const (
	ErrorCodeInvalidModel       = "invalid_model"
	ErrorCodeInvalidMessages    = "invalid_messages"
	ErrorCodeInvalidAPIKey      = "invalid_api_key"
	ErrorCodeMissingAPIKey      = "missing_api_key"
	ErrorCodeRateLimitExceeded  = "rate_limit_exceeded"
	ErrorCodeBackendUnavailable = "backend_unavailable"
	ErrorCodeBackendTimeout     = "backend_timeout"
	ErrorCodeServiceUnavailable = "service_unavailable"
)

func NewAPIError(message, errorType, code string, status int) *APIError {
	return &APIError{
		ErrorInfo: ErrorDetail{
			Message: message,
			Type:    errorType,
			Code:    code,
			Status:  status,
		},
	}
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorInfo.Code, e.ErrorInfo.Message)
}

func (e *APIError) GetStatus() int {
	return e.ErrorInfo.Status
}

func ErrInvalidModel(model string) *APIError {
	return NewAPIError(
		fmt.Sprintf("Model '%s' not found", model),
		ErrorTypeInvalidRequest,
		ErrorCodeInvalidModel,
		400,
	)
}

func ErrInvalidMessages() *APIError {
	return NewAPIError(
		"Invalid messages format",
		ErrorTypeInvalidRequest,
		ErrorCodeInvalidMessages,
		400,
	)
}

func ErrInvalidAPIKey() *APIError {
	return NewAPIError(
		"Invalid API key provided",
		ErrorTypeAuthentication,
		ErrorCodeInvalidAPIKey,
		401,
	)
}

func ErrMissingAPIKey() *APIError {
	return NewAPIError(
		"Missing API key. Include 'Authorization: Bearer <key>' header",
		ErrorTypeAuthentication,
		ErrorCodeMissingAPIKey,
		401,
	)
}

func ErrRateLimitExceeded() *APIError {
	return NewAPIError(
		"Rate limit exceeded. Please slow down",
		ErrorTypeRateLimit,
		ErrorCodeRateLimitExceeded,
		429,
	)
}

func ErrBackendUnavailable(backend string) *APIError {
	return NewAPIError(
		fmt.Sprintf("Backend '%s' is unavailable", backend),
		ErrorTypeBackend,
		ErrorCodeBackendUnavailable,
		503,
	)
}

func ErrBackendTimeout(backend string) *APIError {
	return NewAPIError(
		fmt.Sprintf("Backend '%s' request timed out", backend),
		ErrorTypeBackend,
		ErrorCodeBackendTimeout,
		504,
	)
}
