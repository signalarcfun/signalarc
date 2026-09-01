package repository

import (
	"context"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/database"
)

type Position struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	MarketID          string    `json:"market_id"`
	Outcome           string    `json:"outcome"`
	Quantity          string    `json:"quantity"`
	AverageEntryPrice string    `json:"average_entry_price"`
	RealizedPnL       string    `json:"realized_pnl"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type PositionsRepository struct {
	db *database.DB
}

func NewPositionsRepository(db *database.DB) *PositionsRepository {
	return &PositionsRepository{db: db}
}

func (r *PositionsRepository) GetPositionByID(ctx context.Context, id string) (Position, error) {
	var position Position
	err := r.db.QueryRow(ctx, positionSelectSQL+`
		WHERE id = $1
	`, id).Scan(
		&position.ID,
		&position.UserID,
		&position.MarketID,
		&position.Outcome,
		&position.Quantity,
		&position.AverageEntryPrice,
		&position.RealizedPnL,
		&position.CreatedAt,
		&position.UpdatedAt,
	)

	return position, err
}

func (r *PositionsRepository) ListPositionsByUserID(ctx context.Context, userID string, limit int) ([]Position, error) {
	rows, err := r.db.Query(ctx, positionSelectSQL+`
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	positions := []Position{}
	for rows.Next() {
		var position Position
		if err := rows.Scan(
			&position.ID,
			&position.UserID,
			&position.MarketID,
			&position.Outcome,
			&position.Quantity,
			&position.AverageEntryPrice,
			&position.RealizedPnL,
			&position.CreatedAt,
			&position.UpdatedAt,
		); err != nil {
			return nil, err
		}
		positions = append(positions, position)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return positions, nil
}

func (r *PositionsRepository) ListPositionsByMarketID(ctx context.Context, marketID string, limit int) ([]Position, error) {
	rows, err := r.db.Query(ctx, positionSelectSQL+`
		WHERE market_id = $1
		ORDER BY updated_at DESC
		LIMIT $2
	`, marketID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	positions := []Position{}
	for rows.Next() {
		var position Position
		if err := rows.Scan(
			&position.ID,
			&position.UserID,
			&position.MarketID,
			&position.Outcome,
			&position.Quantity,
			&position.AverageEntryPrice,
			&position.RealizedPnL,
			&position.CreatedAt,
			&position.UpdatedAt,
		); err != nil {
			return nil, err
		}
		positions = append(positions, position)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return positions, nil
}

const positionSelectSQL = `
	SELECT
		id::text,
		user_id::text,
		market_id::text,
		outcome,
		quantity::text,
		average_entry_price::text,
		realized_pnl::text,
		created_at,
		updated_at
	FROM positions
`
