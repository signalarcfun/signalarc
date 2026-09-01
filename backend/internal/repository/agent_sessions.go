package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/database"
)

const (
	AgentOnboardingStatusPendingOTP = "pending_otp"
	AgentOnboardingStatusVerified   = "verified"
	AgentOnboardingStatusExpired    = "expired"
	AgentOnboardingStatusFailed     = "failed"
	AgentOnboardingStatusCancelled  = "cancelled"

	AgentSessionStatusActive   = "active"
	AgentSessionStatusDisabled = "disabled"
	AgentSessionStatusRevoked  = "revoked"
	AgentSessionStatusExpired  = "expired"
)

var ErrInvalidAgentSession = errors.New("invalid agent session")

type AgentOnboardingSession struct {
	ID                          string          `json:"id"`
	OnboardingID                string          `json:"onboarding_id"`
	AgentID                     string          `json:"agent_id"`
	UserEmail                   string          `json:"user_email"`
	UserWallet                  sql.NullString  `json:"user_wallet"`
	RequestedAgentWalletAddress sql.NullString  `json:"requested_agent_wallet_address"`
	SourceClient                sql.NullString  `json:"source_client"`
	Channel                     sql.NullString  `json:"channel"`
	Chain                       string          `json:"chain"`
	WalletProvider              string          `json:"wallet_provider"`
	Status                      string          `json:"status"`
	CircleRequestIDHash         sql.NullString  `json:"circle_request_id_hash"`
	CircleRequestExpiresAt      sql.NullTime    `json:"circle_request_expires_at"`
	FailureReason               sql.NullString  `json:"failure_reason"`
	PolicyMetadata              json.RawMessage `json:"policy_metadata"`
	CreatedAt                   time.Time       `json:"created_at"`
	UpdatedAt                   time.Time       `json:"updated_at"`
}

type CreateAgentOnboardingSessionInput struct {
	OnboardingID                string
	AgentID                     string
	UserEmail                   string
	UserWallet                  sql.NullString
	RequestedAgentWalletAddress sql.NullString
	SourceClient                sql.NullString
	Channel                     sql.NullString
	Chain                       string
	WalletProvider              string
	Status                      string
	CircleRequestIDHash         sql.NullString
	CircleRequestExpiresAt      sql.NullTime
	FailureReason               sql.NullString
	PolicyMetadata              json.RawMessage
}

type AgentSession struct {
	ID                 string          `json:"id"`
	SessionID          string          `json:"session_id"`
	AgentID            string          `json:"agent_id"`
	UserEmail          string          `json:"user_email"`
	UserWallet         string          `json:"user_wallet"`
	AgentWalletAddress string          `json:"agent_wallet_address"`
	WalletProvider     string          `json:"wallet_provider"`
	Chain              string          `json:"chain"`
	Status             string          `json:"status"`
	AllowedActions     []string        `json:"allowed_actions"`
	AllowedChannels    []string        `json:"allowed_channels"`
	SessionMetadata    json.RawMessage `json:"session_metadata"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type CreateAgentSessionInput struct {
	SessionID          string
	AgentID            string
	UserEmail          string
	UserWallet         string
	AgentWalletAddress string
	WalletProvider     string
	Chain              string
	Status             string
	AllowedActions     []string
	AllowedChannels    []string
	SessionMetadata    json.RawMessage
}

type AgentSessionsRepository struct {
	db *database.DB
}

func NewAgentSessionsRepository(db *database.DB) *AgentSessionsRepository {
	return &AgentSessionsRepository{db: db}
}

func (r *AgentSessionsRepository) CreateAgentOnboardingSession(ctx context.Context, input CreateAgentOnboardingSessionInput) (AgentOnboardingSession, error) {
	var session AgentOnboardingSession
	err := r.db.QueryRow(ctx, `
		INSERT INTO agent_onboarding_sessions (
			onboarding_id,
			agent_id,
			user_email,
			user_wallet,
			requested_agent_wallet_address,
			source_client,
			channel,
			chain,
			wallet_provider,
			status,
			circle_request_id_hash,
			circle_request_expires_at,
			failure_reason,
			policy_metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING
			id::text,
			onboarding_id,
			agent_id,
			user_email,
			user_wallet,
			requested_agent_wallet_address,
			source_client,
			channel,
			chain,
			wallet_provider,
			status,
			circle_request_id_hash,
			circle_request_expires_at,
			failure_reason,
			COALESCE(policy_metadata, '{}'::jsonb),
			created_at,
			updated_at
	`,
		input.OnboardingID,
		input.AgentID,
		input.UserEmail,
		input.UserWallet,
		input.RequestedAgentWalletAddress,
		input.SourceClient,
		input.Channel,
		input.Chain,
		input.WalletProvider,
		input.Status,
		input.CircleRequestIDHash,
		input.CircleRequestExpiresAt,
		input.FailureReason,
		nullableJSON(input.PolicyMetadata),
	).Scan(agentOnboardingSessionScanDestinations(&session)...)

	return session, err
}

func (r *AgentSessionsRepository) GetAgentOnboardingSessionByOnboardingID(ctx context.Context, onboardingID string) (AgentOnboardingSession, error) {
	var session AgentOnboardingSession
	err := r.db.QueryRow(ctx, agentOnboardingSessionSelectSQL+`
		WHERE onboarding_id = $1
	`, onboardingID).Scan(agentOnboardingSessionScanDestinations(&session)...)

	return session, err
}

func (r *AgentSessionsRepository) UpdateAgentOnboardingSessionStatus(ctx context.Context, onboardingID string, status string, failureReason sql.NullString) (AgentOnboardingSession, error) {
	var session AgentOnboardingSession
	err := r.db.QueryRow(ctx, `
		UPDATE agent_onboarding_sessions
		SET status = $2,
			failure_reason = $3,
			updated_at = now()
		WHERE onboarding_id = $1
		RETURNING
			id::text,
			onboarding_id,
			agent_id,
			user_email,
			user_wallet,
			requested_agent_wallet_address,
			source_client,
			channel,
			chain,
			wallet_provider,
			status,
			circle_request_id_hash,
			circle_request_expires_at,
			failure_reason,
			COALESCE(policy_metadata, '{}'::jsonb),
			created_at,
			updated_at
	`, onboardingID, status, failureReason).Scan(agentOnboardingSessionScanDestinations(&session)...)

	return session, err
}

func (r *AgentSessionsRepository) UpdateAgentOnboardingSessionOTPStart(ctx context.Context, onboardingID string, requestIDHash string, expiresAt time.Time) (AgentOnboardingSession, error) {
	var session AgentOnboardingSession
	err := r.db.QueryRow(ctx, `
		UPDATE agent_onboarding_sessions
		SET circle_request_id_hash = $2,
			circle_request_expires_at = $3,
			updated_at = now()
		WHERE onboarding_id = $1
		RETURNING
			id::text,
			onboarding_id,
			agent_id,
			user_email,
			user_wallet,
			requested_agent_wallet_address,
			source_client,
			channel,
			chain,
			wallet_provider,
			status,
			circle_request_id_hash,
			circle_request_expires_at,
			failure_reason,
			COALESCE(policy_metadata, '{}'::jsonb),
			created_at,
			updated_at
	`, onboardingID, requestIDHash, expiresAt).Scan(agentOnboardingSessionScanDestinations(&session)...)

	return session, err
}

func (r *AgentSessionsRepository) CreateAgentSession(ctx context.Context, input CreateAgentSessionInput) (AgentSession, error) {
	if err := validateCreateAgentSessionInput(input); err != nil {
		return AgentSession{}, err
	}

	var session AgentSession
	err := r.db.QueryRow(ctx, `
		INSERT INTO agent_sessions (
			session_id,
			agent_id,
			user_email,
			user_wallet,
			agent_wallet_address,
			wallet_provider,
			chain,
			status,
			allowed_actions,
			allowed_channels,
			session_metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING
			id::text,
			session_id,
			agent_id,
			user_email,
			user_wallet,
			agent_wallet_address,
			wallet_provider,
			chain,
			status,
			allowed_actions,
			allowed_channels,
			COALESCE(session_metadata, '{}'::jsonb),
			created_at,
			updated_at
	`,
		input.SessionID,
		input.AgentID,
		input.UserEmail,
		input.UserWallet,
		input.AgentWalletAddress,
		input.WalletProvider,
		input.Chain,
		input.Status,
		input.AllowedActions,
		input.AllowedChannels,
		nullableJSON(input.SessionMetadata),
	).Scan(agentSessionScanDestinations(&session)...)

	return session, err
}

func (r *AgentSessionsRepository) GetAgentSessionByAgentID(ctx context.Context, agentID string) (AgentSession, error) {
	var session AgentSession
	err := r.db.QueryRow(ctx, agentSessionSelectSQL+`
		WHERE agent_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, agentID).Scan(agentSessionScanDestinations(&session)...)

	return session, err
}

func (r *AgentSessionsRepository) GetAgentSessionBySessionID(ctx context.Context, sessionID string) (AgentSession, error) {
	var session AgentSession
	err := r.db.QueryRow(ctx, agentSessionSelectSQL+`
		WHERE session_id = $1
	`, sessionID).Scan(agentSessionScanDestinations(&session)...)

	return session, err
}

func validateCreateAgentSessionInput(input CreateAgentSessionInput) error {
	if strings.TrimSpace(input.SessionID) == "" ||
		strings.TrimSpace(input.AgentID) == "" ||
		strings.TrimSpace(input.UserEmail) == "" ||
		strings.TrimSpace(input.AgentWalletAddress) == "" {
		return ErrInvalidAgentSession
	}
	if !strings.EqualFold(strings.TrimSpace(input.Chain), "ARC-TESTNET") {
		return ErrInvalidAgentSession
	}
	if strings.TrimSpace(input.WalletProvider) != "circle_agent_wallet" {
		return ErrInvalidAgentSession
	}
	if strings.TrimSpace(input.Status) == "" {
		return ErrInvalidAgentSession
	}
	if len(input.AllowedActions) == 0 || len(input.AllowedChannels) == 0 {
		return ErrInvalidAgentSession
	}
	if strings.TrimSpace(input.UserWallet) != "" && strings.EqualFold(strings.TrimSpace(input.AgentWalletAddress), strings.TrimSpace(input.UserWallet)) {
		return ErrInvalidAgentSession
	}
	return nil
}

const agentOnboardingSessionSelectSQL = `
	SELECT
		id::text,
		onboarding_id,
		agent_id,
		user_email,
		user_wallet,
		requested_agent_wallet_address,
		source_client,
		channel,
		chain,
		wallet_provider,
		status,
		circle_request_id_hash,
		circle_request_expires_at,
		failure_reason,
		COALESCE(policy_metadata, '{}'::jsonb),
		created_at,
		updated_at
	FROM agent_onboarding_sessions
`

func agentOnboardingSessionScanDestinations(session *AgentOnboardingSession) []any {
	return []any{
		&session.ID,
		&session.OnboardingID,
		&session.AgentID,
		&session.UserEmail,
		&session.UserWallet,
		&session.RequestedAgentWalletAddress,
		&session.SourceClient,
		&session.Channel,
		&session.Chain,
		&session.WalletProvider,
		&session.Status,
		&session.CircleRequestIDHash,
		&session.CircleRequestExpiresAt,
		&session.FailureReason,
		&session.PolicyMetadata,
		&session.CreatedAt,
		&session.UpdatedAt,
	}
}

const agentSessionSelectSQL = `
	SELECT
		id::text,
		session_id,
		agent_id,
		user_email,
		user_wallet,
		agent_wallet_address,
		wallet_provider,
		chain,
		status,
		allowed_actions,
		allowed_channels,
		COALESCE(session_metadata, '{}'::jsonb),
		created_at,
		updated_at
	FROM agent_sessions
`

func agentSessionScanDestinations(session *AgentSession) []any {
	return []any{
		&session.ID,
		&session.SessionID,
		&session.AgentID,
		&session.UserEmail,
		&session.UserWallet,
		&session.AgentWalletAddress,
		&session.WalletProvider,
		&session.Chain,
		&session.Status,
		&session.AllowedActions,
		&session.AllowedChannels,
		&session.SessionMetadata,
		&session.CreatedAt,
		&session.UpdatedAt,
	}
}
