package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/database"
)

type Market struct {
	ID                      string         `json:"id"`
	CreatorUserID           string         `json:"creator_user_id"`
	Title                   string         `json:"title"`
	Description             sql.NullString `json:"description"`
	Category                sql.NullString `json:"category"`
	CoverImageURL           sql.NullString `json:"cover_image_url"`
	Status                  string         `json:"status"`
	OutcomeYesLabel         string         `json:"outcome_yes_label"`
	OutcomeNoLabel          string         `json:"outcome_no_label"`
	CollateralAsset         string         `json:"collateral_asset"`
	Chain                   string         `json:"chain"`
	ResolutionSource        sql.NullString `json:"resolution_source"`
	OpensAt                 sql.NullTime   `json:"opens_at"`
	ClosesAt                time.Time      `json:"closes_at"`
	ResolvedAt              sql.NullTime   `json:"resolved_at"`
	SettledAt               sql.NullTime   `json:"settled_at"`
	WinningOutcome          sql.NullString `json:"winning_outcome"`
	MarketContractAddress   sql.NullString `json:"market_contract_address"`
	MarketDeploymentTxHash  sql.NullString `json:"market_deployment_tx_hash"`
	MarketFactoryAddress    sql.NullString `json:"market_factory_address"`
	ResolverAddress         sql.NullString `json:"resolver_address"`
	OnchainDeploymentStatus string         `json:"onchain_deployment_status"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
}

type MarketsRepository struct {
	db *database.DB
}

type CreateMarketInput struct {
	ID                     string
	CreatorUserID          string
	Title                  string
	Description            sql.NullString
	Category               sql.NullString
	CoverImageURL          sql.NullString
	Status                 string
	OutcomeYesLabel        string
	OutcomeNoLabel         string
	CollateralAsset        string
	Chain                  string
	ResolutionSource       sql.NullString
	OpensAt                sql.NullTime
	ClosesAt               time.Time
	MarketContractAddress  string
	MarketDeploymentTxHash string
	MarketFactoryAddress   string
	ResolverAddress        string
}

type AttachMarketContractInput struct {
	MarketContractAddress  string
	MarketDeploymentTxHash string
	MarketFactoryAddress   string
	ResolverAddress        string
}

func NewMarketsRepository(db *database.DB) *MarketsRepository {
	return &MarketsRepository{db: db}
}

func (r *MarketsRepository) GetMarketByID(ctx context.Context, id string) (Market, error) {
	var market Market
	err := r.db.QueryRow(ctx, marketSelectSQL+`
		WHERE id = $1
			AND onchain_deployment_status = 'DEPLOYED'
			AND market_contract_address IS NOT NULL
	`, id).Scan(
		marketScanDestinations(&market)...,
	)

	return market, err
}

func (r *MarketsRepository) CreateMarket(ctx context.Context, input CreateMarketInput) (Market, error) {
	var market Market
	err := r.db.QueryRow(ctx, `
		INSERT INTO markets (
			id,
			creator_user_id,
			title,
			description,
			category,
			cover_image_url,
			status,
			outcome_yes_label,
			outcome_no_label,
			collateral_asset,
			chain,
			resolution_source,
			opens_at,
			closes_at,
			market_contract_address,
			market_deployment_tx_hash,
			market_factory_address,
			resolver_address,
			onchain_deployment_status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, 'DEPLOYED')
		RETURNING
			id::text,
			creator_user_id::text,
			title,
			description,
			category,
			cover_image_url,
			status,
			outcome_yes_label,
			outcome_no_label,
			collateral_asset,
			chain,
			resolution_source,
			opens_at,
			closes_at,
			resolved_at,
			settled_at,
			winning_outcome,
			market_contract_address,
			market_deployment_tx_hash,
			market_factory_address,
			resolver_address,
			onchain_deployment_status,
			created_at,
			updated_at
	`,
		input.ID,
		input.CreatorUserID,
		input.Title,
		input.Description,
		input.Category,
		input.CoverImageURL,
		input.Status,
		input.OutcomeYesLabel,
		input.OutcomeNoLabel,
		input.CollateralAsset,
		input.Chain,
		input.ResolutionSource,
		input.OpensAt,
		input.ClosesAt,
		input.MarketContractAddress,
		input.MarketDeploymentTxHash,
		input.MarketFactoryAddress,
		input.ResolverAddress,
	).Scan(
		marketScanDestinations(&market)...,
	)

	return market, err
}

func (r *MarketsRepository) AttachMarketContract(ctx context.Context, id string, input AttachMarketContractInput) (Market, error) {
	var market Market
	err := r.db.QueryRow(ctx, `
		UPDATE markets
		SET
			market_contract_address = $2,
			market_deployment_tx_hash = $3,
			market_factory_address = $4,
			resolver_address = $5,
			onchain_deployment_status = 'DEPLOYED',
			updated_at = now()
		WHERE id = $1
		RETURNING
			id::text,
			creator_user_id::text,
			title,
			description,
			category,
			cover_image_url,
			status,
			outcome_yes_label,
			outcome_no_label,
			collateral_asset,
			chain,
			resolution_source,
			opens_at,
			closes_at,
			resolved_at,
			settled_at,
			winning_outcome,
			market_contract_address,
			market_deployment_tx_hash,
			market_factory_address,
			resolver_address,
			onchain_deployment_status,
			created_at,
			updated_at
	`,
		id,
		input.MarketContractAddress,
		input.MarketDeploymentTxHash,
		input.MarketFactoryAddress,
		input.ResolverAddress,
	).Scan(
		marketScanDestinations(&market)...,
	)

	return market, err
}

func (r *MarketsRepository) ListMarkets(ctx context.Context, limit int) ([]Market, error) {
	rows, err := r.db.Query(ctx, marketSelectSQL+`
		WHERE onchain_deployment_status = 'DEPLOYED'
			AND market_contract_address IS NOT NULL
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	markets := []Market{}
	for rows.Next() {
		var market Market
		if err := rows.Scan(
			marketScanDestinations(&market)...,
		); err != nil {
			return nil, err
		}
		markets = append(markets, market)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return markets, nil
}

const marketSelectSQL = `
	SELECT
		id::text,
		creator_user_id::text,
		title,
		description,
		category,
		cover_image_url,
		status,
		outcome_yes_label,
		outcome_no_label,
		collateral_asset,
		chain,
		resolution_source,
		opens_at,
		closes_at,
		resolved_at,
		settled_at,
		winning_outcome,
		market_contract_address,
		market_deployment_tx_hash,
		market_factory_address,
		resolver_address,
		onchain_deployment_status,
		created_at,
		updated_at
	FROM markets
`

func marketScanDestinations(market *Market) []any {
	return []any{
		&market.ID,
		&market.CreatorUserID,
		&market.Title,
		&market.Description,
		&market.Category,
		&market.CoverImageURL,
		&market.Status,
		&market.OutcomeYesLabel,
		&market.OutcomeNoLabel,
		&market.CollateralAsset,
		&market.Chain,
		&market.ResolutionSource,
		&market.OpensAt,
		&market.ClosesAt,
		&market.ResolvedAt,
		&market.SettledAt,
		&market.WinningOutcome,
		&market.MarketContractAddress,
		&market.MarketDeploymentTxHash,
		&market.MarketFactoryAddress,
		&market.ResolverAddress,
		&market.OnchainDeploymentStatus,
		&market.CreatedAt,
		&market.UpdatedAt,
	}
}
