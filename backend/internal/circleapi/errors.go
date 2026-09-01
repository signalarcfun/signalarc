package circleapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var (
	ErrConfigInvalid       = errors.New("circle api config is invalid")
	ErrTransactionFailed   = errors.New("circle transaction failed")
	ErrTransactionTimedOut = errors.New("circle transaction polling timed out")
)

type Error struct {
	StatusCode      int
	CircleCode      string
	CircleMessage   string
	ErrorReason     string
	ErrorDetails    string
	TransactionID   string
	LastStatus      string
	TransactionHash string
	Message         string
	Err             error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("circle api request failed with status %d", e.StatusCode)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *Error) ErrorClass() string {
	if e == nil {
		return "unknown"
	}
	allText := strings.ToLower(strings.Join([]string{e.CircleCode, e.CircleMessage, e.ErrorReason, e.ErrorDetails, e.Message}, " "))
	switch {
	case errors.Is(e.Err, ErrTransactionTimedOut):
		return "transaction_timeout"
	case errors.Is(e.Err, context.DeadlineExceeded):
		return "circle_api_timeout"
	case strings.Contains(allText, "estimat") && strings.Contains(allText, "revert"):
		return "estimation_reverted"
	case strings.Contains(allText, "execution reverted") || strings.Contains(allText, "evm revert") || strings.Contains(allText, "reverted"):
		return "evm_revert"
	case e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden:
		return "auth_failed"
	case e.StatusCode >= 400 && e.StatusCode < 500:
		return "circle_request_invalid"
	case e.StatusCode >= 500:
		return "circle_api_error"
	case e.Err != nil:
		return "circle_api_error"
	default:
		return "unknown"
	}
}

func (e *Error) SanitizedSummary() string {
	if e == nil {
		return ""
	}
	parts := []string{}
	if e.StatusCode > 0 {
		parts = append(parts, fmt.Sprintf("status=%d", e.StatusCode))
	}
	if strings.TrimSpace(e.CircleCode) != "" {
		parts = append(parts, "code="+e.CircleCode)
	}
	if strings.TrimSpace(e.CircleMessage) != "" {
		parts = append(parts, "message="+e.CircleMessage)
	}
	if strings.TrimSpace(e.ErrorReason) != "" {
		parts = append(parts, "reason="+e.ErrorReason)
	}
	if strings.TrimSpace(e.ErrorDetails) != "" {
		parts = append(parts, "details="+e.ErrorDetails)
	}
	if strings.TrimSpace(e.TransactionID) != "" {
		parts = append(parts, "transaction_id="+e.TransactionID)
	}
	if strings.TrimSpace(e.LastStatus) != "" {
		parts = append(parts, "last_status="+e.LastStatus)
	}
	if strings.TrimSpace(e.TransactionHash) != "" {
		parts = append(parts, "transaction_hash="+e.TransactionHash)
	}
	if len(parts) == 0 && strings.TrimSpace(e.Message) != "" {
		parts = append(parts, e.Message)
	}
	if len(parts) == 0 && e.Err != nil {
		parts = append(parts, e.Err.Error())
	}
	return sanitizeSummary(strings.Join(parts, " "))
}

type errorResponseBody struct {
	Code         string          `json:"code"`
	Message      string          `json:"message"`
	ErrorReason  string          `json:"errorReason"`
	ErrorDetails json.RawMessage `json:"errorDetails"`
	Error        json.RawMessage `json:"error"`
}

func newHTTPError(statusCode int, body []byte) error {
	decoded := parseErrorResponseBody(body)
	return &Error{
		StatusCode:    statusCode,
		CircleCode:    decoded.Code,
		CircleMessage: decoded.Message,
		ErrorReason:   decoded.ErrorReason,
		ErrorDetails:  decoded.errorDetailsString(),
		Message:       http.StatusText(statusCode),
		Err:           errors.New("circle api request failed"),
	}
}

func newTransportError(err error) error {
	if err == nil {
		return nil
	}
	return &Error{
		Message: "circle api request failed",
		Err:     err,
	}
}

func newPollingTimeoutError(last Transaction) error {
	transactionID := strings.TrimSpace(last.ID)
	lastStatus := strings.TrimSpace(last.Status)
	transactionHash := strings.TrimSpace(last.TransactionHash)
	return &Error{
		Message:         "circle transaction polling timed out",
		TransactionID:   transactionID,
		LastStatus:      lastStatus,
		TransactionHash: transactionHash,
		Err:             ErrTransactionTimedOut,
	}
}

func parseErrorResponseBody(body []byte) errorResponseBody {
	body = []byte(strings.TrimSpace(string(body)))
	if len(body) == 0 {
		return errorResponseBody{}
	}
	var decoded errorResponseBody
	if err := json.Unmarshal(body, &decoded); err == nil {
		if len(decoded.Error) > 0 && (decoded.Code == "" || decoded.Message == "") {
			var nested errorResponseBody
			if err := json.Unmarshal(decoded.Error, &nested); err == nil {
				if decoded.Code == "" {
					decoded.Code = nested.Code
				}
				if decoded.Message == "" {
					decoded.Message = nested.Message
				}
				if decoded.ErrorReason == "" {
					decoded.ErrorReason = nested.ErrorReason
				}
				if len(decoded.ErrorDetails) == 0 {
					decoded.ErrorDetails = nested.ErrorDetails
				}
			}
		}
		return decoded
	}
	return errorResponseBody{
		Message: sanitizeSummary(string(body)),
	}
}

func (body errorResponseBody) errorDetailsString() string {
	if len(body.ErrorDetails) > 0 && string(body.ErrorDetails) != "null" {
		return sanitizeSummary(string(body.ErrorDetails))
	}
	if len(body.Error) > 0 && string(body.Error) != "null" {
		return sanitizeSummary(string(body.Error))
	}
	return ""
}

var (
	bearerTokenPattern       = regexp.MustCompile(`(?i)Bearer\s+[A-Za-z0-9._~+/=-]+`)
	jsonSecretFieldPattern   = regexp.MustCompile(`(?i)("(?:entitySecretCiphertext|entitySecret|apiKey|authorization|recoveryFile|privateKey)"\s*:\s*)"[^"]*"`)
	plainSecretFieldPattern  = regexp.MustCompile(`(?i)\b(entitySecretCiphertext|entitySecret|apiKey|authorization|recoveryFile|privateKey)\b\s*[:=]\s*[^,\s}]+`)
	longHexSecretPattern     = regexp.MustCompile(`\b[0-9a-fA-F]{64,}\b`)
	longBase64SecretPattern  = regexp.MustCompile(`\b[A-Za-z0-9+/]{80,}={0,2}\b`)
	multiWhitespacePattern   = regexp.MustCompile(`\s+`)
	maxSanitizedErrorSummary = 512
)

func sanitizeSummary(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = bearerTokenPattern.ReplaceAllString(value, "Bearer [redacted]")
	value = jsonSecretFieldPattern.ReplaceAllString(value, `${1}"[redacted]"`)
	value = plainSecretFieldPattern.ReplaceAllString(value, `${1}=[redacted]`)
	value = longHexSecretPattern.ReplaceAllString(value, "[hex-redacted]")
	value = longBase64SecretPattern.ReplaceAllString(value, "[blob-redacted]")
	value = multiWhitespacePattern.ReplaceAllString(value, " ")
	value = strings.TrimSpace(value)
	if len(value) <= maxSanitizedErrorSummary {
		return value
	}
	return value[:maxSanitizedErrorSummary] + "..."
}
