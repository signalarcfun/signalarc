package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/repository"
)

type testWalletAuthSessionReader struct {
	session repository.WalletAuthSession
	err     error
}

func (reader testWalletAuthSessionReader) GetActiveSessionByTokenHash(_ context.Context, _ string) (repository.WalletAuthSession, error) {
	return reader.session, reader.err
}

func TestIsCORSOriginAllowedIncludesDefaultFrontendOrigins(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	for _, origin := range []string{
		"http://localhost:3000",
		"http://127.0.0.1:3000",
		"https://signalarc.fun",
	} {
		if !isCORSOriginAllowed(origin) {
			t.Fatalf("expected default frontend origin %q to be allowed", origin)
		}
	}
}

func TestIsCORSOriginAllowedUsesConfiguredExactOrigins(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://example.signalarc.fun, https://agents.signalarc.fun")

	if !isCORSOriginAllowed("https://example.signalarc.fun") {
		t.Fatal("expected configured origin to be allowed")
	}
	if !isCORSOriginAllowed("https://agents.signalarc.fun") {
		t.Fatal("expected configured origin with surrounding whitespace to be allowed")
	}
}

func TestIsCORSOriginAllowedDeniesUnknownAndWildcardOrigins(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "*,https://example.signalarc.fun")

	if isCORSOriginAllowed("https://evil.example") {
		t.Fatal("expected unknown origin to be denied")
	}
	if isCORSOriginAllowed("*") {
		t.Fatal("expected wildcard origin to be denied")
	}
}

func TestLocalCORSMiddlewareAllowsPatchPreflight(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	handlerCalled := false
	handler := localCORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}))
	request := httptest.NewRequest(http.MethodOptions, "/markets/test-id/contract", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	request.Header.Set("Access-Control-Request-Method", http.MethodPatch)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if handlerCalled {
		t.Fatal("expected OPTIONS request to end in CORS middleware")
	}
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected localhost allow origin header, got %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, http.MethodPatch) {
		t.Fatalf("expected allow methods to include PATCH, got %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "Authorization") {
		t.Fatalf("expected allow headers to include Authorization, got %q", got)
	}
}

func TestAuthenticatedWalletUserMiddlewareProvidesVerifiedIdentity(t *testing.T) {
	expected := repository.WalletAuthSession{
		UserID:        "10000000-0000-4000-8000-000000000001",
		WalletAddress: "0xb592d99cb3c98b77777d6288e5e5782ac2ce919a",
		ExpiresAt:     time.Now().Add(time.Hour),
	}
	handler := authenticatedWalletUserMiddleware(testWalletAuthSessionReader{session: expected}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := walletAuthIdentityFromContext(r.Context())
		if !ok || identity != expected {
			t.Fatalf("expected authenticated wallet identity, got %#v", identity)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodPost, "/markets", nil)
	request.Header.Set("Authorization", "Bearer session-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
}
