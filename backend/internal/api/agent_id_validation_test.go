package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/wahyu241205/SignalArc/backend/internal/agent"
)

// TestStartAgentOnboardingRejectsGenericAgentIDValues asserts that the
// SignalArc Custom GPT integration cannot start onboarding with any of the
// documented generic placeholder agent_id values that have been observed to
// collide across users.
func TestStartAgentOnboardingRejectsGenericAgentIDValues(t *testing.T) {
	cases := []string{
		"signalarc-gpt-agent",
		"signalarc_gpt_agent",
		"agent_desi_001",
		"default",
		"test",
		"demo",
		"user",
		"agent",
		"chatgpt",
		"DEFAULT",
		"SignalArc-GPT-Agent",
	}

	for _, agentID := range cases {
		agentID := agentID
		t.Run(agentID, func(t *testing.T) {
			sessionRegistry := newTestAgentSessionRegistry()
			router := chi.NewRouter()
			registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry)

			body := `{"agent_id":"` + agentID + `","user_email":"sanatarau21@gmail.com"}`
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/start", bytes.NewBufferString(body))
			router.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d for agent_id %q: %s", http.StatusBadRequest, response.Code, agentID, response.Body.String())
			}
			if !bytes.Contains(response.Body.Bytes(), []byte("agent_onboarding_invalid")) {
				t.Fatalf("expected agent_onboarding_invalid for agent_id %q, got %s", agentID, response.Body.String())
			}
		})
	}
}

// TestStartAgentOnboardingRejectsShortAgentID asserts that very short
// agent_id values are rejected even when they technically start with agent_.
func TestStartAgentOnboardingRejectsShortAgentID(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/start", bytes.NewBufferString(`{
		"agent_id": "agent_a",
		"user_email": "sanatarau21@gmail.com"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("at least 10")) {
		t.Fatalf("expected length-related error message, got %s", response.Body.String())
	}
}

// TestStartAgentOnboardingAcceptsRecommendedAgentIDShape asserts that the
// recommended SignalArc agent_id shape used by the Custom GPT is accepted by
// the onboarding/start handler.
func TestStartAgentOnboardingAcceptsRecommendedAgentIDShape(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/start", bytes.NewBufferString(`{
		"agent_id": "agent_sanatarau21_chatgpt_001",
		"user_email": "sanatarau21@gmail.com",
		"source_client": "chatgpt_custom_action",
		"channel": "chatgpt"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
}

// TestRegisterAgentWalletRejectsGenericAgentID asserts that the explicit
// /agent/wallets registration path also rejects generic placeholder agent_id
// values, so the validation cannot be bypassed.
func TestRegisterAgentWalletRejectsGenericAgentID(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/wallets", bytes.NewBufferString(`{
		"agent_id": "signalarc-gpt-agent",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"agent_wallet_address": "0x9999999999999999999999999999999999999999",
		"wallet_provider": "circle_agent_wallet",
		"chain": "ARC-TESTNET",
		"allowed_actions": ["create_market"],
		"status": "active"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("agent_wallet_invalid")) {
		t.Fatalf("expected agent_wallet_invalid, got %s", response.Body.String())
	}
}

// TestCreateAgentIntentRejectsInvalidJSON asserts that the create intent
// handler returns a stable 400 invalid_json error with a guiding message
// when the request body is not valid JSON, including when natural-language
// dates such as "default" are sent for close_timestamp.
func TestCreateAgentIntentRejectsInvalidJSON(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents", bytes.NewBufferString(`{
		"action": "create_market",
		"close_timestamp": end of May 2026
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("invalid_json")) {
		t.Fatalf("expected invalid_json code, got %s", response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("RFC3339")) {
		t.Fatalf("expected guidance about RFC3339, got %s", response.Body.String())
	}
}

// TestCreateAgentIntentRejectsNonRFC3339CloseTimestamp asserts that a
// create_market intent body with valid JSON shape but a non-RFC3339,
// non-integer close_timestamp value is rejected with a stable 400 validation
// error after the body decodes successfully.
func TestCreateAgentIntentRejectsNonRFC3339CloseTimestamp(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents", bytes.NewBufferString(`{
		"action": "create_market",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "agent-market-1",
		"question": "Will BTC be above 50k in May 2026?",
		"close_timestamp": "default",
		"resolver": "0x2222222222222222222222222222222222222222",
		"collateral_token": "0x3333333333333333333333333333333333333333"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("close_timestamp")) {
		t.Fatalf("expected close_timestamp error, got %s", response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("RFC3339")) {
		t.Fatalf("expected RFC3339 guidance, got %s", response.Body.String())
	}
}

// TestCreateAgentIntentAcceptsRFC3339CloseTimestamp asserts that the create
// intent handler accepts an RFC3339 close_timestamp and normalizes it to the
// unix-seconds string form expected by the SignalArc executor and the
// SignalArcAgentMarketFactory.createMarket(uint256) signature.
func TestCreateAgentIntentAcceptsRFC3339CloseTimestamp(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents", bytes.NewBufferString(`{
		"action": "create_market",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "agent-market-rfc3339",
		"question": "Will BTC be above 50k in May 2026?",
		"close_timestamp": "2026-05-31T23:59:00Z",
		"resolver": "0x2222222222222222222222222222222222222222",
		"collateral_token": "0x3333333333333333333333333333333333333333"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	body := response.Body.String()
	if !strings.Contains(body, `"close_timestamp":"1780271940"`) {
		t.Fatalf("expected normalized unix-seconds close_timestamp in response, got %s", body)
	}
}

// TestCreateAgentIntentAcceptsUnixSecondsCloseTimestamp asserts that
// existing unix-seconds string clients are not regressed by the new
// normalization helper.
func TestCreateAgentIntentAcceptsUnixSecondsCloseTimestamp(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents", bytes.NewBufferString(`{
		"action": "create_market",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "agent-market-int",
		"question": "Will BTC be above 50k in May 2026?",
		"close_timestamp": "1780271940",
		"resolver": "0x2222222222222222222222222222222222222222",
		"collateral_token": "0x3333333333333333333333333333333333333333"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
}

// TestValidateAgentIDDirect runs the underlying helper for completeness.
func TestValidateAgentIDDirect(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid_chatgpt_shape", input: "agent_sanatarau21_chatgpt_001", wantErr: false},
		{name: "valid_live_shape", input: "agent_adenhusen65_live_002", wantErr: false},
		{name: "empty", input: "", wantErr: true},
		{name: "blocklist_signalarc_dash", input: "signalarc-gpt-agent", wantErr: true},
		{name: "blocklist_signalarc_underscore", input: "signalarc_gpt_agent", wantErr: true},
		{name: "blocklist_agent_desi_001", input: "agent_desi_001", wantErr: true},
		{name: "blocklist_default", input: "default", wantErr: true},
		{name: "blocklist_test", input: "test", wantErr: true},
		{name: "blocklist_demo", input: "demo", wantErr: true},
		{name: "blocklist_user", input: "user", wantErr: true},
		{name: "blocklist_chatgpt", input: "chatgpt", wantErr: true},
		{name: "blocklist_uppercase", input: "DEFAULT", wantErr: true},
		{name: "too_short", input: "agent_a", wantErr: true},
		{name: "wrong_prefix", input: "user_sanatarau21_chatgpt_001", wantErr: true},
		{name: "trailing_underscore", input: "agent_sanatarau_", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, errs := validateAgentID(tc.input)
			if tc.wantErr && len(errs) == 0 {
				t.Fatalf("expected validation error for %q", tc.input)
			}
			if !tc.wantErr && len(errs) != 0 {
				t.Fatalf("unexpected validation error %v for %q", errs, tc.input)
			}
		})
	}
}
