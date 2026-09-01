package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/wahyu241205/SignalArc/backend/internal/agent"
)

type stubCircleFaucetRunner struct {
	result    agent.CircleAgentWalletFaucetResult
	err       error
	called    bool
	gotAddr   string
	callCount int
}

func (runner *stubCircleFaucetRunner) RequestFaucet(_ context.Context, address string) (agent.CircleAgentWalletFaucetResult, error) {
	runner.called = true
	runner.callCount++
	runner.gotAddr = address
	if runner.err != nil {
		return agent.CircleAgentWalletFaucetResult{}, runner.err
	}
	return runner.result, nil
}

type recordingFaucetEnvCommandRunner struct {
	output  []byte
	err     error
	gotName string
	gotArgs []string
	gotEnv  []string
}

func (runner *recordingFaucetEnvCommandRunner) RunWithEnv(_ context.Context, name string, args []string, env []string) ([]byte, error) {
	runner.gotName = name
	runner.gotArgs = append([]string{}, args...)
	runner.gotEnv = append([]string{}, env...)
	if runner.err != nil {
		return runner.output, runner.err
	}
	return runner.output, nil
}

func newFaucetRouter(walletRegistry *testAgentWalletRegistry, faucet agent.CircleAgentWalletFaucet) http.Handler {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), walletRegistry, nil, newTestAgentSessionRegistry(), agent.CircleOnboardingStarter{}, &stubCircleWalletResolver{}, faucet)
	return router
}

func TestFaucetSuccessUsesRegisteredAddressAndExactCommandShape(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	command := &recordingFaucetEnvCommandRunner{
		output: []byte(`(node:123) [DEP0040] DeprecationWarning: example
{"data":{"transactionHash":"0xabc"}}`),
	}
	runner := agent.NewCircleCLIFaucetRunner(agent.CircleCLIFaucetRunnerConfig{
		Enabled:       true,
		CLIPath:       "circle",
		Chain:         agent.ChainArcTestnet,
		CommandRunner: command,
	})
	router := newFaucetRouter(walletRegistry, runner)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/wallets/agent_test_1/faucet", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if command.gotName != "circle" {
		t.Fatalf("expected CLI path circle, got %q", command.gotName)
	}
	expectedArgs := []string{
		"wallet", "fund",
		"--address", "0x9999999999999999999999999999999999999999",
		"--chain", "ARC-TESTNET",
		"--token", "usdc",
		"--output", "json",
	}
	assertStringSliceEqual(t, command.gotArgs, expectedArgs)
	for _, blocked := range []string{"--amount", "--method", "--open", "--export", "transfer", "swap", "execute"} {
		for _, arg := range command.gotArgs {
			if arg == blocked {
				t.Fatalf("argument %q must not be present in faucet command", blocked)
			}
		}
	}

	var body struct {
		AgentWalletFaucet agentWalletFaucetResponse `json:"agent_wallet_faucet"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.AgentWalletFaucet.AgentID != "agent_test_1" {
		t.Fatalf("unexpected agent id %q", body.AgentWalletFaucet.AgentID)
	}
	if body.AgentWalletFaucet.AgentWalletAddress != "0x9999999999999999999999999999999999999999" {
		t.Fatalf("unexpected wallet address %q", body.AgentWalletFaucet.AgentWalletAddress)
	}
	if body.AgentWalletFaucet.Chain != agent.ChainArcTestnet {
		t.Fatalf("unexpected chain %q", body.AgentWalletFaucet.Chain)
	}
	if body.AgentWalletFaucet.Token != agent.FaucetTokenUSDC {
		t.Fatalf("unexpected token %q", body.AgentWalletFaucet.Token)
	}
	if body.AgentWalletFaucet.Status != agent.FaucetStatusRequested {
		t.Fatalf("unexpected status %q", body.AgentWalletFaucet.Status)
	}
	if body.AgentWalletFaucet.Result == nil {
		t.Fatal("expected result payload")
	}
}

func TestFaucetMissingWalletReturnsNotFound(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	runner := &stubCircleFaucetRunner{}
	router := newFaucetRouter(walletRegistry, runner)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/wallets/agent_missing/faucet", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("agent_wallet_not_found")) {
		t.Fatalf("expected agent_wallet_not_found, got %s", response.Body.String())
	}
	if runner.called {
		t.Fatal("faucet runner must not be called when wallet missing")
	}
}

func TestFaucetDisabledWalletReturnsConflict(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	if _, err := walletRegistry.DisableAgentWallet(context.Background(), "agent_test_1"); err != nil {
		t.Fatalf("disable wallet: %v", err)
	}
	runner := &stubCircleFaucetRunner{}
	router := newFaucetRouter(walletRegistry, runner)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/wallets/agent_test_1/faucet", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("agent_wallet_status_invalid")) {
		t.Fatalf("expected agent_wallet_status_invalid, got %s", response.Body.String())
	}
	if runner.called {
		t.Fatal("faucet runner must not be called when wallet is disabled")
	}
}

func TestFaucetWrongChainReturnsConflict(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	wallet := walletRegistry.wallets["agent_test_1"]
	wallet.Chain = "BASE"
	walletRegistry.wallets["agent_test_1"] = wallet
	runner := &stubCircleFaucetRunner{}
	router := newFaucetRouter(walletRegistry, runner)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/wallets/agent_test_1/faucet", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("agent_wallet_chain_invalid")) {
		t.Fatalf("expected agent_wallet_chain_invalid, got %s", response.Body.String())
	}
	if runner.called {
		t.Fatal("faucet runner must not be called when chain is invalid")
	}
}

func TestFaucetCLIFailureReturnsBadGatewayAndDoesNotLeakSecrets(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	command := &recordingFaucetEnvCommandRunner{
		output: []byte("circle session token CIRCLE_SECRET_TOKEN_123 request_id req_secret_456 /home/user/.circle/session.json desi@example.com"),
		err:    errors.New("exit status 1 with token CIRCLE_SECRET_TOKEN_123 and request_id req_secret_456"),
	}
	runner := agent.NewCircleCLIFaucetRunner(agent.CircleCLIFaucetRunnerConfig{
		Enabled:       true,
		CLIPath:       "circle",
		Chain:         agent.ChainArcTestnet,
		CommandRunner: command,
	})
	router := newFaucetRouter(walletRegistry, runner)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/wallets/agent_test_1/faucet", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadGateway, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("circle_agent_wallet_faucet_failed")) {
		t.Fatalf("expected circle_agent_wallet_faucet_failed, got %s", response.Body.String())
	}
	leaks := []string{
		"CIRCLE_SECRET_TOKEN_123",
		"req_secret_456",
		"/home/user/.circle/session.json",
		"desi@example.com",
		"exit status 1",
	}
	for _, secret := range leaks {
		if bytes.Contains(response.Body.Bytes(), []byte(secret)) {
			t.Fatalf("response must not expose %q: %s", secret, response.Body.String())
		}
	}
}

func TestFaucetNotConfiguredReturnsNotImplemented(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), walletRegistry, nil, newTestAgentSessionRegistry(), agent.CircleOnboardingStarter{}, &stubCircleWalletResolver{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/wallets/agent_test_1/faucet", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotImplemented {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotImplemented, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("circle_agent_wallet_faucet_not_configured")) {
		t.Fatalf("expected circle_agent_wallet_faucet_not_configured, got %s", response.Body.String())
	}
}

func TestFaucetDisabledRunnerReturnsNotImplemented(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	runner := agent.NewCircleCLIFaucetRunner(agent.CircleCLIFaucetRunnerConfig{
		Enabled:       false,
		CLIPath:       "circle",
		Chain:         agent.ChainArcTestnet,
		CommandRunner: &recordingFaucetEnvCommandRunner{},
	})
	router := newFaucetRouter(walletRegistry, runner)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/wallets/agent_test_1/faucet", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotImplemented {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotImplemented, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("circle_agent_wallet_faucet_not_configured")) {
		t.Fatalf("expected circle_agent_wallet_faucet_not_configured, got %s", response.Body.String())
	}
}

func TestFaucetIgnoresArbitraryRecipientFromRequestBody(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	command := &recordingFaucetEnvCommandRunner{output: []byte(`{"ok":true}`)}
	runner := agent.NewCircleCLIFaucetRunner(agent.CircleCLIFaucetRunnerConfig{
		Enabled:       true,
		CLIPath:       "circle",
		Chain:         agent.ChainArcTestnet,
		CommandRunner: command,
	})
	router := newFaucetRouter(walletRegistry, runner)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/agent/wallets/agent_test_1/faucet",
		bytes.NewBufferString(`{"address":"0xattackercccccccccccccccccccccccccccccccc","chain":"ETHEREUM","token":"weth"}`),
	)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	for _, addr := range command.gotArgs {
		if strings.EqualFold(addr, "0xattackercccccccccccccccccccccccccccccccc") {
			t.Fatalf("faucet must not accept attacker-supplied address; got args %#v", command.gotArgs)
		}
	}
	if command.gotArgs[3] != "0x9999999999999999999999999999999999999999" {
		t.Fatalf("expected registered wallet address as --address arg, got %#v", command.gotArgs)
	}
	if command.gotArgs[5] != agent.ChainArcTestnet {
		t.Fatalf("expected ARC-TESTNET as --chain arg, got %#v", command.gotArgs)
	}
	if command.gotArgs[7] != agent.FaucetTokenUSDC {
		t.Fatalf("expected usdc as --token arg, got %#v", command.gotArgs)
	}
}

func TestFaucetParserAcceptsCleanJSON(t *testing.T) {
	output := []byte(`{"data":{"transactionHash":"0xabc"}}`)
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	runner := agent.NewCircleCLIFaucetRunner(agent.CircleCLIFaucetRunnerConfig{
		Enabled:       true,
		CLIPath:       "circle",
		Chain:         agent.ChainArcTestnet,
		CommandRunner: &recordingFaucetEnvCommandRunner{output: output},
	})
	router := newFaucetRouter(walletRegistry, runner)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/wallets/agent_test_1/faucet", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(`"transactionHash":"0xabc"`)) {
		t.Fatalf("expected JSON result to surface, got %s", response.Body.String())
	}
}

func TestFaucetParserStripsNodeWarningPrefix(t *testing.T) {
	output := []byte(`(node:123) [DEP0040] DeprecationWarning: punycode
(node:123) Use 'node --trace-warnings'
{"status":"submitted"}`)
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	runner := agent.NewCircleCLIFaucetRunner(agent.CircleCLIFaucetRunnerConfig{
		Enabled:       true,
		CLIPath:       "circle",
		Chain:         agent.ChainArcTestnet,
		CommandRunner: &recordingFaucetEnvCommandRunner{output: output},
	})
	router := newFaucetRouter(walletRegistry, runner)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/wallets/agent_test_1/faucet", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(`"status":"submitted"`)) {
		t.Fatalf("expected JSON result to surface after warning prefix, got %s", response.Body.String())
	}
}

func TestFaucetParserReturnsSanitizedTextWhenNotJSON(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	runner := agent.NewCircleCLIFaucetRunner(agent.CircleCLIFaucetRunnerConfig{
		Enabled:       true,
		CLIPath:       "circle",
		Chain:         agent.ChainArcTestnet,
		CommandRunner: &recordingFaucetEnvCommandRunner{output: []byte("Faucet request submitted for 0x9999999999999999999999999999999999999999")},
	})
	router := newFaucetRouter(walletRegistry, runner)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/wallets/agent_test_1/faucet", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(`"message"`)) {
		t.Fatalf("expected message field for text-only output, got %s", response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), []byte("0x9999999999999999999999999999999999999999")) {
		// the registered wallet address itself is not secret, but other secrets must be redacted.
		// Verify text-only path still returns text content.
		t.Logf("note: address present in sanitized text is acceptable since it is the registered address")
	}
}

func TestFaucetParserEmptyOutputIsTreatedAsFailure(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	runner := agent.NewCircleCLIFaucetRunner(agent.CircleCLIFaucetRunnerConfig{
		Enabled:       true,
		CLIPath:       "circle",
		Chain:         agent.ChainArcTestnet,
		CommandRunner: &recordingFaucetEnvCommandRunner{output: []byte("")},
	})
	router := newFaucetRouter(walletRegistry, runner)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/wallets/agent_test_1/faucet", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadGateway, response.Code, response.Body.String())
	}
}
