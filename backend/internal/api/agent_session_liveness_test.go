package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wahyu241205/SignalArc/backend/internal/agent"
	"github.com/wahyu241205/SignalArc/backend/internal/repository"
)

// stubLivenessResolver implements both CircleWalletResolver and the optional
// CircleAgentSessionLivenessChecker so the GET /agent/sessions handler can
// downgrade DB-backed "active" status when the local CLI session is missing.
type stubLivenessResolver struct {
	wallet           agent.CircleAgentWallet
	balances         agent.CircleAgentWalletBalances
	balanceErr       error
	resolveErr       error
	livenessResult   agent.AgentSessionLivenessResult
	livenessAddress  string
	livenessCalled   bool
	livenessOverride bool
}

func (resolver *stubLivenessResolver) ResolveAgentWallet(_ context.Context, _ string) (agent.CircleAgentWallet, error) {
	return resolver.wallet, resolver.resolveErr
}

func (resolver *stubLivenessResolver) GetAgentWalletBalances(_ context.Context, _ string) (agent.CircleAgentWalletBalances, error) {
	return resolver.balances, resolver.balanceErr
}

func (resolver *stubLivenessResolver) CheckAgentSessionLiveness(_ context.Context, address string) agent.AgentSessionLivenessResult {
	resolver.livenessCalled = true
	resolver.livenessAddress = address
	if resolver.livenessOverride {
		return resolver.livenessResult
	}
	return agent.AgentSessionLivenessResult{State: agent.AgentSessionLivenessLive}
}

func newAgentSessionsRouter(sessionRegistry *testAgentSessionRegistry, walletRegistry *testAgentWalletRegistry, resolver agent.CircleWalletResolver) http.Handler {
	router := chi.NewRouter()
	registerAgentIntentRoutes(
		router,
		agent.NewStore(),
		walletRegistry,
		nil,
		sessionRegistry,
		agent.CircleOnboardingStarter{},
		resolver,
	)
	return router
}

func insertActiveAgentSession(t *testing.T, sessionRegistry *testAgentSessionRegistry, agentID string, agentWallet string) {
	t.Helper()
	_, err := sessionRegistry.CreateAgentSession(context.Background(), repository.CreateAgentSessionInput{
		SessionID:          "agent_session_test_1",
		AgentID:            agentID,
		UserEmail:          "desi@example.com",
		UserWallet:         "0x1111111111111111111111111111111111111111",
		AgentWalletAddress: agentWallet,
		WalletProvider:     agent.WalletProviderCircleAgentWallet,
		Chain:              agent.ChainArcTestnet,
		Status:             repository.AgentSessionStatusActive,
		AllowedActions:     []string{agent.ActionCreateMarket},
		AllowedChannels:    []string{"chatgpt"},
		SessionMetadata:    json.RawMessage(`{"note":"test"}`),
	})
	if err != nil {
		t.Fatalf("create test agent session: %v", err)
	}
}

func TestGetAgentSessionDowngradesWhenLivenessReportsAuthRequired(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	walletRegistry := newTestAgentWalletRegistry()
	insertActiveAgentSession(t, sessionRegistry, "agent_test_liveness_auth", "0xa9914bca9123ba0079be8c968f632c0db6400fe7")

	resolver := &stubLivenessResolver{
		livenessOverride: true,
		livenessResult: agent.AgentSessionLivenessResult{
			State:      agent.AgentSessionLivenessAuthRequired,
			ErrorClass: agent.CircleErrorClassAuthRequired,
			Reason:     "Circle CLI agent session is not active on this backend instance; OTP onboarding must be re-run",
		},
	}
	router := newAgentSessionsRouter(sessionRegistry, walletRegistry, resolver)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/sessions/agent_test_liveness_auth", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !resolver.livenessCalled {
		t.Fatal("expected CLI liveness probe to run")
	}
	if resolver.livenessAddress != "0xa9914bca9123ba0079be8c968f632c0db6400fe7" {
		t.Fatalf("liveness probe used wrong address: %q", resolver.livenessAddress)
	}

	var body struct {
		AgentSession agentSessionResponse `json:"agent_session"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.AgentSession.Status != "cli_session_unavailable" {
		t.Fatalf("expected cli_session_unavailable, got %q", body.AgentSession.Status)
	}
	if body.AgentSession.LivenessReason == "" {
		t.Fatal("expected sanitized liveness reason in response")
	}
	for _, secret := range []string{"desi@example.com", "AUTH_REQUIRED"} {
		if bytes.Contains(response.Body.Bytes(), []byte(secret)) {
			t.Fatalf("response leaked sensitive content %q: %s", secret, response.Body.String())
		}
	}
}

func TestGetAgentSessionDowngradesToUnknownOnUnclassifiedFailure(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	walletRegistry := newTestAgentWalletRegistry()
	insertActiveAgentSession(t, sessionRegistry, "agent_test_liveness_unknown", "0xa9914bca9123ba0079be8c968f632c0db6400fe7")

	resolver := &stubLivenessResolver{
		livenessOverride: true,
		livenessResult: agent.AgentSessionLivenessResult{
			State:      agent.AgentSessionLivenessUnknown,
			ErrorClass: agent.CircleErrorClassUnknown,
			Reason:     "Circle CLI agent wallet liveness probe failed",
		},
	}
	router := newAgentSessionsRouter(sessionRegistry, walletRegistry, resolver)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/sessions/agent_test_liveness_unknown", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var body struct {
		AgentSession agentSessionResponse `json:"agent_session"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.AgentSession.Status != "cli_session_unknown" {
		t.Fatalf("expected cli_session_unknown, got %q", body.AgentSession.Status)
	}
}

func TestGetAgentSessionStaysActiveWhenLivenessLive(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	walletRegistry := newTestAgentWalletRegistry()
	insertActiveAgentSession(t, sessionRegistry, "agent_test_liveness_live", "0xa9914bca9123ba0079be8c968f632c0db6400fe7")

	resolver := &stubLivenessResolver{}
	router := newAgentSessionsRouter(sessionRegistry, walletRegistry, resolver)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/sessions/agent_test_liveness_live", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !resolver.livenessCalled {
		t.Fatal("expected liveness probe to run")
	}

	var body struct {
		AgentSession agentSessionResponse `json:"agent_session"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.AgentSession.Status != repository.AgentSessionStatusActive {
		t.Fatalf("expected active status, got %q", body.AgentSession.Status)
	}
	if body.AgentSession.LivenessReason != "" {
		t.Fatalf("expected empty liveness_reason for live status, got %q", body.AgentSession.LivenessReason)
	}
}

func TestGetAgentSessionWithoutLivenessCheckerKeepsActiveStatus(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	walletRegistry := newTestAgentWalletRegistry()
	insertActiveAgentSession(t, sessionRegistry, "agent_test_liveness_no_checker", "0xa9914bca9123ba0079be8c968f632c0db6400fe7")

	router := newAgentSessionsRouter(sessionRegistry, walletRegistry, &stubCircleWalletResolver{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/sessions/agent_test_liveness_no_checker", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var body struct {
		AgentSession agentSessionResponse `json:"agent_session"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.AgentSession.Status != repository.AgentSessionStatusActive {
		t.Fatalf("expected active status when no liveness checker is wired, got %q", body.AgentSession.Status)
	}
}

func TestGetAgentSessionDoesNotProbeForNonCircleProvider(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	walletRegistry := newTestAgentWalletRegistry()
	_, err := sessionRegistry.CreateAgentSession(context.Background(), repository.CreateAgentSessionInput{
		SessionID:          "agent_session_non_circle",
		AgentID:            "agent_test_non_circle",
		UserEmail:          "desi@example.com",
		UserWallet:         "0x1111111111111111111111111111111111111111",
		AgentWalletAddress: "0xa9914bca9123ba0079be8c968f632c0db6400fe7",
		WalletProvider:     agent.WalletProviderTemporaryTestnetAgentEOA,
		Chain:              agent.ChainArcTestnet,
		Status:             repository.AgentSessionStatusActive,
		AllowedActions:     []string{agent.ActionCreateMarket},
		AllowedChannels:    []string{"chatgpt"},
		SessionMetadata:    json.RawMessage(`{"note":"test"}`),
	})
	if err != nil {
		t.Fatalf("create non-circle session: %v", err)
	}

	resolver := &stubLivenessResolver{
		livenessOverride: true,
		livenessResult: agent.AgentSessionLivenessResult{
			State:      agent.AgentSessionLivenessAuthRequired,
			ErrorClass: agent.CircleErrorClassAuthRequired,
			Reason:     "should not be returned",
		},
	}
	router := newAgentSessionsRouter(sessionRegistry, walletRegistry, resolver)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/sessions/agent_test_non_circle", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if resolver.livenessCalled {
		t.Fatal("liveness probe must not run for non-Circle wallet providers")
	}
}

func TestBalanceFailureLogsClassifiedSummary(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	wrapped := &agent.CircleCLIError{
		Operation:        "circle_agent_wallet_balance",
		ErrorClass:       agent.CircleErrorClassAuthRequired,
		SanitizedSummary: "exit status 1",
		Err:              agent.ErrCircleAgentWalletBalanceFailed,
	}
	resolver := &stubLivenessResolver{balanceErr: wrapped}
	router := newAgentSessionsRouter(newTestAgentSessionRegistry(), walletRegistry, resolver)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/wallets/agent_test_1/balance", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadGateway, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("circle_agent_wallet_balance_failed")) {
		t.Fatalf("expected preserved public error code, got %s", response.Body.String())
	}
	if !errors.Is(wrapped, agent.ErrCircleAgentWalletBalanceFailed) {
		t.Fatal("classified error must still satisfy the public sentinel")
	}
}

// _ keeps time import alive for parity with similar tests in this package.
var _ = time.Second
