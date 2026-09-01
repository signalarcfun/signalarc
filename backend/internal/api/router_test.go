package api

import (
	"testing"

	"github.com/wahyu241205/SignalArc/backend/internal/agent"
	"github.com/wahyu241205/SignalArc/backend/internal/config"
)

func TestNewCircleAgentWalletExecutorSelectsCLIByDefault(t *testing.T) {
	executor := newCircleAgentWalletExecutor(config.Config{})
	if _, ok := executor.(*agent.CircleCLIExecutor); !ok {
		t.Fatalf("expected Circle CLI executor, got %T", executor)
	}
}

func TestNewCircleAgentWalletExecutorSelectsAPI(t *testing.T) {
	executor := newCircleAgentWalletExecutor(config.Config{
		CircleAgentWalletExecutionEnabled:     true,
		AppEnv:                                "development",
		CircleAgentWalletExecutor:             "api",
		CircleAPIKey:                          "test-key",
		CircleStaticDevEntitySecretCiphertext: "test-ciphertext",
		CircleAPIBaseURL:                      "http://127.0.0.1:9999",
		CircleAgentWalletTimeoutSeconds:       1,
	})
	if _, ok := executor.(*agent.CircleAPIExecutor); !ok {
		t.Fatalf("expected Circle API executor, got %T", executor)
	}
}

func TestNewCircleAgentWalletBalanceReaderSelectsAPI(t *testing.T) {
	reader := newCircleAgentWalletBalanceReader(config.Config{
		CircleAgentWalletExecutor:       "api",
		CircleAPIKey:                    "test-key",
		CircleAPIBaseURL:                "http://127.0.0.1:9999",
		CircleAgentWalletTimeoutSeconds: 1,
	})
	if _, ok := reader.(*agent.CircleAPIBalanceReader); !ok {
		t.Fatalf("expected Circle API balance reader, got %T", reader)
	}
}

func TestNewCircleAgentWalletBalanceReaderCLIIsNil(t *testing.T) {
	reader := newCircleAgentWalletBalanceReader(config.Config{
		CircleAgentWalletExecutor:       "cli",
		CircleAgentWalletTimeoutSeconds: 1,
	})
	if reader != nil {
		t.Fatalf("expected nil balance reader for CLI executor, got %T", reader)
	}
}

func TestNewCircleAgentWalletBalanceReaderMissingConfigFailsClosed(t *testing.T) {
	reader := newCircleAgentWalletBalanceReader(config.Config{
		CircleAgentWalletExecutor: "api",
	})
	if reader != nil {
		t.Fatalf("expected nil balance reader for invalid API config, got %T", reader)
	}
}

func TestNewCircleAgentWalletExecutorProductionStaticCiphertextFailsClosed(t *testing.T) {
	executor := newCircleAgentWalletExecutor(config.Config{
		AppEnv:                                "production",
		CircleAgentWalletExecutionEnabled:     true,
		CircleAgentWalletExecutor:             "api",
		CircleAPIKey:                          "test-key",
		CircleStaticDevEntitySecretCiphertext: "test-ciphertext",
		CircleAPIBaseURL:                      "http://127.0.0.1:9999",
		CircleAgentWalletTimeoutSeconds:       1,
	})
	if executor != nil {
		t.Fatalf("expected nil executor for production static ciphertext without override, got %T", executor)
	}
}

func TestNewCircleAgentWalletExecutorProductionRawEntitySecretPrefersRawProvider(t *testing.T) {
	executor := newCircleAgentWalletExecutor(config.Config{
		AppEnv:                                "production",
		CircleAgentWalletExecutionEnabled:     true,
		CircleAgentWalletExecutor:             "api",
		CircleAPIKey:                          "test-key",
		CircleEntitySecret:                    "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
		CircleStaticDevEntitySecretCiphertext: "test-ciphertext",
		CircleAPIBaseURL:                      "http://127.0.0.1:9999",
		CircleAgentWalletTimeoutSeconds:       1,
	})
	if _, ok := executor.(*agent.CircleAPIExecutor); !ok {
		t.Fatalf("expected Circle API executor, got %T", executor)
	}
}

func TestNewCircleAgentWalletExecutorProductionStaticCiphertextOverrideAllowsAPI(t *testing.T) {
	executor := newCircleAgentWalletExecutor(config.Config{
		AppEnv:                                  "production",
		CircleAgentWalletExecutionEnabled:       true,
		CircleAgentWalletExecutor:               "api",
		CircleAPIKey:                            "test-key",
		CircleStaticDevEntitySecretCiphertext:   "test-ciphertext",
		CircleAllowStaticEntitySecretCiphertext: true,
		CircleAPIBaseURL:                        "http://127.0.0.1:9999",
		CircleAgentWalletTimeoutSeconds:         1,
	})
	if _, ok := executor.(*agent.CircleAPIExecutor); !ok {
		t.Fatalf("expected Circle API executor, got %T", executor)
	}
}

func TestNewCircleAgentWalletExecutorDevelopmentStaticCiphertextAllowsAPI(t *testing.T) {
	executor := newCircleAgentWalletExecutor(config.Config{
		AppEnv:                                "development",
		CircleAgentWalletExecutionEnabled:     true,
		CircleAgentWalletExecutor:             "api",
		CircleAPIKey:                          "test-key",
		CircleStaticDevEntitySecretCiphertext: "test-ciphertext",
		CircleAPIBaseURL:                      "http://127.0.0.1:9999",
		CircleAgentWalletTimeoutSeconds:       1,
	})
	if _, ok := executor.(*agent.CircleAPIExecutor); !ok {
		t.Fatalf("expected Circle API executor, got %T", executor)
	}
}

func TestNewCircleAgentWalletExecutorAPIMissingConfigFailsClosed(t *testing.T) {
	executor := newCircleAgentWalletExecutor(config.Config{
		CircleAgentWalletExecutionEnabled: true,
		CircleAgentWalletExecutor:         "api",
	})
	if executor != nil {
		t.Fatalf("expected nil executor for invalid API config, got %T", executor)
	}
}
