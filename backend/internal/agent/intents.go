package agent

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ActionCreateMarket  = "create_market"
	ActionBuyYes        = "buy_yes"
	ActionBuyNo         = "buy_no"
	ActionCancelMarket  = "cancel_market"
	ActionCloseMarket   = "close_market"
	ActionResolveMarket = "resolve_market"
	ActionClaimRefund   = "claim_refund"
	ActionClaimPayout   = "claim_payout"

	StatusPreview   = "preview"
	StatusConfirmed = "confirmed"
	StatusExecuted  = "executed"

	ExecutionModeAgentContract            = "agent_contract"
	ExecutionModeCircleAgentWalletCLI     = "circle_agent_wallet_cli"
	ExecutionModeCircleDeveloperWalletAPI = "circle_developer_wallet_api"
	NetworkArcTestnet                     = "arc_testnet"
	AgentFactoryAddress                   = "0x69aE770e8b2F96297101FeC4dc123B3801dA7d80"

	WalletProviderCircleAgentWallet        = "circle_agent_wallet"
	WalletProviderTemporaryTestnetAgentEOA = "temporary_testnet_agent_eoa"
	WalletStatusActive                     = "active"
	ChainArcTestnet                        = "ARC-TESTNET"
)

var (
	ErrIntentNotFound       = errors.New("agent intent not found")
	ErrIntentInvalid        = errors.New("agent intent is invalid")
	ErrAgentWalletNotFound  = errors.New("agent wallet not found")
	ErrAgentWalletForbidden = errors.New("agent wallet is forbidden")
)

type CreateIntentInput struct {
	AgentID               string
	AgentWalletAddress    string
	WalletProvider        string
	AllowedActions        []string
	SourceClient          string
	ClientRequestID       string
	Action                string
	UserWallet            string
	MarketID              string
	MarketContractAddress string
	Amount                string
	Outcome               string
	Resolver              string
	CollateralToken       string
	CloseTimestamp        string
	Question              string
	PolicyMetadata        map[string]string
}

type AgentWallet struct {
	AgentID            string
	UserWallet         string
	AgentWalletAddress string
	WalletProvider     string
	Chain              string
	AllowedActions     []string
	Status             string
	PolicyMetadata     map[string]string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type Intent struct {
	ID                    string
	AgentID               string
	AgentWalletAddress    string
	WalletProvider        string
	AllowedActions        []string
	SourceClient          string
	ClientRequestID       string
	Action                string
	Status                string
	RequiresConfirmation  bool
	UserWallet            string
	MarketID              string
	MarketContractAddress string
	Amount                string
	Outcome               string
	Resolver              string
	CollateralToken       string
	CloseTimestamp        string
	Question              string
	PolicyMetadata        map[string]string
	ValidationResult      ValidationResult
	Warnings              []string
	CreatedAt             time.Time
}

type ExecutionPlan struct {
	IntentID            string
	Action              string
	Status              string
	ExecutionMode       string
	Network             string
	AgentFactoryAddress string
	RequiresSignature   bool
	BroadcastPerformed  bool
	TransactionHash     *string
	TransactionRequest  TransactionRequest
	Warnings            []string
}

type TransactionRequest struct {
	To                 string
	Contract           string
	Function           string
	Args               []string
	Value              string
	Chain              string
	BroadcastPerformed bool
}

type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}

type Store struct {
	mu      sync.RWMutex
	intents map[string]Intent
	wallets map[string]AgentWallet
	now     func() time.Time
	newID   func() (string, error)
}

func NewStore() *Store {
	return &Store{
		intents: make(map[string]Intent),
		wallets: make(map[string]AgentWallet),
		now:     time.Now,
		newID:   newIntentID,
	}
}

func (store *Store) RegisterAgentWallet(wallet AgentWallet) (AgentWallet, error) {
	normalized := normalizeAgentWallet(wallet)
	validation := validateAgentWallet(normalized)
	if !validation.Valid {
		return AgentWallet{}, ErrIntentInvalid
	}

	now := store.now().UTC()
	if normalized.CreatedAt.IsZero() {
		normalized.CreatedAt = now
	}
	normalized.UpdatedAt = now

	store.mu.Lock()
	defer store.mu.Unlock()
	store.wallets[normalized.AgentID] = normalized

	return normalized, nil
}

func (store *Store) GetAgentWallet(agentID string) (AgentWallet, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	wallet, ok := store.wallets[strings.TrimSpace(agentID)]
	if !ok {
		return AgentWallet{}, ErrAgentWalletNotFound
	}

	return wallet, nil
}

func (store *Store) CreateIntent(input CreateIntentInput) (Intent, error) {
	normalized := normalizeInput(input)
	validationResult := validateIntent(normalized)
	if normalized.AgentID != "" {
		if wallet, err := store.GetAgentWallet(normalized.AgentID); err == nil {
			if normalized.AgentWalletAddress == "" {
				normalized.AgentWalletAddress = wallet.AgentWalletAddress
			}
			if normalized.WalletProvider == "" {
				normalized.WalletProvider = wallet.WalletProvider
			}
			if len(normalized.AllowedActions) == 0 {
				normalized.AllowedActions = append([]string{}, wallet.AllowedActions...)
			}
			if len(normalized.PolicyMetadata) == 0 {
				normalized.PolicyMetadata = normalizeStringMap(wallet.PolicyMetadata)
			}
		}
	}
	intentID, err := store.newID()
	if err != nil {
		return Intent{}, err
	}

	intent := Intent{
		ID:                    intentID,
		AgentID:               normalized.AgentID,
		AgentWalletAddress:    normalized.AgentWalletAddress,
		WalletProvider:        normalized.WalletProvider,
		AllowedActions:        append([]string{}, normalized.AllowedActions...),
		SourceClient:          normalized.SourceClient,
		ClientRequestID:       normalized.ClientRequestID,
		Action:                normalized.Action,
		Status:                StatusPreview,
		RequiresConfirmation:  true,
		UserWallet:            normalized.UserWallet,
		MarketID:              normalized.MarketID,
		MarketContractAddress: normalized.MarketContractAddress,
		Amount:                normalized.Amount,
		Outcome:               normalized.Outcome,
		Resolver:              normalized.Resolver,
		CollateralToken:       normalized.CollateralToken,
		CloseTimestamp:        normalized.CloseTimestamp,
		Question:              normalized.Question,
		PolicyMetadata:        normalizeStringMap(normalized.PolicyMetadata),
		ValidationResult:      validationResult,
		Warnings: []string{
			"preview only; no transaction has been executed",
			"execution requires a registered agent wallet controlled separately from the user/deployer wallet",
			"Circle Agent Wallet contract execution must be enabled and configured before backend execution is available",
		},
		CreatedAt: store.now().UTC(),
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	store.intents[intent.ID] = intent

	return intent, nil
}

func (store *Store) GetIntent(id string) (Intent, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	intent, ok := store.intents[strings.TrimSpace(id)]
	if !ok {
		return Intent{}, ErrIntentNotFound
	}

	return intent, nil
}

func (store *Store) ConfirmIntent(id string) (ExecutionPlan, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	intent, ok := store.intents[strings.TrimSpace(id)]
	if !ok {
		return ExecutionPlan{}, ErrIntentNotFound
	}

	if !intent.ValidationResult.Valid {
		return ExecutionPlan{}, ErrIntentInvalid
	}

	intent.Status = StatusConfirmed
	store.intents[intent.ID] = intent

	return NewExecutionPlan(intent), nil
}

func NewExecutionPlan(intent Intent) ExecutionPlan {
	return ExecutionPlan{
		IntentID:            intent.ID,
		Action:              intent.Action,
		Status:              StatusConfirmed,
		ExecutionMode:       ExecutionModeAgentContract,
		Network:             NetworkArcTestnet,
		AgentFactoryAddress: AgentFactoryAddress,
		RequiresSignature:   true,
		BroadcastPerformed:  false,
		TransactionHash:     nil,
		TransactionRequest:  NewTransactionRequest(intent),
		Warnings: []string{
			"confirmation produced an execution plan only; no transaction has been executed",
			"no private key, signing, RPC call, or broadcast was performed",
			"execution must be performed by the registered agent wallet provider, not by the deployer private key",
		},
	}
}

func NewTransactionRequest(intent Intent) TransactionRequest {
	transactionRequest := TransactionRequest{
		Value:              "0",
		Chain:              NetworkArcTestnet,
		BroadcastPerformed: false,
	}

	switch intent.Action {
	case ActionCreateMarket:
		transactionRequest.To = AgentFactoryAddress
		transactionRequest.Contract = "SignalArcAgentMarketFactory"
		transactionRequest.Function = "createMarket"
		transactionRequest.Args = []string{
			intent.MarketID,
			intent.Question,
			intent.CloseTimestamp,
			intent.Resolver,
			intent.CollateralToken,
		}
	case ActionBuyYes:
		transactionRequest.To = intent.MarketContractAddress
		transactionRequest.Contract = "SignalArcAgentMarket"
		transactionRequest.Function = "buyYes"
		transactionRequest.Args = []string{intent.Amount}
	case ActionBuyNo:
		transactionRequest.To = intent.MarketContractAddress
		transactionRequest.Contract = "SignalArcAgentMarket"
		transactionRequest.Function = "buyNo"
		transactionRequest.Args = []string{intent.Amount}
	case ActionCancelMarket:
		transactionRequest.To = intent.MarketContractAddress
		transactionRequest.Contract = "SignalArcAgentMarket"
		transactionRequest.Function = "cancelMarket"
		transactionRequest.Args = []string{}
	case ActionCloseMarket:
		transactionRequest.To = intent.MarketContractAddress
		transactionRequest.Contract = "SignalArcAgentMarket"
		transactionRequest.Function = "closeMarket"
		transactionRequest.Args = []string{}
	case ActionResolveMarket:
		transactionRequest.To = intent.MarketContractAddress
		transactionRequest.Contract = "SignalArcAgentMarket"
		transactionRequest.Function = "resolve"
		transactionRequest.Args = []string{intent.Outcome}
	case ActionClaimRefund:
		transactionRequest.To = intent.MarketContractAddress
		transactionRequest.Contract = "SignalArcAgentMarket"
		transactionRequest.Function = "claimRefund"
		transactionRequest.Args = []string{}
	case ActionClaimPayout:
		transactionRequest.To = intent.MarketContractAddress
		transactionRequest.Contract = "SignalArcAgentMarket"
		transactionRequest.Function = "claimPayout"
		transactionRequest.Args = []string{}
	}

	return transactionRequest
}

func normalizeInput(input CreateIntentInput) CreateIntentInput {
	return CreateIntentInput{
		AgentID:               strings.TrimSpace(input.AgentID),
		AgentWalletAddress:    strings.TrimSpace(input.AgentWalletAddress),
		WalletProvider:        strings.TrimSpace(input.WalletProvider),
		AllowedActions:        normalizeStringSlice(input.AllowedActions),
		SourceClient:          strings.TrimSpace(input.SourceClient),
		ClientRequestID:       strings.TrimSpace(input.ClientRequestID),
		Action:                strings.TrimSpace(input.Action),
		UserWallet:            strings.TrimSpace(input.UserWallet),
		MarketID:              strings.TrimSpace(input.MarketID),
		MarketContractAddress: strings.TrimSpace(input.MarketContractAddress),
		Amount:                strings.TrimSpace(input.Amount),
		Outcome:               strings.TrimSpace(input.Outcome),
		Resolver:              strings.TrimSpace(input.Resolver),
		CollateralToken:       strings.TrimSpace(input.CollateralToken),
		CloseTimestamp:        normalizeCloseTimestamp(input.CloseTimestamp),
		Question:              strings.TrimSpace(input.Question),
		PolicyMetadata:        normalizeStringMap(input.PolicyMetadata),
	}
}

// normalizeCloseTimestamp accepts the raw close_timestamp string from an
// agent intent request and returns it normalized to a base-10 Unix-seconds
// string when possible. Inputs that already look like a base-10 integer are
// returned unchanged so existing Custom GPT and CLI clients continue to work.
// Inputs in RFC3339 (or RFC3339 with nanoseconds) form are converted to the
// integer Unix-seconds form expected by the existing on-chain executor and
// SignalArcAgentMarketFactory.createMarket(uint256) signature. Inputs that
// cannot be normalized are returned trimmed so downstream validation can
// produce a stable error.
//
// This helper does not alter contract behavior. It only adds a backwards
// compatible parser for RFC3339 input that can be produced by natural
// language conversion before the request reaches the API layer.
func normalizeCloseTimestamp(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	if isAllDigits(trimmed) {
		return trimmed
	}

	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return strconv.FormatInt(parsed.UTC().Unix(), 10)
		}
	}

	return trimmed
}

func isAllDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func normalizeStringSlice(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			normalized = append(normalized, value)
		}
	}
	return normalized
}

func normalizeStringMap(values map[string]string) map[string]string {
	normalized := make(map[string]string, len(values))
	for key, value := range values {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		normalized[key] = value
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func validateIntent(input CreateIntentInput) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []string{}}

	if input.Action == "" {
		result.Errors = append(result.Errors, "action is required")
	} else if !isSupportedAction(input.Action) {
		result.Errors = append(result.Errors, "action must be one of the supported agent intent actions")
	}

	if isTransactionAction(input.Action) && input.UserWallet == "" {
		result.Errors = append(result.Errors, "user_wallet is required for transaction actions")
	}
	if isMarketSpecificAction(input.Action) && input.MarketID == "" {
		result.Errors = append(result.Errors, "market_id is required for market-specific actions")
	}

	if input.Action == ActionCreateMarket {
		if input.Question == "" {
			result.Errors = append(result.Errors, "question is required for create_market")
		}
		if input.CloseTimestamp == "" {
			result.Errors = append(result.Errors, "close_timestamp is required for create_market")
		} else if !isAllDigits(input.CloseTimestamp) {
			result.Errors = append(result.Errors, "close_timestamp must be a unix-seconds integer or an RFC3339 timestamp such as 2026-05-31T23:59:00Z")
		}
		if input.Resolver == "" {
			result.Errors = append(result.Errors, "resolver is required for create_market")
		}
		if input.CollateralToken == "" {
			result.Errors = append(result.Errors, "collateral_token is required for create_market")
		}
	}

	if requiresMarketContractAddress(input.Action) && input.MarketContractAddress == "" {
		result.Errors = append(result.Errors, "market_contract_address is required for existing market contract actions")
	}

	if input.Action == ActionBuyYes || input.Action == ActionBuyNo {
		if input.Amount == "" {
			result.Errors = append(result.Errors, "amount is required for buy_yes and buy_no")
		} else if !isPositiveDecimal(input.Amount) {
			result.Errors = append(result.Errors, "amount must be positive")
		}
	}

	if input.Action == ActionResolveMarket && input.Outcome == "" {
		result.Errors = append(result.Errors, "outcome is required for resolve_market")
	}

	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result
}

func normalizeAgentWallet(wallet AgentWallet) AgentWallet {
	allowedActions := make([]string, 0, len(wallet.AllowedActions))
	for _, action := range wallet.AllowedActions {
		action = strings.TrimSpace(action)
		if action != "" {
			allowedActions = append(allowedActions, action)
		}
	}

	normalized := AgentWallet{
		AgentID:            strings.TrimSpace(wallet.AgentID),
		UserWallet:         strings.TrimSpace(wallet.UserWallet),
		AgentWalletAddress: strings.TrimSpace(wallet.AgentWalletAddress),
		WalletProvider:     strings.TrimSpace(wallet.WalletProvider),
		Chain:              strings.TrimSpace(wallet.Chain),
		AllowedActions:     allowedActions,
		Status:             strings.TrimSpace(wallet.Status),
		PolicyMetadata:     normalizeStringMap(wallet.PolicyMetadata),
		CreatedAt:          wallet.CreatedAt,
		UpdatedAt:          wallet.UpdatedAt,
	}
	if normalized.Chain == "" {
		normalized.Chain = ChainArcTestnet
	}
	if normalized.Status == "" {
		normalized.Status = WalletStatusActive
	}
	return normalized
}

func validateAgentWallet(wallet AgentWallet) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []string{}}
	if wallet.AgentID == "" {
		result.Errors = append(result.Errors, "agent_id is required")
	}
	if wallet.UserWallet == "" {
		result.Errors = append(result.Errors, "user_wallet is required")
	}
	if wallet.AgentWalletAddress == "" {
		result.Errors = append(result.Errors, "agent_wallet_address is required")
	}
	if wallet.WalletProvider != WalletProviderCircleAgentWallet && wallet.WalletProvider != WalletProviderTemporaryTestnetAgentEOA {
		result.Errors = append(result.Errors, "wallet_provider is unsupported")
	}
	if wallet.Chain != ChainArcTestnet {
		result.Errors = append(result.Errors, "chain must be ARC-TESTNET")
	}
	if len(wallet.AllowedActions) == 0 {
		result.Errors = append(result.Errors, "allowed_actions is required")
	}
	for _, action := range wallet.AllowedActions {
		if !isSupportedAction(action) {
			result.Errors = append(result.Errors, "allowed_actions contains unsupported action")
		}
	}
	if len(result.Errors) > 0 {
		result.Valid = false
	}
	return result
}

func AgentWalletAllowsAction(wallet AgentWallet, action string) bool {
	for _, allowedAction := range wallet.AllowedActions {
		if allowedAction == action {
			return true
		}
	}
	return false
}

func isSupportedAction(action string) bool {
	switch action {
	case ActionCreateMarket,
		ActionBuyYes,
		ActionBuyNo,
		ActionCancelMarket,
		ActionCloseMarket,
		ActionResolveMarket,
		ActionClaimRefund,
		ActionClaimPayout:
		return true
	default:
		return false
	}
}

func isTransactionAction(action string) bool {
	return isSupportedAction(action)
}

func isMarketSpecificAction(action string) bool {
	switch action {
	case ActionBuyYes,
		ActionBuyNo,
		ActionCancelMarket,
		ActionCloseMarket,
		ActionResolveMarket,
		ActionClaimRefund,
		ActionClaimPayout:
		return true
	default:
		return false
	}
}

func requiresMarketContractAddress(action string) bool {
	switch action {
	case ActionBuyYes,
		ActionBuyNo,
		ActionCancelMarket,
		ActionCloseMarket,
		ActionResolveMarket,
		ActionClaimRefund,
		ActionClaimPayout:
		return true
	default:
		return false
	}
}

func isPositiveDecimal(value string) bool {
	decimal, ok := new(big.Rat).SetString(value)
	return ok && decimal.Sign() > 0
}

func newIntentID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return "agent_intent_" + hex.EncodeToString(bytes), nil
}
