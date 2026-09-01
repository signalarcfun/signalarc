package agent

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type fakeEnvCommandRunner struct {
	name   string
	args   []string
	env    []string
	output []byte
	err    error
}

func (runner *fakeEnvCommandRunner) RunWithEnv(_ context.Context, name string, args []string, env []string) ([]byte, error) {
	runner.name = name
	runner.args = append([]string{}, args...)
	runner.env = append([]string{}, env...)
	if runner.err != nil {
		return runner.output, runner.err
	}
	if runner.output != nil {
		return runner.output, nil
	}
	return []byte(`{"ok":true}`), nil
}

type failingEnvCommandRunner struct {
	output []byte
	err    error
}

func (runner failingEnvCommandRunner) RunWithEnv(_ context.Context, _ string, _ []string, _ []string) ([]byte, error) {
	return runner.output, runner.err
}

func TestCircleCLIOnboardingRunnerVerifyOTPUsesDocumentedCommandShape(t *testing.T) {
	commandRunner := &fakeEnvCommandRunner{}
	runner := NewCircleCLIOnboardingRunner(CircleCLIOnboardingRunnerConfig{
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		CommandRunner: commandRunner,
	})

	if err := runner.VerifyOTP(context.Background(), "request-123", "SIN-232794"); err != nil {
		t.Fatalf("verify otp: %v", err)
	}

	if commandRunner.name != "circle" {
		t.Fatalf("expected circle command, got %q", commandRunner.name)
	}
	expectedArgs := []string{"wallet", "login", "--request", "request-123", "--otp", "SIN-232794"}
	if !slices.Equal(commandRunner.args, expectedArgs) {
		t.Fatalf("unexpected args %#v", commandRunner.args)
	}
	if !slices.Contains(commandRunner.env, "CIRCLE_ACCEPT_TERMS=1") {
		t.Fatalf("expected CIRCLE_ACCEPT_TERMS=1 env, got %#v", commandRunner.env)
	}
}

func TestFindCircleOTPRequestIDFromJSONRequestID(t *testing.T) {
	requestID, ok := findCircleOTPRequestID([]byte(`{"request_id":"request-json-123"}`))
	if !ok {
		t.Fatal("expected request id")
	}
	if requestID != "request-json-123" {
		t.Fatalf("unexpected request id %q", requestID)
	}
}

func TestFindCircleOTPRequestIDFromJSONRequestId(t *testing.T) {
	requestID, ok := findCircleOTPRequestID([]byte(`{"requestId":"request-camel-123"}`))
	if !ok {
		t.Fatal("expected request id")
	}
	if requestID != "request-camel-123" {
		t.Fatalf("unexpected request id %q", requestID)
	}
}

func TestFindCircleOTPRequestIDFromTextLabel(t *testing.T) {
	requestID, ok := findCircleOTPRequestID([]byte("OTP email sent\nRequest ID: request-text-123\nExpires in 10 minutes"))
	if !ok {
		t.Fatal("expected request id")
	}
	if requestID != "request-text-123" {
		t.Fatalf("unexpected request id %q", requestID)
	}
}

func TestFindCircleOTPRequestIDFromPlainTextID(t *testing.T) {
	requestID, ok := findCircleOTPRequestID([]byte("request-plain-123"))
	if !ok {
		t.Fatal("expected request id")
	}
	if requestID != "request-plain-123" {
		t.Fatalf("unexpected request id %q", requestID)
	}
}

func TestFindCircleOTPRequestIDFromNestedJSONCompletionCommand(t *testing.T) {
	output := []byte(`{"data":{"message":"OTP code sent to desi@example.com\nPlease run: circle wallet login --request request-nested-123 --otp <code>"}}`)
	requestID, ok := findCircleOTPRequestID(output)
	if !ok {
		t.Fatal("expected request id")
	}
	if requestID != "request-nested-123" {
		t.Fatalf("unexpected request id %q", requestID)
	}
}

func TestFindCircleOTPRequestIDFromTextCompletionCommand(t *testing.T) {
	output := []byte("Please run: circle wallet login --request request-command-123 --otp <code>")
	requestID, ok := findCircleOTPRequestID(output)
	if !ok {
		t.Fatal("expected request id")
	}
	if requestID != "request-command-123" {
		t.Fatalf("unexpected request id %q", requestID)
	}
}

func TestCircleCLIOnboardingRunnerStartOTPSucceedsWhenFailedCommandPrintedRequestID(t *testing.T) {
	requestID := "request-start-command-123"
	email := "desi@example.com"
	commandRunner := &fakeEnvCommandRunner{
		output: []byte(`{"data":{"message":"OTP code sent to desi@example.com\nPlease run: circle wallet login --request request-start-command-123 --otp <code>"}}`),
		err:    errors.New("exit status 1 for desi@example.com request-start-command-123"),
	}
	runner := NewCircleCLIOnboardingRunner(CircleCLIOnboardingRunnerConfig{
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		CommandRunner: commandRunner,
	})

	var logs bytes.Buffer
	previousLogger := log.Logger
	log.Logger = zerolog.New(&logs)
	defer func() {
		log.Logger = previousLogger
	}()

	result, err := runner.StartOTP(context.Background(), email)
	if err != nil {
		t.Fatalf("start otp: %v", err)
	}
	if result.RequestID != requestID {
		t.Fatalf("unexpected request id %q", result.RequestID)
	}

	logText := logs.String()
	if strings.Contains(logText, requestID) {
		t.Fatalf("sanitized diagnostics exposed request ID: %s", logText)
	}
	if strings.Contains(logText, email) {
		t.Fatalf("sanitized diagnostics exposed email: %s", logText)
	}
	if !strings.Contains(logText, "Circle CLI OTP start failed") {
		t.Fatalf("expected start diagnostic log, got %s", logText)
	}
}

func TestSanitizeCircleOnboardingTextRedactsCommandRequestID(t *testing.T) {
	sanitized := sanitizeCircleOnboardingText("Please run: circle wallet login --request request-secret-123 --otp SIN-232794")
	if strings.Contains(sanitized, "request-secret-123") {
		t.Fatalf("sanitized text exposed request ID: %s", sanitized)
	}
	if strings.Contains(sanitized, "SIN-232794") {
		t.Fatalf("sanitized text exposed OTP: %s", sanitized)
	}
	if !strings.Contains(sanitized, "--request [redacted]") {
		t.Fatalf("expected redacted request id, got %s", sanitized)
	}
}

func TestCircleCLIWalletResolverResolveAgentWalletUsesReadOnlyList(t *testing.T) {
	commandRunner := &fakeEnvCommandRunner{
		output: []byte(`{"data":{"wallets":[{"address":"0xa9914bca9123ba0079be8c968f632c0db6400fe7","chain":"ARC-TESTNET"}]}}`),
	}
	resolver := NewCircleCLIWalletResolver(CircleCLIWalletResolverConfig{
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		CommandRunner: commandRunner,
	})

	wallet, err := resolver.ResolveAgentWallet(context.Background(), "desi@example.com")
	if err != nil {
		t.Fatalf("resolve wallet: %v", err)
	}
	if wallet.Address != "0xa9914bca9123ba0079be8c968f632c0db6400fe7" {
		t.Fatalf("unexpected wallet address %q", wallet.Address)
	}
	expectedArgs := []string{"wallet", "list", "--type", "agent", "--chain", ChainArcTestnet, "--output", "json"}
	if !slices.Equal(commandRunner.args, expectedArgs) {
		t.Fatalf("unexpected args %#v", commandRunner.args)
	}
	if slices.Contains(commandRunner.args, "execute") {
		t.Fatalf("resolver must not use write commands: %#v", commandRunner.args)
	}
}

func TestCircleCLIWalletResolverAmbiguousAgentWallets(t *testing.T) {
	commandRunner := &fakeEnvCommandRunner{
		output: []byte(`{"data":{"wallets":[{"address":"0xa9914bca9123ba0079be8c968f632c0db6400fe7"},{"address":"0x96d5051a005547eba149f71604ccf58ae1a7c950"}]}}`),
	}
	resolver := NewCircleCLIWalletResolver(CircleCLIWalletResolverConfig{
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		CommandRunner: commandRunner,
	})

	_, err := resolver.ResolveAgentWallet(context.Background(), "desi@example.com")
	if !errors.Is(err, ErrCircleAgentWalletResolutionAmbiguous) {
		t.Fatalf("expected ambiguous wallet error, got %v", err)
	}
}

func TestCircleCLIWalletResolverBalancesUsesReadOnlyBalance(t *testing.T) {
	commandRunner := &fakeEnvCommandRunner{
		output: []byte(`{"data":{"balances":[]}}`),
	}
	resolver := NewCircleCLIWalletResolver(CircleCLIWalletResolverConfig{
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		CommandRunner: commandRunner,
	})

	balances, err := resolver.GetAgentWalletBalances(context.Background(), "0xa9914bca9123ba0079be8c968f632c0db6400fe7")
	if err != nil {
		t.Fatalf("get balances: %v", err)
	}
	if len(balances.Balances) != 0 {
		t.Fatalf("expected empty balances, got %#v", balances.Balances)
	}
	expectedArgs := []string{"wallet", "balance", "--address", "0xa9914bca9123ba0079be8c968f632c0db6400fe7", "--chain", ChainArcTestnet, "--output", "json"}
	if !slices.Equal(commandRunner.args, expectedArgs) {
		t.Fatalf("unexpected args %#v", commandRunner.args)
	}
	if slices.Contains(commandRunner.args, "execute") {
		t.Fatalf("balance resolver must not use write commands: %#v", commandRunner.args)
	}
}

func TestCircleCLIOnboardingRunnerStartOTPSanitizesFailureDiagnostics(t *testing.T) {
	email := "desi@example.com"
	commandRunner := &fakeEnvCommandRunner{
		output: []byte("OTP start failed for desi@example.com without request id"),
		err:    errors.New("exit status 1 for desi@example.com"),
	}
	runner := NewCircleCLIOnboardingRunner(CircleCLIOnboardingRunnerConfig{
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		CommandRunner: commandRunner,
	})

	var logs bytes.Buffer
	previousLogger := log.Logger
	log.Logger = zerolog.New(&logs)
	defer func() {
		log.Logger = previousLogger
	}()

	_, err := runner.StartOTP(context.Background(), email)
	if err == nil {
		t.Fatal("expected start error")
	}
	if !errors.Is(err, ErrCircleOnboardingCommandFailed) {
		t.Fatalf("expected command failed error, got %v", err)
	}

	errorText := err.Error()
	logText := logs.String()
	for _, text := range []string{errorText, logText} {
		if strings.Contains(text, email) {
			t.Fatalf("sanitized diagnostics exposed email: %s", text)
		}
		if !strings.Contains(text, "OTP start failed for [redacted]") {
			t.Fatalf("expected sanitized start output detail, got %s", text)
		}
	}
	if !strings.Contains(logText, "Circle CLI OTP start failed") {
		t.Fatalf("expected start failure log message, got %s", logText)
	}
}

func TestCircleCLIOnboardingRunnerVerifyOTPSanitizesFailureDiagnostics(t *testing.T) {
	requestID := "request-secret-123"
	otp := "SIN-232794"
	commandOutput := []byte("circle wallet login --request request-secret-123 --otp SIN-232794 failed: invalid or expired OTP")
	commandRunner := failingEnvCommandRunner{
		output: commandOutput,
		err:    errors.New("exit status 1 for request-secret-123 with SIN-232794"),
	}
	runner := NewCircleCLIOnboardingRunner(CircleCLIOnboardingRunnerConfig{
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		CommandRunner: commandRunner,
	})

	var logs bytes.Buffer
	previousLogger := log.Logger
	log.Logger = zerolog.New(&logs)
	defer func() {
		log.Logger = previousLogger
	}()

	err := runner.VerifyOTP(context.Background(), requestID, otp)
	if err == nil {
		t.Fatal("expected verify error")
	}
	if !errors.Is(err, ErrCircleOnboardingCommandFailed) {
		t.Fatalf("expected command failed error, got %v", err)
	}

	errorText := err.Error()
	logText := logs.String()
	for _, text := range []string{errorText, logText} {
		if strings.Contains(text, requestID) {
			t.Fatalf("sanitized diagnostics exposed request ID: %s", text)
		}
		if strings.Contains(text, otp) {
			t.Fatalf("sanitized diagnostics exposed OTP: %s", text)
		}
		if !strings.Contains(text, "invalid or expired OTP") {
			t.Fatalf("expected sanitized diagnostic detail, got %s", text)
		}
	}
	if strings.Count(errorText, "[redacted]") < 2 {
		t.Fatalf("expected redacted request ID and OTP in error text, got %s", errorText)
	}
	if strings.Count(logText, "[redacted]") < 2 {
		t.Fatalf("expected redacted request ID and OTP in log text, got %s", logText)
	}
}

func TestParseCircleAgentWalletsWithNodeWarningPrefix(t *testing.T) {
	output := []byte("(node:12345) [DEP0040] DeprecationWarning: The `punycode` module is deprecated. Please use a userland alternative instead.\n(Use `node --trace-deprecation ...` to show where the warning was created)\n{\"data\": {\"wallets\": [{\"type\": \"agent\",\"address\": \"0xa9914bca9123ba0079be8c968f632c0db6400fe7\",\"blockchain\": \"ARC-TESTNET\",\"createDate\": \"2026-05-22T16:04:04Z\"}]}}")
	wallets, err := parseCircleAgentWallets(output)
	if err != nil {
		t.Fatalf("expected successful parse with warning prefix, got error: %v", err)
	}
	if len(wallets) != 1 {
		t.Fatalf("expected 1 wallet, got %d", len(wallets))
	}
	if wallets[0].Address != "0xa9914bca9123ba0079be8c968f632c0db6400fe7" {
		t.Fatalf("unexpected wallet address %q", wallets[0].Address)
	}
}

func TestParseCircleAgentWalletBalancesWithNodeWarningPrefix(t *testing.T) {
	output := []byte("(node:67890) [DEP0040] DeprecationWarning: The `punycode` module is deprecated.\n{\"data\":{\"balances\":[{\"token\":\"USDC\",\"amount\":\"20.00\"}]}}")
	balances, err := parseCircleAgentWalletBalances(output)
	if err != nil {
		t.Fatalf("expected successful parse with warning prefix, got error: %v", err)
	}
	if len(balances) != 1 {
		t.Fatalf("expected 1 balance entry, got %d", len(balances))
	}
}

func TestParseCircleAgentWalletsInvalidOutputWithoutJSON(t *testing.T) {
	output := []byte("(node:12345) [DEP0040] DeprecationWarning: something\nError: authentication required\nPlease login first.")
	_, err := parseCircleAgentWallets(output)
	if err == nil {
		t.Fatal("expected parse error for output without JSON")
	}
}

func TestParseCircleAgentWalletBalancesInvalidOutputWithoutJSON(t *testing.T) {
	output := []byte("Error: wallet not found\nPlease check the address.")
	_, err := parseCircleAgentWalletBalances(output)
	if err == nil {
		t.Fatal("expected parse error for output without JSON")
	}
}

func TestParseCircleAgentWalletsCleanJSONStillWorks(t *testing.T) {
	output := []byte(`{"data":{"wallets":[{"address":"0xa9914bca9123ba0079be8c968f632c0db6400fe7","chain":"ARC-TESTNET"}]}}`)
	wallets, err := parseCircleAgentWallets(output)
	if err != nil {
		t.Fatalf("expected successful parse for clean JSON, got error: %v", err)
	}
	if len(wallets) != 1 {
		t.Fatalf("expected 1 wallet, got %d", len(wallets))
	}
}

func TestParseCircleAgentWalletBalancesCleanJSONStillWorks(t *testing.T) {
	output := []byte(`{"data":{"balances":[{"token":"USDC","amount":"10.00"}]}}`)
	balances, err := parseCircleAgentWalletBalances(output)
	if err != nil {
		t.Fatalf("expected successful parse for clean JSON, got error: %v", err)
	}
	if len(balances) != 1 {
		t.Fatalf("expected 1 balance entry, got %d", len(balances))
	}
}
