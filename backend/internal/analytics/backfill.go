package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/repository"
)

type Backfiller struct {
	client Client
	store  Store
	now    func() time.Time
}

func NewBackfiller(client Client, store Store) *Backfiller {
	return &Backfiller{
		client: client,
		store:  store,
		now:    time.Now,
	}
}

func (backfiller *Backfiller) Run(ctx context.Context, opts BackfillOptions) (BackfillResult, error) {
	if backfiller == nil || backfiller.client == nil {
		return BackfillResult{}, fmt.Errorf("analytics backfiller client is required")
	}
	if !opts.DryRun && backfiller.store == nil {
		return BackfillResult{}, fmt.Errorf("analytics backfiller store is required for write mode")
	}
	if opts.IncludeMarketEvents && backfiller.store == nil {
		return BackfillResult{}, fmt.Errorf("analytics backfiller store is required to discover market contracts")
	}
	if opts.FactoryAddress == "" {
		return BackfillResult{}, fmt.Errorf("factory address is required")
	}
	chainID := opts.ChainID
	if chainID == 0 {
		chainID = DefaultChainID
	}

	result := BackfillResult{
		FactoryAddress:      opts.FactoryAddress,
		DryRun:              opts.DryRun,
		IncludeMarketEvents: opts.IncludeMarketEvents,
	}
	pageParams := map[string]string{}

	for {
		if opts.PageLimit > 0 && result.PagesFetched >= opts.PageLimit {
			break
		}

		page, err := backfiller.client.FetchAddressLogs(ctx, opts.FactoryAddress, pageParams)
		if err != nil {
			return result, err
		}
		result.PagesFetched++
		result.LogsSeen += len(page.Items)

		for _, log := range page.Items {
			if opts.FromBlock > 0 && log.BlockNumber < opts.FromBlock {
				continue
			}

			event, matched, err := ParseMarketDeployed(opts.FactoryAddress, log)
			if err != nil {
				return result, err
			}
			if !matched {
				continue
			}

			result.EventsParsed++
			if event.BlockNumber > 0 && (!result.LatestBlock.Valid || event.BlockNumber > result.LatestBlock.Int64) {
				result.LatestBlock = sql.NullInt64{Int64: event.BlockNumber, Valid: true}
			}
			if !event.BlockTimestamp.IsZero() && (!result.LatestEventAt.Valid || event.BlockTimestamp.After(result.LatestEventAt.Time)) {
				result.LatestEventAt = sql.NullTime{Time: event.BlockTimestamp, Valid: true}
			}

			if opts.DryRun {
				continue
			}

			if err := backfiller.store.UpsertAnalyticsMarket(ctx, marketInput(event)); err != nil {
				return result, err
			}
			result.MarketsUpserted++

			inserted, err := backfiller.store.InsertAnalyticsEvent(ctx, eventInput(chainID, event))
			if err != nil {
				return result, err
			}
			if inserted {
				result.EventsInserted++
			}
		}

		if len(page.NextPageParams) == 0 {
			break
		}
		pageParams = page.NextPageParams
	}

	if opts.IncludeMarketEvents {
		if err := backfiller.ingestMarketEvents(ctx, opts, chainID, &result); err != nil {
			return result, err
		}
	}

	if !opts.DryRun {
		if result.LatestBlock.Valid {
			source := IndexerSourceFactory
			if opts.IncludeMarketEvents {
				source = IndexerSourceMarkets
			}
			if err := backfiller.store.UpdateAnalyticsIndexerState(ctx, repository.UpdateAnalyticsIndexerStateInput{
				Source:            source,
				FactoryAddress:    opts.FactoryAddress,
				LastIndexedBlock:  result.LatestBlock.Int64,
				LastIndexedLogKey: latestLogKey(result),
				LastSuccessAt:     backfiller.now().UTC(),
			}); err != nil {
				return result, err
			}
		}

		summary, err := backfiller.store.RebuildAnalyticsSummaryCache(ctx, opts.FactoryAddress)
		if err != nil {
			return result, err
		}
		result.Summary = summary
	}

	return result, nil
}

func (backfiller *Backfiller) ingestMarketEvents(ctx context.Context, opts BackfillOptions, chainID int, result *BackfillResult) error {
	markets, err := backfiller.store.ListAnalyticsMarketsByFactory(ctx, opts.FactoryAddress)
	if err != nil {
		return err
	}

	for _, market := range markets {
		pageParams := map[string]string{}
		pagesFetchedForMarket := 0

		for {
			if opts.PageLimit > 0 && pagesFetchedForMarket >= opts.PageLimit {
				break
			}

			page, err := backfiller.client.FetchAddressLogs(ctx, market.MarketAddress, pageParams)
			if err != nil {
				return err
			}
			pagesFetchedForMarket++
			result.PagesFetched++
			result.LogsSeen += len(page.Items)

			for _, log := range page.Items {
				if opts.FromBlock > 0 && log.BlockNumber < opts.FromBlock {
					continue
				}

				event, matched, err := ParseMarketEvent(opts.FactoryAddress, market.MarketAddress, log)
				if err != nil {
					return err
				}
				if !matched {
					continue
				}

				result.EventsParsed++
				updateLatest(result, event.BlockNumber, event.BlockTimestamp)

				if opts.DryRun {
					continue
				}

				inserted, err := backfiller.store.InsertAnalyticsEvent(ctx, childEventInput(chainID, event))
				if err != nil {
					return err
				}
				if inserted {
					result.EventsInserted++
				}

				if event.Status != "" {
					if err := backfiller.store.UpdateAnalyticsMarketLifecycle(ctx, lifecycleInput(event)); err != nil {
						return err
					}
				}
			}

			if len(page.NextPageParams) == 0 {
				break
			}
			pageParams = page.NextPageParams
		}
	}

	return nil
}

func marketInput(event MarketDeployed) repository.UpsertAnalyticsMarketInput {
	return repository.UpsertAnalyticsMarketInput{
		MarketAddress:          event.MarketAddress,
		FactoryAddress:         event.FactoryAddress,
		MarketIDHash:           event.MarketIDHash,
		CreatorAddress:         event.CreatorAddress,
		ResolverAddress:        event.ResolverAddress,
		CollateralTokenAddress: event.CollateralTokenAddress,
		Question:               event.Question,
		CloseTimestamp:         unixSecondsNullTime(event.CloseTimestamp),
		DeploymentTxHash:       event.TransactionHash,
		DeploymentBlock:        sql.NullInt64{Int64: event.BlockNumber, Valid: event.BlockNumber > 0},
		DeploymentTimestamp:    sql.NullTime{Time: event.BlockTimestamp, Valid: !event.BlockTimestamp.IsZero()},
		LastIndexedBlock:       sql.NullInt64{Int64: event.BlockNumber, Valid: event.BlockNumber > 0},
	}
}

func eventInput(chainID int, event MarketDeployed) repository.InsertAnalyticsEventInput {
	return repository.InsertAnalyticsEventInput{
		ChainID:         chainID,
		ContractAddress: event.FactoryAddress,
		MarketAddress:   event.MarketAddress,
		FactoryAddress:  event.FactoryAddress,
		EventName:       MarketDeployedEvent,
		TransactionHash: event.TransactionHash,
		BlockNumber:     event.BlockNumber,
		LogIndex:        event.LogIndex,
		BlockTimestamp:  sql.NullTime{Time: event.BlockTimestamp, Valid: !event.BlockTimestamp.IsZero()},
		Raw:             event.Raw,
	}
}

func childEventInput(chainID int, event MarketEvent) repository.InsertAnalyticsEventInput {
	return repository.InsertAnalyticsEventInput{
		ChainID:         chainID,
		ContractAddress: event.MarketAddress,
		MarketAddress:   event.MarketAddress,
		FactoryAddress:  event.FactoryAddress,
		EventName:       event.EventName,
		TransactionHash: event.TransactionHash,
		BlockNumber:     event.BlockNumber,
		LogIndex:        event.LogIndex,
		BlockTimestamp:  sql.NullTime{Time: event.BlockTimestamp, Valid: !event.BlockTimestamp.IsZero()},
		WalletAddress:   event.WalletAddress,
		Side:            event.Side,
		AmountBaseUnits: event.AmountBaseUnits,
		Raw:             event.Raw,
	}
}

func lifecycleInput(event MarketEvent) repository.UpdateAnalyticsMarketLifecycleInput {
	return repository.UpdateAnalyticsMarketLifecycleInput{
		MarketAddress:    event.MarketAddress,
		Status:           event.Status,
		WinningOutcome:   event.WinningOutcome,
		LastIndexedBlock: sql.NullInt64{Int64: event.BlockNumber, Valid: event.BlockNumber > 0},
	}
}

func updateLatest(result *BackfillResult, blockNumber int64, blockTimestamp time.Time) {
	if blockNumber > 0 && (!result.LatestBlock.Valid || blockNumber > result.LatestBlock.Int64) {
		result.LatestBlock = sql.NullInt64{Int64: blockNumber, Valid: true}
	}
	if !blockTimestamp.IsZero() && (!result.LatestEventAt.Valid || blockTimestamp.After(result.LatestEventAt.Time)) {
		result.LatestEventAt = sql.NullTime{Time: blockTimestamp, Valid: true}
	}
}

func unixSecondsNullTime(value string) sql.NullTime {
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil || seconds <= 0 {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: time.Unix(seconds, 0).UTC(), Valid: true}
}

func latestLogKey(result BackfillResult) string {
	if !result.LatestBlock.Valid {
		return ""
	}
	return strconv.FormatInt(result.LatestBlock.Int64, 10)
}
