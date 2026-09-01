package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/wahyu241205/SignalArc/backend/internal/database"
)

type Settlement struct {
	ID           string         `json:"id"`
	MarketID     string         `json:"market_id"`
	UserID       sql.NullString `json:"user_id"`
	ResolutionID sql.NullString `json:"resolution_id"`
	Outcome      sql.NullString `json:"outcome"`
	Amount       string         `json:"amount"`
	Status       string         `json:"status"`
	TxHash       sql.NullString `json:"tx_hash"`
	SettledAt    sql.NullTime   `json:"settled_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type SettlementsRepository struct {
	db *database.DB
}

func NewSettlementsRepository(db *database.DB) *SettlementsRepository {
	return &SettlementsRepository{db: db}
}

func (r *SettlementsRepository) GetSettlementByID(ctx context.Context, id string) (Settlement, error) {
	var settlement Settlement
	err := r.db.QueryRow(ctx, settlementSelectSQL+`
		WHERE id = $1
	`, id).Scan(
		&settlement.ID,
		&settlement.MarketID,
		&settlement.UserID,
		&settlement.ResolutionID,
		&settlement.Outcome,
		&settlement.Amount,
		&settlement.Status,
		&settlement.TxHash,
		&settlement.SettledAt,
		&settlement.CreatedAt,
		&settlement.UpdatedAt,
	)

	return settlement, err
}

func (r *SettlementsRepository) ListSettlementsByUserID(ctx context.Context, userID string, limit int) ([]Settlement, error) {
	rows, err := r.db.Query(ctx, settlementSelectSQL+`
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSettlements(rows)
}

func (r *SettlementsRepository) ListSettlementsByMarketID(ctx context.Context, marketID string, limit int) ([]Settlement, error) {
	rows, err := r.db.Query(ctx, settlementSelectSQL+`
		WHERE market_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, marketID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSettlements(rows)
}

func scanSettlements(rows pgx.Rows) ([]Settlement, error) {
	settlements := []Settlement{}
	for rows.Next() {
		var settlement Settlement
		if err := rows.Scan(
			&settlement.ID,
			&settlement.MarketID,
			&settlement.UserID,
			&settlement.ResolutionID,
			&settlement.Outcome,
			&settlement.Amount,
			&settlement.Status,
			&settlement.TxHash,
			&settlement.SettledAt,
			&settlement.CreatedAt,
			&settlement.UpdatedAt,
		); err != nil {
			return nil, err
		}
		settlements = append(settlements, settlement)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return settlements, nil
}

const settlementSelectSQL = `
	SELECT
		id::text,
		market_id::text,
		CASE WHEN user_id IS NULL THEN NULL ELSE user_id::text END,
		CASE WHEN resolution_id IS NULL THEN NULL ELSE resolution_id::text END,
		outcome,
		amount::text,
		status,
		tx_hash,
		settled_at,
		created_at,
		updated_at
	FROM settlements
`
