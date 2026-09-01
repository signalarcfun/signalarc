package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsCORSOriginAllowedIncludesLocalDefaults(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	if !isCORSOriginAllowed("http://localhost:3000") {
		t.Fatal("expected localhost frontend origin to be allowed")
	}
	if !isCORSOriginAllowed("http://127.0.0.1:3000") {
		t.Fatal("expected loopback frontend origin to be allowed")
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
}
