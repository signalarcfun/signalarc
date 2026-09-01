package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/database"
)

type AgentWallet struct {
	ID                 string          `json:"id"`
	AgentID            string          `json:"agent_id"`
	UserWallet         sql.NullString  `json:"user_wallet"`
	UserEmail          sql.NullString  `json:"user_email"`
	AgentWalletAddress string          `json:"agent_wallet_address"`
	WalletProvider     string          `json:"wallet_provider"`
	Chain              string          `json:"chain"`
	Status             string          `json:"status"`
	AllowedActions     []string        `json:"allowed_actions"`
	PolicyMetadata     json.RawMessage `json:"policy_metadata"`
	SourceClient       sql.NullString  `json:"source_client"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type UpsertAgentWalletInput struct {
	AgentID            string
	UserWallet         string
	UserEmail          sql.NullString
	AgentWalletAddress string
	WalletProvider     string
	Chain              string
	Status             string
	AllowedActions     []string
	PolicyMetadata     json.RawMessage
	SourceClient       sql.NullString
}

type AgentWalletsRepository struct {
	db *database.DB
}

func NewAgentWalletsRepository(db *database.DB) *AgentWalletsRepository {
	return &AgentWalletsRepository{db: db}
}

func (r *AgentWalletsRepository) RegisterAgentWallet(ctx context.Context, input UpsertAgentWalletInput) (AgentWallet, error) {
	var wallet AgentWallet
	err := r.db.QueryRow(ctx, `
		INSERT INTO agent_wallets (
			agent_id,
			user_wallet,
			user_email,
			agent_wallet_address,
			wallet_provider,
			chain,
			status,
			allowed_actions,
			policy_metadata,
			source_client
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (agent_id) DO UPDATE
		SET
			user_wallet = EXCLUDED.user_wallet,
			user_email = EXCLUDED.user_email,
			agent_wallet_address = EXCLUDED.agent_wallet_address,
			wallet_provider = EXCLUDED.wallet_provider,
			chain = EXCLUDED.chain,
			status = EXCLUDED.status,
			allowed_actions = EXCLUDED.allowed_actions,
			policy_metadata = EXCLUDED.policy_metadata,
			source_client = EXCLUDED.source_client,
			updated_at = now()
		RETURNING
			id::text,
			agent_id,
			user_wallet,
			user_email,
			agent_wallet_address,
			wallet_provider,
			chain,
			status,
			allowed_actions,
			COALESCE(policy_metadata, '{}'::jsonb),
			source_client,
			created_at,
			updated_at
	`,
		input.AgentID,
		nullableText(input.UserWallet),
		input.UserEmail,
		input.AgentWalletAddress,
		input.WalletProvider,
		input.Chain,
		input.Status,
		input.AllowedActions,
		nullableJSON(input.PolicyMetadata),
		input.SourceClient,
	).Scan(agentWalletScanDestinations(&wallet)...)

	return wallet, err
}

func (r *AgentWalletsRepository) GetAgentWalletByAgentID(ctx context.Context, agentID string) (AgentWallet, error) {
	var wallet AgentWallet
	err := r.db.QueryRow(ctx, agentWalletSelectSQL+`
		WHERE agent_id = $1
	`, agentID).Scan(agentWalletScanDestinations(&wallet)...)

	return wallet, err
}

func (r *AgentWalletsRepository) DisableAgentWallet(ctx context.Context, agentID string) (AgentWallet, error) {
	var wallet AgentWallet
	err := r.db.QueryRow(ctx, agentWalletSelectSQL+`
		WHERE agent_id = $1
		FOR UPDATE
	`, agentID).Scan(agentWalletScanDestinations(&wallet)...)
	if err != nil {
		return wallet, err
	}

	err = r.db.QueryRow(ctx, `
		UPDATE agent_wallets
		SET status = 'disabled', updated_at = now()
		WHERE agent_id = $1
		RETURNING
			id::text,
			agent_id,
			user_wallet,
			user_email,
			agent_wallet_address,
			wallet_provider,
			chain,
			status,
			allowed_actions,
			COALESCE(policy_metadata, '{}'::jsonb),
			source_client,
			created_at,
			updated_at
	`, agentID).Scan(agentWalletScanDestinations(&wallet)...)

	return wallet, err
}

const agentWalletSelectSQL = `
	SELECT
		id::text,
		agent_id,
		user_wallet,
		user_email,
		agent_wallet_address,
		wallet_provider,
		chain,
		status,
		allowed_actions,
		COALESCE(policy_metadata, '{}'::jsonb),
		source_client,
		created_at,
		updated_at
	FROM agent_wallets
`

func agentWalletScanDestinations(wallet *AgentWallet) []any {
	return []any{
		&wallet.ID,
		&wallet.AgentID,
		&wallet.UserWallet,
		&wallet.UserEmail,
		&wallet.AgentWalletAddress,
		&wallet.WalletProvider,
		&wallet.Chain,
		&wallet.Status,
		&wallet.AllowedActions,
		&wallet.PolicyMetadata,
		&wallet.SourceClient,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	}
}

func nullableJSON(value json.RawMessage) any {
	if len(value) == 0 {
		return nil
	}
	return value
}

func nullableText(value string) sql.NullString {
	value = strings.TrimSpace(value)
	return sql.NullString{String: value, Valid: value != ""}
}
