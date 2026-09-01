package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/database"
)

type Resolution struct {
	ID                string         `json:"id"`
	MarketID          string         `json:"market_id"`
	WinningOutcome    sql.NullString `json:"winning_outcome"`
	Status            string         `json:"status"`
	ResolverType      sql.NullString `json:"resolver_type"`
	EvidenceReference sql.NullString `json:"evidence_reference"`
	ResolvedAt        sql.NullTime   `json:"resolved_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

type ResolutionsRepository struct {
	db *database.DB
}

func NewResolutionsRepository(db *database.DB) *ResolutionsRepository {
	return &ResolutionsRepository{db: db}
}

func (r *ResolutionsRepository) GetResolutionByID(ctx context.Context, id string) (Resolution, error) {
	var resolution Resolution
	err := r.db.QueryRow(ctx, resolutionSelectSQL+`
		WHERE id = $1
	`, id).Scan(
		&resolution.ID,
		&resolution.MarketID,
		&resolution.WinningOutcome,
		&resolution.Status,
		&resolution.ResolverType,
		&resolution.EvidenceReference,
		&resolution.ResolvedAt,
		&resolution.CreatedAt,
		&resolution.UpdatedAt,
	)

	return resolution, err
}

func (r *ResolutionsRepository) GetResolutionByMarketID(ctx context.Context, marketID string) (Resolution, error) {
	var resolution Resolution
	err := r.db.QueryRow(ctx, resolutionSelectSQL+`
		WHERE market_id = $1
	`, marketID).Scan(
		&resolution.ID,
		&resolution.MarketID,
		&resolution.WinningOutcome,
		&resolution.Status,
		&resolution.ResolverType,
		&resolution.EvidenceReference,
		&resolution.ResolvedAt,
		&resolution.CreatedAt,
		&resolution.UpdatedAt,
	)

	return resolution, err
}

func (r *ResolutionsRepository) ListResolutionsByMarketID(ctx context.Context, marketID string, limit int) ([]Resolution, error) {
	rows, err := r.db.Query(ctx, resolutionSelectSQL+`
		WHERE market_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, marketID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resolutions := []Resolution{}
	for rows.Next() {
		var resolution Resolution
		if err := rows.Scan(
			&resolution.ID,
			&resolution.MarketID,
			&resolution.WinningOutcome,
			&resolution.Status,
			&resolution.ResolverType,
			&resolution.EvidenceReference,
			&resolution.ResolvedAt,
			&resolution.CreatedAt,
			&resolution.UpdatedAt,
		); err != nil {
			return nil, err
		}
		resolutions = append(resolutions, resolution)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resolutions, nil
}

const resolutionSelectSQL = `
	SELECT
		id::text,
		market_id::text,
		winning_outcome,
		status,
		resolver_type,
		evidence_reference,
		resolved_at,
		created_at,
		updated_at
	FROM resolutions
`
