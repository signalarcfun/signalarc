package circleapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientNon2xxJSONErrorReturnsSanitizedDiagnostics(t *testing.T) {
	secretValue := strings.Repeat("a", 96)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code":"bad_request",
			"message":"estimation reverted while processing request",
			"errorReason":"execution reverted",
			"errorDetails":{"entitySecretCiphertext":"` + secretValue + `","safe":"insufficient gas"}
		}`))
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{APIKey: "test-key", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	_, err = client.GetTransaction(context.Background(), "tx_123")
	if err == nil {
		t.Fatal("expected error")
	}
	var circleErr *Error
	if !errors.As(err, &circleErr) {
		t.Fatalf("expected circle api error, got %T", err)
	}
	if circleErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", circleErr.StatusCode)
	}
	if circleErr.ErrorClass() != "estimation_reverted" {
		t.Fatalf("expected estimation_reverted class, got %q", circleErr.ErrorClass())
	}
	summary := circleErr.SanitizedSummary()
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
	if !strings.Contains(summary, "bad_request") || !strings.Contains(summary, "insufficient gas") {
		t.Fatalf("summary missing safe diagnostic detail: %q", summary)
	}
	assertNoSecretLikeValue(t, summary, secretValue)
}

func TestCircleAPIErrorClassification(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{name: "auth", err: &Error{StatusCode: http.StatusUnauthorized}, want: "auth_failed"},
		{name: "request", err: &Error{StatusCode: http.StatusBadRequest}, want: "circle_request_invalid"},
		{name: "server", err: &Error{StatusCode: http.StatusBadGateway}, want: "circle_api_error"},
		{name: "api timeout", err: &Error{Err: context.DeadlineExceeded}, want: "circle_api_timeout"},
		{name: "transaction timeout", err: &Error{Err: ErrTransactionTimedOut}, want: "transaction_timeout"},
		{name: "evm revert", err: &Error{CircleMessage: "execution reverted: market closed"}, want: "evm_revert"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.ErrorClass(); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
			if summary := tt.err.SanitizedSummary(); summary == "" {
				t.Fatal("expected non-empty summary")
			}
		})
	}
}

func TestPollTransactionTimeoutReturnsDiagnosticError(t *testing.T) {
	secretValue := strings.Repeat("c", 96)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"id":"tx_123","status":"PENDING","transactionHash":"0xabc123","debug":"` + secretValue + `"}}`))
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		APIKey:       "test-key",
		BaseURL:      server.URL,
		PollInterval: time.Millisecond,
		PollTimeout:  25 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	_, err = client.PollTransaction(context.Background(), "tx_123")
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !errors.Is(err, ErrTransactionTimedOut) {
		t.Fatalf("expected transaction timeout sentinel, got %v", err)
	}
	var circleErr *Error
	if !errors.As(err, &circleErr) {
		t.Fatalf("expected diagnostic circle api error, got %T", err)
	}
	if class := circleErr.ErrorClass(); class != "transaction_timeout" {
		t.Fatalf("expected transaction_timeout class, got %q", class)
	}
	summary := circleErr.SanitizedSummary()
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
	if !strings.Contains(summary, "transaction_id=tx_123") || !strings.Contains(summary, "last_status=PENDING") {
		t.Fatalf("summary missing transaction metadata: %q", summary)
	}
	if !strings.Contains(summary, "transaction_hash=0xabc123") {
		t.Fatalf("summary missing transaction hash: %q", summary)
	}
	assertNoSecretLikeValue(t, summary, secretValue)
}

func TestSanitizedSummaryRedactsSecretLikeValues(t *testing.T) {
	secretValue := strings.Repeat("b", 96)
	err := &Error{
		StatusCode:    http.StatusBadRequest,
		CircleMessage: `Authorization: Bearer tokenvalue entitySecretCiphertext=` + secretValue,
		ErrorDetails:  `{"apiKey":"` + secretValue + `","entitySecret":"` + secretValue + `"}`,
	}
	summary := err.SanitizedSummary()
	assertNoSecretLikeValue(t, summary, secretValue)
	if strings.Contains(strings.ToLower(summary), "bearer tokenvalue") {
		t.Fatalf("summary leaked bearer token: %q", summary)
	}
}

func assertNoSecretLikeValue(t *testing.T, summary string, secretValue string) {
	t.Helper()
	if strings.Contains(summary, secretValue) {
		t.Fatalf("summary leaked secret-like value: %q", summary)
	}
	if strings.Contains(summary, "entitySecretCiphertext="+secretValue) {
		t.Fatalf("summary leaked ciphertext field: %q", summary)
	}
}
