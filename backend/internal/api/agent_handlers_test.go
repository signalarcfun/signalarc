package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wahyu241205/SignalArc/backend/internal/agent"
	"github.com/wahyu241205/SignalArc/backend/internal/repository"
)

type stubAgentExecutor struct {
	result                agent.ExecutionResult
	err                   error
	requiredAllowedAction string
	intent                agent.Intent
	called                bool
}

type stubCircleOnboardingRunner struct {
	result       agent.CircleOTPStartResult
	err          error
	email        string
	requestID    string
	otp          string
	startCalled  bool
	verifyCalled bool
}

type stubCircleWalletResolver struct {
	wallet         agent.CircleAgentWallet
	balances       agent.CircleAgentWalletBalances
	err            error
	balanceErr     error
	resolveCalled  bool
	balanceCalled  bool
	email          string
	balanceAddress string
}

type stubCircleBalanceReader struct {
	balances agent.CircleAgentWalletBalances
	err      error
	request  agent.CircleAgentWalletBalanceRequest
	called   bool
}

func (runner *stubCircleOnboardingRunner) StartOTP(_ context.Context, email string) (agent.CircleOTPStartResult, error) {
	runner.startCalled = true
	runner.email = email
	if runner.err != nil {
		return agent.CircleOTPStartResult{}, runner.err
	}
	return runner.result, nil
}

func (runner *stubCircleOnboardingRunner) VerifyOTP(_ context.Context, requestID string, otp string) error {
	runner.verifyCalled = true
	runner.requestID = requestID
	runner.otp = otp
	if runner.err != nil {
		return runner.err
	}
	return nil
}

func (resolver *stubCircleWalletResolver) ResolveAgentWallet(_ context.Context, email string) (agent.CircleAgentWallet, error) {
	resolver.resolveCalled = true
	resolver.email = email
	if resolver.err != nil {
		return agent.CircleAgentWallet{}, resolver.err
	}
	return resolver.wallet, nil
}

func (resolver *stubCircleWalletResolver) GetAgentWalletBalances(_ context.Context, address string) (agent.CircleAgentWalletBalances, error) {
	resolver.balanceCalled = true
	resolver.balanceAddress = address
	if resolver.balanceErr != nil {
		return agent.CircleAgentWalletBalances{}, resolver.balanceErr
	}
	return resolver.balances, nil
}

func (reader *stubCircleBalanceReader) GetAgentWalletBalances(_ context.Context, request agent.CircleAgentWalletBalanceRequest) (agent.CircleAgentWalletBalances, error) {
	reader.called = true
	reader.request = request
	if reader.err != nil {
		return agent.CircleAgentWalletBalances{}, reader.err
	}
	return reader.balances, nil
}

type testEnvCommandRunner struct {
	output []byte
	err    error
}

func (runner testEnvCommandRunner) RunWithEnv(_ context.Context, _ string, _ []string, _ []string) ([]byte, error) {
	return runner.output, runner.err
}

func (executor *stubAgentExecutor) ExecuteCreateMarket(_ context.Context, intent agent.Intent) (agent.ExecutionResult, error) {
	executor.called = true
	executor.intent = intent
	if executor.err != nil {
		return agent.ExecutionResult{}, executor.err
	}
	return executor.result, nil
}

func (executor *stubAgentExecutor) ExecuteBuyYes(_ context.Context, intent agent.Intent) (agent.ExecutionResult, error) {
	executor.called = true
	executor.intent = intent
	if executor.requiredAllowedAction != "" && !agent.AgentWalletAllowsAction(agent.AgentWallet{AllowedActions: intent.AllowedActions}, executor.requiredAllowedAction) {
		return agent.ExecutionResult{}, agent.ErrIntentInvalid
	}
	if executor.err != nil {
		return agent.ExecutionResult{}, executor.err
	}
	return executor.result, nil
}

func (executor *stubAgentExecutor) ExecuteBuyNo(_ context.Context, intent agent.Intent) (agent.ExecutionResult, error) {
	executor.called = true
	executor.intent = intent
	if executor.err != nil {
		return agent.ExecutionResult{}, executor.err
	}
	return executor.result, nil
}

func (executor *stubAgentExecutor) ExecuteCloseMarket(_ context.Context, intent agent.Intent) (agent.ExecutionResult, error) {
	executor.called = true
	executor.intent = intent
	if executor.err != nil {
		return agent.ExecutionResult{}, executor.err
	}
	return executor.result, nil
}

func (executor *stubAgentExecutor) ExecuteResolveMarket(_ context.Context, intent agent.Intent) (agent.ExecutionResult, error) {
	executor.called = true
	executor.intent = intent
	if executor.err != nil {
		return agent.ExecutionResult{}, executor.err
	}
	return executor.result, nil
}

func (executor *stubAgentExecutor) ExecuteClaimPayout(_ context.Context, intent agent.Intent) (agent.ExecutionResult, error) {
	executor.called = true
	executor.intent = intent
	if executor.err != nil {
		return agent.ExecutionResult{}, executor.err
	}
	return executor.result, nil
}

func (executor *stubAgentExecutor) ExecuteCancelMarket(_ context.Context, intent agent.Intent) (agent.ExecutionResult, error) {
	executor.called = true
	executor.intent = intent
	if executor.err != nil {
		return agent.ExecutionResult{}, executor.err
	}
	return executor.result, nil
}

func (executor *stubAgentExecutor) ExecuteClaimRefund(_ context.Context, intent agent.Intent) (agent.ExecutionResult, error) {
	executor.called = true
	executor.intent = intent
	if executor.err != nil {
		return agent.ExecutionResult{}, executor.err
	}
	return executor.result, nil
}

type testAgentWalletRegistry struct {
	wallets map[string]repository.AgentWallet
}

type testAgentSessionRegistry struct {
	onboardingSessions map[string]repository.AgentOnboardingSession
	agentSessions      map[string]repository.AgentSession
	sessionsByID       map[string]repository.AgentSession
	failStatusUpdate   bool
}

func TestHydrateAgentIntentForExecutionIncludesPolicyMetadata(t *testing.T) {
	wallet := repository.AgentWallet{
		AgentID:            "agent_test",
		AgentWalletAddress: "0x9999999999999999999999999999999999999999",
		WalletProvider:     agent.WalletProviderCircleAgentWallet,
		AllowedActions:     []string{agent.ActionBuyYes},
		PolicyMetadata:     json.RawMessage(`{"circle_wallet_id":"wallet-1","max_trade_amount":"10"}`),
	}

	intent := hydrateAgentIntentForExecution(agent.Intent{AgentID: "agent_test"}, wallet)
	if intent.PolicyMetadata["circle_wallet_id"] != "wallet-1" {
		t.Fatalf("expected circle_wallet_id hydration, got %#v", intent.PolicyMetadata)
	}
	if intent.PolicyMetadata["max_trade_amount"] != "10" {
		t.Fatalf("expected max_trade_amount hydration, got %#v", intent.PolicyMetadata)
	}
}

type testDurableAgentIntentRegistry struct {
	intents    map[string]repository.AgentIntent
	executions map[string]repository.AgentExecution
	nextExecID int
}

func newTestAgentWalletRegistry() *testAgentWalletRegistry {
	return &testAgentWalletRegistry{wallets: map[string]repository.AgentWallet{}}
}

func newTestDurableAgentIntentRegistry() *testDurableAgentIntentRegistry {
	return &testDurableAgentIntentRegistry{
		intents:    map[string]repository.AgentIntent{},
		executions: map[string]repository.AgentExecution{},
	}
}

func newTestAgentSessionRegistry() *testAgentSessionRegistry {
	return &testAgentSessionRegistry{
		onboardingSessions: map[string]repository.AgentOnboardingSession{},
		agentSessions:      map[string]repository.AgentSession{},
		sessionsByID:       map[string]repository.AgentSession{},
	}
}

func (registry *testAgentWalletRegistry) RegisterAgentWallet(_ context.Context, input repository.UpsertAgentWalletInput) (repository.AgentWallet, error) {
	now := time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)
	wallet := repository.AgentWallet{
		ID:                 "agent_wallet_test_1",
		AgentID:            input.AgentID,
		UserWallet:         sql.NullString{String: input.UserWallet, Valid: strings.TrimSpace(input.UserWallet) != ""},
		UserEmail:          input.UserEmail,
		AgentWalletAddress: input.AgentWalletAddress,
		WalletProvider:     input.WalletProvider,
		Chain:              input.Chain,
		Status:             input.Status,
		AllowedActions:     input.AllowedActions,
		PolicyMetadata:     input.PolicyMetadata,
		SourceClient:       input.SourceClient,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	registry.wallets[wallet.AgentID] = wallet
	return wallet, nil
}

func (registry *testAgentWalletRegistry) GetAgentWalletByAgentID(_ context.Context, agentID string) (repository.AgentWallet, error) {
	wallet, ok := registry.wallets[agentID]
	if !ok {
		return repository.AgentWallet{}, sql.ErrNoRows
	}
	return wallet, nil
}

func (registry *testAgentWalletRegistry) DisableAgentWallet(_ context.Context, agentID string) (repository.AgentWallet, error) {
	wallet, ok := registry.wallets[agentID]
	if !ok {
		return repository.AgentWallet{}, sql.ErrNoRows
	}
	wallet.Status = "disabled"
	wallet.UpdatedAt = wallet.UpdatedAt.Add(time.Minute)
	registry.wallets[agentID] = wallet
	return wallet, nil
}

func (registry *testAgentSessionRegistry) CreateAgentOnboardingSession(_ context.Context, input repository.CreateAgentOnboardingSessionInput) (repository.AgentOnboardingSession, error) {
	now := time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC)
	session := repository.AgentOnboardingSession{
		ID:                          "agent_onboarding_row_1",
		OnboardingID:                input.OnboardingID,
		AgentID:                     input.AgentID,
		UserEmail:                   input.UserEmail,
		UserWallet:                  input.UserWallet,
		RequestedAgentWalletAddress: input.RequestedAgentWalletAddress,
		SourceClient:                input.SourceClient,
		Channel:                     input.Channel,
		Chain:                       input.Chain,
		WalletProvider:              input.WalletProvider,
		Status:                      input.Status,
		CircleRequestIDHash:         input.CircleRequestIDHash,
		CircleRequestExpiresAt:      input.CircleRequestExpiresAt,
		FailureReason:               input.FailureReason,
		PolicyMetadata:              input.PolicyMetadata,
		CreatedAt:                   now,
		UpdatedAt:                   now,
	}
	registry.onboardingSessions[session.OnboardingID] = session
	return session, nil
}

func (registry *testAgentSessionRegistry) GetAgentOnboardingSessionByOnboardingID(_ context.Context, onboardingID string) (repository.AgentOnboardingSession, error) {
	session, ok := registry.onboardingSessions[onboardingID]
	if !ok {
		return repository.AgentOnboardingSession{}, sql.ErrNoRows
	}
	return session, nil
}

func (registry *testAgentSessionRegistry) UpdateAgentOnboardingSessionStatus(_ context.Context, onboardingID string, status string, failureReason sql.NullString) (repository.AgentOnboardingSession, error) {
	if registry.failStatusUpdate {
		return repository.AgentOnboardingSession{}, errors.New("status update failed")
	}
	session, ok := registry.onboardingSessions[onboardingID]
	if !ok {
		return repository.AgentOnboardingSession{}, sql.ErrNoRows
	}
	session.Status = status
	session.FailureReason = failureReason
	session.UpdatedAt = session.UpdatedAt.Add(time.Minute)
	registry.onboardingSessions[onboardingID] = session
	return session, nil
}

func (registry *testAgentSessionRegistry) UpdateAgentOnboardingSessionOTPStart(_ context.Context, onboardingID string, requestIDHash string, expiresAt time.Time) (repository.AgentOnboardingSession, error) {
	session, ok := registry.onboardingSessions[onboardingID]
	if !ok {
		return repository.AgentOnboardingSession{}, sql.ErrNoRows
	}
	session.CircleRequestIDHash = sql.NullString{String: requestIDHash, Valid: requestIDHash != ""}
	session.CircleRequestExpiresAt = sql.NullTime{Time: expiresAt, Valid: !expiresAt.IsZero()}
	session.UpdatedAt = session.UpdatedAt.Add(time.Minute)
	registry.onboardingSessions[onboardingID] = session
	return session, nil
}

func (registry *testAgentSessionRegistry) CreateAgentSession(_ context.Context, input repository.CreateAgentSessionInput) (repository.AgentSession, error) {
	now := time.Date(2026, 5, 22, 1, 0, 0, 0, time.UTC)
	session := repository.AgentSession{
		ID:                 "agent_session_row_1",
		SessionID:          input.SessionID,
		AgentID:            input.AgentID,
		UserEmail:          input.UserEmail,
		UserWallet:         input.UserWallet,
		AgentWalletAddress: input.AgentWalletAddress,
		WalletProvider:     input.WalletProvider,
		Chain:              input.Chain,
		Status:             input.Status,
		AllowedActions:     input.AllowedActions,
		AllowedChannels:    input.AllowedChannels,
		SessionMetadata:    input.SessionMetadata,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	registry.agentSessions[session.AgentID] = session
	registry.sessionsByID[session.SessionID] = session
	return session, nil
}

func (registry *testAgentSessionRegistry) GetAgentSessionByAgentID(_ context.Context, agentID string) (repository.AgentSession, error) {
	session, ok := registry.agentSessions[agentID]
	if !ok {
		return repository.AgentSession{}, sql.ErrNoRows
	}
	return session, nil
}

func (registry *testAgentSessionRegistry) GetAgentSessionBySessionID(_ context.Context, sessionID string) (repository.AgentSession, error) {
	session, ok := registry.sessionsByID[sessionID]
	if !ok {
		return repository.AgentSession{}, sql.ErrNoRows
	}
	return session, nil
}

func (registry *testDurableAgentIntentRegistry) CreateAgentIntent(_ context.Context, input repository.CreateAgentIntentInput) (repository.AgentIntent, error) {
	for _, intent := range registry.intents {
		if intent.AgentID.String == input.AgentID && intent.SourceClient.String == input.SourceClient && intent.ClientRequestID.String == input.ClientRequestID && input.AgentID != "" && input.SourceClient != "" && input.ClientRequestID != "" {
			return intent, nil
		}
	}
	now := time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)
	intent := repository.AgentIntent{
		ID:                    "agent_intent_row_" + input.IntentID,
		IntentID:              input.IntentID,
		AgentID:               nullableString(input.AgentID),
		AgentWalletAddress:    nullableString(input.AgentWalletAddress),
		WalletProvider:        nullableString(input.WalletProvider),
		SourceClient:          nullableString(input.SourceClient),
		ClientRequestID:       nullableString(input.ClientRequestID),
		Action:                input.Action,
		Status:                input.Status,
		RequiresConfirmation:  input.RequiresConfirmation,
		UserWallet:            nullableString(input.UserWallet),
		MarketID:              nullableString(input.MarketID),
		MarketContractAddress: nullableString(input.MarketContractAddress),
		Amount:                nullableString(input.Amount),
		Outcome:               nullableString(input.Outcome),
		Resolver:              nullableString(input.Resolver),
		CollateralToken:       nullableString(input.CollateralToken),
		CloseTimestamp:        nullableString(input.CloseTimestamp),
		Question:              nullableString(input.Question),
		ValidationResult:      input.ValidationResult,
		Warnings:              input.Warnings,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	registry.intents[intent.IntentID] = intent
	return intent, nil
}

func (registry *testDurableAgentIntentRegistry) GetAgentIntentByIntentID(_ context.Context, intentID string) (repository.AgentIntent, error) {
	intent, ok := registry.intents[intentID]
	if !ok {
		return repository.AgentIntent{}, sql.ErrNoRows
	}
	return intent, nil
}

func (registry *testDurableAgentIntentRegistry) ConfirmAgentIntent(_ context.Context, intentID string) (repository.AgentIntent, error) {
	intent, ok := registry.intents[intentID]
	if !ok {
		return repository.AgentIntent{}, sql.ErrNoRows
	}
	intent.Status = agent.StatusConfirmed
	intent.ConfirmedAt = sql.NullTime{Time: intent.UpdatedAt.Add(time.Minute), Valid: true}
	intent.UpdatedAt = intent.ConfirmedAt.Time
	registry.intents[intentID] = intent
	return intent, nil
}

func (registry *testDurableAgentIntentRegistry) MarkAgentIntentExecuted(_ context.Context, intentID string) (repository.AgentIntent, error) {
	return registry.markIntentTerminal(intentID, agent.StatusExecuted)
}

func (registry *testDurableAgentIntentRegistry) MarkAgentIntentFailed(_ context.Context, intentID string) (repository.AgentIntent, error) {
	return registry.markIntentTerminal(intentID, "failed")
}

func (registry *testDurableAgentIntentRegistry) CreateAgentExecution(_ context.Context, input repository.CreateAgentExecutionInput) (repository.AgentExecution, error) {
	registry.nextExecID++
	now := time.Date(2026, 6, 28, 1, 0, 0, 0, time.UTC)
	execution := repository.AgentExecution{
		ID:                    "agent_execution_row_" + strconv.Itoa(registry.nextExecID),
		IntentID:              input.IntentID,
		AgentID:               nullableString(input.AgentID),
		Action:                input.Action,
		Status:                repository.AgentExecutionStatusPending,
		ExecutionMode:         nullableString(input.ExecutionMode),
		Network:               nullableString(input.Network),
		AgentFactoryAddress:   nullableString(input.AgentFactoryAddress),
		MarketContractAddress: nullableString(input.MarketContractAddress),
		Readback:              json.RawMessage(`{}`),
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	registry.executions[execution.ID] = execution
	return execution, nil
}

func (registry *testDurableAgentIntentRegistry) MarkAgentExecutionExecuted(_ context.Context, id string, input repository.CompleteAgentExecutionInput) (repository.AgentExecution, error) {
	execution, ok := registry.executions[id]
	if !ok {
		return repository.AgentExecution{}, sql.ErrNoRows
	}
	execution.Status = repository.AgentExecutionStatusExecuted
	execution.ExecutionMode = nullableString(input.ExecutionMode)
	execution.Network = nullableString(input.Network)
	execution.AgentFactoryAddress = nullableString(input.AgentFactoryAddress)
	execution.MarketContractAddress = nullableString(input.MarketContractAddress)
	execution.ApproveTransactionHash = nullableString(input.ApproveTransactionHash)
	execution.TransactionHash = nullableString(input.TransactionHash)
	execution.BroadcastPerformed = input.BroadcastPerformed
	execution.Readback = input.Readback
	execution.CompletedAt = sql.NullTime{Time: execution.UpdatedAt.Add(time.Minute), Valid: true}
	execution.UpdatedAt = execution.CompletedAt.Time
	registry.executions[id] = execution
	return execution, nil
}

func (registry *testDurableAgentIntentRegistry) MarkAgentExecutionFailed(_ context.Context, id string, input repository.FailAgentExecutionInput) (repository.AgentExecution, error) {
	execution, ok := registry.executions[id]
	if !ok {
		return repository.AgentExecution{}, sql.ErrNoRows
	}
	execution.Status = repository.AgentExecutionStatusFailed
	execution.ErrorCode = nullableString(input.ErrorCode)
	execution.ErrorMessage = nullableString(input.ErrorMessage)
	execution.Readback = input.Readback
	execution.CompletedAt = sql.NullTime{Time: execution.UpdatedAt.Add(time.Minute), Valid: true}
	execution.UpdatedAt = execution.CompletedAt.Time
	registry.executions[id] = execution
	return execution, nil
}

func (registry *testDurableAgentIntentRegistry) ListAgentIntentsByAgentID(_ context.Context, agentID string, _ int) ([]repository.AgentIntent, error) {
	intents := []repository.AgentIntent{}
	for _, intent := range registry.intents {
		if intent.AgentID.String == agentID {
			intents = append(intents, intent)
		}
	}
	return intents, nil
}

func (registry *testDurableAgentIntentRegistry) ListAgentExecutionsByAgentID(_ context.Context, agentID string, _ int) ([]repository.AgentExecution, error) {
	executions := []repository.AgentExecution{}
	for _, execution := range registry.executions {
		if execution.AgentID.String == agentID {
			executions = append(executions, execution)
		}
	}
	return executions, nil
}

func (registry *testDurableAgentIntentRegistry) ListAgentExecutionsByIntentID(_ context.Context, intentID string, _ int) ([]repository.AgentExecution, error) {
	executions := []repository.AgentExecution{}
	for _, execution := range registry.executions {
		if execution.IntentID == intentID {
			executions = append(executions, execution)
		}
	}
	return executions, nil
}

func (registry *testDurableAgentIntentRegistry) markIntentTerminal(intentID string, status string) (repository.AgentIntent, error) {
	intent, ok := registry.intents[intentID]
	if !ok {
		return repository.AgentIntent{}, sql.ErrNoRows
	}
	intent.Status = status
	intent.ExecutedAt = sql.NullTime{Time: intent.UpdatedAt.Add(time.Minute), Valid: true}
	intent.UpdatedAt = intent.ExecutedAt.Time
	registry.intents[intentID] = intent
	return intent, nil
}

func TestCreateAgentIntentPreview(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents", bytes.NewBufferString(`{
		"action": "buy_yes",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "market-1",
		"market_contract_address": "0x3333333333333333333333333333333333333333",
		"amount": "12.5"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var body struct {
		Intent agentIntentResponse `json:"intent"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Intent.IntentID == "" {
		t.Fatal("expected intent_id")
	}
	if body.Intent.Action != "buy_yes" {
		t.Fatalf("expected action buy_yes, got %q", body.Intent.Action)
	}
	if body.Intent.Status != "preview" {
		t.Fatalf("expected preview status, got %q", body.Intent.Status)
	}
	if !body.Intent.RequiresConfirmation {
		t.Fatal("expected requires_confirmation true")
	}
	if body.Intent.UserWallet != "0x1111111111111111111111111111111111111111" {
		t.Fatalf("expected user wallet to be echoed, got %q", body.Intent.UserWallet)
	}
	if !body.Intent.ValidationResult.Valid {
		t.Fatalf("expected valid validation result, got %#v", body.Intent.ValidationResult)
	}
	if len(body.Intent.Warnings) == 0 {
		t.Fatal("expected preview warnings")
	}
}

func TestCreateAgentIntentPersistsToDurableRegistry(t *testing.T) {
	durableRegistry := newTestDurableAgentIntentRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, durableRegistry)

	intentID := createAgentIntent(t, router, `{
		"agent_id": "agent_test_1",
		"source_client": "test_client",
		"client_request_id": "client_req_1",
		"action": "buy_yes",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "market-1",
		"market_contract_address": "0x3333333333333333333333333333333333333333",
		"amount": "12.5"
	}`)

	if _, ok := durableRegistry.intents[intentID]; !ok {
		t.Fatalf("expected durable intent %q", intentID)
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/intents/"+intentID, nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected get status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var body struct {
		Intent agentIntentResponse `json:"intent"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if body.Intent.IntentID != intentID {
		t.Fatalf("expected durable intent id %q, got %q", intentID, body.Intent.IntentID)
	}
}

func TestCreateAgentIntentDuplicateClientRequestReturnsDurableIntent(t *testing.T) {
	durableRegistry := newTestDurableAgentIntentRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, durableRegistry)
	payload := `{
		"agent_id": "agent_test_1",
		"source_client": "test_client",
		"client_request_id": "client_req_duplicate",
		"action": "buy_yes",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "market-1",
		"market_contract_address": "0x3333333333333333333333333333333333333333",
		"amount": "12.5"
	}`

	firstIntentID := createAgentIntent(t, router, payload)
	secondIntentID := createAgentIntent(t, router, payload)

	if secondIntentID != firstIntentID {
		t.Fatalf("expected duplicate idempotency key to return %q, got %q", firstIntentID, secondIntentID)
	}
	if len(durableRegistry.intents) != 1 {
		t.Fatalf("expected one durable intent, got %d", len(durableRegistry.intents))
	}
}

func TestConfirmAgentIntentPersistsDurableStatus(t *testing.T) {
	durableRegistry := newTestDurableAgentIntentRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, durableRegistry)

	intentID := createAgentIntent(t, router, `{
		"agent_id": "agent_test_1",
		"source_client": "test_client",
		"client_request_id": "client_req_confirm",
		"action": "buy_yes",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "market-1",
		"market_contract_address": "0x3333333333333333333333333333333333333333",
		"amount": "12.5"
	}`)

	confirmAgentIntent(t, router, intentID)

	intent := durableRegistry.intents[intentID]
	if intent.Status != agent.StatusConfirmed {
		t.Fatalf("expected durable confirmed status, got %q", intent.Status)
	}
	if !intent.ConfirmedAt.Valid {
		t.Fatal("expected confirmed_at to be set")
	}
}

func TestAgentPortfolioRequiresValidAgentID(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, newTestDurableAgentIntentRegistry())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/portfolio/not-valid", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("agent_id_invalid")) {
		t.Fatalf("expected agent_id_invalid, got %s", response.Body.String())
	}
}

func TestAgentEndpointsRejectGenericAgentID(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, newTestDurableAgentIntentRegistry())

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "create intent",
			method: http.MethodPost,
			path:   "/agent/intents",
			body: `{
				"agent_id": "agent_desi_001",
				"action": "buy_yes",
				"user_wallet": "0x1111111111111111111111111111111111111111",
				"market_id": "market-1",
				"market_contract_address": "0x3333333333333333333333333333333333333333",
				"amount": "12.5"
			}`,
		},
		{name: "get wallet", method: http.MethodGet, path: "/agent/wallets/agent_desi_001"},
		{name: "get balance", method: http.MethodGet, path: "/agent/wallets/agent_desi_001/balance"},
		{name: "request faucet", method: http.MethodPost, path: "/agent/wallets/agent_desi_001/faucet"},
		{name: "portfolio", method: http.MethodGet, path: "/agent/portfolio/agent_desi_001"},
		{name: "activity", method: http.MethodGet, path: "/agent/activity/agent_desi_001"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.path, bytes.NewBufferString(test.body))
			router.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
			}
			if !bytes.Contains(response.Body.Bytes(), []byte("agent_id_invalid")) {
				t.Fatalf("expected agent_id_invalid, got %s", response.Body.String())
			}
		})
	}
}

func TestAgentPortfolioReturnsWalletMetadataAndEmptyArrays(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionBuyYes)
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), walletRegistry, nil, newTestDurableAgentIntentRegistry())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/portfolio/agent_test_1", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var body struct {
		Portfolio agentPortfolioResponse `json:"portfolio"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode portfolio response: %v", err)
	}
	if body.Portfolio.AgentID != "agent_test_1" {
		t.Fatalf("expected agent_test_1, got %q", body.Portfolio.AgentID)
	}
	if body.Portfolio.AgentWalletAddress != "0x9999999999999999999999999999999999999999" {
		t.Fatalf("unexpected agent wallet %q", body.Portfolio.AgentWalletAddress)
	}
	if len(body.Portfolio.Positions) != 0 {
		t.Fatalf("expected empty positions, got %#v", body.Portfolio.Positions)
	}
	if len(body.Portfolio.Settlements) != 0 {
		t.Fatalf("expected empty settlements, got %#v", body.Portfolio.Settlements)
	}
	if len(body.Portfolio.UnavailableFields) == 0 {
		t.Fatal("expected unavailable_fields explaining data limitations")
	}
}

func TestAgentActivityHandlesNoActivity(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionBuyYes)
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), walletRegistry, nil, newTestDurableAgentIntentRegistry())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/activity/agent_test_1", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var body struct {
		Activity agentActivityResponse `json:"activity"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode activity response: %v", err)
	}
	if body.Activity.AgentID != "agent_test_1" {
		t.Fatalf("expected agent_test_1, got %q", body.Activity.AgentID)
	}
	if len(body.Activity.Items) != 0 {
		t.Fatalf("expected no activity items, got %#v", body.Activity.Items)
	}
}

func TestAgentMarketResponseIncludesContractAddress(t *testing.T) {
	responses := newAgentMarketResponses([]repository.Market{
		{
			ID:                    "market-1",
			Title:                 "Will SignalArc keep agent markets readable?",
			Status:                "OPEN",
			CollateralAsset:       "USDC",
			Chain:                 "Arc Testnet",
			ClosesAt:              time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			MarketContractAddress: sql.NullString{String: "0x3333333333333333333333333333333333333333", Valid: true},
		},
	})

	if len(responses) != 1 {
		t.Fatalf("expected one response, got %d", len(responses))
	}
	if responses[0].MarketContractAddress == nil || *responses[0].MarketContractAddress != "0x3333333333333333333333333333333333333333" {
		t.Fatalf("expected market contract address, got %#v", responses[0].MarketContractAddress)
	}
	if responses[0].Title == "" || responses[0].Status == "" {
		t.Fatalf("expected existing compact market fields to remain populated: %#v", responses[0])
	}
}

func TestRegisterAgentWallet(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/wallets", bytes.NewBufferString(`{
		"agent_id": "agent_test_1",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"agent_wallet_address": "0x9999999999999999999999999999999999999999",
		"wallet_provider": "circle_agent_wallet",
		"chain": "ARC-TESTNET",
		"allowed_actions": ["create_market", "buy_yes"],
		"status": "active",
		"policy_metadata": {
			"per_tx_usdc_cap": "unknown / not documented until Circle policy is configured"
		}
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var body struct {
		AgentWallet agentWalletResponse `json:"agent_wallet"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.AgentWallet.AgentID != "agent_test_1" {
		t.Fatalf("expected agent id, got %q", body.AgentWallet.AgentID)
	}
	if body.AgentWallet.AgentWalletAddress != "0x9999999999999999999999999999999999999999" {
		t.Fatalf("unexpected agent wallet address %q", body.AgentWallet.AgentWalletAddress)
	}
	if body.AgentWallet.WalletProvider != agent.WalletProviderCircleAgentWallet {
		t.Fatalf("unexpected wallet provider %q", body.AgentWallet.WalletProvider)
	}
}

func TestStartAgentOnboardingMinimalPayload(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/start", bytes.NewBufferString(`{
		"agent_id": "agent_start_1",
		"user_email": "desi@example.com"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var body struct {
		Onboarding agentOnboardingSessionResponse `json:"onboarding"`
		NextStep   string                         `json:"next_step"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Onboarding.OnboardingID == "" {
		t.Fatal("expected onboarding_id")
	}
	if body.Onboarding.AgentID != "agent_start_1" {
		t.Fatalf("expected agent id, got %q", body.Onboarding.AgentID)
	}
	if body.Onboarding.Status != repository.AgentOnboardingStatusPendingOTP {
		t.Fatalf("expected pending_otp status, got %q", body.Onboarding.Status)
	}
	if body.Onboarding.UserWallet != "" {
		t.Fatalf("expected empty user wallet when omitted, got %q", body.Onboarding.UserWallet)
	}
	if body.NextStep != "circle_otp_verification_not_implemented" {
		t.Fatalf("unexpected next step %q", body.NextStep)
	}
}

func TestStartAgentOnboardingAppliesDefaults(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/start", bytes.NewBufferString(`{
		"agent_id": "agent_start_defaults",
		"user_email": "desi@example.com"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var body struct {
		Onboarding agentOnboardingSessionResponse `json:"onboarding"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Onboarding.Chain != agent.ChainArcTestnet {
		t.Fatalf("expected ARC-TESTNET chain, got %q", body.Onboarding.Chain)
	}
	if body.Onboarding.WalletProvider != agent.WalletProviderCircleAgentWallet {
		t.Fatalf("expected circle agent wallet provider, got %q", body.Onboarding.WalletProvider)
	}
	if body.Onboarding.PolicyMetadata["note"] != "pending Circle Agent Wallet OTP onboarding" {
		t.Fatalf("unexpected policy metadata %#v", body.Onboarding.PolicyMetadata)
	}
}

func TestStartAgentOnboardingPreservesOptionalUserWallet(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/start", bytes.NewBufferString(`{
		"agent_id": "agent_start_with_wallet",
		"user_email": "desi@example.com",
		"user_wallet": "0x1111111111111111111111111111111111111111"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var body struct {
		Onboarding agentOnboardingSessionResponse `json:"onboarding"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Onboarding.UserWallet != "0x1111111111111111111111111111111111111111" {
		t.Fatalf("unexpected user wallet %q", body.Onboarding.UserWallet)
	}
}

func TestStartAgentOnboardingPreservesSourceClientAndChannel(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/start", bytes.NewBufferString(`{
		"agent_id": "agent_start_channel",
		"user_email": "desi@example.com",
		"source_client": "chatgpt_custom_action",
		"channel": "chatgpt"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var body struct {
		Onboarding agentOnboardingSessionResponse `json:"onboarding"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Onboarding.SourceClient != "chatgpt_custom_action" {
		t.Fatalf("unexpected source client %q", body.Onboarding.SourceClient)
	}
	if body.Onboarding.Channel != "chatgpt" {
		t.Fatalf("unexpected channel %q", body.Onboarding.Channel)
	}
}

func TestStartAgentOnboardingRejectsMissingRequiredFields(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/start", bytes.NewBufferString(`{}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestStartAgentOnboardingRejectsMissingAgentID(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/start", bytes.NewBufferString(`{
		"user_email": "desi@example.com"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestStartAgentOnboardingRejectsMissingUserEmail(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/start", bytes.NewBufferString(`{
		"agent_id": "agent_missing_email"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestStartAgentOnboardingDoesNotCallExecutor(t *testing.T) {
	executor := &stubAgentExecutor{}
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), executor, sessionRegistry)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/start", bytes.NewBufferString(`{
		"agent_id": "agent_start_no_execute",
		"user_email": "desi@example.com"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	if executor.called {
		t.Fatal("executor should not be called during onboarding start")
	}
}

func TestStartAgentOnboardingDisabledDoesNotCallRunner(t *testing.T) {
	runner := &stubCircleOnboardingRunner{}
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry, agent.CircleOnboardingStarter{
		Enabled: false,
		Runner:  runner,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/start", bytes.NewBufferString(`{
		"agent_id": "agent_start_disabled",
		"user_email": "desi@example.com"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	if runner.startCalled {
		t.Fatal("runner should not be called while OTP start is disabled")
	}

	var body struct {
		NextStep string `json:"next_step"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.NextStep != "circle_otp_verification_not_implemented" {
		t.Fatalf("unexpected next step %q", body.NextStep)
	}
}

func TestStartAgentOnboardingEnabledCallsRunner(t *testing.T) {
	expiresAt := time.Date(2026, 5, 22, 2, 10, 0, 0, time.UTC)
	runner := &stubCircleOnboardingRunner{
		result: agent.CircleOTPStartResult{
			RequestID: "circle_request_secret_123",
			ExpiresAt: expiresAt,
		},
	}
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry, agent.CircleOnboardingStarter{
		Enabled: true,
		Runner:  runner,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/start", bytes.NewBufferString(`{
		"agent_id": "agent_start_enabled",
		"user_email": "desi@example.com"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	if !runner.startCalled {
		t.Fatal("expected runner to be called")
	}
	if runner.email != "desi@example.com" {
		t.Fatalf("expected runner email, got %q", runner.email)
	}

	var body struct {
		Onboarding       agentOnboardingSessionResponse `json:"onboarding"`
		NextStep         string                         `json:"next_step"`
		ExpiresAt        string                         `json:"expires_at"`
		RequestReference string                         `json:"request_reference"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.NextStep != "circle_otp_required" {
		t.Fatalf("unexpected next step %q", body.NextStep)
	}
	if body.ExpiresAt == "" {
		t.Fatal("expected expires_at")
	}
	if body.RequestReference != body.Onboarding.OnboardingID {
		t.Fatalf("expected onboarding id request reference, got %q", body.RequestReference)
	}
	if bytes.Contains(response.Body.Bytes(), []byte("circle_request_secret_123")) {
		t.Fatal("response must not expose raw request_id")
	}

	stored := sessionRegistry.onboardingSessions[body.Onboarding.OnboardingID]
	if !stored.CircleRequestIDHash.Valid {
		t.Fatal("expected stored request id hash")
	}
	if stored.CircleRequestIDHash.String != agent.HashCircleRequestID("circle_request_secret_123") {
		t.Fatalf("unexpected stored request hash %q", stored.CircleRequestIDHash.String)
	}
	if strings.Contains(stored.CircleRequestIDHash.String, "circle_request_secret_123") {
		t.Fatal("stored request reference must be hashed")
	}
	if !stored.CircleRequestExpiresAt.Valid || !stored.CircleRequestExpiresAt.Time.Equal(expiresAt) {
		t.Fatalf("unexpected stored expiry %#v", stored.CircleRequestExpiresAt)
	}
}

func TestStartAgentOnboardingEnabledAcceptsTextRequestIDOutput(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	requestStore := agent.NewCircleOTPRequestStore()
	onboardingRunner := agent.NewCircleCLIOnboardingRunner(agent.CircleCLIOnboardingRunnerConfig{
		CLIPath: "circle",
		Chain:   agent.ChainArcTestnet,
		CommandRunner: testEnvCommandRunner{
			output: []byte(`{"data":{"message":"OTP code sent to desi@example.com\nPlease run: circle wallet login --request circle_request_secret_text_123 --otp <code>"}}`),
			err:    errors.New("exit status 1 after sending OTP"),
		},
	})
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry, agent.CircleOnboardingStarter{
		Enabled:      true,
		Runner:       onboardingRunner,
		RequestStore: requestStore,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/start", bytes.NewBufferString(`{
		"agent_id": "agent_start_text_output",
		"user_email": "desi@example.com"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var body struct {
		Onboarding       agentOnboardingSessionResponse `json:"onboarding"`
		NextStep         string                         `json:"next_step"`
		RequestReference string                         `json:"request_reference"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Onboarding.OnboardingID == "" {
		t.Fatal("expected onboarding_id")
	}
	if body.NextStep != "circle_otp_required" {
		t.Fatalf("unexpected next step %q", body.NextStep)
	}
	if body.RequestReference != body.Onboarding.OnboardingID {
		t.Fatalf("expected onboarding id request reference, got %q", body.RequestReference)
	}
	if bytes.Contains(response.Body.Bytes(), []byte("circle_request_secret_text_123")) {
		t.Fatalf("response must not expose raw request_id: %s", response.Body.String())
	}

	stored := sessionRegistry.onboardingSessions[body.Onboarding.OnboardingID]
	if stored.CircleRequestIDHash.String != agent.HashCircleRequestID("circle_request_secret_text_123") {
		t.Fatalf("unexpected stored request hash %q", stored.CircleRequestIDHash.String)
	}
	if requestID, ok := requestStore.Get(body.Onboarding.OnboardingID); !ok || requestID != "circle_request_secret_text_123" {
		t.Fatalf("expected in-memory request id for verify, got %q, %v", requestID, ok)
	}
}

func TestStartAgentOnboardingEnabledFailureIsSanitized(t *testing.T) {
	runner := &stubCircleOnboardingRunner{err: errors.New("raw request_id circle_request_secret_123 otp B1X-123456")}
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry, agent.CircleOnboardingStarter{
		Enabled: true,
		Runner:  runner,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/start", bytes.NewBufferString(`{
		"agent_id": "agent_start_failure",
		"user_email": "desi@example.com",
		"user_wallet": "0x1111111111111111111111111111111111111111"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadGateway, response.Code, response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), []byte("circle_request_secret_123")) || bytes.Contains(response.Body.Bytes(), []byte("B1X-123456")) {
		t.Fatalf("response should not expose raw request_id or OTP: %s", response.Body.String())
	}
}

func TestVerifyAgentOnboardingDisabledReturnsNotEnabled(t *testing.T) {
	runner := &stubCircleOnboardingRunner{}
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry, agent.CircleOnboardingStarter{
		Enabled: false,
		Runner:  runner,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/verify", bytes.NewBufferString(`{
		"onboarding_id": "agent_onboarding_missing",
		"otp": "B1X-123456"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotImplemented {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotImplemented, response.Code, response.Body.String())
	}
	if runner.verifyCalled {
		t.Fatal("verifier should not be called while disabled")
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("circle_otp_verify_not_enabled")) {
		t.Fatalf("expected not-enabled error code, got %s", response.Body.String())
	}
}

func TestVerifyAgentOnboardingRejectsMissingOnboardingID(t *testing.T) {
	router := newVerifyEnabledRouter(newTestAgentSessionRegistry(), &stubCircleOnboardingRunner{}, agent.NewCircleOTPRequestStore())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/verify", bytes.NewBufferString(`{"otp":"B1X-123456"}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestVerifyAgentOnboardingRejectsMissingOTP(t *testing.T) {
	router := newVerifyEnabledRouter(newTestAgentSessionRegistry(), &stubCircleOnboardingRunner{}, agent.NewCircleOTPRequestStore())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/verify", bytes.NewBufferString(`{"onboarding_id":"agent_onboarding_1"}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestVerifyAgentOnboardingUnknownIDReturnsNotFound(t *testing.T) {
	router := newVerifyEnabledRouter(newTestAgentSessionRegistry(), &stubCircleOnboardingRunner{}, agent.NewCircleOTPRequestStore())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/verify", bytes.NewBufferString(`{
		"onboarding_id": "agent_onboarding_missing",
		"otp": "B1X-123456"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, response.Code, response.Body.String())
	}
}

func TestVerifyAgentOnboardingRejectsNonPendingStatus(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	onboarding := insertTestOnboardingSession(sessionRegistry, "agent_onboarding_verified")
	onboarding.Status = repository.AgentOnboardingStatusVerified
	sessionRegistry.onboardingSessions[onboarding.OnboardingID] = onboarding
	store := agent.NewCircleOTPRequestStore()
	store.Save(onboarding.OnboardingID, "circle_request_secret_123")
	router := newVerifyEnabledRouter(sessionRegistry, &stubCircleOnboardingRunner{}, store)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/verify", bytes.NewBufferString(`{
		"onboarding_id": "agent_onboarding_verified",
		"otp": "B1X-123456"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, response.Code, response.Body.String())
	}
}

func TestVerifyAgentOnboardingRejectsExpiredRequest(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	onboarding := insertTestOnboardingSession(sessionRegistry, "agent_onboarding_expired")
	onboarding.CircleRequestExpiresAt = sql.NullTime{Time: time.Now().UTC().Add(-time.Minute), Valid: true}
	sessionRegistry.onboardingSessions[onboarding.OnboardingID] = onboarding
	store := agent.NewCircleOTPRequestStore()
	store.Save(onboarding.OnboardingID, "circle_request_secret_123")
	runner := &stubCircleOnboardingRunner{}
	router := newVerifyEnabledRouter(sessionRegistry, runner, store)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/verify", bytes.NewBufferString(`{
		"onboarding_id": "agent_onboarding_expired",
		"otp": "B1X-123456"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, response.Code, response.Body.String())
	}
	if runner.verifyCalled {
		t.Fatal("verifier should not be called for expired request")
	}
	if sessionRegistry.onboardingSessions[onboarding.OnboardingID].Status != repository.AgentOnboardingStatusExpired {
		t.Fatalf("expected expired status, got %q", sessionRegistry.onboardingSessions[onboarding.OnboardingID].Status)
	}
}

func TestVerifyAgentOnboardingMissingMemoryRequestID(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	onboarding := insertTestOnboardingSession(sessionRegistry, "agent_onboarding_no_request")
	onboarding.CircleRequestIDHash = sql.NullString{String: agent.HashCircleRequestID("circle_request_secret_123"), Valid: true}
	sessionRegistry.onboardingSessions[onboarding.OnboardingID] = onboarding
	router := newVerifyEnabledRouter(sessionRegistry, &stubCircleOnboardingRunner{}, agent.NewCircleOTPRequestStore())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/verify", bytes.NewBufferString(`{
		"onboarding_id": "agent_onboarding_no_request",
		"otp": "B1X-123456"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("circle_otp_request_not_available")) {
		t.Fatalf("expected request unavailable code, got %s", response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), []byte("circle_request_secret_123")) || bytes.Contains(response.Body.Bytes(), []byte(onboarding.CircleRequestIDHash.String)) {
		t.Fatalf("response should not expose raw request id or stored hash: %s", response.Body.String())
	}
}

func TestVerifyAgentOnboardingEnabledSuccess(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	onboarding := insertTestOnboardingSession(sessionRegistry, "agent_onboarding_verify_success")
	onboarding.UserWallet = sql.NullString{}
	onboarding.SourceClient = sql.NullString{String: "chatgpt_custom_action", Valid: true}
	onboarding.Channel = sql.NullString{String: "chatgpt", Valid: true}
	sessionRegistry.onboardingSessions[onboarding.OnboardingID] = onboarding
	store := agent.NewCircleOTPRequestStore()
	store.Save(onboarding.OnboardingID, "circle_request_secret_123")
	runner := &stubCircleOnboardingRunner{}
	walletRegistry := newTestAgentWalletRegistry()
	resolver := &stubCircleWalletResolver{
		wallet: agent.CircleAgentWallet{
			Address: "0xa9914bca9123ba0079be8c968f632c0db6400fe7",
			Chain:   agent.ChainArcTestnet,
		},
	}
	router := newVerifyEnabledRouterWithWalletRegistryAndResolver(sessionRegistry, walletRegistry, runner, store, resolver)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/verify", bytes.NewBufferString(`{
		"onboarding_id": "agent_onboarding_verify_success",
		"otp": "B1X-123456"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !runner.verifyCalled {
		t.Fatal("expected verifier to be called")
	}
	if runner.requestID != "circle_request_secret_123" {
		t.Fatalf("unexpected request id %q", runner.requestID)
	}
	if runner.otp != "B1X-123456" {
		t.Fatalf("unexpected otp %q", runner.otp)
	}
	if _, ok := store.Get(onboarding.OnboardingID); ok {
		t.Fatal("expected request id to be consumed after successful verification")
	}
	if !resolver.resolveCalled {
		t.Fatal("expected wallet resolver to be called")
	}
	if resolver.email != "desi@example.com" {
		t.Fatalf("expected resolver email, got %q", resolver.email)
	}

	var body struct {
		Onboarding   agentOnboardingSessionResponse `json:"onboarding"`
		AgentWallet  agentWalletResponse            `json:"agent_wallet"`
		AgentSession agentSessionResponse           `json:"agent_session"`
		NextStep     string                         `json:"next_step"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Onboarding.Status != repository.AgentOnboardingStatusVerified {
		t.Fatalf("expected verified status, got %q", body.Onboarding.Status)
	}
	if body.NextStep != "agent_session_active" {
		t.Fatalf("unexpected next step %q", body.NextStep)
	}
	if body.AgentWallet.AgentWalletAddress != "0xa9914bca9123ba0079be8c968f632c0db6400fe7" {
		t.Fatalf("unexpected agent wallet address %q", body.AgentWallet.AgentWalletAddress)
	}
	if body.AgentSession.Status != repository.AgentSessionStatusActive {
		t.Fatalf("expected active session, got %q", body.AgentSession.Status)
	}
	if body.AgentSession.AgentWalletAddress != body.AgentWallet.AgentWalletAddress {
		t.Fatalf("session wallet mismatch %q", body.AgentSession.AgentWalletAddress)
	}
	if bytes.Contains(response.Body.Bytes(), []byte("circle_request_secret_123")) || bytes.Contains(response.Body.Bytes(), []byte("B1X-123456")) {
		t.Fatalf("response should not expose raw request id or otp: %s", response.Body.String())
	}

	sessionResponse := httptest.NewRecorder()
	sessionRequest := httptest.NewRequest(http.MethodGet, "/agent/sessions/agent_verify_test", nil)
	router.ServeHTTP(sessionResponse, sessionRequest)
	if sessionResponse.Code != http.StatusOK {
		t.Fatalf("expected session status %d, got %d: %s", http.StatusOK, sessionResponse.Code, sessionResponse.Body.String())
	}

	walletResponse := httptest.NewRecorder()
	walletRequest := httptest.NewRequest(http.MethodGet, "/agent/wallets/agent_verify_test", nil)
	router.ServeHTTP(walletResponse, walletRequest)
	if walletResponse.Code != http.StatusOK {
		t.Fatalf("expected wallet status %d, got %d: %s", http.StatusOK, walletResponse.Code, walletResponse.Body.String())
	}
	if !bytes.Contains(walletResponse.Body.Bytes(), []byte("0xa9914bca9123ba0079be8c968f632c0db6400fe7")) {
		t.Fatalf("expected wallet address response, got %s", walletResponse.Body.String())
	}
}

func TestVerifyAgentOnboardingKeepsRequestIDWhenVerifiedUpdateFails(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	onboarding := insertTestOnboardingSession(sessionRegistry, "agent_onboarding_verify_update_fails")
	sessionRegistry.failStatusUpdate = true
	store := agent.NewCircleOTPRequestStore()
	store.Save(onboarding.OnboardingID, "circle_request_secret_123")
	runner := &stubCircleOnboardingRunner{}
	router := newVerifyEnabledRouter(sessionRegistry, runner, store)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/verify", bytes.NewBufferString(`{
		"onboarding_id": "agent_onboarding_verify_update_fails",
		"otp": "B1X-123456"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d: %s", http.StatusInternalServerError, response.Code, response.Body.String())
	}
	if !runner.verifyCalled {
		t.Fatal("expected verifier to be called before status update")
	}
	if requestID, ok := store.Get(onboarding.OnboardingID); !ok || requestID != "circle_request_secret_123" {
		t.Fatalf("expected request id to remain available after failed verified update, got %q, %v", requestID, ok)
	}
	if bytes.Contains(response.Body.Bytes(), []byte("circle_request_secret_123")) || bytes.Contains(response.Body.Bytes(), []byte("B1X-123456")) {
		t.Fatalf("response should not expose raw request id or otp: %s", response.Body.String())
	}
}

func TestVerifyAgentOnboardingMissingResolvedWalletDoesNotCreateSession(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	onboarding := insertTestOnboardingSession(sessionRegistry, "agent_onboarding_missing_wallet")
	store := agent.NewCircleOTPRequestStore()
	store.Save(onboarding.OnboardingID, "circle_request_secret_123")
	runner := &stubCircleOnboardingRunner{}
	walletRegistry := newTestAgentWalletRegistry()
	resolver := &stubCircleWalletResolver{err: agent.ErrCircleAgentWalletNotFound}
	router := newVerifyEnabledRouterWithWalletRegistryAndResolver(sessionRegistry, walletRegistry, runner, store, resolver)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/verify", bytes.NewBufferString(`{
		"onboarding_id": "agent_onboarding_missing_wallet",
		"otp": "B1X-123456"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("circle_agent_wallet_not_found")) {
		t.Fatalf("expected wallet not found code, got %s", response.Body.String())
	}
	if _, err := sessionRegistry.GetAgentSessionByAgentID(context.Background(), onboarding.AgentID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected no agent session, got %v", err)
	}
	if _, err := walletRegistry.GetAgentWalletByAgentID(context.Background(), onboarding.AgentID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected no agent wallet, got %v", err)
	}
	if sessionRegistry.onboardingSessions[onboarding.OnboardingID].Status != repository.AgentOnboardingStatusVerified {
		t.Fatalf("expected onboarding to remain verified after OTP success, got %q", sessionRegistry.onboardingSessions[onboarding.OnboardingID].Status)
	}
}

func TestVerifyAgentOnboardingAmbiguousResolvedWalletDoesNotCreateSession(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	onboarding := insertTestOnboardingSession(sessionRegistry, "agent_onboarding_ambiguous_wallet")
	store := agent.NewCircleOTPRequestStore()
	store.Save(onboarding.OnboardingID, "circle_request_secret_123")
	runner := &stubCircleOnboardingRunner{}
	walletRegistry := newTestAgentWalletRegistry()
	resolver := &stubCircleWalletResolver{err: agent.ErrCircleAgentWalletResolutionAmbiguous}
	router := newVerifyEnabledRouterWithWalletRegistryAndResolver(sessionRegistry, walletRegistry, runner, store, resolver)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/verify", bytes.NewBufferString(`{
		"onboarding_id": "agent_onboarding_ambiguous_wallet",
		"otp": "B1X-123456"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("circle_agent_wallet_resolution_ambiguous")) {
		t.Fatalf("expected ambiguous wallet code, got %s", response.Body.String())
	}
	if _, err := sessionRegistry.GetAgentSessionByAgentID(context.Background(), onboarding.AgentID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected no agent session, got %v", err)
	}
	if sessionRegistry.onboardingSessions[onboarding.OnboardingID].Status != repository.AgentOnboardingStatusVerified {
		t.Fatalf("expected onboarding to remain verified after OTP success, got %q", sessionRegistry.onboardingSessions[onboarding.OnboardingID].Status)
	}
}

func TestVerifyAgentOnboardingFailureIsSanitized(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	onboarding := insertTestOnboardingSession(sessionRegistry, "agent_onboarding_verify_failure")
	store := agent.NewCircleOTPRequestStore()
	store.Save(onboarding.OnboardingID, "circle_request_secret_123")
	runner := &stubCircleOnboardingRunner{err: errors.New("raw request_id circle_request_secret_123 otp B1X-123456")}
	router := newVerifyEnabledRouter(sessionRegistry, runner, store)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/verify", bytes.NewBufferString(`{
		"onboarding_id": "agent_onboarding_verify_failure",
		"otp": "B1X-123456"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadGateway, response.Code, response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), []byte("circle_request_secret_123")) || bytes.Contains(response.Body.Bytes(), []byte("B1X-123456")) {
		t.Fatalf("response should not expose raw request id or otp: %s", response.Body.String())
	}
	if sessionRegistry.onboardingSessions[onboarding.OnboardingID].Status != repository.AgentOnboardingStatusFailed {
		t.Fatalf("expected failed status, got %q", sessionRegistry.onboardingSessions[onboarding.OnboardingID].Status)
	}
}

func TestGetAgentOnboardingByOnboardingID(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry)

	createResponse := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/agent/onboarding/start", bytes.NewBufferString(`{
		"agent_id": "agent_start_fetch",
		"user_email": "desi@example.com",
		"user_wallet": "0x1111111111111111111111111111111111111111"
	}`))
	router.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	var createBody struct {
		Onboarding agentOnboardingSessionResponse `json:"onboarding"`
	}
	if err := json.NewDecoder(createResponse.Body).Decode(&createBody); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	getResponse := httptest.NewRecorder()
	getRequest := httptest.NewRequest(http.MethodGet, "/agent/onboarding/"+createBody.Onboarding.OnboardingID, nil)
	router.ServeHTTP(getResponse, getRequest)

	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected get status %d, got %d: %s", http.StatusOK, getResponse.Code, getResponse.Body.String())
	}

	var getBody struct {
		Onboarding agentOnboardingSessionResponse `json:"onboarding"`
	}
	if err := json.NewDecoder(getResponse.Body).Decode(&getBody); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if getBody.Onboarding.OnboardingID != createBody.Onboarding.OnboardingID {
		t.Fatalf("expected onboarding id %q, got %q", createBody.Onboarding.OnboardingID, getBody.Onboarding.OnboardingID)
	}
}

func TestGetUnknownAgentOnboardingReturnsNotFound(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/onboarding/missing", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, response.Code, response.Body.String())
	}
}

func TestCreateAgentSessionRejectsUserWalletReuse(t *testing.T) {
	sessionsRepository := repository.NewAgentSessionsRepository(nil)
	_, err := sessionsRepository.CreateAgentSession(context.Background(), repository.CreateAgentSessionInput{
		SessionID:          "agent_session_bad",
		AgentID:            "agent_bad",
		UserEmail:          "desi@example.com",
		UserWallet:         "0x9999999999999999999999999999999999999999",
		AgentWalletAddress: "0x9999999999999999999999999999999999999999",
		WalletProvider:     agent.WalletProviderCircleAgentWallet,
		Chain:              agent.ChainArcTestnet,
		Status:             repository.AgentSessionStatusActive,
		AllowedActions:     []string{agent.ActionCreateMarket},
		AllowedChannels:    []string{"chatgpt"},
	})
	if !errors.Is(err, repository.ErrInvalidAgentSession) {
		t.Fatalf("expected invalid agent session error, got %v", err)
	}
}

func TestGetMissingAgentSessionReturnsNotFound(t *testing.T) {
	sessionRegistry := newTestAgentSessionRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/sessions/agent_missing", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, response.Code, response.Body.String())
	}
}

func TestRegisterAgentOnboardingMinimalPayload(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/register", bytes.NewBufferString(`{
		"agent_id": "agent_onboard_1",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"agent_wallet_address": "0x9999999999999999999999999999999999999999"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var body struct {
		AgentWallet agentWalletResponse `json:"agent_wallet"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.AgentWallet.AgentID != "agent_onboard_1" {
		t.Fatalf("expected agent id, got %q", body.AgentWallet.AgentID)
	}
	if body.AgentWallet.UserWallet != "0x1111111111111111111111111111111111111111" {
		t.Fatalf("unexpected user wallet %q", body.AgentWallet.UserWallet)
	}
	if body.AgentWallet.AgentWalletAddress != "0x9999999999999999999999999999999999999999" {
		t.Fatalf("unexpected agent wallet address %q", body.AgentWallet.AgentWalletAddress)
	}
}

func TestRegisterAgentOnboardingAppliesDefaults(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/register", bytes.NewBufferString(`{
		"agent_id": "agent_onboard_defaults",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"agent_wallet_address": "0x9999999999999999999999999999999999999999"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var body struct {
		AgentWallet agentWalletResponse `json:"agent_wallet"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.AgentWallet.Chain != agent.ChainArcTestnet {
		t.Fatalf("expected ARC-TESTNET chain, got %q", body.AgentWallet.Chain)
	}
	if body.AgentWallet.WalletProvider != agent.WalletProviderCircleAgentWallet {
		t.Fatalf("expected circle agent wallet provider, got %q", body.AgentWallet.WalletProvider)
	}
	if body.AgentWallet.Status != agent.WalletStatusActive {
		t.Fatalf("expected active status, got %q", body.AgentWallet.Status)
	}
	for _, action := range defaultAgentWalletAllowedActions() {
		if !containsString(body.AgentWallet.AllowedActions, action) {
			t.Fatalf("expected default action %q in %#v", action, body.AgentWallet.AllowedActions)
		}
	}
	if body.AgentWallet.PolicyMetadata["note"] != "default ARC-TESTNET onboarding policy" {
		t.Fatalf("unexpected policy metadata %#v", body.AgentWallet.PolicyMetadata)
	}
}

func TestRegisterAgentOnboardingPreservesOptionalFields(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/register", bytes.NewBufferString(`{
		"agent_id": "agent_onboard_optional",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"agent_wallet_address": "0x9999999999999999999999999999999999999999",
		"user_email": "desi@example.com",
		"source_client": "chatgpt_custom_action"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var body struct {
		AgentWallet agentWalletResponse `json:"agent_wallet"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.AgentWallet.UserEmail != "desi@example.com" {
		t.Fatalf("unexpected user email %q", body.AgentWallet.UserEmail)
	}
	if body.AgentWallet.SourceClient != "chatgpt_custom_action" {
		t.Fatalf("unexpected source client %q", body.AgentWallet.SourceClient)
	}
}

func TestRegisterAgentOnboardingRejectsMissingRequiredFields(t *testing.T) {
	assertAgentOnboardingRegistrationFails(t, `{}`)
}

func TestRegisterAgentOnboardingRejectsUserWalletReuse(t *testing.T) {
	assertAgentOnboardingRegistrationFails(t, `{
		"agent_id": "agent_onboard_bad",
		"user_wallet": "0x9999999999999999999999999999999999999999",
		"agent_wallet_address": "0x9999999999999999999999999999999999999999"
	}`)
}

func TestRegisterAgentOnboardingDoesNotCallExecutor(t *testing.T) {
	executor := &stubAgentExecutor{}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), executor)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/register", bytes.NewBufferString(`{
		"agent_id": "agent_onboard_no_execute",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"agent_wallet_address": "0x9999999999999999999999999999999999999999"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	if executor.called {
		t.Fatal("executor should not be called during onboarding")
	}
}

func TestAgentOnboardingDoesNotBreakAgentWalletRegistration(t *testing.T) {
	registry := newTestAgentWalletRegistry()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), registry, nil)

	onboardingResponse := httptest.NewRecorder()
	onboardingRequest := httptest.NewRequest(http.MethodPost, "/agent/onboarding/register", bytes.NewBufferString(`{
		"agent_id": "agent_onboard_then_wallet",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"agent_wallet_address": "0x9999999999999999999999999999999999999999"
	}`))
	router.ServeHTTP(onboardingResponse, onboardingRequest)
	if onboardingResponse.Code != http.StatusCreated {
		t.Fatalf("expected onboarding status %d, got %d: %s", http.StatusCreated, onboardingResponse.Code, onboardingResponse.Body.String())
	}

	walletResponse := httptest.NewRecorder()
	walletRequest := httptest.NewRequest(http.MethodPost, "/agent/wallets", bytes.NewBufferString(`{
		"agent_id": "agent_onboard_then_wallet",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"agent_wallet_address": "0x9999999999999999999999999999999999999999",
		"wallet_provider": "circle_agent_wallet",
		"chain": "ARC-TESTNET",
		"allowed_actions": ["create_market", "buy_yes"],
		"status": "active",
		"policy_metadata": {
			"source": "explicit registry update"
		}
	}`))
	router.ServeHTTP(walletResponse, walletRequest)
	if walletResponse.Code != http.StatusCreated {
		t.Fatalf("expected wallet registration status %d, got %d: %s", http.StatusCreated, walletResponse.Code, walletResponse.Body.String())
	}

	var body struct {
		AgentWallet agentWalletResponse `json:"agent_wallet"`
	}
	if err := json.NewDecoder(walletResponse.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !containsString(body.AgentWallet.AllowedActions, agent.ActionBuyYes) {
		t.Fatalf("expected explicit wallet registration action, got %#v", body.AgentWallet.AllowedActions)
	}
	if containsString(body.AgentWallet.AllowedActions, agent.ActionClaimRefund) {
		t.Fatalf("expected explicit /agent/wallets action set to be preserved, got %#v", body.AgentWallet.AllowedActions)
	}
}

func TestGetAgentWalletByAgentID(t *testing.T) {
	registry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, registry, agent.ActionCreateMarket, agent.ActionBuyYes)
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), registry, nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/wallets/agent_test_1", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var body struct {
		AgentWallet agentWalletResponse `json:"agent_wallet"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.AgentWallet.AgentID != "agent_test_1" {
		t.Fatalf("expected agent id, got %q", body.AgentWallet.AgentID)
	}
	if !containsString(body.AgentWallet.AllowedActions, agent.ActionBuyYes) {
		t.Fatalf("expected buy_yes in allowed actions, got %#v", body.AgentWallet.AllowedActions)
	}
}

func TestGetAgentWalletBalanceReturnsEmptyBalances(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	resolver := &stubCircleWalletResolver{
		balances: agent.CircleAgentWalletBalances{Balances: []any{}},
	}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), walletRegistry, nil, newTestAgentSessionRegistry(), agent.CircleOnboardingStarter{}, resolver)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/wallets/agent_test_1/balance", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !resolver.balanceCalled {
		t.Fatal("expected balance resolver to be called")
	}
	if resolver.balanceAddress != "0x9999999999999999999999999999999999999999" {
		t.Fatalf("unexpected balance address %q", resolver.balanceAddress)
	}

	var body struct {
		Balance agentWalletBalanceResponse `json:"agent_wallet_balance"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode balance response: %v", err)
	}
	if body.Balance.AgentWalletAddress != "0x9999999999999999999999999999999999999999" {
		t.Fatalf("unexpected balance wallet %q", body.Balance.AgentWalletAddress)
	}
	if len(body.Balance.Balances) != 0 {
		t.Fatalf("expected empty balances, got %#v", body.Balance.Balances)
	}
}

func TestGetAgentWalletBalancePrefersDirectAPIReader(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	wallet := walletRegistry.wallets["agent_test_1"]
	wallet.PolicyMetadata = json.RawMessage(`{"circle_wallet_id":"wallet-1"}`)
	walletRegistry.wallets["agent_test_1"] = wallet
	resolver := &stubCircleWalletResolver{
		balances: agent.CircleAgentWalletBalances{Balances: []any{map[string]any{"source": "cli"}}},
	}
	reader := &stubCircleBalanceReader{
		balances: agent.CircleAgentWalletBalances{Balances: []any{map[string]any{"source": "api"}}},
	}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), walletRegistry, nil, newTestAgentSessionRegistry(), agent.CircleOnboardingStarter{}, resolver, reader)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/wallets/agent_test_1/balance", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !reader.called {
		t.Fatal("expected direct API balance reader to be called")
	}
	if resolver.balanceCalled {
		t.Fatal("CLI balance resolver must not be called when direct API reader is configured")
	}
	if reader.request.AgentID != "agent_test_1" || reader.request.AgentWalletAddress != "0x9999999999999999999999999999999999999999" {
		t.Fatalf("unexpected direct API request: %#v", reader.request)
	}
	if reader.request.PolicyMetadata["circle_wallet_id"] != "wallet-1" {
		t.Fatalf("expected circle wallet id in policy metadata, got %#v", reader.request.PolicyMetadata)
	}

	var body struct {
		Balance agentWalletBalanceResponse `json:"agent_wallet_balance"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode balance response: %v", err)
	}
	if len(body.Balance.Balances) != 1 {
		t.Fatalf("expected one API balance, got %#v", body.Balance.Balances)
	}
}

func TestGetAgentWalletBalanceUnknownAgentReturnsNotFound(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, newTestAgentSessionRegistry(), agent.CircleOnboardingStarter{}, &stubCircleBalanceReader{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/wallets/agent_test_missing/balance", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, response.Code, response.Body.String())
	}
}

func TestGetAgentWalletBalanceInactiveWalletReturnsConflict(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	wallet := walletRegistry.wallets["agent_test_1"]
	wallet.Status = "disabled"
	walletRegistry.wallets["agent_test_1"] = wallet
	reader := &stubCircleBalanceReader{}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), walletRegistry, nil, newTestAgentSessionRegistry(), agent.CircleOnboardingStarter{}, reader)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/wallets/agent_test_1/balance", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, response.Code, response.Body.String())
	}
	if reader.called {
		t.Fatal("balance provider must not be called for inactive wallet")
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("agent_wallet_status_invalid")) {
		t.Fatalf("expected status invalid error, got %s", response.Body.String())
	}
}

func TestGetAgentWalletBalanceWrongChainReturnsConflict(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	wallet := walletRegistry.wallets["agent_test_1"]
	wallet.Chain = "ETH-SEPOLIA"
	walletRegistry.wallets["agent_test_1"] = wallet
	reader := &stubCircleBalanceReader{}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), walletRegistry, nil, newTestAgentSessionRegistry(), agent.CircleOnboardingStarter{}, reader)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/wallets/agent_test_1/balance", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, response.Code, response.Body.String())
	}
	if reader.called {
		t.Fatal("balance provider must not be called for wrong chain")
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("agent_wallet_chain_invalid")) {
		t.Fatalf("expected chain invalid error, got %s", response.Body.String())
	}
}

func TestGetAgentWalletBalanceFailureIsGeneric(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	resolver := &stubCircleWalletResolver{balanceErr: agent.ErrCircleAgentWalletBalanceFailed}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), walletRegistry, nil, newTestAgentSessionRegistry(), agent.CircleOnboardingStarter{}, resolver)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/wallets/agent_test_1/balance", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadGateway, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("circle_agent_wallet_balance_failed")) {
		t.Fatalf("expected generic balance failure, got %s", response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), []byte("token")) || bytes.Contains(response.Body.Bytes(), []byte("request_id")) {
		t.Fatalf("response should not expose secret-like diagnostics: %s", response.Body.String())
	}
}

func TestDisableAgentWallet(t *testing.T) {
	registry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, registry, agent.ActionBuyYes)
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), registry, nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/wallets/agent_test_1/disable", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var body struct {
		AgentWallet agentWalletResponse `json:"agent_wallet"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.AgentWallet.Status != "disabled" {
		t.Fatalf("expected disabled status, got %q", body.AgentWallet.Status)
	}
}

func TestRegisterAgentWalletRejectsDeployerWallet(t *testing.T) {
	assertAgentWalletRegistrationFails(t, `{
		"agent_id": "agent_bad",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"agent_wallet_address": "0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC",
		"wallet_provider": "circle_agent_wallet",
		"chain": "ARC-TESTNET",
		"allowed_actions": ["buy_yes"]
	}`)
}

func TestRegisterAgentWalletRejectsUserWalletReuse(t *testing.T) {
	assertAgentWalletRegistrationFails(t, `{
		"agent_id": "agent_bad",
		"user_wallet": "0x9999999999999999999999999999999999999999",
		"agent_wallet_address": "0x9999999999999999999999999999999999999999",
		"wallet_provider": "circle_agent_wallet",
		"chain": "ARC-TESTNET",
		"allowed_actions": ["buy_yes"]
	}`)
}

func TestRegisterAgentWalletRejectsUnsupportedProvider(t *testing.T) {
	assertAgentWalletRegistrationFails(t, `{
		"agent_id": "agent_bad",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"agent_wallet_address": "0x9999999999999999999999999999999999999999",
		"wallet_provider": "temporary_testnet_agent_eoa",
		"chain": "ARC-TESTNET",
		"allowed_actions": ["buy_yes"]
	}`)
}

func TestRegisterAgentWalletRejectsWrongChain(t *testing.T) {
	assertAgentWalletRegistrationFails(t, `{
		"agent_id": "agent_bad",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"agent_wallet_address": "0x9999999999999999999999999999999999999999",
		"wallet_provider": "circle_agent_wallet",
		"chain": "BASE",
		"allowed_actions": ["buy_yes"]
	}`)
}

func TestGetAgentIntentPreview(t *testing.T) {
	store := agent.NewStore()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, newTestAgentWalletRegistry(), nil)

	createResponse := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/agent/intents", bytes.NewBufferString(`{
		"action": "claim_payout",
		"user_wallet": "0x2222222222222222222222222222222222222222",
		"market_id": "market-2",
		"market_contract_address": "0x4444444444444444444444444444444444444444"
	}`))
	router.ServeHTTP(createResponse, createRequest)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d", http.StatusCreated, createResponse.Code)
	}

	var createBody struct {
		Intent agentIntentResponse `json:"intent"`
	}
	if err := json.NewDecoder(createResponse.Body).Decode(&createBody); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	getResponse := httptest.NewRecorder()
	getRequest := httptest.NewRequest(http.MethodGet, "/agent/intents/"+createBody.Intent.IntentID, nil)
	router.ServeHTTP(getResponse, getRequest)

	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected get status %d, got %d", http.StatusOK, getResponse.Code)
	}

	var getBody struct {
		Intent agentIntentResponse `json:"intent"`
	}
	if err := json.NewDecoder(getResponse.Body).Decode(&getBody); err != nil {
		t.Fatalf("decode get response: %v", err)
	}

	if getBody.Intent.IntentID != createBody.Intent.IntentID {
		t.Fatalf("expected intent id %q, got %q", createBody.Intent.IntentID, getBody.Intent.IntentID)
	}
}

func TestCreateAgentIntentValidationErrors(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents", bytes.NewBufferString(`{
		"action": "buy_no",
		"market_id": "market-1",
		"market_contract_address": "0x5555555555555555555555555555555555555555",
		"amount": "0"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	var body struct {
		Intent agentIntentResponse `json:"intent"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Intent.ValidationResult.Valid {
		t.Fatal("expected invalid validation result")
	}
	if len(body.Intent.ValidationResult.Errors) != 2 {
		t.Fatalf("expected two validation errors, got %#v", body.Intent.ValidationResult.Errors)
	}
}

func TestCreateAgentIntentBindsCircleAgentWalletWhenUserWalletMissing(t *testing.T) {
	registry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, registry, agent.ActionBuyYes)
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), registry, nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents", bytes.NewBufferString(`{
		"action": "buy_yes",
		"agent_id": "agent_test_1",
		"source_client": "hermes_telegram",
		"client_request_id": "client_req_agent_wallet_binding",
		"market_id": "market-1",
		"market_contract_address": "0x3333333333333333333333333333333333333333",
		"amount": "5"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var body struct {
		Intent agentIntentResponse `json:"intent"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Intent.UserWallet != "0x9999999999999999999999999999999999999999" {
		t.Fatalf("expected user_wallet to bind to agent wallet address, got %q", body.Intent.UserWallet)
	}
	if !body.Intent.ValidationResult.Valid {
		t.Fatalf("expected valid validation result, got %#v", body.Intent.ValidationResult)
	}
}

func TestCreateAgentIntentIncludesRegisteredAgentMetadata(t *testing.T) {
	registry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, registry, agent.ActionBuyYes)
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), registry, nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents", bytes.NewBufferString(`{
		"action": "buy_yes",
		"agent_id": "agent_test_1",
		"source_client": "chatgpt_custom_action",
		"client_request_id": "client_req_1",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "market-1",
		"market_contract_address": "0x3333333333333333333333333333333333333333",
		"amount": "1000000"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var body struct {
		Intent agentIntentResponse `json:"intent"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Intent.AgentID != "agent_test_1" {
		t.Fatalf("expected agent metadata, got %q", body.Intent.AgentID)
	}
	if body.Intent.AgentWalletAddress != "0x9999999999999999999999999999999999999999" {
		t.Fatalf("unexpected agent wallet address %q", body.Intent.AgentWalletAddress)
	}
	if body.Intent.WalletProvider != agent.WalletProviderCircleAgentWallet {
		t.Fatalf("unexpected wallet provider %q", body.Intent.WalletProvider)
	}
	if body.Intent.SourceClient != "chatgpt_custom_action" {
		t.Fatalf("unexpected source client %q", body.Intent.SourceClient)
	}
	if body.Intent.ClientRequestID != "client_req_1" {
		t.Fatalf("unexpected client request id %q", body.Intent.ClientRequestID)
	}
}

func TestGetAgentIntentNotFound(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/intents/missing", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestConfirmValidAgentIntent(t *testing.T) {
	store := agent.NewStore()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, newTestAgentWalletRegistry(), nil)
	intentID := createValidAgentIntent(t, router)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/confirm", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var body struct {
		ExecutionPlan agentExecutionPlanResponse `json:"execution_plan"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.ExecutionPlan.IntentID != intentID {
		t.Fatalf("expected intent id %q, got %q", intentID, body.ExecutionPlan.IntentID)
	}
	if body.ExecutionPlan.Status != "confirmed" {
		t.Fatalf("expected confirmed status, got %q", body.ExecutionPlan.Status)
	}
	if body.ExecutionPlan.ExecutionMode != "agent_contract" {
		t.Fatalf("expected agent_contract execution mode, got %q", body.ExecutionPlan.ExecutionMode)
	}
	if body.ExecutionPlan.Network != "arc_testnet" {
		t.Fatalf("expected arc_testnet network, got %q", body.ExecutionPlan.Network)
	}
	if body.ExecutionPlan.AgentFactoryAddress != "0x69aE770e8b2F96297101FeC4dc123B3801dA7d80" {
		t.Fatalf("unexpected agent factory address %q", body.ExecutionPlan.AgentFactoryAddress)
	}
	if !body.ExecutionPlan.RequiresSignature {
		t.Fatal("expected requires_signature true")
	}
	if body.ExecutionPlan.BroadcastPerformed {
		t.Fatal("expected broadcast_performed false")
	}
	if body.ExecutionPlan.TransactionHash != nil {
		t.Fatalf("expected nil transaction hash, got %q", *body.ExecutionPlan.TransactionHash)
	}
	if len(body.ExecutionPlan.Warnings) == 0 {
		t.Fatal("expected warnings")
	}
}

func TestConfirmMissingAgentIntentReturnsNotFound(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/missing/confirm", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestConfirmInvalidAgentIntentReturnsBadRequest(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	createResponse := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/agent/intents", bytes.NewBufferString(`{
		"action": "buy_yes",
		"market_id": "market-1",
		"amount": "0"
	}`))
	router.ServeHTTP(createResponse, createRequest)

	if createResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected create status %d, got %d", http.StatusBadRequest, createResponse.Code)
	}

	var createBody struct {
		Intent agentIntentResponse `json:"intent"`
	}
	if err := json.NewDecoder(createResponse.Body).Decode(&createBody); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+createBody.Intent.IntentID+"/confirm", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected confirm status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestConfirmAgentIntentIsIdempotent(t *testing.T) {
	store := agent.NewStore()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, newTestAgentWalletRegistry(), nil)
	intentID := createValidAgentIntent(t, router)

	firstResponse := httptest.NewRecorder()
	firstRequest := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/confirm", nil)
	router.ServeHTTP(firstResponse, firstRequest)

	secondResponse := httptest.NewRecorder()
	secondRequest := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/confirm", nil)
	router.ServeHTTP(secondResponse, secondRequest)

	if firstResponse.Code != http.StatusOK {
		t.Fatalf("expected first status %d, got %d", http.StatusOK, firstResponse.Code)
	}
	if secondResponse.Code != http.StatusOK {
		t.Fatalf("expected second status %d, got %d", http.StatusOK, secondResponse.Code)
	}

	var firstBody struct {
		ExecutionPlan agentExecutionPlanResponse `json:"execution_plan"`
	}
	if err := json.NewDecoder(firstResponse.Body).Decode(&firstBody); err != nil {
		t.Fatalf("decode first response: %v", err)
	}

	var secondBody struct {
		ExecutionPlan agentExecutionPlanResponse `json:"execution_plan"`
	}
	if err := json.NewDecoder(secondResponse.Body).Decode(&secondBody); err != nil {
		t.Fatalf("decode second response: %v", err)
	}

	if firstBody.ExecutionPlan.IntentID != secondBody.ExecutionPlan.IntentID {
		t.Fatalf("expected same intent id, got %q and %q", firstBody.ExecutionPlan.IntentID, secondBody.ExecutionPlan.IntentID)
	}
	if secondBody.ExecutionPlan.Status != "confirmed" {
		t.Fatalf("expected confirmed status, got %q", secondBody.ExecutionPlan.Status)
	}
}

func TestConfirmAgentIntentResponseSaysBroadcastNotPerformed(t *testing.T) {
	store := agent.NewStore()
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, newTestAgentWalletRegistry(), nil)
	intentID := createValidAgentIntent(t, router)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/confirm", nil)
	router.ServeHTTP(response, request)

	var body struct {
		ExecutionPlan agentExecutionPlanResponse `json:"execution_plan"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.ExecutionPlan.BroadcastPerformed {
		t.Fatal("expected broadcast_performed false")
	}
	if body.ExecutionPlan.TransactionHash != nil {
		t.Fatalf("expected transaction_hash null, got %q", *body.ExecutionPlan.TransactionHash)
	}
	if body.ExecutionPlan.TransactionRequest.BroadcastPerformed {
		t.Fatal("expected transaction_request broadcast_performed false")
	}
}

func TestConfirmCreateMarketReturnsFactoryTransactionRequest(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	intentID := createAgentIntent(t, router, `{
		"action": "create_market",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "agent-market-1",
		"question": "Will SignalArc create an agent market?",
		"close_timestamp": "1767225600",
		"resolver": "0x2222222222222222222222222222222222222222",
		"collateral_token": "0x3333333333333333333333333333333333333333"
	}`)

	executionPlan := confirmAgentIntent(t, router, intentID)

	if executionPlan.TransactionRequest.To != "0x69aE770e8b2F96297101FeC4dc123B3801dA7d80" {
		t.Fatalf("expected factory address as to, got %q", executionPlan.TransactionRequest.To)
	}
	if executionPlan.TransactionRequest.Contract != "SignalArcAgentMarketFactory" {
		t.Fatalf("expected factory contract, got %q", executionPlan.TransactionRequest.Contract)
	}
	if executionPlan.TransactionRequest.Function != "createMarket" {
		t.Fatalf("expected createMarket function, got %q", executionPlan.TransactionRequest.Function)
	}
	expectedArgs := []string{
		"agent-market-1",
		"Will SignalArc create an agent market?",
		"1767225600",
		"0x2222222222222222222222222222222222222222",
		"0x3333333333333333333333333333333333333333",
	}
	assertStringSliceEqual(t, executionPlan.TransactionRequest.Args, expectedArgs)
	assertNoExecutionClaim(t, executionPlan)
}

func TestExecuteConfirmedCreateMarketReturnsRealExecutionShape(t *testing.T) {
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	isMarket := true
	executor := &stubAgentExecutor{
		result: agent.ExecutionResult{
			IntentID:            "set-by-test",
			AgentID:             "agent_test_1",
			AgentWalletAddress:  "0x9999999999999999999999999999999999999999",
			WalletProvider:      agent.WalletProviderCircleAgentWallet,
			Action:              agent.ActionCreateMarket,
			Status:              agent.StatusExecuted,
			ExecutionMode:       agent.ExecutionModeAgentContract,
			Network:             agent.NetworkArcTestnet,
			AgentFactoryAddress: "0x69aE770e8b2F96297101FeC4dc123B3801dA7d80",
			BroadcastPerformed:  true,
			TransactionHash:     "0x1111111111111111111111111111111111111111111111111111111111111111",
			Readback: agent.ExecutionReadback{
				MarketCount:   "5",
				CreatedMarket: "0x2222222222222222222222222222222222222222",
				IsMarket:      &isMarket,
			},
		},
	}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor)

	intentID := createAgentIntent(t, router, `{
		"action": "create_market",
		"agent_id": "agent_test_1",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "agent-market-execute-1",
		"question": "Will SignalArc execute an agent market?",
		"close_timestamp": "1767225600",
		"resolver": "0x2222222222222222222222222222222222222222",
		"collateral_token": "0x3333333333333333333333333333333333333333"
	}`)
	confirmAgentIntent(t, router, intentID)
	executor.result.IntentID = intentID

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !executor.called {
		t.Fatal("expected executor to be called")
	}
	if executor.intent.ID != intentID {
		t.Fatalf("expected executor intent %q, got %q", intentID, executor.intent.ID)
	}

	var body struct {
		Execution agentExecutionResponse `json:"execution"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode execute response: %v", err)
	}

	if body.Execution.Status != agent.StatusExecuted {
		t.Fatalf("expected executed status, got %q", body.Execution.Status)
	}
	if !body.Execution.BroadcastPerformed {
		t.Fatal("expected broadcast_performed true")
	}
	if body.Execution.TransactionHash != executor.result.TransactionHash {
		t.Fatalf("unexpected transaction hash %q", body.Execution.TransactionHash)
	}
	if body.Execution.Network != agent.NetworkArcTestnet {
		t.Fatalf("expected arc_testnet network, got %q", body.Execution.Network)
	}
	if body.Execution.Readback.MarketCount != "5" {
		t.Fatalf("expected market count 5, got %q", body.Execution.Readback.MarketCount)
	}
	if body.Execution.Readback.CreatedMarket != "0x2222222222222222222222222222222222222222" {
		t.Fatalf("unexpected created market %q", body.Execution.Readback.CreatedMarket)
	}
	if body.Execution.Readback.IsMarket == nil || !*body.Execution.Readback.IsMarket {
		t.Fatalf("expected is_market true, got %#v", body.Execution.Readback.IsMarket)
	}
}

func TestExecuteConfirmedIntentPersistsDurableExecutionSuccess(t *testing.T) {
	durableRegistry := newTestDurableAgentIntentRegistry()
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	isMarket := true
	executor := &stubAgentExecutor{
		result: agent.ExecutionResult{
			AgentID:             "agent_test_1",
			AgentWalletAddress:  "0x9999999999999999999999999999999999999999",
			WalletProvider:      agent.WalletProviderCircleAgentWallet,
			Action:              agent.ActionCreateMarket,
			Status:              agent.StatusExecuted,
			ExecutionMode:       agent.ExecutionModeCircleAgentWalletCLI,
			Network:             agent.NetworkArcTestnet,
			AgentFactoryAddress: agent.AgentFactoryAddress,
			BroadcastPerformed:  true,
			TransactionHash:     "0x1111111111111111111111111111111111111111111111111111111111111111",
			Readback: agent.ExecutionReadback{
				MarketCount:   "5",
				CreatedMarket: "0x2222222222222222222222222222222222222222",
				IsMarket:      &isMarket,
			},
		},
	}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor, durableRegistry)

	intentID := createAgentIntent(t, router, `{
		"agent_id": "agent_test_1",
		"source_client": "test_client",
		"client_request_id": "client_req_execute_success",
		"action": "create_market",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "agent-market-execute-durable",
		"question": "Will SignalArc persist durable execution success?",
		"close_timestamp": "1767225600",
		"resolver": "0x2222222222222222222222222222222222222222",
		"collateral_token": "0x3333333333333333333333333333333333333333"
	}`)
	confirmAgentIntent(t, router, intentID)
	executor.result.IntentID = intentID

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if durableRegistry.intents[intentID].Status != agent.StatusExecuted {
		t.Fatalf("expected durable intent executed, got %q", durableRegistry.intents[intentID].Status)
	}
	var execution repository.AgentExecution
	for _, value := range durableRegistry.executions {
		execution = value
	}
	if execution.Status != repository.AgentExecutionStatusExecuted {
		t.Fatalf("expected durable execution executed, got %q", execution.Status)
	}
	if execution.TransactionHash.String != executor.result.TransactionHash {
		t.Fatalf("expected transaction hash %q, got %q", executor.result.TransactionHash, execution.TransactionHash.String)
	}
	if !execution.BroadcastPerformed {
		t.Fatal("expected durable execution broadcast_performed true")
	}
}

func TestExecuteDurableBuyYesHydratesAllowedActionsFromRegisteredWallet(t *testing.T) {
	durableRegistry := newTestDurableAgentIntentRegistry()
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionBuyYes)
	executor := &stubAgentExecutor{
		requiredAllowedAction: agent.ActionBuyYes,
		result: agent.ExecutionResult{
			AgentID:                "agent_test_1",
			AgentWalletAddress:     "0x9999999999999999999999999999999999999999",
			WalletProvider:         agent.WalletProviderCircleAgentWallet,
			Action:                 agent.ActionBuyYes,
			Status:                 agent.StatusExecuted,
			ExecutionMode:          agent.ExecutionModeCircleAgentWalletCLI,
			Network:                agent.NetworkArcTestnet,
			MarketContractAddress:  "0x3333333333333333333333333333333333333333",
			BroadcastPerformed:     true,
			ApproveTransactionHash: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			TransactionHash:        "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			Readback: agent.ExecutionReadback{
				YesPositions:    "1000000",
				TotalYes:        "1000000",
				TotalCollateral: "1000000",
				USDCBalance:     "19000000",
			},
		},
	}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor, durableRegistry)

	intentID := createAgentIntent(t, router, `{
		"agent_id": "agent_test_1",
		"source_client": "hermes-telegram",
		"client_request_id": "buy-yes-durable-hydration",
		"action": "buy_yes",
		"market_id": "546af477-7d88-48b4-8358-5d64f1b51e90",
		"market_contract_address": "0x3333333333333333333333333333333333333333",
		"amount": "1"
	}`)
	confirmAgentIntent(t, router, intentID)
	executor.result.IntentID = intentID

	reloadedIntent := newAgentIntentFromRepository(durableRegistry.intents[intentID])
	if len(reloadedIntent.AllowedActions) != 0 {
		t.Fatalf("expected durable reload to omit allowed actions, got %#v", reloadedIntent.AllowedActions)
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !executor.called {
		t.Fatal("expected executor to be called")
	}
	if !agent.AgentWalletAllowsAction(agent.AgentWallet{AllowedActions: executor.intent.AllowedActions}, agent.ActionBuyYes) {
		t.Fatalf("expected executor intent allowed actions to include buy_yes, got %#v", executor.intent.AllowedActions)
	}
	if executor.intent.AgentWalletAddress != "0x9999999999999999999999999999999999999999" {
		t.Fatalf("unexpected executor agent wallet %q", executor.intent.AgentWalletAddress)
	}
	if executor.intent.WalletProvider != agent.WalletProviderCircleAgentWallet {
		t.Fatalf("unexpected executor wallet provider %q", executor.intent.WalletProvider)
	}
	if executor.intent.UserWallet != "0x9999999999999999999999999999999999999999" {
		t.Fatalf("expected missing user_wallet to bind to agent wallet, got %q", executor.intent.UserWallet)
	}
	if durableRegistry.intents[intentID].Status != agent.StatusExecuted {
		t.Fatalf("expected durable intent executed, got %q", durableRegistry.intents[intentID].Status)
	}
}

func TestExecuteConfirmedIntentPersistsDurableExecutionFailure(t *testing.T) {
	durableRegistry := newTestDurableAgentIntentRegistry()
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	executor := &stubAgentExecutor{err: errors.New("rpc unavailable with upstream detail")}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor, durableRegistry)

	intentID := createAgentIntent(t, router, `{
		"agent_id": "agent_test_1",
		"source_client": "test_client",
		"client_request_id": "client_req_execute_failure",
		"action": "create_market",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "agent-market-execute-failure",
		"question": "Will SignalArc persist durable execution failure?",
		"close_timestamp": "1767225600",
		"resolver": "0x2222222222222222222222222222222222222222",
		"collateral_token": "0x3333333333333333333333333333333333333333"
	}`)
	confirmAgentIntent(t, router, intentID)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusBadGateway, response.Code, response.Body.String())
	}
	if durableRegistry.intents[intentID].Status != "failed" {
		t.Fatalf("expected durable intent failed, got %q", durableRegistry.intents[intentID].Status)
	}
	var execution repository.AgentExecution
	for _, value := range durableRegistry.executions {
		execution = value
	}
	if execution.Status != repository.AgentExecutionStatusFailed {
		t.Fatalf("expected durable execution failed, got %q", execution.Status)
	}
	if execution.ErrorCode.String != "agent_execution_failed" {
		t.Fatalf("expected sanitized error code, got %q", execution.ErrorCode.String)
	}
	if strings.Contains(execution.ErrorMessage.String, "rpc unavailable") {
		t.Fatalf("durable error message should be sanitized, got %q", execution.ErrorMessage.String)
	}
}

func TestAgentActivityReturnsDurableIntentAndExecutionActivity(t *testing.T) {
	durableRegistry := newTestDurableAgentIntentRegistry()
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionBuyYes)
	executor := &stubAgentExecutor{
		result: agent.ExecutionResult{
			AgentID:                "agent_test_1",
			AgentWalletAddress:     "0x9999999999999999999999999999999999999999",
			WalletProvider:         agent.WalletProviderCircleAgentWallet,
			Action:                 agent.ActionBuyYes,
			Status:                 agent.StatusExecuted,
			ExecutionMode:          agent.ExecutionModeCircleAgentWalletCLI,
			Network:                agent.NetworkArcTestnet,
			MarketContractAddress:  "0x3333333333333333333333333333333333333333",
			BroadcastPerformed:     true,
			ApproveTransactionHash: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			TransactionHash:        "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			Readback: agent.ExecutionReadback{
				YesPositions:    "1000000",
				TotalYes:        "1000000",
				TotalCollateral: "1000000",
				USDCBalance:     "1000000",
			},
		},
	}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor, durableRegistry)

	intentID := createValidAgentIntent(t, router)
	confirmAgentIntent(t, router, intentID)
	executor.result.IntentID = intentID

	executeResponse := httptest.NewRecorder()
	executeRequest := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(executeResponse, executeRequest)
	if executeResponse.Code != http.StatusOK {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusOK, executeResponse.Code, executeResponse.Body.String())
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/activity/agent_test_1", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected activity status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var body struct {
		Activity agentActivityResponse `json:"activity"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode activity response: %v", err)
	}
	if len(body.Activity.Items) < 2 {
		t.Fatalf("expected intent and execution activity, got %#v", body.Activity.Items)
	}
	var executionItem agentActivityItemResponse
	for _, item := range body.Activity.Items {
		if item.Type == "execution" {
			executionItem = item
			break
		}
	}
	if executionItem.Type != "execution" {
		t.Fatalf("expected execution item, got %#v", body.Activity.Items)
	}
	if executionItem.TransactionHash != executor.result.TransactionHash {
		t.Fatalf("expected transaction hash %q, got %q", executor.result.TransactionHash, executionItem.TransactionHash)
	}
	if executionItem.ApproveTransactionHash != executor.result.ApproveTransactionHash {
		t.Fatalf("expected approve hash %q, got %q", executor.result.ApproveTransactionHash, executionItem.ApproveTransactionHash)
	}
	if executionItem.Readback["yes_positions"] != "1000000" {
		t.Fatalf("expected readback yes_positions, got %#v", executionItem.Readback)
	}
}

func TestAgentIntentExecutionsIncludesFailureFields(t *testing.T) {
	durableRegistry := newTestDurableAgentIntentRegistry()
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	executor := &stubAgentExecutor{err: errors.New("upstream secret detail")}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor, durableRegistry)

	intentID := createAgentIntent(t, router, `{
		"agent_id": "agent_test_1",
		"source_client": "test_client",
		"client_request_id": "client_req_execution_lookup_failure",
		"action": "create_market",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "agent-market-execution-lookup",
		"question": "Will SignalArc expose execution failure history?",
		"close_timestamp": "1767225600",
		"resolver": "0x2222222222222222222222222222222222222222",
		"collateral_token": "0x3333333333333333333333333333333333333333"
	}`)
	confirmAgentIntent(t, router, intentID)

	executeResponse := httptest.NewRecorder()
	executeRequest := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(executeResponse, executeRequest)
	if executeResponse.Code != http.StatusBadGateway {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusBadGateway, executeResponse.Code, executeResponse.Body.String())
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/agent/intents/"+intentID+"/executions", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected executions status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var body struct {
		Executions []agentActivityItemResponse `json:"executions"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode executions response: %v", err)
	}
	if len(body.Executions) != 1 {
		t.Fatalf("expected one execution, got %#v", body.Executions)
	}
	if body.Executions[0].ErrorCode != "agent_execution_failed" {
		t.Fatalf("expected sanitized error code, got %q", body.Executions[0].ErrorCode)
	}
	if strings.Contains(body.Executions[0].ErrorMessage, "secret detail") {
		t.Fatalf("expected sanitized error message, got %q", body.Executions[0].ErrorMessage)
	}
}

func TestExecuteUnconfirmedIntentReturnsConflict(t *testing.T) {
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	executor := &stubAgentExecutor{}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor)

	intentID := createAgentIntent(t, router, `{
		"action": "create_market",
		"agent_id": "agent_test_1",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "agent-market-execute-2",
		"question": "Will SignalArc reject unconfirmed execution?",
		"close_timestamp": "1767225600",
		"resolver": "0x2222222222222222222222222222222222222222",
		"collateral_token": "0x3333333333333333333333333333333333333333"
	}`)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusConflict, response.Code, response.Body.String())
	}
	if executor.called {
		t.Fatal("executor should not be called for unconfirmed intent")
	}
}

func TestExecuteMissingAgentWalletReturnsBadRequest(t *testing.T) {
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	executor := &stubAgentExecutor{}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor)

	intentID := createAgentIntent(t, router, `{
		"action": "buy_yes",
		"agent_id": "agent_missing",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "market-1",
		"market_contract_address": "0x3333333333333333333333333333333333333333",
		"amount": "1000000"
	}`)
	confirmAgentIntent(t, router, intentID)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	if executor.called {
		t.Fatal("executor should not be called when the agent wallet is missing")
	}
}

func TestExecuteRejectsDeployerWalletAsAgentWallet(t *testing.T) {
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	_, err := walletRegistry.RegisterAgentWallet(context.Background(), repository.UpsertAgentWalletInput{
		AgentID:            "agent_test_1",
		UserWallet:         "0x1111111111111111111111111111111111111111",
		AgentWalletAddress: knownDeployerResolverWallet(),
		WalletProvider:     agent.WalletProviderCircleAgentWallet,
		Chain:              agent.ChainArcTestnet,
		AllowedActions:     []string{agent.ActionBuyYes},
		Status:             agent.WalletStatusActive,
	})
	if err != nil {
		t.Fatalf("register agent wallet: %v", err)
	}
	executor := &stubAgentExecutor{}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor)

	intentID := createValidAgentIntent(t, router)
	confirmAgentIntent(t, router, intentID)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusForbidden, response.Code, response.Body.String())
	}
	if executor.called {
		t.Fatal("executor should not be called for deployer wallet")
	}
}

func TestExecuteRejectsInactiveAgentWallet(t *testing.T) {
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionBuyYes)
	wallet, err := walletRegistry.DisableAgentWallet(context.Background(), "agent_test_1")
	if err != nil {
		t.Fatalf("disable test wallet: %v", err)
	}
	if wallet.Status != "disabled" {
		t.Fatalf("expected disabled wallet, got %q", wallet.Status)
	}
	executor := &stubAgentExecutor{}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor)

	intentID := createValidAgentIntent(t, router)
	confirmAgentIntent(t, router, intentID)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusForbidden, response.Code, response.Body.String())
	}
	if executor.called {
		t.Fatal("executor should not be called for inactive wallet")
	}
}

func TestExecuteRejectsDisallowedAction(t *testing.T) {
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionBuyYes)
	executor := &stubAgentExecutor{}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor)

	intentID := createValidAgentIntent(t, router)
	confirmAgentIntent(t, router, intentID)
	wallet := walletRegistry.wallets["agent_test_1"]
	wallet.AllowedActions = []string{agent.ActionBuyNo}
	walletRegistry.wallets["agent_test_1"] = wallet

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusForbidden, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("agent_action_forbidden")) {
		t.Fatalf("expected agent_action_forbidden, got %s", response.Body.String())
	}
	if executor.called {
		t.Fatal("executor should not be called for disallowed action")
	}
}

func TestExecuteDurableIntentRejectsDisallowedRegisteredWalletAction(t *testing.T) {
	durableRegistry := newTestDurableAgentIntentRegistry()
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionBuyYes)
	executor := &stubAgentExecutor{requiredAllowedAction: agent.ActionBuyYes}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor, durableRegistry)

	intentID := createAgentIntent(t, router, `{
		"agent_id": "agent_test_1",
		"source_client": "hermes-telegram",
		"client_request_id": "buy-yes-durable-disallowed",
		"action": "buy_yes",
		"market_id": "market-1",
		"market_contract_address": "0x3333333333333333333333333333333333333333",
		"amount": "1"
	}`)
	confirmAgentIntent(t, router, intentID)
	wallet := walletRegistry.wallets["agent_test_1"]
	wallet.AllowedActions = []string{agent.ActionBuyNo}
	walletRegistry.wallets["agent_test_1"] = wallet

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusForbidden, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("agent_action_forbidden")) {
		t.Fatalf("expected agent_action_forbidden, got %s", response.Body.String())
	}
	if executor.called {
		t.Fatal("executor should not be called for durable disallowed action")
	}
}

func TestCreateIntentRejectsDisallowedAction(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionBuyNo)
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), walletRegistry, nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents", bytes.NewBufferString(`{
		"action": "buy_yes",
		"agent_id": "agent_test_1",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "market-1",
		"market_contract_address": "0x3333333333333333333333333333333333333333",
		"amount": "12.5"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("agent_action_forbidden")) {
		t.Fatalf("expected agent_action_forbidden, got %s", response.Body.String())
	}
}

func TestCreateIntentRejectsInvalidBuyAmount(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents", bytes.NewBufferString(`{
		"action": "buy_yes",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "market-1",
		"market_contract_address": "0x3333333333333333333333333333333333333333",
		"amount": "0"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}

	var body struct {
		Intent agentIntentResponse `json:"intent"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Intent.ValidationResult.Valid {
		t.Fatal("expected invalid validation result")
	}
	if !containsString(body.Intent.ValidationResult.Errors, "amount must be positive") {
		t.Fatalf("expected amount validation error, got %#v", body.Intent.ValidationResult.Errors)
	}
}

func TestCreateIntentRejectsMaxTradeAmountPolicyViolation(t *testing.T) {
	walletRegistry := newTestAgentWalletRegistry()
	_, err := walletRegistry.RegisterAgentWallet(context.Background(), repository.UpsertAgentWalletInput{
		AgentID:            "agent_test_1",
		UserWallet:         "0x1111111111111111111111111111111111111111",
		AgentWalletAddress: "0x9999999999999999999999999999999999999999",
		WalletProvider:     agent.WalletProviderCircleAgentWallet,
		Chain:              agent.ChainArcTestnet,
		AllowedActions:     []string{agent.ActionBuyYes},
		Status:             agent.WalletStatusActive,
		PolicyMetadata:     json.RawMessage(`{"max_trade_amount":"10"}`),
	})
	if err != nil {
		t.Fatalf("register agent wallet: %v", err)
	}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), walletRegistry, nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents", bytes.NewBufferString(`{
		"action": "buy_yes",
		"agent_id": "agent_test_1",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "market-1",
		"market_contract_address": "0x3333333333333333333333333333333333333333",
		"amount": "12.5"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("agent_policy_violation")) {
		t.Fatalf("expected agent_policy_violation, got %s", response.Body.String())
	}
}

func TestExecuteCircleProviderDisabledReturnsServiceUnavailable(t *testing.T) {
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionBuyYes)
	executor := agent.NewCircleCLIExecutor(agent.CircleCLIExecutorConfig{
		Enabled: false,
	})
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor)

	intentID := createValidAgentIntent(t, router)
	confirmAgentIntent(t, router, intentID)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusServiceUnavailable, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("agent_execution_provider_disabled")) {
		t.Fatalf("expected provider disabled code, got %s", response.Body.String())
	}
}

func TestExecuteConfirmedBuyYesReturnsRealExecutionShape(t *testing.T) {
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionBuyYes)
	executor := &stubAgentExecutor{
		result: agent.ExecutionResult{
			IntentID:               "set-by-test",
			AgentID:                "agent_test_1",
			AgentWalletAddress:     "0x9999999999999999999999999999999999999999",
			WalletProvider:         agent.WalletProviderCircleAgentWallet,
			Action:                 agent.ActionBuyYes,
			Status:                 agent.StatusExecuted,
			ExecutionMode:          agent.ExecutionModeAgentContract,
			Network:                agent.NetworkArcTestnet,
			MarketContractAddress:  "0x3333333333333333333333333333333333333333",
			BroadcastPerformed:     true,
			ApproveTransactionHash: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			TransactionHash:        "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			Readback: agent.ExecutionReadback{
				YesPositions:    "1000000",
				TotalYes:        "1000000",
				TotalCollateral: "1000000",
				USDCBalance:     "1000000",
			},
		},
	}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor)

	intentID := createValidAgentIntent(t, router)
	confirmAgentIntent(t, router, intentID)
	executor.result.IntentID = intentID

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !executor.called {
		t.Fatal("expected executor to be called")
	}
	if executor.intent.Action != agent.ActionBuyYes {
		t.Fatalf("expected buy_yes intent, got %q", executor.intent.Action)
	}

	var body struct {
		Execution agentExecutionResponse `json:"execution"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode execute response: %v", err)
	}

	if body.Execution.Action != agent.ActionBuyYes {
		t.Fatalf("expected buy_yes action, got %q", body.Execution.Action)
	}
	if body.Execution.Status != agent.StatusExecuted {
		t.Fatalf("expected executed status, got %q", body.Execution.Status)
	}
	if !body.Execution.BroadcastPerformed {
		t.Fatal("expected broadcast_performed true")
	}
	if body.Execution.ApproveTransactionHash != executor.result.ApproveTransactionHash {
		t.Fatalf("unexpected approve transaction hash %q", body.Execution.ApproveTransactionHash)
	}
	if body.Execution.TransactionHash != executor.result.TransactionHash {
		t.Fatalf("unexpected buyYes transaction hash %q", body.Execution.TransactionHash)
	}
	if body.Execution.MarketContractAddress != executor.result.MarketContractAddress {
		t.Fatalf("unexpected market address %q", body.Execution.MarketContractAddress)
	}
	if body.Execution.Readback.YesPositions != "1000000" {
		t.Fatalf("expected yes positions 1000000, got %q", body.Execution.Readback.YesPositions)
	}
	if body.Execution.Readback.TotalYes != "1000000" {
		t.Fatalf("expected total yes 1000000, got %q", body.Execution.Readback.TotalYes)
	}
	if body.Execution.Readback.TotalCollateral != "1000000" {
		t.Fatalf("expected total collateral 1000000, got %q", body.Execution.Readback.TotalCollateral)
	}
	if body.Execution.Readback.USDCBalance != "1000000" {
		t.Fatalf("expected usdc balance 1000000, got %q", body.Execution.Readback.USDCBalance)
	}
}

func TestExecuteConfirmedBuyNoReturnsRealExecutionShape(t *testing.T) {
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionBuyNo)
	executor := &stubAgentExecutor{
		result: agent.ExecutionResult{
			IntentID:               "set-by-test",
			AgentID:                "agent_test_1",
			AgentWalletAddress:     "0x9999999999999999999999999999999999999999",
			WalletProvider:         agent.WalletProviderCircleAgentWallet,
			Action:                 agent.ActionBuyNo,
			Status:                 agent.StatusExecuted,
			ExecutionMode:          agent.ExecutionModeAgentContract,
			Network:                agent.NetworkArcTestnet,
			MarketContractAddress:  "0x3333333333333333333333333333333333333333",
			BroadcastPerformed:     true,
			ApproveTransactionHash: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			TransactionHash:        "0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
			Readback: agent.ExecutionReadback{
				NoPositions:     "1000000",
				TotalNo:         "1000000",
				TotalCollateral: "1000000",
				USDCBalance:     "1000000",
				USDCAllowance:   "0",
			},
		},
	}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor)

	intentID := createAgentIntent(t, router, `{
		"action": "buy_no",
		"agent_id": "agent_test_1",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "market-1",
		"market_contract_address": "0x3333333333333333333333333333333333333333",
		"amount": "1000000"
	}`)
	confirmAgentIntent(t, router, intentID)
	executor.result.IntentID = intentID

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !executor.called {
		t.Fatal("expected executor to be called")
	}
	if executor.intent.Action != agent.ActionBuyNo {
		t.Fatalf("expected buy_no intent, got %q", executor.intent.Action)
	}

	var body struct {
		Execution agentExecutionResponse `json:"execution"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode execute response: %v", err)
	}

	if body.Execution.Action != agent.ActionBuyNo {
		t.Fatalf("expected buy_no action, got %q", body.Execution.Action)
	}
	if body.Execution.Status != agent.StatusExecuted {
		t.Fatalf("expected executed status, got %q", body.Execution.Status)
	}
	if !body.Execution.BroadcastPerformed {
		t.Fatal("expected broadcast_performed true")
	}
	if body.Execution.ApproveTransactionHash != executor.result.ApproveTransactionHash {
		t.Fatalf("unexpected approve transaction hash %q", body.Execution.ApproveTransactionHash)
	}
	if body.Execution.TransactionHash != executor.result.TransactionHash {
		t.Fatalf("unexpected buyNo transaction hash %q", body.Execution.TransactionHash)
	}
	if body.Execution.MarketContractAddress != executor.result.MarketContractAddress {
		t.Fatalf("unexpected market address %q", body.Execution.MarketContractAddress)
	}
	if body.Execution.Readback.NoPositions != "1000000" {
		t.Fatalf("expected no positions 1000000, got %q", body.Execution.Readback.NoPositions)
	}
	if body.Execution.Readback.TotalNo != "1000000" {
		t.Fatalf("expected total no 1000000, got %q", body.Execution.Readback.TotalNo)
	}
	if body.Execution.Readback.TotalCollateral != "1000000" {
		t.Fatalf("expected total collateral 1000000, got %q", body.Execution.Readback.TotalCollateral)
	}
	if body.Execution.Readback.USDCBalance != "1000000" {
		t.Fatalf("expected usdc balance 1000000, got %q", body.Execution.Readback.USDCBalance)
	}
	if body.Execution.Readback.USDCAllowance != "0" {
		t.Fatalf("expected usdc allowance 0, got %q", body.Execution.Readback.USDCAllowance)
	}
}

func TestExecuteConfirmedCancelMarketReturnsExecutionShape(t *testing.T) {
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCancelMarket)
	hasClaimed := false
	executor := &stubAgentExecutor{
		result: agent.ExecutionResult{
			IntentID:              "set-by-test",
			AgentID:               "agent_test_1",
			AgentWalletAddress:    "0x9999999999999999999999999999999999999999",
			WalletProvider:        agent.WalletProviderCircleAgentWallet,
			Action:                agent.ActionCancelMarket,
			Status:                agent.StatusExecuted,
			ExecutionMode:         agent.ExecutionModeCircleAgentWalletCLI,
			Network:               agent.NetworkArcTestnet,
			MarketContractAddress: "0x3333333333333333333333333333333333333333",
			BroadcastPerformed:    true,
			TransactionHash:       "0xdddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
			Readback: agent.ExecutionReadback{
				MarketStatus:    "3",
				ClaimableRefund: "2000000",
				HasClaimed:      &hasClaimed,
				USDCBalance:     "2000000",
			},
		},
	}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor)

	intentID := createAgentIntent(t, router, `{
		"action": "cancel_market",
		"agent_id": "agent_test_1",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "market-1",
		"market_contract_address": "0x3333333333333333333333333333333333333333"
	}`)
	confirmAgentIntent(t, router, intentID)
	executor.result.IntentID = intentID

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !executor.called {
		t.Fatal("expected executor to be called for lifecycle action")
	}
	if executor.intent.Action != agent.ActionCancelMarket {
		t.Fatalf("expected cancel_market intent, got %q", executor.intent.Action)
	}

	var body struct {
		Execution agentExecutionResponse `json:"execution"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode execute response: %v", err)
	}

	if body.Execution.Action != agent.ActionCancelMarket {
		t.Fatalf("expected cancel_market action, got %q", body.Execution.Action)
	}
	if body.Execution.Readback.MarketStatus != "3" {
		t.Fatalf("expected market status 3, got %q", body.Execution.Readback.MarketStatus)
	}
	if body.Execution.Readback.ClaimableRefund != "2000000" {
		t.Fatalf("expected claimable refund 2000000, got %q", body.Execution.Readback.ClaimableRefund)
	}
	if body.Execution.Readback.HasClaimed == nil || *body.Execution.Readback.HasClaimed {
		t.Fatalf("expected has_claimed false, got %#v", body.Execution.Readback.HasClaimed)
	}
}

func TestExecuteReportsConfigErrorWithoutSecretDetails(t *testing.T) {
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	executor := &stubAgentExecutor{err: agent.ErrExecutionConfigInvalid}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor)

	intentID := createAgentIntent(t, router, `{
		"action": "create_market",
		"agent_id": "agent_test_1",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "agent-market-execute-3",
		"question": "Will SignalArc hide execution config details?",
		"close_timestamp": "1767225600",
		"resolver": "0x2222222222222222222222222222222222222222",
		"collateral_token": "0x3333333333333333333333333333333333333333"
	}`)
	confirmAgentIntent(t, router, intentID)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusServiceUnavailable, response.Code, response.Body.String())
	}
	if !executor.called {
		t.Fatal("expected executor to be called")
	}
	if bytes.Contains(response.Body.Bytes(), []byte("private")) {
		t.Fatalf("response should not expose private-key details: %s", response.Body.String())
	}
}

func TestExecuteMapsExecutorNotImplemented(t *testing.T) {
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	executor := &stubAgentExecutor{err: agent.ErrExecutionNotImplemented}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor)

	intentID := createAgentIntent(t, router, `{
		"action": "create_market",
		"agent_id": "agent_test_1",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "agent-market-execute-4",
		"question": "Will SignalArc map executor not implemented?",
		"close_timestamp": "1767225600",
		"resolver": "0x2222222222222222222222222222222222222222",
		"collateral_token": "0x3333333333333333333333333333333333333333"
	}`)
	confirmAgentIntent(t, router, intentID)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotImplemented {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusNotImplemented, response.Code, response.Body.String())
	}
}

func TestExecuteMapsUnexpectedExecutorError(t *testing.T) {
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	executor := &stubAgentExecutor{err: errors.New("rpc unavailable")}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor)

	intentID := createAgentIntent(t, router, `{
		"action": "create_market",
		"agent_id": "agent_test_1",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "agent-market-execute-5",
		"question": "Will SignalArc map executor errors?",
		"close_timestamp": "1767225600",
		"resolver": "0x2222222222222222222222222222222222222222",
		"collateral_token": "0x3333333333333333333333333333333333333333"
	}`)
	confirmAgentIntent(t, router, intentID)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("expected execute status %d, got %d: %s", http.StatusBadGateway, response.Code, response.Body.String())
	}
}

func TestExecuteStaleCloseTimestampReturnsBadRequest(t *testing.T) {
	store := agent.NewStore()
	walletRegistry := newTestAgentWalletRegistry()
	registerTestAgentWallet(t, walletRegistry, agent.ActionCreateMarket)
	executor := &stubAgentExecutor{err: agent.ErrCreateMarketCloseTimestampStale}
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, store, walletRegistry, executor)

	intentID := createAgentIntent(t, router, `{
		"action": "create_market",
		"agent_id": "agent_test_1",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "agent-market-stale-ts",
		"question": "Will this market have a stale timestamp?",
		"close_timestamp": "1767225600",
		"resolver": "0x2222222222222222222222222222222222222222",
		"collateral_token": "0x3333333333333333333333333333333333333333"
	}`)
	confirmAgentIntent(t, router, intentID)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/execute", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d for stale timestamp, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	if !executor.called {
		t.Fatal("expected executor to be called")
	}

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Error.Code != "create_market_close_timestamp_stale" {
		t.Fatalf("expected error code create_market_close_timestamp_stale, got %q", body.Error.Code)
	}
}

func TestConfirmBuyYesReturnsMarketTransactionRequest(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)
	marketAddress := "0x4444444444444444444444444444444444444444"

	intentID := createAgentIntent(t, router, `{
		"action": "buy_yes",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "market-1",
		"market_contract_address": "`+marketAddress+`",
		"amount": "42.5"
	}`)

	executionPlan := confirmAgentIntent(t, router, intentID)

	if executionPlan.TransactionRequest.To != marketAddress {
		t.Fatalf("expected market address as to, got %q", executionPlan.TransactionRequest.To)
	}
	if executionPlan.TransactionRequest.Contract != "SignalArcAgentMarket" {
		t.Fatalf("expected agent market contract, got %q", executionPlan.TransactionRequest.Contract)
	}
	if executionPlan.TransactionRequest.Function != "buyYes" {
		t.Fatalf("expected buyYes function, got %q", executionPlan.TransactionRequest.Function)
	}
	assertStringSliceEqual(t, executionPlan.TransactionRequest.Args, []string{"42.5"})
	assertNoExecutionClaim(t, executionPlan)
}

func TestConfirmClaimRefundReturnsMarketTransactionRequest(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)
	marketAddress := "0x5555555555555555555555555555555555555555"

	intentID := createAgentIntent(t, router, `{
		"action": "claim_refund",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "market-1",
		"market_contract_address": "`+marketAddress+`"
	}`)

	executionPlan := confirmAgentIntent(t, router, intentID)

	if executionPlan.TransactionRequest.To != marketAddress {
		t.Fatalf("expected market address as to, got %q", executionPlan.TransactionRequest.To)
	}
	if executionPlan.TransactionRequest.Contract != "SignalArcAgentMarket" {
		t.Fatalf("expected agent market contract, got %q", executionPlan.TransactionRequest.Contract)
	}
	if executionPlan.TransactionRequest.Function != "claimRefund" {
		t.Fatalf("expected claimRefund function, got %q", executionPlan.TransactionRequest.Function)
	}
	assertStringSliceEqual(t, executionPlan.TransactionRequest.Args, []string{})
	assertNoExecutionClaim(t, executionPlan)
}

func TestBuyYesMissingMarketContractAddressFailsValidation(t *testing.T) {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents", bytes.NewBufferString(`{
		"action": "buy_yes",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "market-1",
		"amount": "12.5"
	}`))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	var body struct {
		Intent agentIntentResponse `json:"intent"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Intent.ValidationResult.Valid {
		t.Fatal("expected invalid validation result")
	}
	if !containsString(body.Intent.ValidationResult.Errors, "market_contract_address is required for existing market contract actions") {
		t.Fatalf("expected market_contract_address validation error, got %#v", body.Intent.ValidationResult.Errors)
	}
}

func createValidAgentIntent(t *testing.T, router http.Handler) string {
	t.Helper()

	return createAgentIntent(t, router, `{
		"action": "buy_yes",
		"agent_id": "agent_test_1",
		"user_wallet": "0x1111111111111111111111111111111111111111",
		"market_id": "market-1",
		"market_contract_address": "0x3333333333333333333333333333333333333333",
		"amount": "12.5"
	}`)
}

func createAgentIntent(t *testing.T, router http.Handler, payload string) string {
	t.Helper()

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents", bytes.NewBufferString(payload))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var body struct {
		Intent agentIntentResponse `json:"intent"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	return body.Intent.IntentID
}

func assertAgentWalletRegistrationFails(t *testing.T, payload string) {
	t.Helper()

	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/wallets", bytes.NewBufferString(payload))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func assertAgentOnboardingRegistrationFails(t *testing.T, payload string) {
	t.Helper()

	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/onboarding/register", bytes.NewBufferString(payload))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func newVerifyEnabledRouter(sessionRegistry *testAgentSessionRegistry, runner *stubCircleOnboardingRunner, store *agent.CircleOTPRequestStore) http.Handler {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), newTestAgentWalletRegistry(), nil, sessionRegistry, agent.CircleOnboardingStarter{
		Enabled:      true,
		Runner:       runner,
		RequestStore: store,
	})
	return router
}

func newVerifyEnabledRouterWithWalletRegistryAndResolver(sessionRegistry *testAgentSessionRegistry, walletRegistry *testAgentWalletRegistry, runner *stubCircleOnboardingRunner, store *agent.CircleOTPRequestStore, resolver agent.CircleWalletResolver) http.Handler {
	router := chi.NewRouter()
	registerAgentIntentRoutes(router, agent.NewStore(), walletRegistry, nil, sessionRegistry, agent.CircleOnboardingStarter{
		Enabled:      true,
		Runner:       runner,
		RequestStore: store,
	}, resolver)
	return router
}

func insertTestOnboardingSession(registry *testAgentSessionRegistry, onboardingID string) repository.AgentOnboardingSession {
	onboarding, err := registry.CreateAgentOnboardingSession(context.Background(), repository.CreateAgentOnboardingSessionInput{
		OnboardingID:   onboardingID,
		AgentID:        "agent_verify_test",
		UserEmail:      "desi@example.com",
		UserWallet:     nullableString("0x1111111111111111111111111111111111111111"),
		Chain:          agent.ChainArcTestnet,
		WalletProvider: agent.WalletProviderCircleAgentWallet,
		Status:         repository.AgentOnboardingStatusPendingOTP,
		PolicyMetadata: json.RawMessage(`{"note":"test"}`),
	})
	if err != nil {
		panic(err)
	}
	return onboarding
}

func registerTestAgentWallet(t *testing.T, registry *testAgentWalletRegistry, allowedActions ...string) {
	t.Helper()

	_, err := registry.RegisterAgentWallet(context.Background(), repository.UpsertAgentWalletInput{
		AgentID:            "agent_test_1",
		UserWallet:         "0x1111111111111111111111111111111111111111",
		AgentWalletAddress: "0x9999999999999999999999999999999999999999",
		WalletProvider:     agent.WalletProviderCircleAgentWallet,
		Chain:              agent.ChainArcTestnet,
		AllowedActions:     allowedActions,
		Status:             agent.WalletStatusActive,
		PolicyMetadata:     json.RawMessage(`{"source":"test"}`),
	})
	if err != nil {
		t.Fatalf("register test agent wallet: %v", err)
	}
}

func confirmAgentIntent(t *testing.T, router http.Handler, intentID string) agentExecutionPlanResponse {
	t.Helper()

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/agent/intents/"+intentID+"/confirm", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected confirm status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var body struct {
		ExecutionPlan agentExecutionPlanResponse `json:"execution_plan"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode confirm response: %v", err)
	}

	return body.ExecutionPlan
}

func assertNoExecutionClaim(t *testing.T, executionPlan agentExecutionPlanResponse) {
	t.Helper()

	if executionPlan.BroadcastPerformed {
		t.Fatal("expected broadcast_performed false")
	}
	if executionPlan.TransactionHash != nil {
		t.Fatalf("expected transaction_hash null, got %q", *executionPlan.TransactionHash)
	}
	if executionPlan.TransactionRequest.BroadcastPerformed {
		t.Fatal("expected transaction_request broadcast_performed false")
	}
	if executionPlan.TransactionRequest.Value != "0" {
		t.Fatalf("expected transaction value 0, got %q", executionPlan.TransactionRequest.Value)
	}
	if executionPlan.TransactionRequest.Chain != "arc_testnet" {
		t.Fatalf("expected transaction chain arc_testnet, got %q", executionPlan.TransactionRequest.Chain)
	}
}

func assertStringSliceEqual(t *testing.T, actual []string, expected []string) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("expected args %#v, got %#v", expected, actual)
	}
	for index := range expected {
		if actual[index] != expected[index] {
			t.Fatalf("expected args %#v, got %#v", expected, actual)
		}
	}
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}

	return false
}
