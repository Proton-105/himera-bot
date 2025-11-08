package errors

import "fmt"

type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

type AppError struct {
	Code        string
	Message     string
	UserMessage string
	Severity    Severity
	Retryable   bool
	cause       error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}

	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.cause
}

func (e *AppError) Cause() error {
	return e.Unwrap()
}

func NewValidationError(msg string) *AppError {
	return &AppError{
		Code:        "E100",
		Message:     msg,
		UserMessage: fmt.Sprintf("Неверный формат данных. %s", msg),
		Severity:    SeverityLow,
		Retryable:   false,
		cause:       nil,
	}
}

func NewDatabaseError(cause error) *AppError {
	var underlyingMsg string
	if cause != nil {
		underlyingMsg = cause.Error()
	}

	return &AppError{
		Code:        "E200",
		Message:     fmt.Sprintf("Database error: %s", underlyingMsg),
		UserMessage: "Временная проблема, попробуйте позже",
		Severity:    SeverityHigh,
		Retryable:   true,
		cause:       cause,
	}
}

func NewExternalAPIError(apiName string, cause error) *AppError {
	return &AppError{
		Code:        "E300",
		Message:     fmt.Sprintf("External API error: %s", apiName),
		UserMessage: "Сервис временно недоступен",
		Severity:    SeverityMedium,
		Retryable:   true,
		cause:       cause,
	}
}

func NewStateError(msg string) *AppError {
	return &AppError{
		Code:        "E400",
		Message:     msg,
		UserMessage: "Операция невозможна в текущем состоянии",
		Severity:    SeverityMedium,
		Retryable:   false,
		cause:       nil,
	}
}

func NewRateLimitError(retryAfter int) *AppError {
	return &AppError{
		Code:        "E500",
		Message:     fmt.Sprintf("Rate limit exceeded: retry after %d seconds", retryAfter),
		UserMessage: fmt.Sprintf("Слишком много запросов. Попробуйте через %d секунд", retryAfter),
		Severity:    SeverityLow,
		Retryable:   false,
		cause:       nil,
	}
}
