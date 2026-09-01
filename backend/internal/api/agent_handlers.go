package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/wahyu241205/SignalArc/backend/internal/agent"
	"github.com/wahyu241205/SignalArc/backend/internal/httpjson"
	"github.com/wahyu241205/SignalArc/backend/internal/repository"
)

type createAgentIntentRequest struct {
	AgentID               string `json:"agent_id"`
	AgentWalletAddress    string `json:"agent_wallet_address"`
	SourceClient          string `json:"source_client"`
	ClientRequestID       string `json:"client_request_id"`
	Action                string `json:"action"`
	UserWallet            string `json:"user_wallet"`
	MarketID              string `json:"market_id"`
	MarketContractAddress string `json:"market_contract_address"`
	Amount                string `json:"amount"`
	Outcome               string `json:"outcome"`
	Resolver              string `json:"resolver"`
	CollateralToken       string `json:"collateral_token"`
	CloseTimestamp        string `json:"close_timestamp"`
	Question              string `json:"question"`
}

type registerAgentWalletRequest struct {
	AgentID            string            `json:"agent_id"`
	UserWallet         string            `json:"user_wallet"`
	UserEmail          string            `json:"user_email"`
	AgentWalletAddress string            `json:"agent_wallet_address"`
	WalletProvider     string            `json:"wallet_provider"`
	Chain              string            `json:"chain"`
	AllowedActions     []string          `json:"allowed_actions"`
	Status             string            `json:"status"`
	PolicyMetadata     map[string]string `json:"policy_metadata"`
	SourceClient       string            `json:"source_client"`
}

type registerAgentOnboardingRequest struct {
	AgentID            string `json:"agent_id"`
	UserWallet         string `json:"user_wallet"`
	AgentWalletAddress string `json:"agent_wallet_address"`
	UserEmail          string `json:"user_email"`
	SourceClient       string `json:"source_client"`
}

type startAgentOnboardingRequest struct {
	AgentID      string `json:"agent_id"`
	UserEmail    string `json:"user_email"`
	UserWallet   string `json:"user_wallet"`
	SourceClient string `json:"source_client"`
	Channel      string `json:"channel"`
}

type verifyAgentOnboardingRequest struct {
	OnboardingID string `json:"onboarding_id"`
	OTP          string `json:"otp"`
}

type agentWalletRegistry interface {
	RegisterAgentWallet(context.Context, repository.UpsertAgentWalletInput) (repository.AgentWallet, error)
	GetAgentWalletByAgentID(context.Context, string) (repository.AgentWallet, error)
	DisableAgentWallet(context.Context, string) (repository.AgentWallet, error)
}

type agentSessionRegistry interface {
	CreateAgentOnboardingSession(context.Context, repository.CreateAgentOnboardingSessionInput) (repository.AgentOnboardingSession, error)
	GetAgentOnboardingSessionByOnboardingID(context.Context, string) (repository.AgentOnboardingSession, error)
	UpdateAgentOnboardingSessionStatus(context.Context, string, string, sql.NullString) (repository.AgentOnboardingSession, error)
	UpdateAgentOnboardingSessionOTPStart(context.Context, string, string, time.Time) (repository.AgentOnboardingSession, error)
	CreateAgentSession(context.Context, repository.CreateAgentSessionInput) (repository.AgentSession, error)
	GetAgentSessionByAgentID(context.Context, string) (repository.AgentSession, error)
	GetAgentSessionBySessionID(context.Context, string) (repository.AgentSession, error)
}

type durableAgentIntentRegistry interface {
	CreateAgentIntent(context.Context, repository.CreateAgentIntentInput) (repository.AgentIntent, error)
	GetAgentIntentByIntentID(context.Context, string) (repository.AgentIntent, error)
	ConfirmAgentIntent(context.Context, string) (repository.AgentIntent, error)
	MarkAgentIntentExecuted(context.Context, string) (repository.AgentIntent, error)
	MarkAgentIntentFailed(context.Context, string) (repository.AgentIntent, error)
	CreateAgentExecution(context.Context, repository.CreateAgentExecutionInput) (repository.AgentExecution, error)
	MarkAgentExecutionExecuted(context.Context, string, repository.CompleteAgentExecutionInput) (repository.AgentExecution, error)
	MarkAgentExecutionFailed(context.Context, string, repository.FailAgentExecutionInput) (repository.AgentExecution, error)
	ListAgentIntentsByAgentID(context.Context, string, int) ([]repository.AgentIntent, error)
	ListAgentExecutionsByAgentID(context.Context, string, int) ([]repository.AgentExecution, error)
	ListAgentExecutionsByIntentID(context.Context, string, int) ([]repository.AgentExecution, error)
}

type agentWalletBalanceResponse struct {
	AgentID            string `json:"agent_id"`
	AgentWalletAddress string `json:"agent_wallet_address"`
	Chain              string `json:"chain"`
	Balances           []any  `json:"balances"`
}

type agentWalletFaucetResponse struct {
	AgentID            string `json:"agent_id"`
	AgentWalletAddress string `json:"agent_wallet_address"`
	Chain              string `json:"chain"`
	Token              string `json:"token"`
	Status             string `json:"status"`
	Result             any    `json:"result"`
}

type agentWalletResponse struct {
	ID                 string            `json:"id"`
	AgentID            string            `json:"agent_id"`
	UserWallet         string            `json:"user_wallet"`
	UserEmail          string            `json:"user_email,omitempty"`
	AgentWalletAddress string            `json:"agent_wallet_address"`
	WalletProvider     string            `json:"wallet_provider"`
	Chain              string            `json:"chain"`
	AllowedActions     []string          `json:"allowed_actions"`
	Status             string            `json:"status"`
	PolicyMetadata     map[string]string `json:"policy_metadata,omitempty"`
	SourceClient       string            `json:"source_client,omitempty"`
	CreatedAt          string            `json:"created_at"`
	UpdatedAt          string            `json:"updated_at"`
}

type agentOnboardingSessionResponse struct {
	OnboardingID                string            `json:"onboarding_id"`
	AgentID                     string            `json:"agent_id"`
	UserEmail                   string            `json:"user_email"`
	UserWallet                  string            `json:"user_wallet,omitempty"`
	RequestedAgentWalletAddress string            `json:"requested_agent_wallet_address,omitempty"`
	SourceClient                string            `json:"source_client,omitempty"`
	Channel                     string            `json:"channel,omitempty"`
	Chain                       string            `json:"chain"`
	WalletProvider              string            `json:"wallet_provider"`
	Status                      string            `json:"status"`
	FailureReason               string            `json:"failure_reason,omitempty"`
	PolicyMetadata              map[string]string `json:"policy_metadata,omitempty"`
	CreatedAt                   string            `json:"created_at"`
	UpdatedAt                   string            `json:"updated_at"`
}

type agentSessionResponse struct {
	SessionID          string            `json:"session_id"`
	AgentID            string            `json:"agent_id"`
	UserEmail          string            `json:"user_email"`
	UserWallet         string            `json:"user_wallet"`
	AgentWalletAddress string            `json:"agent_wallet_address"`
	WalletProvider     string            `json:"wallet_provider"`
	Chain              string            `json:"chain"`
	Status             string            `json:"status"`
	AllowedActions     []string          `json:"allowed_actions"`
	AllowedChannels    []string          `json:"allowed_channels"`
	SessionMetadata    map[string]string `json:"session_metadata,omitempty"`
	LivenessReason     string            `json:"liveness_reason,omitempty"`
	CreatedAt          string            `json:"created_at"`
	UpdatedAt          string            `json:"updated_at"`
}

type agentIntentResponse struct {
	IntentID              string                 `json:"intent_id"`
	AgentID               string                 `json:"agent_id,omitempty"`
	AgentWalletAddress    string                 `json:"agent_wallet_address,omitempty"`
	WalletProvider        string                 `json:"wallet_provider,omitempty"`
	AllowedActions        []string               `json:"allowed_actions,omitempty"`
	SourceClient          string                 `json:"source_client,omitempty"`
	ClientRequestID       string                 `json:"client_request_id,omitempty"`
	Action                string                 `json:"action"`
	Status                string                 `json:"status"`
	RequiresConfirmation  bool                   `json:"requires_confirmation"`
	UserWallet            string                 `json:"user_wallet,omitempty"`
	Address               string                 `json:"address,omitempty"`
	MarketID              string                 `json:"market_id,omitempty"`
	MarketContractAddress string                 `json:"market_contract_address,omitempty"`
	Amount                string                 `json:"amount,omitempty"`
	Outcome               string                 `json:"outcome,omitempty"`
	Resolver              string                 `json:"resolver,omitempty"`
	CollateralToken       string                 `json:"collateral_token,omitempty"`
	CloseTimestamp        string                 `json:"close_timestamp,omitempty"`
	Question              string                 `json:"question,omitempty"`
	ValidationResult      agent.ValidationResult `json:"validation_result"`
	Warnings              []string               `json:"warnings"`
	CreatedAt             string                 `json:"created_at"`
}

type agentExecutionPlanResponse struct {
	IntentID            string                     `json:"intent_id"`
	Action              string                     `json:"action"`
	Status              string                     `json:"status"`
	ExecutionMode       string                     `json:"execution_mode"`
	Network             string                     `json:"network"`
	AgentFactoryAddress string                     `json:"agent_factory_address"`
	RequiresSignature   bool                       `json:"requires_signature"`
	BroadcastPerformed  bool                       `json:"broadcast_performed"`
	TransactionHash     *string                    `json:"transaction_hash"`
	TransactionRequest  transactionRequestResponse `json:"transaction_request"`
	Warnings            []string                   `json:"warnings"`
}

type agentExecutionResponse struct {
	IntentID               string                `json:"intent_id"`
	AgentID                string                `json:"agent_id"`
	AgentWalletAddress     string                `json:"agent_wallet_address"`
	WalletProvider         string                `json:"wallet_provider"`
	Action                 string                `json:"action"`
	Status                 string                `json:"status"`
	ExecutionMode          string                `json:"execution_mode"`
	Network                string                `json:"network"`
	AgentFactoryAddress    string                `json:"agent_factory_address"`
	MarketContractAddress  string                `json:"market_contract_address,omitempty"`
	BroadcastPerformed     bool                  `json:"broadcast_performed"`
	ApproveTransactionHash string                `json:"approve_transaction_hash,omitempty"`
	TransactionHash        string                `json:"transaction_hash"`
	Readback               agentReadbackResponse `json:"readback"`
}

type agentPortfolioResponse struct {
	AgentID                        string                           `json:"agent_id"`
	AgentWalletAddress             string                           `json:"agent_wallet_address"`
	Chain                          string                           `json:"chain"`
	WalletProvider                 string                           `json:"wallet_provider"`
	ActivePositionsCount           int                              `json:"active_positions_count"`
	ResolvedOrClosedPositionsCount int                              `json:"resolved_or_closed_positions_count"`
	ClaimableRefundableCount       int                              `json:"claimable_refundable_count"`
	TotalExposure                  string                           `json:"total_exposure"`
	Positions                      []agentPortfolioPositionResponse `json:"positions"`
	Settlements                    []any                            `json:"settlements"`
	UnavailableFields              []string                         `json:"unavailable_fields"`
}

type agentPortfolioPositionResponse struct {
	Source                string `json:"source"`
	IntentID              string `json:"intent_id"`
	Action                string `json:"action"`
	Status                string `json:"status"`
	MarketID              string `json:"market_id,omitempty"`
	MarketContractAddress string `json:"market_contract_address,omitempty"`
	Amount                string `json:"amount,omitempty"`
	Outcome               string `json:"outcome,omitempty"`
	TransactionHash       string `json:"transaction_hash,omitempty"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
}

type agentActivityResponse struct {
	AgentID string                      `json:"agent_id"`
	Items   []agentActivityItemResponse `json:"items"`
}

type agentActivityItemResponse struct {
	ID                     string         `json:"id"`
	Type                   string         `json:"type"`
	Action                 string         `json:"action"`
	Status                 string         `json:"status"`
	IntentID               string         `json:"intent_id,omitempty"`
	MarketID               string         `json:"market_id,omitempty"`
	MarketTitle            string         `json:"market_title,omitempty"`
	MarketContractAddress  string         `json:"market_contract_address,omitempty"`
	Amount                 string         `json:"amount,omitempty"`
	Outcome                string         `json:"outcome,omitempty"`
	Side                   string         `json:"side,omitempty"`
	TransactionHash        string         `json:"transaction_hash,omitempty"`
	ApproveTransactionHash string         `json:"approve_transaction_hash,omitempty"`
	ErrorCode              string         `json:"error_code,omitempty"`
	ErrorMessage           string         `json:"error_message,omitempty"`
	Readback               map[string]any `json:"readback,omitempty"`
	CreatedAt              string         `json:"created_at"`
	UpdatedAt              string         `json:"updated_at"`
	CompletedAt            string         `json:"completed_at,omitempty"`
}

type agentReadbackResponse struct {
	MarketCount     string `json:"market_count"`
	CreatedMarket   string `json:"created_market,omitempty"`
	IsMarket        *bool  `json:"is_market,omitempty"`
	MarketStatus    string `json:"market_status,omitempty"`
	WinningOutcome  string `json:"winning_outcome,omitempty"`
	YesPositions    string `json:"yes_positions,omitempty"`
	NoPositions     string `json:"no_positions,omitempty"`
	TotalYes        string `json:"total_yes,omitempty"`
	TotalNo         string `json:"total_no,omitempty"`
	TotalCollateral string `json:"total_collateral,omitempty"`
	ClaimablePayout string `json:"claimable_payout,omitempty"`
	ClaimableRefund string `json:"claimable_refund,omitempty"`
	HasClaimed      *bool  `json:"has_claimed,omitempty"`
	IsOpen          *bool  `json:"is_open,omitempty"`
	USDCBalance     string `json:"usdc_balance,omitempty"`
	USDCAllowance   string `json:"usdc_allowance,omitempty"`
}

type transactionRequestResponse struct {
	To                 string   `json:"to"`
	Contract           string   `json:"contract"`
	Function           string   `json:"function"`
	Args               []string `json:"args"`
	Value              string   `json:"value"`
	Chain              string   `json:"chain"`
	BroadcastPerformed bool     `json:"broadcast_performed"`
}

func registerAgentIntentRoutes(router chi.Router, store *agent.Store, walletRegistry agentWalletRegistry, executor agent.Executor, extras ...any) {
	var sessionRegistry agentSessionRegistry
	var onboardingStarter agent.CircleOnboardingStarter
	var walletResolver agent.CircleWalletResolver
	var balanceReader agent.CircleAgentWalletBalanceReader
	var faucetRunner agent.CircleAgentWalletFaucet
	var durableIntents durableAgentIntentRegistry
	for _, extra := range extras {
		if registry, ok := extra.(agentSessionRegistry); ok {
			sessionRegistry = registry
			continue
		}
		if registry, ok := extra.(durableAgentIntentRegistry); ok {
			durableIntents = registry
			continue
		}
		if starter, ok := extra.(agent.CircleOnboardingStarter); ok {
			onboardingStarter = starter
			continue
		}
		if resolver, ok := extra.(agent.CircleWalletResolver); ok {
			walletResolver = resolver
			continue
		}
		if reader, ok := extra.(agent.CircleAgentWalletBalanceReader); ok {
			balanceReader = reader
			continue
		}
		if runner, ok := extra.(agent.CircleAgentWalletFaucet); ok {
			faucetRunner = runner
		}
	}

	router.Post("/agent/onboarding/start", func(w http.ResponseWriter, r *http.Request) {
		if sessionRegistry == nil {
			httpjson.WriteError(w, http.StatusNotImplemented, "agent_onboarding_sessions_not_configured", "agent onboarding session storage is not configured")
			return
		}

		var request startAgentOnboardingRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON request body")
			return
		}

		input, validationErrors, err := newAgentOnboardingSessionInput(request)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_onboarding_id_failed", "failed to create onboarding id")
			return
		}
		if len(validationErrors) > 0 {
			httpjson.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"error": map[string]any{
					"code":    "agent_onboarding_invalid",
					"message": "agent onboarding validation failed",
					"details": validationErrors,
				},
			})
			return
		}

		onboarding, err := sessionRegistry.CreateAgentOnboardingSession(r.Context(), input)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_onboarding_create_failed", "failed to create agent onboarding session")
			return
		}

		otpResult, requestIDHash, attempted, err := onboardingStarter.StartOTP(r.Context(), onboarding.OnboardingID, onboarding.UserEmail)
		if err != nil && attempted {
			_, _ = sessionRegistry.UpdateAgentOnboardingSessionStatus(
				r.Context(),
				onboarding.OnboardingID,
				repository.AgentOnboardingStatusFailed,
				nullableString("Circle Agent Wallet OTP start failed"),
			)
			httpjson.WriteError(w, http.StatusBadGateway, "circle_otp_start_failed", "Circle Agent Wallet OTP start failed")
			return
		}
		if err == nil && attempted {
			updated, updateErr := sessionRegistry.UpdateAgentOnboardingSessionOTPStart(r.Context(), onboarding.OnboardingID, requestIDHash, otpResult.ExpiresAt)
			if updateErr != nil {
				_, _ = sessionRegistry.UpdateAgentOnboardingSessionStatus(
					r.Context(),
					onboarding.OnboardingID,
					repository.AgentOnboardingStatusFailed,
					nullableString("Circle Agent Wallet OTP request metadata update failed"),
				)
				httpjson.WriteError(w, http.StatusInternalServerError, "circle_otp_start_update_failed", "failed to store Circle Agent Wallet OTP request metadata")
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, map[string]any{
				"onboarding":        newAgentOnboardingSessionResponse(updated),
				"next_step":         "circle_otp_required",
				"expires_at":        otpResult.ExpiresAt.Format("2006-01-02T15:04:05.000000000Z07:00"),
				"request_reference": updated.OnboardingID,
			})
			return
		}

		httpjson.WriteJSON(w, http.StatusCreated, map[string]any{
			"onboarding": newAgentOnboardingSessionResponse(onboarding),
			"next_step":  "circle_otp_verification_not_implemented",
		})
	})

	router.Post("/agent/onboarding/verify", func(w http.ResponseWriter, r *http.Request) {
		if !onboardingStarter.Enabled {
			httpjson.WriteError(w, http.StatusNotImplemented, "circle_otp_verify_not_enabled", "Circle Agent Wallet OTP verification is not enabled")
			return
		}
		if sessionRegistry == nil {
			httpjson.WriteError(w, http.StatusNotImplemented, "agent_onboarding_sessions_not_configured", "agent onboarding session storage is not configured")
			return
		}

		var request verifyAgentOnboardingRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON request body")
			return
		}

		onboardingID := strings.TrimSpace(request.OnboardingID)
		otp := strings.TrimSpace(request.OTP)
		validationErrors := []string{}
		if onboardingID == "" {
			validationErrors = append(validationErrors, "onboarding_id is required")
		}
		if otp == "" {
			validationErrors = append(validationErrors, "otp is required")
		}
		if len(validationErrors) > 0 {
			httpjson.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"error": map[string]any{
					"code":    "agent_onboarding_verify_invalid",
					"message": "agent onboarding verification validation failed",
					"details": validationErrors,
				},
			})
			return
		}

		onboarding, err := sessionRegistry.GetAgentOnboardingSessionByOnboardingID(r.Context(), onboardingID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				httpjson.WriteError(w, http.StatusNotFound, "agent_onboarding_not_found", "agent onboarding session not found")
				return
			}
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_onboarding_get_failed", "failed to get agent onboarding session")
			return
		}
		if onboarding.Status != repository.AgentOnboardingStatusPendingOTP {
			httpjson.WriteError(w, http.StatusConflict, "circle_otp_status_invalid", "agent onboarding session is not pending OTP verification")
			return
		}
		if onboarding.CircleRequestExpiresAt.Valid && time.Now().UTC().After(onboarding.CircleRequestExpiresAt.Time) {
			_, _ = sessionRegistry.UpdateAgentOnboardingSessionStatus(r.Context(), onboarding.OnboardingID, repository.AgentOnboardingStatusExpired, nullableString("Circle Agent Wallet OTP request expired"))
			httpjson.WriteError(w, http.StatusConflict, "circle_otp_request_expired", "Circle Agent Wallet OTP request expired")
			return
		}

		attempted, err := onboardingStarter.VerifyOTP(r.Context(), onboarding.OnboardingID, otp)
		if err != nil {
			if errors.Is(err, agent.ErrCircleOnboardingRequestIDNotAvailable) {
				httpjson.WriteError(w, http.StatusConflict, "circle_otp_request_not_available", "Circle OTP request is not available; backend restart requires onboarding restart")
				return
			}
			if attempted {
				_, _ = sessionRegistry.UpdateAgentOnboardingSessionStatus(
					r.Context(),
					onboarding.OnboardingID,
					repository.AgentOnboardingStatusFailed,
					nullableString("Circle Agent Wallet OTP verification failed"),
				)
			}
			httpjson.WriteError(w, http.StatusBadGateway, "circle_otp_verify_failed", "Circle Agent Wallet OTP verification failed")
			return
		}

		verified, err := sessionRegistry.UpdateAgentOnboardingSessionStatus(r.Context(), onboarding.OnboardingID, repository.AgentOnboardingStatusVerified, sql.NullString{})
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_onboarding_verify_update_failed", "failed to update agent onboarding verification status")
			return
		}
		if onboardingStarter.RequestStore != nil {
			onboardingStarter.RequestStore.Delete(onboarding.OnboardingID)
		}

		if walletResolver == nil {
			httpjson.WriteJSON(w, http.StatusOK, map[string]any{
				"onboarding": newAgentOnboardingSessionResponse(verified),
				"next_step":  "agent_wallet_resolution_not_implemented",
			})
			return
		}

		resolvedWallet, err := walletResolver.ResolveAgentWallet(r.Context(), verified.UserEmail)
		if err != nil {
			switch {
			case errors.Is(err, agent.ErrCircleAgentWalletNotFound):
				httpjson.WriteError(w, http.StatusConflict, "circle_agent_wallet_not_found", "Circle Agent Wallet not found after OTP verification")
			case errors.Is(err, agent.ErrCircleAgentWalletResolutionAmbiguous):
				httpjson.WriteError(w, http.StatusConflict, "circle_agent_wallet_resolution_ambiguous", "Circle Agent Wallet resolution is ambiguous")
			default:
				httpjson.WriteError(w, http.StatusBadGateway, "circle_agent_wallet_resolution_failed", "Circle Agent Wallet resolution failed")
			}
			return
		}

		registeredWallet, err := walletRegistry.RegisterAgentWallet(r.Context(), newAgentWalletActivationInput(verified, resolvedWallet.Address))
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_wallet_register_failed", "failed to register agent wallet")
			return
		}
		agentSessionInput, err := newAgentSessionActivationInput(verified, registeredWallet.AgentWalletAddress)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_session_id_failed", "failed to create agent session id")
			return
		}
		agentSession, err := sessionRegistry.CreateAgentSession(r.Context(), agentSessionInput)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_session_create_failed", "failed to create agent session")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"onboarding":    newAgentOnboardingSessionResponse(verified),
			"agent_wallet":  newAgentWalletResponse(registeredWallet),
			"agent_session": newAgentSessionResponse(agentSession),
			"next_step":     "agent_session_active",
		})
	})

	router.Post("/agent/onboarding/register", func(w http.ResponseWriter, r *http.Request) {
		var request registerAgentOnboardingRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON request body")
			return
		}

		input, validationErrors := newAgentOnboardingRegistrationInput(request)
		if len(validationErrors) > 0 {
			httpjson.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"error": map[string]any{
					"code":    "agent_wallet_invalid",
					"message": "agent wallet registration validation failed",
					"details": validationErrors,
				},
			})
			return
		}

		wallet, err := walletRegistry.RegisterAgentWallet(r.Context(), input)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_wallet_register_failed", "failed to register agent wallet")
			return
		}

		httpjson.WriteJSON(w, http.StatusCreated, map[string]any{
			"agent_wallet": newAgentWalletResponse(wallet),
		})
	})

	router.Get("/agent/onboarding/{onboarding_id}", func(w http.ResponseWriter, r *http.Request) {
		if sessionRegistry == nil {
			httpjson.WriteError(w, http.StatusNotImplemented, "agent_onboarding_sessions_not_configured", "agent onboarding session storage is not configured")
			return
		}

		onboarding, err := sessionRegistry.GetAgentOnboardingSessionByOnboardingID(r.Context(), chi.URLParam(r, "onboarding_id"))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				httpjson.WriteError(w, http.StatusNotFound, "agent_onboarding_not_found", "agent onboarding session not found")
				return
			}
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_onboarding_get_failed", "failed to get agent onboarding session")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"onboarding": newAgentOnboardingSessionResponse(onboarding),
		})
	})

	router.Get("/agent/sessions/{agent_id}", func(w http.ResponseWriter, r *http.Request) {
		if sessionRegistry == nil {
			httpjson.WriteError(w, http.StatusNotImplemented, "agent_sessions_not_configured", "agent session storage is not configured")
			return
		}

		agentID, ok := validateAgentIDPath(w, chi.URLParam(r, "agent_id"))
		if !ok {
			return
		}

		session, err := sessionRegistry.GetAgentSessionByAgentID(r.Context(), agentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				httpjson.WriteError(w, http.StatusNotFound, "agent_session_not_found", "agent session not found")
				return
			}
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_session_get_failed", "failed to get agent session")
			return
		}

		response := newAgentSessionResponse(session)
		// Database "active" reflects SignalArc bookkeeping only. The Circle
		// CLI agent session lives on the host filesystem of whichever
		// instance handled OTP verify, which on Cloud Run may not be the
		// instance handling this request. Probe the local CLI before
		// returning "active" so we do not mislead callers.
		if session.Status == repository.AgentSessionStatusActive && session.WalletProvider == agent.WalletProviderCircleAgentWallet {
			if checker, ok := walletResolver.(agent.CircleAgentSessionLivenessChecker); ok {
				liveness := checker.CheckAgentSessionLiveness(r.Context(), session.AgentWalletAddress)
				switch liveness.State {
				case agent.AgentSessionLivenessLive:
					// keep DB-backed status as is
				case agent.AgentSessionLivenessAuthRequired:
					response.Status = "cli_session_unavailable"
					response.LivenessReason = liveness.Reason
					logCircleSessionLivenessGap(r.Context(), session.AgentID, liveness)
				default:
					response.Status = "cli_session_unknown"
					response.LivenessReason = liveness.Reason
					logCircleSessionLivenessGap(r.Context(), session.AgentID, liveness)
				}
			}
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"agent_session": response,
		})
	})

	router.Post("/agent/wallets", func(w http.ResponseWriter, r *http.Request) {
		var request registerAgentWalletRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON request body")
			return
		}

		input, validationErrors := newAgentWalletRegistrationInput(request)
		if len(validationErrors) > 0 {
			httpjson.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"error": map[string]any{
					"code":    "agent_wallet_invalid",
					"message": "agent wallet registration validation failed",
					"details": validationErrors,
				},
			})
			return
		}

		wallet, err := walletRegistry.RegisterAgentWallet(r.Context(), input)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_wallet_register_failed", "failed to register agent wallet")
			return
		}

		httpjson.WriteJSON(w, http.StatusCreated, map[string]any{
			"agent_wallet": newAgentWalletResponse(wallet),
		})
	})

	router.Get("/agent/wallets/{agent_id}", func(w http.ResponseWriter, r *http.Request) {
		agentID, ok := validateAgentIDPath(w, chi.URLParam(r, "agent_id"))
		if !ok {
			return
		}

		wallet, err := walletRegistry.GetAgentWalletByAgentID(r.Context(), agentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				httpjson.WriteError(w, http.StatusNotFound, "agent_wallet_not_found", "agent wallet not found")
				return
			}
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_wallet_get_failed", "failed to get agent wallet")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"agent_wallet": newAgentWalletResponse(wallet),
		})
	})

	router.Get("/agent/wallets/{agent_id}/balance", func(w http.ResponseWriter, r *http.Request) {
		agentID, ok := validateAgentIDPath(w, chi.URLParam(r, "agent_id"))
		if !ok {
			return
		}
		if balanceReader == nil && walletResolver == nil {
			httpjson.WriteError(w, http.StatusNotImplemented, "circle_agent_wallet_balance_not_configured", "Circle Agent Wallet balance lookup is not configured")
			return
		}

		wallet, err := walletRegistry.GetAgentWalletByAgentID(r.Context(), agentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				httpjson.WriteError(w, http.StatusNotFound, "agent_wallet_not_found", "agent wallet not found")
				return
			}
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_wallet_get_failed", "failed to get agent wallet")
			return
		}
		if wallet.Status != agent.WalletStatusActive {
			httpjson.WriteError(w, http.StatusConflict, "agent_wallet_status_invalid", "agent wallet is not active")
			return
		}
		if wallet.Chain != agent.ChainArcTestnet {
			httpjson.WriteError(w, http.StatusConflict, "agent_wallet_chain_invalid", "agent wallet chain is not ARC-TESTNET")
			return
		}

		balances, err := readAgentWalletBalances(r.Context(), balanceReader, walletResolver, wallet)
		if err != nil {
			logCircleProviderFailure(r.Context(), "circle_agent_wallet_balance", wallet.AgentID, "", err)
			httpjson.WriteError(w, http.StatusBadGateway, "circle_agent_wallet_balance_failed", "Circle Agent Wallet balance lookup failed")
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"agent_wallet_balance": agentWalletBalanceResponse{
				AgentID:            wallet.AgentID,
				AgentWalletAddress: wallet.AgentWalletAddress,
				Chain:              wallet.Chain,
				Balances:           balances.Balances,
			},
		})
	})

	router.Post("/agent/wallets/{agent_id}/faucet", func(w http.ResponseWriter, r *http.Request) {
		agentID, ok := validateAgentIDPath(w, chi.URLParam(r, "agent_id"))
		if !ok {
			return
		}
		if faucetRunner == nil {
			httpjson.WriteError(w, http.StatusNotImplemented, "circle_agent_wallet_faucet_not_configured", "Circle Agent Wallet faucet is not configured")
			return
		}

		wallet, err := walletRegistry.GetAgentWalletByAgentID(r.Context(), agentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				httpjson.WriteError(w, http.StatusNotFound, "agent_wallet_not_found", "agent wallet not found")
				return
			}
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_wallet_get_failed", "failed to get agent wallet")
			return
		}
		if wallet.Status != agent.WalletStatusActive {
			httpjson.WriteError(w, http.StatusConflict, "agent_wallet_status_invalid", "agent wallet is not active")
			return
		}
		if wallet.Chain != agent.ChainArcTestnet {
			httpjson.WriteError(w, http.StatusConflict, "agent_wallet_chain_invalid", "agent wallet chain is not ARC-TESTNET")
			return
		}

		result, err := faucetRunner.RequestFaucet(r.Context(), wallet.AgentWalletAddress)
		if err != nil {
			if errors.Is(err, agent.ErrCircleAgentWalletFaucetNotConfigured) {
				httpjson.WriteError(w, http.StatusNotImplemented, "circle_agent_wallet_faucet_not_configured", "Circle Agent Wallet faucet is not configured")
				return
			}
			httpjson.WriteError(w, http.StatusBadGateway, "circle_agent_wallet_faucet_failed", "Circle Agent Wallet faucet request failed")
			return
		}

		response := agentWalletFaucetResponse{
			AgentID:            wallet.AgentID,
			AgentWalletAddress: wallet.AgentWalletAddress,
			Chain:              wallet.Chain,
			Token:              agent.FaucetTokenUSDC,
			Status:             agent.FaucetStatusRequested,
		}
		switch {
		case result.JSON != nil:
			response.Result = result.JSON
		case result.Message != "":
			response.Result = map[string]any{"message": result.Message}
		default:
			response.Result = map[string]any{}
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"agent_wallet_faucet": response,
		})
	})

	router.Post("/agent/wallets/{agent_id}/disable", func(w http.ResponseWriter, r *http.Request) {
		agentID, ok := validateAgentIDPath(w, chi.URLParam(r, "agent_id"))
		if !ok {
			return
		}

		wallet, err := walletRegistry.DisableAgentWallet(r.Context(), agentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				httpjson.WriteError(w, http.StatusNotFound, "agent_wallet_not_found", "agent wallet not found")
				return
			}
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_wallet_disable_failed", "failed to disable agent wallet")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"agent_wallet": newAgentWalletResponse(wallet),
		})
	})

	router.Get("/agent/portfolio/{agent_id}", func(w http.ResponseWriter, r *http.Request) {
		agentID, ok := validateAgentIDPath(w, chi.URLParam(r, "agent_id"))
		if !ok {
			return
		}
		wallet, err := walletRegistry.GetAgentWalletByAgentID(r.Context(), agentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				httpjson.WriteError(w, http.StatusNotFound, "agent_wallet_not_found", "agent wallet not found")
				return
			}
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_portfolio_get_failed", "failed to get agent portfolio")
			return
		}

		intents := []repository.AgentIntent{}
		executions := []repository.AgentExecution{}
		if durableIntents != nil {
			intents, err = durableIntents.ListAgentIntentsByAgentID(r.Context(), agentID, defaultListLimit)
			if err != nil {
				httpjson.WriteError(w, http.StatusInternalServerError, "agent_portfolio_get_failed", "failed to get agent portfolio")
				return
			}
			executions, err = durableIntents.ListAgentExecutionsByAgentID(r.Context(), agentID, defaultListLimit)
			if err != nil {
				httpjson.WriteError(w, http.StatusInternalServerError, "agent_portfolio_get_failed", "failed to get agent portfolio")
				return
			}
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"portfolio": newAgentPortfolioResponse(wallet, intents, executions),
		})
	})

	router.Get("/agent/activity/{agent_id}", func(w http.ResponseWriter, r *http.Request) {
		agentID, ok := validateAgentIDPath(w, chi.URLParam(r, "agent_id"))
		if !ok {
			return
		}
		if _, err := walletRegistry.GetAgentWalletByAgentID(r.Context(), agentID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				httpjson.WriteError(w, http.StatusNotFound, "agent_wallet_not_found", "agent wallet not found")
				return
			}
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_activity_get_failed", "failed to get agent activity")
			return
		}
		if durableIntents == nil {
			httpjson.WriteJSON(w, http.StatusOK, map[string]any{
				"activity": agentActivityResponse{AgentID: agentID, Items: []agentActivityItemResponse{}},
			})
			return
		}

		intents, err := durableIntents.ListAgentIntentsByAgentID(r.Context(), agentID, defaultListLimit)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_activity_get_failed", "failed to get agent activity")
			return
		}
		executions, err := durableIntents.ListAgentExecutionsByAgentID(r.Context(), agentID, defaultListLimit)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_activity_get_failed", "failed to get agent activity")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"activity": newAgentActivityResponse(agentID, intents, executions),
		})
	})

	router.Post("/agent/intents", func(w http.ResponseWriter, r *http.Request) {
		var request createAgentIntentRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON; for create_market send close_timestamp as RFC3339 (for example 2026-05-31T23:59:00Z) or unix-seconds integer string")
			return
		}

		request.AgentID = strings.TrimSpace(request.AgentID)
		if request.AgentID != "" {
			agentID, ok := validateAgentIDPath(w, request.AgentID)
			if !ok {
				return
			}
			request.AgentID = agentID
		}

		var registeredWallet repository.AgentWallet
		if request.AgentID != "" {
			var err error
			registeredWallet, err = walletRegistry.GetAgentWalletByAgentID(r.Context(), request.AgentID)
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					httpjson.WriteError(w, http.StatusInternalServerError, "agent_wallet_get_failed", "failed to get agent wallet")
					return
				}
			}
		}

		intentUserWallet := request.UserWallet
		if shouldBindAgentWalletAsIntentUserWallet(request, registeredWallet) {
			intentUserWallet = registeredWallet.AgentWalletAddress
		}

		intent, err := store.CreateIntent(agent.CreateIntentInput{
			AgentID:               request.AgentID,
			AgentWalletAddress:    firstNonEmpty(request.AgentWalletAddress, registeredWallet.AgentWalletAddress),
			WalletProvider:        registeredWallet.WalletProvider,
			AllowedActions:        registeredWallet.AllowedActions,
			SourceClient:          request.SourceClient,
			ClientRequestID:       request.ClientRequestID,
			Action:                request.Action,
			UserWallet:            intentUserWallet,
			MarketID:              request.MarketID,
			MarketContractAddress: request.MarketContractAddress,
			Amount:                request.Amount,
			Outcome:               request.Outcome,
			Resolver:              request.Resolver,
			CollateralToken:       request.CollateralToken,
			CloseTimestamp:        request.CloseTimestamp,
			Question:              request.Question,
		})
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_intent_create_failed", "failed to create agent intent preview")
			return
		}
		if !intent.ValidationResult.Valid {
			httpjson.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"intent": newAgentIntentResponse(intent),
			})
			return
		}
		if registeredWallet.AgentID != "" && isBackendExecutableAgentAction(intent.Action) && !repositoryAgentWalletAllowsAction(registeredWallet, intent.Action) {
			writeAgentActionForbidden(w)
			return
		}
		if policyViolations := agentWalletPolicyViolations(intent, registeredWallet); len(policyViolations) > 0 {
			writeAgentPolicyViolation(w, policyViolations)
			return
		}

		if durableIntents != nil {
			persistedIntent, err := durableIntents.CreateAgentIntent(r.Context(), newDurableAgentIntentInput(intent))
			if err != nil {
				httpjson.WriteError(w, http.StatusInternalServerError, "agent_intent_create_failed", "failed to create agent intent preview")
				return
			}
			intent = newAgentIntentFromRepository(persistedIntent)
		}

		httpjson.WriteJSON(w, http.StatusCreated, map[string]any{
			"intent": newAgentIntentResponse(intent),
		})
	})

	router.Get("/agent/intents/{id}", func(w http.ResponseWriter, r *http.Request) {
		intentID := chi.URLParam(r, "id")
		if durableIntents != nil {
			intent, err := durableIntents.GetAgentIntentByIntentID(r.Context(), intentID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					httpjson.WriteError(w, http.StatusNotFound, "agent_intent_not_found", "agent intent not found")
					return
				}
				httpjson.WriteError(w, http.StatusInternalServerError, "agent_intent_get_failed", "failed to get agent intent")
				return
			}
			responseIntent := newAgentIntentFromRepository(intent)
			if responseIntent.AgentID != "" {
				if _, ok := validateAgentIDPath(w, responseIntent.AgentID); !ok {
					return
				}
			}
			httpjson.WriteJSON(w, http.StatusOK, map[string]any{
				"intent": newAgentIntentResponse(responseIntent),
			})
			return
		}

		intent, err := store.GetIntent(intentID)
		if err != nil {
			if errors.Is(err, agent.ErrIntentNotFound) {
				httpjson.WriteError(w, http.StatusNotFound, "agent_intent_not_found", "agent intent not found")
				return
			}

			httpjson.WriteError(w, http.StatusInternalServerError, "agent_intent_get_failed", "failed to get agent intent")
			return
		}
		if intent.AgentID != "" {
			if _, ok := validateAgentIDPath(w, intent.AgentID); !ok {
				return
			}
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"intent": newAgentIntentResponse(intent),
		})
	})

	router.Get("/agent/intents/{id}/executions", func(w http.ResponseWriter, r *http.Request) {
		intentID := chi.URLParam(r, "id")
		if durableIntents == nil {
			httpjson.WriteJSON(w, http.StatusOK, map[string]any{
				"executions": []agentActivityItemResponse{},
			})
			return
		}
		intent, err := durableIntents.GetAgentIntentByIntentID(r.Context(), intentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				httpjson.WriteError(w, http.StatusNotFound, "agent_intent_not_found", "agent intent not found")
				return
			}
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_execution_get_failed", "failed to get agent executions")
			return
		}
		executions, err := durableIntents.ListAgentExecutionsByIntentID(r.Context(), intentID, defaultListLimit)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_execution_get_failed", "failed to get agent executions")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"executions": newAgentExecutionActivityItems([]repository.AgentIntent{intent}, executions),
		})
	})

	router.Post("/agent/intents/{id}/confirm", func(w http.ResponseWriter, r *http.Request) {
		intentID := chi.URLParam(r, "id")
		if durableIntents != nil {
			persistedIntent, err := durableIntents.GetAgentIntentByIntentID(r.Context(), intentID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					httpjson.WriteError(w, http.StatusNotFound, "agent_intent_not_found", "agent intent not found")
					return
				}
				httpjson.WriteError(w, http.StatusInternalServerError, "agent_intent_get_failed", "failed to get agent intent")
				return
			}
			intent := newAgentIntentFromRepository(persistedIntent)
			if intent.AgentID != "" {
				if _, ok := validateAgentIDPath(w, intent.AgentID); !ok {
					return
				}
			}
			if !intent.ValidationResult.Valid {
				writeAgentIntentInvalid(w, intent.ValidationResult.Errors)
				return
			}
			if !validateAgentIntentPolicyForConfirmation(w, r, walletRegistry, intent) {
				return
			}
			confirmedIntent, err := durableIntents.ConfirmAgentIntent(r.Context(), intentID)
			if err != nil {
				httpjson.WriteError(w, http.StatusInternalServerError, "agent_intent_confirm_failed", "failed to confirm agent intent")
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, map[string]any{
				"execution_plan": newAgentExecutionPlanResponse(agent.NewExecutionPlan(newAgentIntentFromRepository(confirmedIntent))),
			})
			return
		}

		intent, err := store.GetIntent(intentID)
		if err != nil {
			if errors.Is(err, agent.ErrIntentNotFound) {
				httpjson.WriteError(w, http.StatusNotFound, "agent_intent_not_found", "agent intent not found")
				return
			}
			httpjson.WriteError(w, http.StatusInternalServerError, "agent_intent_get_failed", "failed to get agent intent")
			return
		}
		if intent.AgentID != "" {
			if _, ok := validateAgentIDPath(w, intent.AgentID); !ok {
				return
			}
		}
		if !intent.ValidationResult.Valid {
			writeAgentIntentInvalid(w, intent.ValidationResult.Errors)
			return
		}
		if !validateAgentIntentPolicyForConfirmation(w, r, walletRegistry, intent) {
			return
		}

		executionPlan, err := store.ConfirmIntent(intentID)
		if err != nil {
			if errors.Is(err, agent.ErrIntentNotFound) {
				httpjson.WriteError(w, http.StatusNotFound, "agent_intent_not_found", "agent intent not found")
				return
			}
			if errors.Is(err, agent.ErrIntentInvalid) {
				writeAgentIntentInvalid(w, nil)
				return
			}

			httpjson.WriteError(w, http.StatusInternalServerError, "agent_intent_confirm_failed", "failed to confirm agent intent")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"execution_plan": newAgentExecutionPlanResponse(executionPlan),
		})
	})

	router.Post("/agent/intents/{id}/execute", func(w http.ResponseWriter, r *http.Request) {
		intentID := chi.URLParam(r, "id")
		var intent agent.Intent
		if durableIntents != nil {
			persistedIntent, err := durableIntents.GetAgentIntentByIntentID(r.Context(), intentID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					httpjson.WriteError(w, http.StatusNotFound, "agent_intent_not_found", "agent intent not found")
					return
				}
				httpjson.WriteError(w, http.StatusInternalServerError, "agent_intent_get_failed", "failed to get agent intent")
				return
			}
			intent = newAgentIntentFromRepository(persistedIntent)
		} else {
			var err error
			intent, err = store.GetIntent(intentID)
			if err != nil {
				if errors.Is(err, agent.ErrIntentNotFound) {
					httpjson.WriteError(w, http.StatusNotFound, "agent_intent_not_found", "agent intent not found")
					return
				}

				httpjson.WriteError(w, http.StatusInternalServerError, "agent_intent_get_failed", "failed to get agent intent")
				return
			}
		}

		if intent.Status != agent.StatusConfirmed {
			httpjson.WriteError(w, http.StatusConflict, "agent_intent_not_confirmed", "agent intent must be confirmed before execution")
			return
		}
		if intent.AgentID != "" {
			if _, ok := validateAgentIDPath(w, intent.AgentID); !ok {
				return
			}
		}
		if !intent.ValidationResult.Valid {
			writeAgentIntentInvalid(w, intent.ValidationResult.Errors)
			return
		}
		if !isBackendExecutableAgentAction(intent.Action) {
			httpjson.WriteError(w, http.StatusNotImplemented, "not_implemented", "agent execution action is not implemented")
			return
		}

		agentWallet, err := walletRegistry.GetAgentWalletByAgentID(r.Context(), intent.AgentID)
		if err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "agent_wallet_missing", "agent wallet must be registered before execution")
			return
		}
		if err := validateAgentWalletForExecution(intent, agentWallet); err != nil {
			if errors.Is(err, errAgentActionForbidden) {
				writeAgentActionForbidden(w)
				return
			}
			httpjson.WriteError(w, http.StatusForbidden, "agent_wallet_forbidden", err.Error())
			return
		}
		intent = hydrateAgentIntentForExecution(intent, agentWallet)
		if policyViolations := agentWalletPolicyViolations(intent, agentWallet); len(policyViolations) > 0 {
			writeAgentPolicyViolation(w, policyViolations)
			return
		}
		if executor == nil && agentWallet.WalletProvider == agent.WalletProviderCircleAgentWallet {
			httpjson.WriteError(w, http.StatusNotImplemented, "circle_agent_wallet_execution_not_enabled", "Circle Agent Wallet ARC-TESTNET contract execution requires Circle CLI authentication and live wallet proof before backend execution is enabled")
			return
		}
		if executor == nil && agentWallet.WalletProvider == agent.WalletProviderTemporaryTestnetAgentEOA {
			httpjson.WriteError(w, http.StatusNotImplemented, "temporary_agent_eoa_execution_not_enabled", "temporary_testnet_agent_eoa is a documented fallback name only and is not enabled without a non-deployer, non-user agent key")
			return
		}

		activeExecutor := executor
		if activeExecutor == nil {
			activeExecutor, err = agent.NewArcExecutorFromEnv()
			if err != nil {
				httpjson.WriteError(w, http.StatusServiceUnavailable, "agent_execution_config_invalid", "agent execution environment is not configured")
				return
			}
		}

		var execution repository.AgentExecution
		if durableIntents != nil {
			executionPlan := agent.NewExecutionPlan(intent)
			execution, err = durableIntents.CreateAgentExecution(r.Context(), repository.CreateAgentExecutionInput{
				IntentID:              intent.ID,
				AgentID:               intent.AgentID,
				Action:                intent.Action,
				ExecutionMode:         executionPlan.ExecutionMode,
				Network:               executionPlan.Network,
				AgentFactoryAddress:   executionPlan.AgentFactoryAddress,
				MarketContractAddress: intent.MarketContractAddress,
			})
			if err != nil {
				httpjson.WriteError(w, http.StatusInternalServerError, "agent_execution_create_failed", "failed to create agent execution record")
				return
			}
		}

		var result agent.ExecutionResult
		switch intent.Action {
		case agent.ActionCreateMarket:
			result, err = activeExecutor.ExecuteCreateMarket(r.Context(), intent)
		case agent.ActionBuyYes:
			result, err = activeExecutor.ExecuteBuyYes(r.Context(), intent)
		case agent.ActionBuyNo:
			result, err = activeExecutor.ExecuteBuyNo(r.Context(), intent)
		case agent.ActionCloseMarket:
			result, err = activeExecutor.ExecuteCloseMarket(r.Context(), intent)
		case agent.ActionResolveMarket:
			result, err = activeExecutor.ExecuteResolveMarket(r.Context(), intent)
		case agent.ActionClaimPayout:
			result, err = activeExecutor.ExecuteClaimPayout(r.Context(), intent)
		case agent.ActionCancelMarket:
			result, err = activeExecutor.ExecuteCancelMarket(r.Context(), intent)
		case agent.ActionClaimRefund:
			result, err = activeExecutor.ExecuteClaimRefund(r.Context(), intent)
		default:
			err = agent.ErrExecutionNotImplemented
		}
		if err != nil {
			failure := newAgentExecutionFailure(err)
			if durableIntents != nil && execution.ID != "" {
				_, _ = durableIntents.MarkAgentExecutionFailed(r.Context(), execution.ID, repository.FailAgentExecutionInput{
					ErrorCode:    failure.Code,
					ErrorMessage: failure.Message,
					Readback:     json.RawMessage(`{}`),
				})
				_, _ = durableIntents.MarkAgentIntentFailed(r.Context(), intent.ID)
			}
			logCircleProviderFailure(r.Context(), "agent_execution", intent.AgentID, intent.ID, enrichExecuteErrorAction(err, intent.Action))
			httpjson.WriteError(w, failure.Status, failure.Code, failure.Message)
			return
		}
		if durableIntents != nil && execution.ID != "" {
			_, err = durableIntents.MarkAgentExecutionExecuted(r.Context(), execution.ID, repository.CompleteAgentExecutionInput{
				ExecutionMode:          result.ExecutionMode,
				Network:                result.Network,
				AgentFactoryAddress:    result.AgentFactoryAddress,
				MarketContractAddress:  result.MarketContractAddress,
				ApproveTransactionHash: result.ApproveTransactionHash,
				TransactionHash:        result.TransactionHash,
				BroadcastPerformed:     result.BroadcastPerformed,
				Readback:               newAgentExecutionReadbackJSON(result.Readback),
			})
			if err != nil {
				httpjson.WriteError(w, http.StatusInternalServerError, "agent_execution_update_failed", "failed to update agent execution record")
				return
			}
			_, err = durableIntents.MarkAgentIntentExecuted(r.Context(), intent.ID)
			if err != nil {
				httpjson.WriteError(w, http.StatusInternalServerError, "agent_intent_update_failed", "failed to update agent intent")
				return
			}
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"execution": newAgentExecutionResponse(result),
		})
	})
}

func hydrateAgentIntentForExecution(intent agent.Intent, wallet repository.AgentWallet) agent.Intent {
	if intent.AgentWalletAddress == "" {
		intent.AgentWalletAddress = wallet.AgentWalletAddress
	}
	if intent.WalletProvider == "" {
		intent.WalletProvider = wallet.WalletProvider
	}
	if len(intent.AllowedActions) == 0 {
		intent.AllowedActions = append([]string{}, wallet.AllowedActions...)
	}
	if len(intent.PolicyMetadata) == 0 {
		intent.PolicyMetadata = agentPolicyMetadataMap(wallet.PolicyMetadata)
	}
	return intent
}

func shouldBindAgentWalletAsIntentUserWallet(request createAgentIntentRequest, wallet repository.AgentWallet) bool {
	if strings.TrimSpace(request.UserWallet) != "" {
		return false
	}
	if !isBackendExecutableAgentAction(strings.TrimSpace(request.Action)) {
		return false
	}
	if wallet.AgentID == "" || wallet.AgentWalletAddress == "" {
		return false
	}
	if wallet.WalletProvider != agent.WalletProviderCircleAgentWallet {
		return false
	}
	if wallet.Chain != agent.ChainArcTestnet {
		return false
	}
	if wallet.Status != agent.WalletStatusActive {
		return false
	}
	return repositoryAgentWalletAllowsAction(wallet, strings.TrimSpace(request.Action))
}

func isBackendExecutableAgentAction(action string) bool {
	switch action {
	case agent.ActionCreateMarket,
		agent.ActionBuyYes,
		agent.ActionBuyNo,
		agent.ActionCloseMarket,
		agent.ActionResolveMarket,
		agent.ActionClaimPayout,
		agent.ActionCancelMarket,
		agent.ActionClaimRefund:
		return true
	default:
		return false
	}
}

func newAgentOnboardingSessionInput(request startAgentOnboardingRequest) (repository.CreateAgentOnboardingSessionInput, []string, error) {
	onboardingID, err := newPublicID("agent_onboarding")
	if err != nil {
		return repository.CreateAgentOnboardingSessionInput{}, nil, err
	}
	policyMetadata, _ := json.Marshal(map[string]string{
		"note": "pending Circle Agent Wallet OTP onboarding",
	})

	input := repository.CreateAgentOnboardingSessionInput{
		OnboardingID:   onboardingID,
		AgentID:        strings.TrimSpace(request.AgentID),
		UserEmail:      strings.TrimSpace(request.UserEmail),
		UserWallet:     nullableString(request.UserWallet),
		SourceClient:   nullableString(request.SourceClient),
		Channel:        nullableString(request.Channel),
		Chain:          agent.ChainArcTestnet,
		WalletProvider: agent.WalletProviderCircleAgentWallet,
		Status:         repository.AgentOnboardingStatusPendingOTP,
		PolicyMetadata: policyMetadata,
	}

	validationErrors := validateAgentOnboardingSessionInput(input)
	return input, validationErrors, nil
}

func validateAgentOnboardingSessionInput(input repository.CreateAgentOnboardingSessionInput) []string {
	errors := []string{}
	if _, agentIDErrors := validateAgentID(input.AgentID); len(agentIDErrors) > 0 {
		errors = append(errors, agentIDErrors...)
	}
	if input.UserEmail == "" {
		errors = append(errors, "user_email is required")
	}
	if input.Chain != agent.ChainArcTestnet {
		errors = append(errors, "chain must be ARC-TESTNET")
	}
	if input.WalletProvider != agent.WalletProviderCircleAgentWallet {
		errors = append(errors, "wallet_provider must be circle_agent_wallet")
	}
	if input.Status != repository.AgentOnboardingStatusPendingOTP {
		errors = append(errors, "status must be pending_otp")
	}
	return errors
}

func newAgentOnboardingRegistrationInput(request registerAgentOnboardingRequest) (repository.UpsertAgentWalletInput, []string) {
	policyMetadata, _ := json.Marshal(map[string]string{
		"note": "default ARC-TESTNET onboarding policy",
	})
	input := repository.UpsertAgentWalletInput{
		AgentID:            strings.TrimSpace(request.AgentID),
		UserWallet:         strings.TrimSpace(request.UserWallet),
		UserEmail:          nullableString(request.UserEmail),
		AgentWalletAddress: strings.TrimSpace(request.AgentWalletAddress),
		WalletProvider:     agent.WalletProviderCircleAgentWallet,
		Chain:              agent.ChainArcTestnet,
		Status:             agent.WalletStatusActive,
		AllowedActions:     defaultAgentWalletAllowedActions(),
		PolicyMetadata:     policyMetadata,
		SourceClient:       nullableString(request.SourceClient),
	}

	errors := validateAgentWalletRegistrationInput(input)
	return input, errors
}

func newAgentWalletActivationInput(onboarding repository.AgentOnboardingSession, agentWalletAddress string) repository.UpsertAgentWalletInput {
	policyMetadata, _ := json.Marshal(map[string]string{
		"note": "activated after Circle Agent Wallet OTP verification",
	})
	return repository.UpsertAgentWalletInput{
		AgentID:            onboarding.AgentID,
		UserWallet:         onboarding.UserWallet.String,
		UserEmail:          sql.NullString{String: onboarding.UserEmail, Valid: onboarding.UserEmail != ""},
		AgentWalletAddress: strings.TrimSpace(agentWalletAddress),
		WalletProvider:     agent.WalletProviderCircleAgentWallet,
		Chain:              agent.ChainArcTestnet,
		Status:             agent.WalletStatusActive,
		AllowedActions:     defaultAgentWalletAllowedActions(),
		PolicyMetadata:     policyMetadata,
		SourceClient:       onboarding.SourceClient,
	}
}

func newAgentSessionActivationInput(onboarding repository.AgentOnboardingSession, agentWalletAddress string) (repository.CreateAgentSessionInput, error) {
	sessionID, err := newPublicID("agent_session")
	if err != nil {
		return repository.CreateAgentSessionInput{}, err
	}
	sessionMetadata, _ := json.Marshal(map[string]string{
		"note": "Circle Agent Wallet OTP verified; wallet resolution source is read-only Circle CLI list output",
	})
	return repository.CreateAgentSessionInput{
		SessionID:          sessionID,
		AgentID:            onboarding.AgentID,
		UserEmail:          onboarding.UserEmail,
		UserWallet:         onboarding.UserWallet.String,
		AgentWalletAddress: strings.TrimSpace(agentWalletAddress),
		WalletProvider:     agent.WalletProviderCircleAgentWallet,
		Chain:              agent.ChainArcTestnet,
		Status:             repository.AgentSessionStatusActive,
		AllowedActions:     defaultAgentWalletAllowedActions(),
		AllowedChannels:    defaultAgentSessionAllowedChannels(onboarding),
		SessionMetadata:    sessionMetadata,
	}, nil
}

func defaultAgentSessionAllowedChannels(onboarding repository.AgentOnboardingSession) []string {
	channels := []string{}
	for _, value := range []string{onboarding.Channel.String, onboarding.SourceClient.String} {
		value = strings.TrimSpace(value)
		if value != "" && !stringSliceContains(channels, value) {
			channels = append(channels, value)
		}
	}
	if len(channels) == 0 {
		channels = append(channels, "unknown")
	}
	return channels
}

func stringSliceContains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func newAgentWalletRegistrationInput(request registerAgentWalletRequest) (repository.UpsertAgentWalletInput, []string) {
	input := repository.UpsertAgentWalletInput{
		AgentID:            strings.TrimSpace(request.AgentID),
		UserWallet:         strings.TrimSpace(request.UserWallet),
		UserEmail:          nullableString(request.UserEmail),
		AgentWalletAddress: strings.TrimSpace(request.AgentWalletAddress),
		WalletProvider:     strings.TrimSpace(request.WalletProvider),
		Chain:              strings.TrimSpace(request.Chain),
		Status:             strings.TrimSpace(request.Status),
		AllowedActions:     normalizeActions(request.AllowedActions),
		SourceClient:       nullableString(request.SourceClient),
	}
	if input.Chain == "" {
		input.Chain = agent.ChainArcTestnet
	}
	if input.Status == "" {
		input.Status = agent.WalletStatusActive
	}
	if request.PolicyMetadata != nil {
		if bytes, err := json.Marshal(request.PolicyMetadata); err == nil {
			input.PolicyMetadata = bytes
		}
	}

	errors := validateAgentWalletRegistrationInput(input)
	return input, errors
}

func defaultAgentWalletAllowedActions() []string {
	return []string{
		agent.ActionCreateMarket,
		agent.ActionBuyYes,
		agent.ActionBuyNo,
		agent.ActionCloseMarket,
		agent.ActionResolveMarket,
		agent.ActionClaimPayout,
		agent.ActionCancelMarket,
		agent.ActionClaimRefund,
	}
}

func validateAgentWalletRegistrationInput(input repository.UpsertAgentWalletInput) []string {
	errors := []string{}
	if _, agentIDErrors := validateAgentID(input.AgentID); len(agentIDErrors) > 0 {
		errors = append(errors, agentIDErrors...)
	}
	if input.UserWallet == "" {
		errors = append(errors, "user_wallet is required")
	}
	if input.AgentWalletAddress == "" {
		errors = append(errors, "agent_wallet_address is required")
	}
	if input.WalletProvider != agent.WalletProviderCircleAgentWallet {
		errors = append(errors, "wallet_provider must be circle_agent_wallet")
	}
	if input.Chain != agent.ChainArcTestnet {
		errors = append(errors, "chain must be ARC-TESTNET")
	}
	if input.Status != agent.WalletStatusActive && input.Status != "disabled" {
		errors = append(errors, "status must be active or disabled")
	}
	if len(input.AllowedActions) == 0 {
		errors = append(errors, "allowed_actions is required")
	}
	for _, action := range input.AllowedActions {
		if !isAllowedAgentAction(action) {
			errors = append(errors, "allowed_actions contains unsupported action")
		}
	}
	if equalAddress(input.AgentWalletAddress, knownDeployerResolverWallet()) {
		errors = append(errors, "agent_wallet_address must not equal the deployer/resolver wallet")
	}
	if equalAddress(input.AgentWalletAddress, input.UserWallet) {
		errors = append(errors, "agent_wallet_address must not equal user_wallet unless a documented user-controlled custody link is implemented")
	}
	return errors
}

func newPublicID(prefix string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(bytes), nil
}

func nullStringValue(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func nullableString(value string) sql.NullString {
	value = strings.TrimSpace(value)
	return sql.NullString{String: value, Valid: value != ""}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func normalizeActions(values []string) []string {
	actions := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			actions = append(actions, value)
		}
	}
	return actions
}

func isAllowedAgentAction(action string) bool {
	switch action {
	case agent.ActionCreateMarket,
		agent.ActionBuyYes,
		agent.ActionBuyNo,
		agent.ActionCancelMarket,
		agent.ActionCloseMarket,
		agent.ActionResolveMarket,
		agent.ActionClaimRefund,
		agent.ActionClaimPayout:
		return true
	default:
		return false
	}
}

func newAgentWalletResponse(wallet repository.AgentWallet) agentWalletResponse {
	policyMetadata := map[string]string{}
	if len(wallet.PolicyMetadata) > 0 {
		_ = json.Unmarshal(wallet.PolicyMetadata, &policyMetadata)
	}

	return agentWalletResponse{
		ID:                 wallet.ID,
		AgentID:            wallet.AgentID,
		UserWallet:         nullStringValue(wallet.UserWallet),
		UserEmail:          wallet.UserEmail.String,
		AgentWalletAddress: wallet.AgentWalletAddress,
		WalletProvider:     wallet.WalletProvider,
		Chain:              wallet.Chain,
		AllowedActions:     wallet.AllowedActions,
		Status:             wallet.Status,
		PolicyMetadata:     policyMetadata,
		SourceClient:       wallet.SourceClient.String,
		CreatedAt:          wallet.CreatedAt.Format("2006-01-02T15:04:05.000000000Z07:00"),
		UpdatedAt:          wallet.UpdatedAt.Format("2006-01-02T15:04:05.000000000Z07:00"),
	}
}

func newAgentOnboardingSessionResponse(onboarding repository.AgentOnboardingSession) agentOnboardingSessionResponse {
	policyMetadata := map[string]string{}
	if len(onboarding.PolicyMetadata) > 0 {
		_ = json.Unmarshal(onboarding.PolicyMetadata, &policyMetadata)
	}

	return agentOnboardingSessionResponse{
		OnboardingID:                onboarding.OnboardingID,
		AgentID:                     onboarding.AgentID,
		UserEmail:                   onboarding.UserEmail,
		UserWallet:                  onboarding.UserWallet.String,
		RequestedAgentWalletAddress: onboarding.RequestedAgentWalletAddress.String,
		SourceClient:                onboarding.SourceClient.String,
		Channel:                     onboarding.Channel.String,
		Chain:                       onboarding.Chain,
		WalletProvider:              onboarding.WalletProvider,
		Status:                      onboarding.Status,
		FailureReason:               onboarding.FailureReason.String,
		PolicyMetadata:              policyMetadata,
		CreatedAt:                   onboarding.CreatedAt.Format("2006-01-02T15:04:05.000000000Z07:00"),
		UpdatedAt:                   onboarding.UpdatedAt.Format("2006-01-02T15:04:05.000000000Z07:00"),
	}
}

func newAgentSessionResponse(session repository.AgentSession) agentSessionResponse {
	sessionMetadata := map[string]string{}
	if len(session.SessionMetadata) > 0 {
		_ = json.Unmarshal(session.SessionMetadata, &sessionMetadata)
	}

	return agentSessionResponse{
		SessionID:          session.SessionID,
		AgentID:            session.AgentID,
		UserEmail:          session.UserEmail,
		UserWallet:         session.UserWallet,
		AgentWalletAddress: session.AgentWalletAddress,
		WalletProvider:     session.WalletProvider,
		Chain:              session.Chain,
		Status:             session.Status,
		AllowedActions:     session.AllowedActions,
		AllowedChannels:    session.AllowedChannels,
		SessionMetadata:    sessionMetadata,
		CreatedAt:          session.CreatedAt.Format("2006-01-02T15:04:05.000000000Z07:00"),
		UpdatedAt:          session.UpdatedAt.Format("2006-01-02T15:04:05.000000000Z07:00"),
	}
}

// logCircleSessionLivenessGap emits a structured warning when SignalArc
// downgrades the DB-backed agent session status because the local Circle
// CLI agent session is missing on this backend instance. The reason here
// is the sanitized message produced by the resolver and is safe to log.
func logCircleSessionLivenessGap(ctx context.Context, agentID string, liveness agent.AgentSessionLivenessResult) {
	log.Warn().
		Str("operation", "agent_session_liveness").
		Str("agent_id", agentID).
		Str("request_id", requestIDFromContext(ctx)).
		Str("liveness_state", string(liveness.State)).
		Str("error_class", liveness.ErrorClass).
		Str("reason", liveness.Reason).
		Msg("agent session liveness downgraded; Circle CLI session not available on this backend instance")
}

// enrichExecuteErrorAction sets the Action field on a CircleCLIError's
// ExecCtx if one exists. This allows the handler to attach intent-level
// context that the executor layer does not have. The error is returned
// as-is (same pointer) for chaining convenience.
func enrichExecuteErrorAction(err error, action string) error {
	if err == nil || action == "" {
		return err
	}
	var cliErr *agent.CircleCLIError
	if errors.As(err, &cliErr) && cliErr != nil {
		if cliErr.ExecCtx == nil {
			cliErr.ExecCtx = &agent.ExecuteContext{}
		}
		cliErr.ExecCtx.Action = action
	}
	return err
}

// logCircleProviderFailure emits a structured error log for Circle CLI
// provider failures returned from balance or execute paths. Public HTTP
// error codes are not changed; this only adds sanitized structured detail
// to backend logs to make Cloud Run AUTH_REQUIRED conditions diagnosable.
func logCircleProviderFailure(ctx context.Context, operation string, agentID string, intentID string, err error) {
	if err == nil {
		return
	}

	event := log.Error().
		Str("operation", operation).
		Str("agent_id", agentID).
		Str("intent_id", intentID).
		Str("request_id", requestIDFromContext(ctx)).
		Str("error_class", agent.CircleErrorClassFromError(err)).
		Str("summary", agent.CircleErrorSummaryFromError(err))

	// Attach execution context fields when available (execute failures).
	if execCtx := agent.CircleExecuteContextFromError(err); execCtx != nil {
		if execCtx.Action != "" {
			event = event.Str("action", execCtx.Action)
		}
		if execCtx.FunctionSignature != "" {
			event = event.Str("function_signature", execCtx.FunctionSignature)
		}
		if execCtx.ContractAddress != "" {
			event = event.Str("contract_address", execCtx.ContractAddress)
		}
		if execCtx.WalletAddress != "" {
			event = event.Str("wallet_address", execCtx.WalletAddress)
		}
		if execCtx.Chain != "" {
			event = event.Str("chain", execCtx.Chain)
		}
		if execCtx.CommandCategory != "" {
			event = event.Str("command_category", execCtx.CommandCategory)
		}
		if execCtx.ExitStatus != "" {
			event = event.Str("exit_status", execCtx.ExitStatus)
		}
		event = event.Int("raw_output_len", execCtx.RawOutputLen)
		if execCtx.ProcessError != "" {
			event = event.Str("process_error", execCtx.ProcessError)
		}
	}

	event.Msg("Circle Agent Wallet provider call failed")
}

func newAgentIntentResponse(intent agent.Intent) agentIntentResponse {
	return agentIntentResponse{
		IntentID:              intent.ID,
		AgentID:               intent.AgentID,
		AgentWalletAddress:    intent.AgentWalletAddress,
		WalletProvider:        intent.WalletProvider,
		AllowedActions:        intent.AllowedActions,
		SourceClient:          intent.SourceClient,
		ClientRequestID:       intent.ClientRequestID,
		Action:                intent.Action,
		Status:                intent.Status,
		RequiresConfirmation:  intent.RequiresConfirmation,
		UserWallet:            intent.UserWallet,
		Address:               intent.UserWallet,
		MarketID:              intent.MarketID,
		MarketContractAddress: intent.MarketContractAddress,
		Amount:                intent.Amount,
		Outcome:               intent.Outcome,
		Resolver:              intent.Resolver,
		CollateralToken:       intent.CollateralToken,
		CloseTimestamp:        intent.CloseTimestamp,
		Question:              intent.Question,
		ValidationResult:      intent.ValidationResult,
		Warnings:              intent.Warnings,
		CreatedAt:             intent.CreatedAt.Format("2006-01-02T15:04:05.000000000Z07:00"),
	}
}

func newAgentPortfolioResponse(wallet repository.AgentWallet, intents []repository.AgentIntent, executions []repository.AgentExecution) agentPortfolioResponse {
	executionsByIntentID := map[string]repository.AgentExecution{}
	for _, execution := range executions {
		if execution.IntentID == "" {
			continue
		}
		current, ok := executionsByIntentID[execution.IntentID]
		if !ok || execution.CreatedAt.After(current.CreatedAt) {
			executionsByIntentID[execution.IntentID] = execution
		}
	}

	positions := []agentPortfolioPositionResponse{}
	totalExposure := big.NewRat(0, 1)
	for _, intent := range intents {
		if intent.Action != agent.ActionBuyYes && intent.Action != agent.ActionBuyNo {
			continue
		}
		execution, ok := executionsByIntentID[intent.IntentID]
		if !ok || execution.Status != repository.AgentExecutionStatusExecuted {
			continue
		}

		amount := nullStringValue(intent.Amount)
		if value, ok := new(big.Rat).SetString(amount); ok {
			totalExposure.Add(totalExposure, value)
		}
		positions = append(positions, agentPortfolioPositionResponse{
			Source:                "agent_intent_execution",
			IntentID:              intent.IntentID,
			Action:                intent.Action,
			Status:                execution.Status,
			MarketID:              nullStringValue(intent.MarketID),
			MarketContractAddress: firstNonEmpty(nullStringValue(intent.MarketContractAddress), nullStringValue(execution.MarketContractAddress)),
			Amount:                amount,
			Outcome:               outcomeFromAgentAction(intent.Action, nullStringValue(intent.Outcome)),
			TransactionHash:       nullStringValue(execution.TransactionHash),
			CreatedAt:             intent.CreatedAt.Format("2006-01-02T15:04:05.000000000Z07:00"),
			UpdatedAt:             intent.UpdatedAt.Format("2006-01-02T15:04:05.000000000Z07:00"),
		})
	}

	sort.SliceStable(positions, func(i int, j int) bool {
		return positions[i].CreatedAt > positions[j].CreatedAt
	})

	return agentPortfolioResponse{
		AgentID:                        wallet.AgentID,
		AgentWalletAddress:             wallet.AgentWalletAddress,
		Chain:                          wallet.Chain,
		WalletProvider:                 wallet.WalletProvider,
		ActivePositionsCount:           len(positions),
		ResolvedOrClosedPositionsCount: 0,
		ClaimableRefundableCount:       0,
		TotalExposure:                  ratToDecimalString(totalExposure),
		Positions:                      positions,
		Settlements:                    []any{},
		UnavailableFields: []string{
			"positions are derived from executed agent buy intents and are not live contract balance readbacks",
			"resolved_or_closed_positions_count is unavailable until wallet-indexed position lifecycle records exist",
			"claimable_refundable_count is unavailable until wallet-indexed claim/refund eligibility records exist",
			"settlements are unavailable because existing settlement rows are internal-user keyed, not agent-wallet keyed",
		},
	}
}

func newAgentActivityResponse(agentID string, intents []repository.AgentIntent, executions []repository.AgentExecution) agentActivityResponse {
	items := []agentActivityItemResponse{}
	for _, intent := range intents {
		items = append(items, agentActivityItemResponse{
			ID:                    intent.IntentID,
			Type:                  "intent",
			IntentID:              intent.IntentID,
			Action:                intent.Action,
			Status:                intent.Status,
			MarketID:              nullStringValue(intent.MarketID),
			MarketContractAddress: nullStringValue(intent.MarketContractAddress),
			Amount:                nullStringValue(intent.Amount),
			Outcome:               outcomeFromAgentAction(intent.Action, nullStringValue(intent.Outcome)),
			Side:                  sideFromAgentAction(intent.Action),
			CreatedAt:             intent.CreatedAt.Format("2006-01-02T15:04:05.000000000Z07:00"),
			UpdatedAt:             intent.UpdatedAt.Format("2006-01-02T15:04:05.000000000Z07:00"),
		})
	}
	items = append(items, newAgentExecutionActivityItems(intents, executions)...)
	sort.SliceStable(items, func(i int, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	return agentActivityResponse{AgentID: agentID, Items: items}
}

func newAgentExecutionActivityItems(intents []repository.AgentIntent, executions []repository.AgentExecution) []agentActivityItemResponse {
	intentsByID := map[string]repository.AgentIntent{}
	for _, intent := range intents {
		intentsByID[intent.IntentID] = intent
	}

	items := make([]agentActivityItemResponse, 0, len(executions))
	for _, execution := range executions {
		intent := intentsByID[execution.IntentID]
		item := agentActivityItemResponse{
			ID:                     execution.ID,
			Type:                   "execution",
			IntentID:               execution.IntentID,
			Action:                 execution.Action,
			Status:                 execution.Status,
			MarketID:               nullStringValue(intent.MarketID),
			MarketContractAddress:  firstNonEmpty(nullStringValue(execution.MarketContractAddress), nullStringValue(intent.MarketContractAddress)),
			Amount:                 nullStringValue(intent.Amount),
			Outcome:                outcomeFromAgentAction(execution.Action, nullStringValue(intent.Outcome)),
			Side:                   sideFromAgentAction(execution.Action),
			TransactionHash:        nullStringValue(execution.TransactionHash),
			ApproveTransactionHash: nullStringValue(execution.ApproveTransactionHash),
			ErrorCode:              nullStringValue(execution.ErrorCode),
			ErrorMessage:           nullStringValue(execution.ErrorMessage),
			Readback:               jsonObjectMap(execution.Readback),
			CreatedAt:              execution.CreatedAt.Format("2006-01-02T15:04:05.000000000Z07:00"),
			UpdatedAt:              execution.UpdatedAt.Format("2006-01-02T15:04:05.000000000Z07:00"),
		}
		if execution.CompletedAt.Valid {
			item.CompletedAt = execution.CompletedAt.Time.Format("2006-01-02T15:04:05.000000000Z07:00")
		}
		items = append(items, item)
	}
	sort.SliceStable(items, func(i int, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	return items
}

func newDurableAgentIntentInput(intent agent.Intent) repository.CreateAgentIntentInput {
	validationResult, _ := json.Marshal(intent.ValidationResult)
	warnings, _ := json.Marshal(intent.Warnings)
	return repository.CreateAgentIntentInput{
		IntentID:              intent.ID,
		AgentID:               intent.AgentID,
		AgentWalletAddress:    intent.AgentWalletAddress,
		WalletProvider:        intent.WalletProvider,
		SourceClient:          intent.SourceClient,
		ClientRequestID:       intent.ClientRequestID,
		Action:                intent.Action,
		Status:                intent.Status,
		RequiresConfirmation:  intent.RequiresConfirmation,
		UserWallet:            intent.UserWallet,
		MarketID:              intent.MarketID,
		MarketContractAddress: intent.MarketContractAddress,
		Amount:                intent.Amount,
		Outcome:               intent.Outcome,
		Resolver:              intent.Resolver,
		CollateralToken:       intent.CollateralToken,
		CloseTimestamp:        intent.CloseTimestamp,
		Question:              intent.Question,
		ValidationResult:      validationResult,
		Warnings:              warnings,
	}
}

func newAgentIntentFromRepository(intent repository.AgentIntent) agent.Intent {
	validationResult := agent.ValidationResult{Valid: true, Errors: []string{}}
	if len(intent.ValidationResult) > 0 {
		_ = json.Unmarshal(intent.ValidationResult, &validationResult)
	}
	warnings := []string{}
	if len(intent.Warnings) > 0 {
		_ = json.Unmarshal(intent.Warnings, &warnings)
	}

	return agent.Intent{
		ID:                    intent.IntentID,
		AgentID:               nullStringValue(intent.AgentID),
		AgentWalletAddress:    nullStringValue(intent.AgentWalletAddress),
		WalletProvider:        nullStringValue(intent.WalletProvider),
		SourceClient:          nullStringValue(intent.SourceClient),
		ClientRequestID:       nullStringValue(intent.ClientRequestID),
		Action:                intent.Action,
		Status:                intent.Status,
		RequiresConfirmation:  intent.RequiresConfirmation,
		UserWallet:            nullStringValue(intent.UserWallet),
		MarketID:              nullStringValue(intent.MarketID),
		MarketContractAddress: nullStringValue(intent.MarketContractAddress),
		Amount:                nullStringValue(intent.Amount),
		Outcome:               nullStringValue(intent.Outcome),
		Resolver:              nullStringValue(intent.Resolver),
		CollateralToken:       nullStringValue(intent.CollateralToken),
		CloseTimestamp:        nullStringValue(intent.CloseTimestamp),
		Question:              nullStringValue(intent.Question),
		ValidationResult:      validationResult,
		Warnings:              warnings,
		CreatedAt:             intent.CreatedAt,
	}
}

func newAgentExecutionReadbackJSON(readback agent.ExecutionReadback) json.RawMessage {
	bytes, err := json.Marshal(agentReadbackResponse{
		MarketCount:     readback.MarketCount,
		CreatedMarket:   readback.CreatedMarket,
		IsMarket:        readback.IsMarket,
		MarketStatus:    readback.MarketStatus,
		WinningOutcome:  readback.WinningOutcome,
		YesPositions:    readback.YesPositions,
		NoPositions:     readback.NoPositions,
		TotalYes:        readback.TotalYes,
		TotalNo:         readback.TotalNo,
		TotalCollateral: readback.TotalCollateral,
		ClaimablePayout: readback.ClaimablePayout,
		ClaimableRefund: readback.ClaimableRefund,
		HasClaimed:      readback.HasClaimed,
		IsOpen:          readback.IsOpen,
		USDCBalance:     readback.USDCBalance,
		USDCAllowance:   readback.USDCAllowance,
	})
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return bytes
}

type agentExecutionFailure struct {
	Status  int
	Code    string
	Message string
}

var errAgentActionForbidden = errors.New("agent action forbidden")

func newAgentExecutionFailure(err error) agentExecutionFailure {
	switch {
	case errors.Is(err, agent.ErrExecutionProviderDisabled):
		return agentExecutionFailure{Status: http.StatusServiceUnavailable, Code: "agent_execution_provider_disabled", Message: "Circle Agent Wallet execution provider is disabled"}
	case errors.Is(err, agent.ErrExecutionNotImplemented):
		return agentExecutionFailure{Status: http.StatusNotImplemented, Code: "not_implemented", Message: "agent execution action is not implemented"}
	case errors.Is(err, agent.ErrCreateMarketCloseTimestampStale):
		return agentExecutionFailure{Status: http.StatusBadRequest, Code: "create_market_close_timestamp_stale", Message: "close_timestamp must be in the future before execution"}
	case errors.Is(err, agent.ErrCircleWalletIDMissing):
		return agentExecutionFailure{Status: http.StatusBadRequest, Code: "circle_wallet_id_missing", Message: "circle_wallet_id policy metadata is required for Circle API execution"}
	case errors.Is(err, agent.ErrIntentInvalid):
		return agentExecutionFailure{Status: http.StatusBadRequest, Code: "agent_intent_invalid", Message: "agent intent validation failed"}
	case errors.Is(err, agent.ErrIntentNotConfirmed):
		return agentExecutionFailure{Status: http.StatusConflict, Code: "agent_intent_not_confirmed", Message: "agent intent must be confirmed before execution"}
	case errors.Is(err, agent.ErrExecutionConfigInvalid):
		return agentExecutionFailure{Status: http.StatusServiceUnavailable, Code: "agent_execution_config_invalid", Message: "agent execution environment is not configured"}
	default:
		return agentExecutionFailure{Status: http.StatusBadGateway, Code: "agent_execution_failed", Message: "agent execution failed"}
	}
}

func validateAgentIDPath(w http.ResponseWriter, rawAgentID string) (string, bool) {
	agentID, validationErrors := validateAgentID(rawAgentID)
	if len(validationErrors) > 0 {
		httpjson.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{
				"code":    "agent_id_invalid",
				"message": "agent_id validation failed",
				"details": validationErrors,
			},
		})
		return "", false
	}
	return agentID, true
}

func writeAgentIntentInvalid(w http.ResponseWriter, details []string) {
	errorBody := map[string]any{
		"code":    "agent_intent_invalid",
		"message": "agent intent validation failed",
	}
	if len(details) > 0 {
		errorBody["details"] = details
	}
	httpjson.WriteJSON(w, http.StatusBadRequest, map[string]any{
		"error": errorBody,
	})
}

func writeAgentActionForbidden(w http.ResponseWriter) {
	httpjson.WriteError(w, http.StatusForbidden, "agent_action_forbidden", "agent action is not allowed for this agent wallet")
}

func writeAgentPolicyViolation(w http.ResponseWriter, details []string) {
	errorBody := map[string]any{
		"code":    "agent_policy_violation",
		"message": "agent policy violation",
	}
	if len(details) > 0 {
		errorBody["details"] = details
	}
	httpjson.WriteJSON(w, http.StatusForbidden, map[string]any{
		"error": errorBody,
	})
}

func validateAgentIntentPolicyForConfirmation(w http.ResponseWriter, r *http.Request, walletRegistry agentWalletRegistry, intent agent.Intent) bool {
	if intent.AgentID == "" || !isBackendExecutableAgentAction(intent.Action) {
		return true
	}
	wallet, err := walletRegistry.GetAgentWalletByAgentID(r.Context(), intent.AgentID)
	if err != nil {
		return true
	}
	if !repositoryAgentWalletAllowsAction(wallet, intent.Action) {
		writeAgentActionForbidden(w)
		return false
	}
	if policyViolations := agentWalletPolicyViolations(intent, wallet); len(policyViolations) > 0 {
		writeAgentPolicyViolation(w, policyViolations)
		return false
	}
	return true
}

func agentWalletPolicyViolations(intent agent.Intent, wallet repository.AgentWallet) []string {
	if wallet.AgentID == "" {
		return nil
	}
	if intent.Action != agent.ActionBuyYes && intent.Action != agent.ActionBuyNo {
		return nil
	}

	amount, ok := new(big.Rat).SetString(strings.TrimSpace(intent.Amount))
	if !ok || amount.Sign() <= 0 {
		return []string{"amount must be positive and finite"}
	}

	maxAmount, ok := agentPolicyMetadataString(wallet.PolicyMetadata, "max_trade_amount")
	if !ok || maxAmount == "" {
		return nil
	}
	max, ok := new(big.Rat).SetString(maxAmount)
	if !ok || max.Sign() <= 0 {
		return nil
	}
	if amount.Cmp(max) > 0 {
		return []string{"amount exceeds max_trade_amount policy"}
	}
	return nil
}

func agentPolicyMetadataString(raw json.RawMessage, key string) (string, bool) {
	if len(raw) == 0 {
		return "", false
	}
	metadata := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return "", false
	}
	value, ok := metadata[key]
	if !ok {
		return "", false
	}
	var stringValue string
	if err := json.Unmarshal(value, &stringValue); err == nil {
		return strings.TrimSpace(stringValue), true
	}
	return strings.TrimSpace(string(value)), true
}

func agentPolicyMetadataMap(raw json.RawMessage) map[string]string {
	metadata := map[string]string{}
	if len(raw) == 0 {
		return metadata
	}
	_ = json.Unmarshal(raw, &metadata)
	return metadata
}

func outcomeFromAgentAction(action string, fallback string) string {
	if fallback != "" {
		return fallback
	}
	switch action {
	case agent.ActionBuyYes:
		return "yes"
	case agent.ActionBuyNo:
		return "no"
	default:
		return ""
	}
}

func sideFromAgentAction(action string) string {
	switch action {
	case agent.ActionBuyYes, agent.ActionBuyNo:
		return "buy"
	default:
		return ""
	}
}

func ratToDecimalString(value *big.Rat) string {
	if value == nil || value.Sign() == 0 {
		return "0"
	}
	if value.IsInt() {
		return value.Num().String()
	}
	return value.FloatString(18)
}

func jsonObjectMap(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	decoded := map[string]any{}
	if err := json.Unmarshal(raw, &decoded); err != nil || len(decoded) == 0 {
		return nil
	}
	return decoded
}

func newAgentExecutionPlanResponse(executionPlan agent.ExecutionPlan) agentExecutionPlanResponse {
	return agentExecutionPlanResponse{
		IntentID:            executionPlan.IntentID,
		Action:              executionPlan.Action,
		Status:              executionPlan.Status,
		ExecutionMode:       executionPlan.ExecutionMode,
		Network:             executionPlan.Network,
		AgentFactoryAddress: executionPlan.AgentFactoryAddress,
		RequiresSignature:   executionPlan.RequiresSignature,
		BroadcastPerformed:  executionPlan.BroadcastPerformed,
		TransactionHash:     executionPlan.TransactionHash,
		TransactionRequest:  newTransactionRequestResponse(executionPlan.TransactionRequest),
		Warnings:            executionPlan.Warnings,
	}
}

func newTransactionRequestResponse(transactionRequest agent.TransactionRequest) transactionRequestResponse {
	return transactionRequestResponse{
		To:                 transactionRequest.To,
		Contract:           transactionRequest.Contract,
		Function:           transactionRequest.Function,
		Args:               transactionRequest.Args,
		Value:              transactionRequest.Value,
		Chain:              transactionRequest.Chain,
		BroadcastPerformed: transactionRequest.BroadcastPerformed,
	}
}

func newAgentExecutionResponse(result agent.ExecutionResult) agentExecutionResponse {
	return agentExecutionResponse{
		IntentID:               result.IntentID,
		AgentID:                result.AgentID,
		AgentWalletAddress:     result.AgentWalletAddress,
		WalletProvider:         result.WalletProvider,
		Action:                 result.Action,
		Status:                 result.Status,
		ExecutionMode:          result.ExecutionMode,
		Network:                result.Network,
		AgentFactoryAddress:    result.AgentFactoryAddress,
		MarketContractAddress:  result.MarketContractAddress,
		BroadcastPerformed:     result.BroadcastPerformed,
		ApproveTransactionHash: result.ApproveTransactionHash,
		TransactionHash:        result.TransactionHash,
		Readback: agentReadbackResponse{
			MarketCount:     result.Readback.MarketCount,
			CreatedMarket:   result.Readback.CreatedMarket,
			IsMarket:        result.Readback.IsMarket,
			MarketStatus:    result.Readback.MarketStatus,
			WinningOutcome:  result.Readback.WinningOutcome,
			YesPositions:    result.Readback.YesPositions,
			NoPositions:     result.Readback.NoPositions,
			TotalYes:        result.Readback.TotalYes,
			TotalNo:         result.Readback.TotalNo,
			TotalCollateral: result.Readback.TotalCollateral,
			ClaimablePayout: result.Readback.ClaimablePayout,
			ClaimableRefund: result.Readback.ClaimableRefund,
			HasClaimed:      result.Readback.HasClaimed,
			IsOpen:          result.Readback.IsOpen,
			USDCBalance:     result.Readback.USDCBalance,
			USDCAllowance:   result.Readback.USDCAllowance,
		},
	}
}

func validateAgentWalletForExecution(intent agent.Intent, wallet repository.AgentWallet) error {
	if intent.AgentID == "" {
		return errors.New("agent_id is required for execution")
	}
	if wallet.AgentWalletAddress == "" {
		return errors.New("agent wallet address is required for execution")
	}
	if wallet.Status != agent.WalletStatusActive {
		return errors.New("agent wallet is not active")
	}
	if wallet.Chain != agent.ChainArcTestnet {
		return errors.New("agent wallet chain must be ARC-TESTNET")
	}
	if !repositoryAgentWalletAllowsAction(wallet, intent.Action) {
		return errAgentActionForbidden
	}
	if equalAddress(wallet.AgentWalletAddress, knownDeployerResolverWallet()) {
		return errors.New("agent wallet must not equal the deployer/resolver wallet")
	}
	if wallet.UserWallet.Valid && equalAddress(wallet.AgentWalletAddress, wallet.UserWallet.String) {
		return errors.New("agent wallet must not equal the user wallet unless a documented user-controlled custody link is implemented")
	}
	for _, forbidden := range configuredForbiddenAgentWallets() {
		if equalAddress(wallet.AgentWalletAddress, forbidden) {
			return errors.New("agent wallet matches a configured forbidden wallet")
		}
	}
	if intent.AgentWalletAddress != "" && !equalAddress(intent.AgentWalletAddress, wallet.AgentWalletAddress) {
		return errors.New("intent agent_wallet_address does not match registered agent wallet")
	}
	return nil
}

func repositoryAgentWalletAllowsAction(wallet repository.AgentWallet, action string) bool {
	for _, allowedAction := range wallet.AllowedActions {
		if allowedAction == action {
			return true
		}
	}
	return false
}

func readAgentWalletBalances(ctx context.Context, balanceReader agent.CircleAgentWalletBalanceReader, walletResolver agent.CircleWalletResolver, wallet repository.AgentWallet) (agent.CircleAgentWalletBalances, error) {
	if balanceReader != nil {
		return balanceReader.GetAgentWalletBalances(ctx, agent.CircleAgentWalletBalanceRequest{
			AgentID:            wallet.AgentID,
			AgentWalletAddress: wallet.AgentWalletAddress,
			WalletProvider:     wallet.WalletProvider,
			Chain:              wallet.Chain,
			PolicyMetadata:     agentPolicyMetadataMap(wallet.PolicyMetadata),
		})
	}
	if walletResolver == nil {
		return agent.CircleAgentWalletBalances{}, agent.ErrCircleAgentWalletBalanceFailed
	}
	return walletResolver.GetAgentWalletBalances(ctx, wallet.AgentWalletAddress)
}

func knownDeployerResolverWallet() string {
	return "0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC"
}

func configuredForbiddenAgentWallets() []string {
	values := []string{knownDeployerResolverWallet()}
	for _, value := range strings.Split(os.Getenv("SIGNALARC_FORBIDDEN_AGENT_WALLETS"), ",") {
		value = strings.TrimSpace(value)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func equalAddress(left string, right string) bool {
	return strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right))
}
