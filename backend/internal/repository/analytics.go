package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/wahyu241205/SignalArc/backend/internal/database"
)

const (
	AnalyticsSummaryCacheKey  = "public_analytics_summary"
	AnalyticsStatusOK         = "ok"
	AnalyticsSourceCached     = "cached"
	AnalyticsSourceIndexed    = "indexed"
	AnalyticsSourceNotIndexed = "not_indexed"
)

type AnalyticsMetrics struct {
	MarketsCreated       int64  `json:"markets_created"`
	MarketContractsFound int64  `json:"market_contracts_found"`
	TotalTrades          int64  `json:"total_trades"`
	PositionEvents       int64  `json:"position_events"`
	YesPositionEvents    int64  `json:"yes_position_events"`
	NoPositionEvents     int64  `json:"no_position_events"`
	UniqueWallets        int64  `json:"unique_wallets"`
	TestnetUSDCVolume    string `json:"testnet_usdc_volume"`
	ResolvedMarkets      int64  `json:"resolved_markets"`
	CancelledMarkets     int64  `json:"cancelled_markets"`
	ClaimEvents          int64  `json:"claim_events"`
	PayoutsClaimed       int64  `json:"payouts_claimed"`
	RefundsClaimed       int64  `json:"refunds_claimed"`
}

type AnalyticsSummary struct {
	Status         string           `json:"status"`
	SourceStatus   string           `json:"source_status"`
	FactoryAddress string           `json:"factory_address"`
	GeneratedAt    time.Time        `json:"generated_at"`
	LatestEventAt  sql.NullTime     `json:"latest_event_at"`
	LatestBlock    sql.NullInt64    `json:"latest_block"`
	Metrics        AnalyticsMetrics `json:"metrics"`
}

type AnalyticsSummaryCacheInput struct {
	CacheKey       string
	FactoryAddress string
	Summary        AnalyticsSummary
	LatestBlock    sql.NullInt64
	LatestEventAt  sql.NullTime
	GeneratedAt    time.Time
}

type UpsertAnalyticsMarketInput struct {
	MarketAddress          string
	FactoryAddress         string
	MarketIDHash           string
	CreatorAddress         string
	ResolverAddress        string
	CollateralTokenAddress string
	Question               string
	CloseTimestamp         sql.NullTime
	DeploymentTxHash       string
	DeploymentBlock        sql.NullInt64
	DeploymentTimestamp    sql.NullTime
	LastIndexedBlock       sql.NullInt64
}

type InsertAnalyticsEventInput struct {
	ChainID         int
	ContractAddress string
	MarketAddress   string
	FactoryAddress  string
	EventName       string
	TransactionHash string
	BlockNumber     int64
	LogIndex        int
	BlockTimestamp  sql.NullTime
	WalletAddress   string
	Side            string
	AmountBaseUnits string
	Raw             json.RawMessage
}

type UpdateAnalyticsIndexerStateInput struct {
	Source            string
	FactoryAddress    string
	LastIndexedBlock  int64
	LastIndexedLogKey string
	LastSuccessAt     time.Time
	LastError         string
}

type AnalyticsMarketContract struct {
	MarketAddress  string
	FactoryAddress string
}

type UpdateAnalyticsMarketLifecycleInput struct {
	MarketAddress    string
	Status           string
	WinningOutcome   string
	LastIndexedBlock sql.NullInt64
}

type AnalyticsRepository struct {
	db *database.DB
}

func NewAnalyticsRepository(db *database.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) GetSummary(ctx context.Context, cacheKey string, factoryAddress string) (AnalyticsSummary, error) {
	summary, ok, err := r.GetLatestSummaryCache(ctx, cacheKey)
	if err != nil {
		return AnalyticsSummary{}, err
	}
	if ok {
		summary.SourceStatus = AnalyticsSourceCached
		if summary.Status == "" {
			summary.Status = AnalyticsStatusOK
		}
		if summary.FactoryAddress == "" {
			summary.FactoryAddress = factoryAddress
		}
		return summary, nil
	}

	return r.AggregateSummaryFallback(ctx, factoryAddress)
}

func (r *AnalyticsRepository) GetLatestSummaryCache(ctx context.Context, cacheKey string) (AnalyticsSummary, bool, error) {
	var payload []byte
	var factoryAddress string
	var latestBlock sql.NullInt64
	var latestEventAt sql.NullTime
	var generatedAt time.Time

	err := r.db.QueryRow(ctx, `
		SELECT payload, factory_address, latest_block, latest_event_at, generated_at
		FROM analytics_summary_cache
		WHERE cache_key = $1
	`, cacheKey).Scan(&payload, &factoryAddress, &latestBlock, &latestEventAt, &generatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return AnalyticsSummary{}, false, nil
	}
	if err != nil {
		return AnalyticsSummary{}, false, err
	}

	var summary AnalyticsSummary
	if err := json.Unmarshal(payload, &summary); err != nil {
		return AnalyticsSummary{}, false, err
	}

	if summary.FactoryAddress == "" {
		summary.FactoryAddress = factoryAddress
	}
	if summary.GeneratedAt.IsZero() {
		summary.GeneratedAt = generatedAt
	}
	if !summary.LatestBlock.Valid {
		summary.LatestBlock = latestBlock
	}
	if !summary.LatestEventAt.Valid {
		summary.LatestEventAt = latestEventAt
	}

	return summary, true, nil
}

func (r *AnalyticsRepository) UpsertSummaryCache(ctx context.Context, input AnalyticsSummaryCacheInput) (AnalyticsSummary, error) {
	cacheKey := input.CacheKey
	if cacheKey == "" {
		cacheKey = AnalyticsSummaryCacheKey
	}
	generatedAt := input.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}

	summary := input.Summary
	summary.GeneratedAt = generatedAt
	summary.FactoryAddress = input.FactoryAddress
	summary.LatestBlock = input.LatestBlock
	summary.LatestEventAt = input.LatestEventAt
	if summary.Status == "" {
		summary.Status = AnalyticsStatusOK
	}
	if summary.SourceStatus == "" {
		summary.SourceStatus = AnalyticsSourceCached
	}

	payload, err := json.Marshal(summary)
	if err != nil {
		return AnalyticsSummary{}, err
	}

	var saved AnalyticsSummary
	err = r.db.QueryRow(ctx, `
		INSERT INTO analytics_summary_cache (
			cache_key,
			factory_address,
			payload,
			latest_block,
			latest_event_at,
			generated_at
		)
		VALUES ($1, $2, $3::jsonb, $4, $5, $6)
		ON CONFLICT (cache_key) DO UPDATE
		SET
			factory_address = EXCLUDED.factory_address,
			payload = EXCLUDED.payload,
			latest_block = EXCLUDED.latest_block,
			latest_event_at = EXCLUDED.latest_event_at,
			generated_at = EXCLUDED.generated_at,
			updated_at = now()
		RETURNING payload
	`, cacheKey, input.FactoryAddress, payload, input.LatestBlock, input.LatestEventAt, generatedAt).Scan(&payload)
	if err != nil {
		return AnalyticsSummary{}, err
	}
	if err := json.Unmarshal(payload, &saved); err != nil {
		return AnalyticsSummary{}, err
	}

	return saved, nil
}

func (r *AnalyticsRepository) UpsertAnalyticsMarket(ctx context.Context, input UpsertAnalyticsMarketInput) error {
	return r.db.Exec(ctx, `
		INSERT INTO analytics_markets (
			market_address,
			factory_address,
			market_id_hash,
			creator_address,
			resolver_address,
			collateral_token_address,
			question,
			close_timestamp,
			deployment_tx_hash,
			deployment_block,
			deployment_timestamp,
			last_indexed_block
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (market_address) DO UPDATE
		SET
			factory_address = EXCLUDED.factory_address,
			market_id_hash = COALESCE(EXCLUDED.market_id_hash, analytics_markets.market_id_hash),
			creator_address = COALESCE(EXCLUDED.creator_address, analytics_markets.creator_address),
			resolver_address = COALESCE(EXCLUDED.resolver_address, analytics_markets.resolver_address),
			collateral_token_address = COALESCE(EXCLUDED.collateral_token_address, analytics_markets.collateral_token_address),
			question = COALESCE(EXCLUDED.question, analytics_markets.question),
			close_timestamp = COALESCE(EXCLUDED.close_timestamp, analytics_markets.close_timestamp),
			deployment_tx_hash = COALESCE(EXCLUDED.deployment_tx_hash, analytics_markets.deployment_tx_hash),
			deployment_block = COALESCE(EXCLUDED.deployment_block, analytics_markets.deployment_block),
			deployment_timestamp = COALESCE(EXCLUDED.deployment_timestamp, analytics_markets.deployment_timestamp),
			last_indexed_block = GREATEST(
				COALESCE(EXCLUDED.last_indexed_block, 0),
				COALESCE(analytics_markets.last_indexed_block, 0)
			),
			updated_at = now()
	`,
		input.MarketAddress,
		input.FactoryAddress,
		nullableText(input.MarketIDHash),
		nullableText(input.CreatorAddress),
		nullableText(input.ResolverAddress),
		nullableText(input.CollateralTokenAddress),
		nullableText(input.Question),
		input.CloseTimestamp,
		nullableText(input.DeploymentTxHash),
		input.DeploymentBlock,
		input.DeploymentTimestamp,
		input.LastIndexedBlock,
	)
}

func (r *AnalyticsRepository) InsertAnalyticsEvent(ctx context.Context, input InsertAnalyticsEventInput) (bool, error) {
	raw := input.Raw
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	amountBaseUnits := input.AmountBaseUnits
	if amountBaseUnits == "" {
		amountBaseUnits = "0"
	}

	var inserted bool
	err := r.db.QueryRow(ctx, `
		INSERT INTO analytics_events (
			chain_id,
			contract_address,
			market_address,
			factory_address,
			event_name,
			transaction_hash,
			block_number,
			log_index,
			block_timestamp,
			wallet_address,
			side,
			amount_base_units,
			raw
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13::jsonb)
		ON CONFLICT (chain_id, transaction_hash, log_index) DO NOTHING
		RETURNING true
	`,
		input.ChainID,
		input.ContractAddress,
		nullableText(input.MarketAddress),
		nullableText(input.FactoryAddress),
		input.EventName,
		input.TransactionHash,
		input.BlockNumber,
		input.LogIndex,
		input.BlockTimestamp,
		nullableText(input.WalletAddress),
		nullableText(input.Side),
		amountBaseUnits,
		raw,
	).Scan(&inserted)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return inserted, nil
}

func (r *AnalyticsRepository) ListAnalyticsMarketsByFactory(ctx context.Context, factoryAddress string) ([]AnalyticsMarketContract, error) {
	rows, err := r.db.Query(ctx, `
		SELECT market_address, factory_address
		FROM analytics_markets
		WHERE factory_address = $1
		ORDER BY deployment_block DESC NULLS LAST, market_address ASC
	`, factoryAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	markets := []AnalyticsMarketContract{}
	for rows.Next() {
		var market AnalyticsMarketContract
		if err := rows.Scan(&market.MarketAddress, &market.FactoryAddress); err != nil {
			return nil, err
		}
		markets = append(markets, market)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return markets, nil
}

func (r *AnalyticsRepository) UpdateAnalyticsMarketLifecycle(ctx context.Context, input UpdateAnalyticsMarketLifecycleInput) error {
	return r.db.Exec(ctx, `
		UPDATE analytics_markets
		SET
			status = COALESCE($2, status),
			winning_outcome = COALESCE($3, winning_outcome),
			last_indexed_block = GREATEST(
				COALESCE(last_indexed_block, 0),
				COALESCE($4, 0)
			),
			updated_at = now()
		WHERE market_address = $1
	`,
		input.MarketAddress,
		nullableText(input.Status),
		nullableText(input.WinningOutcome),
		input.LastIndexedBlock,
	)
}

func (r *AnalyticsRepository) UpdateAnalyticsIndexerState(ctx context.Context, input UpdateAnalyticsIndexerStateInput) error {
	lastSuccessAt := sql.NullTime{}
	if !input.LastSuccessAt.IsZero() {
		lastSuccessAt = sql.NullTime{Time: input.LastSuccessAt, Valid: true}
	}

	return r.db.Exec(ctx, `
		INSERT INTO analytics_indexer_state (
			source,
			factory_address,
			last_indexed_block,
			last_indexed_log_key,
			last_success_at,
			last_error
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (source) DO UPDATE
		SET
			factory_address = EXCLUDED.factory_address,
			last_indexed_block = GREATEST(
				analytics_indexer_state.last_indexed_block,
				EXCLUDED.last_indexed_block
			),
			last_indexed_log_key = EXCLUDED.last_indexed_log_key,
			last_success_at = EXCLUDED.last_success_at,
			last_error = EXCLUDED.last_error,
			updated_at = now()
	`,
		input.Source,
		input.FactoryAddress,
		input.LastIndexedBlock,
		nullableText(input.LastIndexedLogKey),
		lastSuccessAt,
		nullableText(input.LastError),
	)
}

func (r *AnalyticsRepository) RebuildAnalyticsSummaryCache(ctx context.Context, factoryAddress string) (AnalyticsSummary, error) {
	summary, err := r.AggregateSummaryFallback(ctx, factoryAddress)
	if err != nil {
		return AnalyticsSummary{}, err
	}
	return r.UpsertSummaryCache(ctx, AnalyticsSummaryCacheInput{
		CacheKey:       AnalyticsSummaryCacheKey,
		FactoryAddress: factoryAddress,
		Summary:        summary,
		LatestBlock:    summary.LatestBlock,
		LatestEventAt:  summary.LatestEventAt,
		GeneratedAt:    summary.GeneratedAt,
	})
}

func (r *AnalyticsRepository) AggregateSummaryFallback(ctx context.Context, factoryAddress string) (AnalyticsSummary, error) {
	var metrics AnalyticsMetrics
	var latestBlock sql.NullInt64
	var latestEventAt sql.NullTime

	err := r.db.QueryRow(ctx, `
		WITH market_counts AS (
			SELECT
				count(*)::bigint AS market_contracts_found,
				count(*) FILTER (WHERE status = 'RESOLVED')::bigint AS resolved_by_status,
				count(*) FILTER (WHERE status = 'CANCELLED')::bigint AS cancelled_by_status
			FROM analytics_markets
			WHERE factory_address = $1
		),
		event_counts AS (
			SELECT
				count(*) FILTER (WHERE event_name = 'MarketDeployed')::bigint AS markets_created,
				count(*) FILTER (WHERE event_name = 'PositionOpened')::bigint AS position_events,
				count(*) FILTER (WHERE event_name = 'PositionOpened' AND side = 'YES')::bigint AS yes_position_events,
				count(*) FILTER (WHERE event_name = 'PositionOpened' AND side = 'NO')::bigint AS no_position_events,
				count(DISTINCT wallet_address) FILTER (WHERE wallet_address IS NOT NULL)::bigint AS unique_wallets,
				COALESCE(sum(amount_base_units) FILTER (WHERE event_name = 'PositionOpened'), 0)::text AS testnet_usdc_volume,
				count(DISTINCT market_address) FILTER (WHERE event_name = 'MarketResolved' AND market_address IS NOT NULL)::bigint AS resolved_by_event,
				count(DISTINCT market_address) FILTER (WHERE event_name = 'MarketCancelled' AND market_address IS NOT NULL)::bigint AS cancelled_by_event,
				count(*) FILTER (WHERE event_name IN ('PayoutClaimed', 'RefundClaimed'))::bigint AS claim_events,
				count(*) FILTER (WHERE event_name = 'PayoutClaimed')::bigint AS payouts_claimed,
				count(*) FILTER (WHERE event_name = 'RefundClaimed')::bigint AS refunds_claimed,
				max(block_number) AS latest_block,
				max(block_timestamp) AS latest_event_at
			FROM analytics_events
			WHERE factory_address = $1
		)
		SELECT
			CASE
				WHEN event_counts.markets_created > 0 THEN event_counts.markets_created
				ELSE market_counts.market_contracts_found
			END,
			market_counts.market_contracts_found,
			event_counts.position_events,
			event_counts.position_events,
			event_counts.yes_position_events,
			event_counts.no_position_events,
			event_counts.unique_wallets,
			event_counts.testnet_usdc_volume,
			CASE
				WHEN event_counts.resolved_by_event > 0 THEN event_counts.resolved_by_event
				ELSE market_counts.resolved_by_status
			END,
			CASE
				WHEN event_counts.cancelled_by_event > 0 THEN event_counts.cancelled_by_event
				ELSE market_counts.cancelled_by_status
			END,
			event_counts.claim_events,
			event_counts.payouts_claimed,
			event_counts.refunds_claimed,
			event_counts.latest_block,
			event_counts.latest_event_at
		FROM market_counts, event_counts
	`, factoryAddress).Scan(
		&metrics.MarketsCreated,
		&metrics.MarketContractsFound,
		&metrics.TotalTrades,
		&metrics.PositionEvents,
		&metrics.YesPositionEvents,
		&metrics.NoPositionEvents,
		&metrics.UniqueWallets,
		&metrics.TestnetUSDCVolume,
		&metrics.ResolvedMarkets,
		&metrics.CancelledMarkets,
		&metrics.ClaimEvents,
		&metrics.PayoutsClaimed,
		&metrics.RefundsClaimed,
		&latestBlock,
		&latestEventAt,
	)
	if err != nil {
		return AnalyticsSummary{}, err
	}
	if metrics.TestnetUSDCVolume == "" {
		metrics.TestnetUSDCVolume = "0"
	}

	sourceStatus := AnalyticsSourceIndexed
	if metrics.MarketsCreated == 0 &&
		metrics.MarketContractsFound == 0 &&
		metrics.PositionEvents == 0 &&
		metrics.ResolvedMarkets == 0 &&
		metrics.CancelledMarkets == 0 &&
		metrics.ClaimEvents == 0 {
		sourceStatus = AnalyticsSourceNotIndexed
	}

	return AnalyticsSummary{
		Status:         AnalyticsStatusOK,
		SourceStatus:   sourceStatus,
		FactoryAddress: factoryAddress,
		GeneratedAt:    time.Now().UTC(),
		LatestEventAt:  latestEventAt,
		LatestBlock:    latestBlock,
		Metrics:        metrics,
	}, nil
}
