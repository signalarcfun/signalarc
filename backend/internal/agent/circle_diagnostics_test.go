package agent

import (
	"context"
	"errors"
	"testing"
)

func TestClassifyCircleErrorOutputDetectsAuthRequiredMarkers(t *testing.T) {
	cases := []string{
		"Error: AUTH_REQUIRED",
		"no agent session is active",
		"No local wallet matches 0xabc on ARC-TESTNET",
		"Run circle wallet login user@example.com --type agent",
	}
	for _, sample := range cases {
		if class := ClassifyCircleErrorOutput(sample); class != CircleErrorClassAuthRequired {
			t.Fatalf("expected auth_required for %q, got %q", sample, class)
		}
	}
}

func TestClassifyCircleErrorOutputUnknownByDefault(t *testing.T) {
	if class := ClassifyCircleErrorOutput("unparseable JSON output"); class != CircleErrorClassUnknown {
		t.Fatalf("expected unknown class, got %q", class)
	}
}

func TestCircleErrorHelpersExtractClassAndSummary(t *testing.T) {
	wrapped := &CircleCLIError{
		Operation:        "circle_agent_wallet_balance",
		ErrorClass:       CircleErrorClassAuthRequired,
		SanitizedSummary: "exit status 1",
		Err:              ErrCircleAgentWalletBalanceFailed,
	}
	if class := CircleErrorClassFromError(wrapped); class != CircleErrorClassAuthRequired {
		t.Fatalf("expected auth_required class, got %q", class)
	}
	if summary := CircleErrorSummaryFromError(wrapped); summary != "exit status 1" {
		t.Fatalf("expected sanitized summary, got %q", summary)
	}
	if !errors.Is(wrapped, ErrCircleAgentWalletBalanceFailed) {
		t.Fatal("expected wrapped error to satisfy errors.Is for the underlying public sentinel")
	}
}

func TestCircleErrorHelpersExtractDiagnosticInterface(t *testing.T) {
	err := fakeCircleDiagnosticError{
		class:   "circle_request_invalid",
		summary: "status=400 code=bad_request message=invalid request",
	}
	if class := CircleErrorClassFromError(err); class != "circle_request_invalid" {
		t.Fatalf("expected circle_request_invalid class, got %q", class)
	}
	if summary := CircleErrorSummaryFromError(err); summary != "status=400 code=bad_request message=invalid request" {
		t.Fatalf("expected diagnostic summary, got %q", summary)
	}
}

func TestCheckAgentSessionLivenessReturnsLiveWhenWalletPresent(t *testing.T) {
	commandRunner := &fakeEnvCommandRunner{
		output: []byte(`{"data":{"wallets":[{"address":"0xa9914bca9123ba0079be8c968f632c0db6400fe7","chain":"ARC-TESTNET"}]}}`),
	}
	resolver := NewCircleCLIWalletResolver(CircleCLIWalletResolverConfig{
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		CommandRunner: commandRunner,
	})

	result := resolver.CheckAgentSessionLiveness(context.Background(), "0xa9914bca9123ba0079be8c968f632c0db6400fe7")
	if result.State != AgentSessionLivenessLive {
		t.Fatalf("expected live, got %#v", result)
	}
}

func TestCheckAgentSessionLivenessReturnsAuthRequiredOnAuthMarkerError(t *testing.T) {
	commandRunner := &fakeEnvCommandRunner{
		output: []byte("Error: AUTH_REQUIRED\nNo local wallet matches 0xabc on ARC-TESTNET, and no agent session is active.\nRun `circle wallet login <email> --type agent`"),
		err:    errors.New("exit status 1"),
	}
	resolver := NewCircleCLIWalletResolver(CircleCLIWalletResolverConfig{
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		CommandRunner: commandRunner,
	})

	result := resolver.CheckAgentSessionLiveness(context.Background(), "0xa9914bca9123ba0079be8c968f632c0db6400fe7")
	if result.State != AgentSessionLivenessAuthRequired {
		t.Fatalf("expected auth_required, got %#v", result)
	}
	if result.ErrorClass != CircleErrorClassAuthRequired {
		t.Fatalf("expected auth_required error class, got %q", result.ErrorClass)
	}
	if result.Reason == "" {
		t.Fatal("expected sanitized reason text")
	}
}

func TestCheckAgentSessionLivenessReturnsAuthRequiredWhenWalletMissing(t *testing.T) {
	commandRunner := &fakeEnvCommandRunner{
		output: []byte(`{"data":{"wallets":[{"address":"0x1111111111111111111111111111111111111111","chain":"ARC-TESTNET"}]}}`),
	}
	resolver := NewCircleCLIWalletResolver(CircleCLIWalletResolverConfig{
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		CommandRunner: commandRunner,
	})

	result := resolver.CheckAgentSessionLiveness(context.Background(), "0xa9914bca9123ba0079be8c968f632c0db6400fe7")
	if result.State != AgentSessionLivenessAuthRequired {
		t.Fatalf("expected auth_required when wallet not in local list, got %#v", result)
	}
}

func TestCheckAgentSessionLivenessReturnsUnknownOnNonAuthError(t *testing.T) {
	commandRunner := &fakeEnvCommandRunner{
		output: []byte("connection refused"),
		err:    errors.New("exit status 7"),
	}
	resolver := NewCircleCLIWalletResolver(CircleCLIWalletResolverConfig{
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		CommandRunner: commandRunner,
	})

	result := resolver.CheckAgentSessionLiveness(context.Background(), "0xa9914bca9123ba0079be8c968f632c0db6400fe7")
	if result.State != AgentSessionLivenessUnknown {
		t.Fatalf("expected unknown state, got %#v", result)
	}
	if result.ErrorClass != CircleErrorClassUnknown {
		t.Fatalf("expected unknown error class, got %q", result.ErrorClass)
	}
}

func TestGetAgentWalletBalancesReturnsClassifiedAuthRequiredError(t *testing.T) {
	commandRunner := &fakeEnvCommandRunner{
		output: []byte("AUTH_REQUIRED\nNo local wallet matches 0xa9914bca9123ba0079be8c968f632c0db6400fe7"),
		err:    errors.New("exit status 1"),
	}
	resolver := NewCircleCLIWalletResolver(CircleCLIWalletResolverConfig{
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		CommandRunner: commandRunner,
	})

	_, err := resolver.GetAgentWalletBalances(context.Background(), "0xa9914bca9123ba0079be8c968f632c0db6400fe7")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrCircleAgentWalletBalanceFailed) {
		t.Fatalf("expected public sentinel preserved, got %v", err)
	}
	if class := CircleErrorClassFromError(err); class != CircleErrorClassAuthRequired {
		t.Fatalf("expected auth_required class on returned error, got %q", class)
	}
}

type fakeCircleDiagnosticError struct {
	class   string
	summary string
}

func (err fakeCircleDiagnosticError) Error() string {
	return "fake circle diagnostic error"
}

func (err fakeCircleDiagnosticError) ErrorClass() string {
	return err.class
}

func (err fakeCircleDiagnosticError) SanitizedSummary() string {
	return err.summary
}
