package config

import "testing"

func TestNormalizeCircleAgentWalletExecutor(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "empty defaults cli", value: "", want: "cli"},
		{name: "api", value: "api", want: "api"},
		{name: "api trimmed case folded", value: " API ", want: "api"},
		{name: "unknown defaults cli", value: "bogus", want: "cli"},
		{name: "cli stays cli", value: "cli", want: "cli"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeCircleAgentWalletExecutor(tt.value); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestLoadCircleAPIFields(t *testing.T) {
	t.Setenv("CIRCLE_AGENT_WALLET_EXECUTOR", "api")
	t.Setenv("CIRCLE_API_KEY", "test-api-key")
	t.Setenv("CIRCLE_ENTITY_SECRET", "00112233445566778899aabbccddeeff")
	t.Setenv("CIRCLE_ENTITY_SECRET_CIPHERTEXT", "test-entity-secret-ciphertext")
	t.Setenv("CIRCLE_ALLOW_STATIC_ENTITY_SECRET_CIPHERTEXT", "true")
	t.Setenv("CIRCLE_API_BASE_URL", "http://127.0.0.1:9999")
	t.Setenv("CIRCLE_AGENT_WALLET_TIMEOUT_SECONDS", "7")

	cfg := Load()
	if cfg.CircleAgentWalletExecutor != "api" {
		t.Fatalf("expected api executor, got %q", cfg.CircleAgentWalletExecutor)
	}
	if cfg.CircleAPIKey != "test-api-key" || cfg.CircleEntitySecret == "" || cfg.CircleStaticDevEntitySecretCiphertext != "test-entity-secret-ciphertext" || cfg.CircleAPIBaseURL != "http://127.0.0.1:9999" {
		t.Fatal("unexpected Circle API config")
	}
	if !cfg.CircleAllowStaticEntitySecretCiphertext {
		t.Fatal("expected static ciphertext override flag to parse true")
	}
	if cfg.CircleAgentWalletTimeoutSeconds != 7 {
		t.Fatalf("expected timeout 7, got %d", cfg.CircleAgentWalletTimeoutSeconds)
	}
}

func TestLoadCircleAPIBaseURLDefault(t *testing.T) {
	t.Setenv("CIRCLE_API_BASE_URL", "")

	cfg := Load()
	if cfg.CircleAPIBaseURL != "https://api.circle.com" {
		t.Fatalf("expected Circle API base URL default, got %q", cfg.CircleAPIBaseURL)
	}
}
