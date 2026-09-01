package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/wahyu241205/SignalArc/backend/internal/database"
)

const (
	AgentExecutionStatusPending  = "pending"
	AgentExecutionStatusExecuted = "executed"
	AgentExecutionStatusFailed   = "failed"
)

type AgentIntent struct {
	ID                    string          `json:"id"`
	IntentID              string          `json:"intent_id"`
	AgentID               sql.NullString  `json:"agent_id"`
	AgentWalletAddress    sql.NullString  `json:"agent_wallet_address"`
	WalletProvider        sql.NullString  `json:"wallet_provider"`
	SourceClient          sql.NullString  `json:"source_client"`
	ClientRequestID       sql.NullString  `json:"client_request_id"`
	Action                string          `json:"action"`
	Status                string          `json:"status"`
	RequiresConfirmation  bool            `json:"requires_confirmation"`
	UserWallet            sql.NullString  `json:"user_wallet"`
	MarketID              sql.NullString  `json:"market_id"`
	MarketContractAddress sql.NullString  `json:"market_contract_address"`
	Amount                sql.NullString  `json:"amount"`
	Outcome               sql.NullString  `json:"outcome"`
	Resolver              sql.NullString  `json:"resolver"`
	CollateralToken       sql.NullString  `json:"collateral_token"`
	CloseTimestamp        sql.NullString  `json:"close_timestamp"`
	Question              sql.NullString  `json:"question"`
	ValidationResult      json.RawMessage `json:"validation_result"`
	Warnings              json.RawMessage `json:"warnings"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	ConfirmedAt           sql.NullTime    `json:"confirmed_at"`
	ExecutedAt            sql.NullTime    `json:"executed_at"`
}

type AgentExecution struct {
	ID                     string          `json:"id"`
	IntentID               string          `json:"intent_id"`
	AgentID                sql.NullString  `json:"agent_id"`
	Action                 string          `json:"action"`
	Status                 string          `json:"status"`
	ExecutionMode          sql.NullString  `json:"execution_mode"`
	Network                sql.NullString  `json:"network"`
	AgentFactoryAddress    sql.NullString  `json:"agent_factory_address"`
	MarketContractAddress  sql.NullString  `json:"market_contract_address"`
	ApproveTransactionHash sql.NullString  `json:"approve_transaction_hash"`
	TransactionHash        sql.NullString  `json:"transaction_hash"`
	BroadcastPerformed     bool            `json:"broadcast_performed"`
	Readback               json.RawMessage `json:"readback"`
	ErrorCode              sql.NullString  `json:"error_code"`
	ErrorMessage           sql.NullString  `json:"error_message"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
	CompletedAt            sql.NullTime    `json:"completed_at"`
}

type CreateAgentIntentInput struct {
	IntentID              string
	AgentID               string
	AgentWalletAddress    string
	WalletProvider        string
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
	ValidationResult      json.RawMessage
	Warnings              json.RawMessage
}

type CreateAgentExecutionInput struct {
	IntentID              string
	AgentID               string
	Action                string
	ExecutionMode         string
	Network               string
	AgentFactoryAddress   string
	MarketContractAddress string
}

type CompleteAgentExecutionInput struct {
	ExecutionMode          string
	Network                string
	AgentFactoryAddress    string
	MarketContractAddress  string
	ApproveTransactionHash string
	TransactionHash        string
	BroadcastPerformed     bool
	Readback               json.RawMessage
}

type FailAgentExecutionInput struct {
	ErrorCode    string
	ErrorMessage string
	Readback     json.RawMessage
}

type AgentIntentsRepository struct {
	db *database.DB
}

func NewAgentIntentsRepository(db *database.DB) *AgentIntentsRepository {
	return &AgentIntentsRepository{db: db}
}

func (r *AgentIntentsRepository) CreateAgentIntent(ctx context.Context, input CreateAgentIntentInput) (AgentIntent, error) {
	if existing, ok, err := r.getIntentByIdempotencyKey(ctx, input.AgentID, input.SourceClient, input.ClientRequestID); err != nil || ok {
		return existing, err
	}

	var intent AgentIntent
	err := r.db.QueryRow(ctx, `
		INSERT INTO agent_intents (
			intent_id,
			agent_id,
			agent_wallet_address,
			wallet_provider,
			source_client,
			client_request_id,
			action,
			status,
			requires_confirmation,
			user_wallet,
			market_id,
			market_contract_address,
			amount,
			outcome,
			resolver,
			collateral_token,
			close_timestamp,
			question,
			validation_result,
			warnings
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		RETURNING
			id::text,
			intent_id,
			agent_id,
			agent_wallet_address,
			wallet_provider,
			source_client,
			client_request_id,
			action,
			status,
			requires_confirmation,
			user_wallet,
			market_id,
			market_contract_address,
			amount,
			outcome,
			resolver,
			collateral_token,
			close_timestamp,
			question,
			validation_result,
			warnings,
			created_at,
			updated_at,
			confirmed_at,
			executed_at
	`,
		input.IntentID,
		nullableText(input.AgentID),
		nullableText(input.AgentWalletAddress),
		nullableText(input.WalletProvider),
		nullableText(input.SourceClient),
		nullableText(input.ClientRequestID),
		input.Action,
		input.Status,
		input.RequiresConfirmation,
		nullableText(input.UserWallet),
		nullableText(input.MarketID),
		nullableText(input.MarketContractAddress),
		nullableText(input.Amount),
		nullableText(input.Outcome),
		nullableText(input.Resolver),
		nullableText(input.CollateralToken),
		nullableText(input.CloseTimestamp),
		nullableText(input.Question),
		nullableJSON(input.ValidationResult),
		nullableJSON(input.Warnings),
	).Scan(agentIntentScanDestinations(&intent)...)

	return intent, err
}

func (r *AgentIntentsRepository) GetAgentIntentByIntentID(ctx context.Context, intentID string) (AgentIntent, error) {
	var intent AgentIntent
	err := r.db.QueryRow(ctx, agentIntentSelectSQL+`
		WHERE intent_id = $1
	`, intentID).Scan(agentIntentScanDestinations(&intent)...)
	return intent, normalizeAgentIntentNoRows(err)
}

func (r *AgentIntentsRepository) ConfirmAgentIntent(ctx context.Context, intentID string) (AgentIntent, error) {
	var intent AgentIntent
	err := r.db.QueryRow(ctx, `
		UPDATE agent_intents
		SET status = 'confirmed',
			confirmed_at = COALESCE(confirmed_at, now()),
			updated_at = now()
		WHERE intent_id = $1
		RETURNING
			id::text,
			intent_id,
			agent_id,
			agent_wallet_address,
			wallet_provider,
			source_client,
			client_request_id,
			action,
			status,
			requires_confirmation,
			user_wallet,
			market_id,
			market_contract_address,
			amount,
			outcome,
			resolver,
			collateral_token,
			close_timestamp,
			question,
			validation_result,
			warnings,
			created_at,
			updated_at,
			confirmed_at,
			executed_at
	`, intentID).Scan(agentIntentScanDestinations(&intent)...)
	return intent, normalizeAgentIntentNoRows(err)
}

func (r *AgentIntentsRepository) MarkAgentIntentExecuted(ctx context.Context, intentID string) (AgentIntent, error) {
	return r.updateIntentTerminalStatus(ctx, intentID, "executed")
}

func (r *AgentIntentsRepository) MarkAgentIntentFailed(ctx context.Context, intentID string) (AgentIntent, error) {
	return r.updateIntentTerminalStatus(ctx, intentID, "failed")
}

func (r *AgentIntentsRepository) CreateAgentExecution(ctx context.Context, input CreateAgentExecutionInput) (AgentExecution, error) {
	var execution AgentExecution
	err := r.db.QueryRow(ctx, `
		INSERT INTO agent_executions (
			intent_id,
			agent_id,
			action,
			status,
			execution_mode,
			network,
			agent_factory_address,
			market_contract_address
		)
		VALUES ($1, $2, $3, 'pending', $4, $5, $6, $7)
		RETURNING
			id::text,
			intent_id,
			agent_id,
			action,
			status,
			execution_mode,
			network,
			agent_factory_address,
			market_contract_address,
			approve_transaction_hash,
			transaction_hash,
			broadcast_performed,
			readback,
			error_code,
			error_message,
			created_at,
			updated_at,
			completed_at
	`, input.IntentID, nullableText(input.AgentID), input.Action, nullableText(input.ExecutionMode), nullableText(input.Network), nullableText(input.AgentFactoryAddress), nullableText(input.MarketContractAddress)).Scan(agentExecutionScanDestinations(&execution)...)
	return execution, err
}

func (r *AgentIntentsRepository) MarkAgentExecutionExecuted(ctx context.Context, id string, input CompleteAgentExecutionInput) (AgentExecution, error) {
	var execution AgentExecution
	err := r.db.QueryRow(ctx, `
		UPDATE agent_executions
		SET status = 'executed',
			execution_mode = $2,
			network = $3,
			agent_factory_address = $4,
			market_contract_address = $5,
			approve_transaction_hash = $6,
			transaction_hash = $7,
			broadcast_performed = $8,
			readback = $9,
			error_code = NULL,
			error_message = NULL,
			completed_at = now(),
			updated_at = now()
		WHERE id = $1
		RETURNING
			id::text,
			intent_id,
			agent_id,
			action,
			status,
			execution_mode,
			network,
			agent_factory_address,
			market_contract_address,
			approve_transaction_hash,
			transaction_hash,
			broadcast_performed,
			readback,
			error_code,
			error_message,
			created_at,
			updated_at,
			completed_at
	`, id, nullableText(input.ExecutionMode), nullableText(input.Network), nullableText(input.AgentFactoryAddress), nullableText(input.MarketContractAddress), nullableText(input.ApproveTransactionHash), nullableText(input.TransactionHash), input.BroadcastPerformed, nullableJSON(input.Readback)).Scan(agentExecutionScanDestinations(&execution)...)
	return execution, err
}

func (r *AgentIntentsRepository) MarkAgentExecutionFailed(ctx context.Context, id string, input FailAgentExecutionInput) (AgentExecution, error) {
	var execution AgentExecution
	err := r.db.QueryRow(ctx, `
		UPDATE agent_executions
		SET status = 'failed',
			readback = $2,
			error_code = $3,
			error_message = $4,
			completed_at = now(),
			updated_at = now()
		WHERE id = $1
		RETURNING
			id::text,
			intent_id,
			agent_id,
			action,
			status,
			execution_mode,
			network,
			agent_factory_address,
			market_contract_address,
			approve_transaction_hash,
			transaction_hash,
			broadcast_performed,
			readback,
			error_code,
			error_message,
			created_at,
			updated_at,
			completed_at
	`, id, nullableJSON(input.Readback), nullableText(input.ErrorCode), nullableText(input.ErrorMessage)).Scan(agentExecutionScanDestinations(&execution)...)
	return execution, err
}

func (r *AgentIntentsRepository) ListAgentIntentsByAgentID(ctx context.Context, agentID string, limit int) ([]AgentIntent, error) {
	rows, err := r.db.Query(ctx, agentIntentSelectSQL+`
		WHERE agent_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, agentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	intents := []AgentIntent{}
	for rows.Next() {
		var intent AgentIntent
		if err := rows.Scan(agentIntentScanDestinations(&intent)...); err != nil {
			return nil, err
		}
		intents = append(intents, intent)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return intents, nil
}

func (r *AgentIntentsRepository) ListAgentExecutionsByAgentID(ctx context.Context, agentID string, limit int) ([]AgentExecution, error) {
	rows, err := r.db.Query(ctx, agentExecutionSelectSQL+`
		WHERE agent_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, agentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAgentExecutions(rows)
}

func (r *AgentIntentsRepository) ListAgentExecutionsByIntentID(ctx context.Context, intentID string, limit int) ([]AgentExecution, error) {
	rows, err := r.db.Query(ctx, agentExecutionSelectSQL+`
		WHERE intent_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, intentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAgentExecutions(rows)
}

func (r *AgentIntentsRepository) updateIntentTerminalStatus(ctx context.Context, intentID string, status string) (AgentIntent, error) {
	var intent AgentIntent
	err := r.db.QueryRow(ctx, `
		UPDATE agent_intents
		SET status = $2,
			executed_at = COALESCE(executed_at, now()),
			updated_at = now()
		WHERE intent_id = $1
		RETURNING
			id::text,
			intent_id,
			agent_id,
			agent_wallet_address,
			wallet_provider,
			source_client,
			client_request_id,
			action,
			status,
			requires_confirmation,
			user_wallet,
			market_id,
			market_contract_address,
			amount,
			outcome,
			resolver,
			collateral_token,
			close_timestamp,
			question,
			validation_result,
			warnings,
			created_at,
			updated_at,
			confirmed_at,
			executed_at
	`, intentID, status).Scan(agentIntentScanDestinations(&intent)...)
	return intent, normalizeAgentIntentNoRows(err)
}

func (r *AgentIntentsRepository) getIntentByIdempotencyKey(ctx context.Context, agentID string, sourceClient string, clientRequestID string) (AgentIntent, bool, error) {
	if strings.TrimSpace(agentID) == "" || strings.TrimSpace(sourceClient) == "" || strings.TrimSpace(clientRequestID) == "" {
		return AgentIntent{}, false, nil
	}

	var intent AgentIntent
	err := r.db.QueryRow(ctx, agentIntentSelectSQL+`
		WHERE agent_id = $1
			AND source_client = $2
			AND client_request_id = $3
		ORDER BY created_at DESC
		LIMIT 1
	`, agentID, sourceClient, clientRequestID).Scan(agentIntentScanDestinations(&intent)...)
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows {
			return AgentIntent{}, false, nil
		}
		return AgentIntent{}, false, err
	}
	return intent, true, nil
}

const agentIntentSelectSQL = `
	SELECT
		id::text,
		intent_id,
		agent_id,
		agent_wallet_address,
		wallet_provider,
		source_client,
		client_request_id,
		action,
		status,
		requires_confirmation,
		user_wallet,
		market_id,
		market_contract_address,
		amount,
		outcome,
		resolver,
		collateral_token,
		close_timestamp,
		question,
		COALESCE(validation_result, '{}'::jsonb),
		COALESCE(warnings, '[]'::jsonb),
		created_at,
		updated_at,
		confirmed_at,
		executed_at
	FROM agent_intents
`

const agentExecutionSelectSQL = `
	SELECT
		id::text,
		intent_id,
		agent_id,
		action,
		status,
		execution_mode,
		network,
		agent_factory_address,
		market_contract_address,
		approve_transaction_hash,
		transaction_hash,
		broadcast_performed,
		COALESCE(readback, '{}'::jsonb),
		error_code,
		error_message,
		created_at,
		updated_at,
		completed_at
	FROM agent_executions
`

func agentIntentScanDestinations(intent *AgentIntent) []any {
	return []any{
		&intent.ID,
		&intent.IntentID,
		&intent.AgentID,
		&intent.AgentWalletAddress,
		&intent.WalletProvider,
		&intent.SourceClient,
		&intent.ClientRequestID,
		&intent.Action,
		&intent.Status,
		&intent.RequiresConfirmation,
		&intent.UserWallet,
		&intent.MarketID,
		&intent.MarketContractAddress,
		&intent.Amount,
		&intent.Outcome,
		&intent.Resolver,
		&intent.CollateralToken,
		&intent.CloseTimestamp,
		&intent.Question,
		&intent.ValidationResult,
		&intent.Warnings,
		&intent.CreatedAt,
		&intent.UpdatedAt,
		&intent.ConfirmedAt,
		&intent.ExecutedAt,
	}
}

func agentExecutionScanDestinations(execution *AgentExecution) []any {
	return []any{
		&execution.ID,
		&execution.IntentID,
		&execution.AgentID,
		&execution.Action,
		&execution.Status,
		&execution.ExecutionMode,
		&execution.Network,
		&execution.AgentFactoryAddress,
		&execution.MarketContractAddress,
		&execution.ApproveTransactionHash,
		&execution.TransactionHash,
		&execution.BroadcastPerformed,
		&execution.Readback,
		&execution.ErrorCode,
		&execution.ErrorMessage,
		&execution.CreatedAt,
		&execution.UpdatedAt,
		&execution.CompletedAt,
	}
}

func scanAgentExecutions(rows pgx.Rows) ([]AgentExecution, error) {
	executions := []AgentExecution{}
	for rows.Next() {
		var execution AgentExecution
		if err := rows.Scan(agentExecutionScanDestinations(&execution)...); err != nil {
			return nil, err
		}
		executions = append(executions, execution)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return executions, nil
}

func normalizeAgentIntentNoRows(err error) error {
	if err == pgx.ErrNoRows {
		return sql.ErrNoRows
	}
	return err
}
