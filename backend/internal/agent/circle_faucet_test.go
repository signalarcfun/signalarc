package agent

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeFaucetEnvCommandRunner struct {
	output  []byte
	err     error
	gotName string
	gotArgs []string
	gotEnv  []string
}

func (runner *fakeFaucetEnvCommandRunner) RunWithEnv(_ context.Context, name string, args []string, env []string) ([]byte, error) {
	runner.gotName = name
	runner.gotArgs = append([]string{}, args...)
	runner.gotEnv = append([]string{}, env...)
	return runner.output, runner.err
}

func TestCircleCLIFaucetRunnerCommandShape(t *testing.T) {
	runner := &fakeFaucetEnvCommandRunner{output: []byte(`{"ok":true}`)}
	faucet := NewCircleCLIFaucetRunner(CircleCLIFaucetRunnerConfig{
		Enabled:       true,
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		CommandRunner: runner,
	})
	if _, err := faucet.RequestFaucet(context.Background(), "0x9999999999999999999999999999999999999999"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{
		"wallet", "fund",
		"--address", "0x9999999999999999999999999999999999999999",
		"--chain", "ARC-TESTNET",
		"--token", "usdc",
		"--output", "json",
	}
	if len(runner.gotArgs) != len(expected) {
		t.Fatalf("expected args %#v, got %#v", expected, runner.gotArgs)
	}
	for index, arg := range expected {
		if runner.gotArgs[index] != arg {
			t.Fatalf("expected arg %d to be %q, got %q (full: %#v)", index, arg, runner.gotArgs[index], runner.gotArgs)
		}
	}
}

func TestCircleCLIFaucetRunnerDisabledReturnsNotConfigured(t *testing.T) {
	faucet := NewCircleCLIFaucetRunner(CircleCLIFaucetRunnerConfig{
		Enabled:       false,
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		CommandRunner: &fakeFaucetEnvCommandRunner{},
	})
	if _, err := faucet.RequestFaucet(context.Background(), "0x9999999999999999999999999999999999999999"); !errors.Is(err, ErrCircleAgentWalletFaucetNotConfigured) {
		t.Fatalf("expected not-configured error, got %v", err)
	}
}

func TestCircleCLIFaucetRunnerWrongChainReturnsFailure(t *testing.T) {
	faucet := NewCircleCLIFaucetRunner(CircleCLIFaucetRunnerConfig{
		Enabled:       true,
		CLIPath:       "circle",
		Chain:         "BASE",
		CommandRunner: &fakeFaucetEnvCommandRunner{},
	})
	if _, err := faucet.RequestFaucet(context.Background(), "0x9999999999999999999999999999999999999999"); !errors.Is(err, ErrCircleAgentWalletFaucetFailed) {
		t.Fatalf("expected failure error, got %v", err)
	}
}

func TestCircleCLIFaucetRunnerEmptyAddressReturnsFailure(t *testing.T) {
	faucet := NewCircleCLIFaucetRunner(CircleCLIFaucetRunnerConfig{
		Enabled:       true,
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		CommandRunner: &fakeFaucetEnvCommandRunner{},
	})
	if _, err := faucet.RequestFaucet(context.Background(), "   "); !errors.Is(err, ErrCircleAgentWalletFaucetFailed) {
		t.Fatalf("expected failure error, got %v", err)
	}
}

func TestParseCircleFaucetOutputCleanJSON(t *testing.T) {
	result, err := parseCircleFaucetOutput([]byte(`{"transactionHash":"0xabc"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.JSON == nil {
		t.Fatal("expected JSON result")
	}
	if result.Message != "" {
		t.Fatalf("expected empty message, got %q", result.Message)
	}
}

func TestParseCircleFaucetOutputNodeWarningPrefix(t *testing.T) {
	output := []byte(`(node:123) [DEP0040] DeprecationWarning: example
{"data":{"status":"submitted"}}`)
	result, err := parseCircleFaucetOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.JSON == nil {
		t.Fatal("expected JSON result after stripping warning prefix")
	}
}

func TestParseCircleFaucetOutputTextOnly(t *testing.T) {
	output := []byte("Faucet request submitted")
	result, err := parseCircleFaucetOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.JSON != nil {
		t.Fatalf("expected nil JSON for text-only output, got %#v", result.JSON)
	}
	if !strings.Contains(result.Message, "Faucet request submitted") {
		t.Fatalf("expected text-only message preserved, got %q", result.Message)
	}
}

func TestParseCircleFaucetOutputTextOnlyRedactsSecretPatterns(t *testing.T) {
	output := []byte("Faucet ok. command: --request circle_request_secret_xyz --otp B1X-123456")
	result, err := parseCircleFaucetOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Message, "circle_request_secret_xyz") {
		t.Fatalf("text-only message must redact request id, got %q", result.Message)
	}
	if strings.Contains(result.Message, "B1X-123456") {
		t.Fatalf("text-only message must redact OTP, got %q", result.Message)
	}
}

func TestParseCircleFaucetOutputEmptyIsFailure(t *testing.T) {
	if _, err := parseCircleFaucetOutput([]byte("")); err == nil {
		t.Fatal("expected error for empty output")
	}
	if _, err := parseCircleFaucetOutput([]byte("    ")); err == nil {
		t.Fatal("expected error for whitespace-only output")
	}
}
