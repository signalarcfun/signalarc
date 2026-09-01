package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                               string
	AppPort                              string
	DatabaseURL                          string
	CircleAgentWalletExecutionEnabled    bool
	CircleAgentOnboardingOTPStartEnabled bool
	CircleAgentWalletFaucetEnabled       bool
	CircleCLIPath                        string
	CircleAgentWalletChain               string
	CircleAgentWalletTimeoutSeconds      int
	CircleAgentWalletExecutor            string
	CircleAPIKey                         string
	// CircleEntitySecret is the raw Circle entity secret. Store it only in a
	// secret manager or protected environment; runtime code generates
	// entitySecretCiphertext per Circle request.
	CircleEntitySecret                      string
	CircleStaticDevEntitySecretCiphertext   string
	CircleAllowStaticEntitySecretCiphertext bool
	CircleAPIBaseURL                        string
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		AppEnv:                                  os.Getenv("APP_ENV"),
		AppPort:                                 os.Getenv("APP_PORT"),
		DatabaseURL:                             os.Getenv("DATABASE_URL"),
		CircleAgentWalletExecutionEnabled:       parseBool(os.Getenv("CIRCLE_AGENT_WALLET_EXECUTION_ENABLED")),
		CircleAgentOnboardingOTPStartEnabled:    parseBool(os.Getenv("CIRCLE_AGENT_ONBOARDING_OTP_START_ENABLED")),
		CircleAgentWalletFaucetEnabled:          parseBool(os.Getenv("CIRCLE_AGENT_WALLET_FAUCET_ENABLED")),
		CircleCLIPath:                           os.Getenv("CIRCLE_CLI_PATH"),
		CircleAgentWalletChain:                  os.Getenv("CIRCLE_AGENT_WALLET_CHAIN"),
		CircleAgentWalletTimeoutSeconds:         parseInt(os.Getenv("CIRCLE_AGENT_WALLET_TIMEOUT_SECONDS"), 120),
		CircleAgentWalletExecutor:               normalizeCircleAgentWalletExecutor(os.Getenv("CIRCLE_AGENT_WALLET_EXECUTOR")),
		CircleAPIKey:                            os.Getenv("CIRCLE_API_KEY"),
		CircleEntitySecret:                      os.Getenv("CIRCLE_ENTITY_SECRET"),
		CircleStaticDevEntitySecretCiphertext:   os.Getenv("CIRCLE_ENTITY_SECRET_CIPHERTEXT"),
		CircleAllowStaticEntitySecretCiphertext: parseBool(os.Getenv("CIRCLE_ALLOW_STATIC_ENTITY_SECRET_CIPHERTEXT")),
		CircleAPIBaseURL:                        os.Getenv("CIRCLE_API_BASE_URL"),
	}

	if cfg.AppEnv == "" {
		cfg.AppEnv = "development"
	}
	if cfg.AppPort == "" {
		cfg.AppPort = "4000"
	}
	if cfg.CircleCLIPath == "" {
		cfg.CircleCLIPath = "circle"
	}
	if cfg.CircleAgentWalletChain == "" {
		cfg.CircleAgentWalletChain = "ARC-TESTNET"
	}
	if cfg.CircleAPIBaseURL == "" {
		cfg.CircleAPIBaseURL = "https://api.circle.com"
	}

	return cfg
}

func (c Config) ValidateDatabaseURL() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	return nil
}

func parseBool(value string) bool {
	parsed, err := strconv.ParseBool(value)
	return err == nil && parsed
}

func parseInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func normalizeCircleAgentWalletExecutor(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "api":
		return "api"
	default:
		return "cli"
	}
}
