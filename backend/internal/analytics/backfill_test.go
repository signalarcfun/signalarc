package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/repository"
)

type fakeLogsClient struct {
	pagesByAddress map[string][]LogsPage
	pages          []LogsPage
	calls          int
	addressCalls   map[string]int
}

func (client *fakeLogsClient) FetchAddressLogs(ctx context.Context, address string, pageParams map[string]string) (LogsPage, error) {
	if client.addressCalls == nil {
		client.addressCalls = map[string]int{}
	}
	if pages, ok := client.pagesByAddress[address]; ok {
		call := client.addressCalls[address]
		client.addressCalls[address] = call + 1
		if call >= len(pages) {
			return LogsPage{}, nil
		}
		return pages[call], nil
	}
	if client.calls >= len(client.pages) {
		return LogsPage{}, nil
	}
	page := client.pages[client.calls]
	client.calls++
	return page, nil
}

type fakeAnalyticsStore struct {
	markets   map[string]repository.UpsertAnalyticsMarketInput
	events    map[string]repository.InsertAnalyticsEventInput
	summary   repository.AnalyticsSummary
	state     repository.UpdateAnalyticsIndexerStateInput
	lifecycle map[string]repository.UpdateAnalyticsMarketLifecycleInput
}

func newFakeAnalyticsStore() *fakeAnalyticsStore {
	return &fakeAnalyticsStore{
		markets:   map[string]repository.UpsertAnalyticsMarketInput{},
		events:    map[string]repository.InsertAnalyticsEventInput{},
		lifecycle: map[string]repository.UpdateAnalyticsMarketLifecycleInput{},
	}
}

func (store *fakeAnalyticsStore) UpsertAnalyticsMarket(ctx context.Context, input repository.UpsertAnalyticsMarketInput) error {
	store.markets[input.MarketAddress] = input
	return nil
}

func (store *fakeAnalyticsStore) InsertAnalyticsEvent(ctx context.Context, input repository.InsertAnalyticsEventInput) (bool, error) {
	key := input.TransactionHash + ":" + strconv.Itoa(input.LogIndex)
	if _, ok := store.events[key]; ok {
		return false, nil
	}
	store.events[key] = input
	return true, nil
}

func (store *fakeAnalyticsStore) ListAnalyticsMarketsByFactory(ctx context.Context, factoryAddress string) ([]repository.AnalyticsMarketContract, error) {
	markets := []repository.AnalyticsMarketContract{}
	for _, market := range store.markets {
		if market.FactoryAddress == factoryAddress {
			markets = append(markets, repository.AnalyticsMarketContract{
				MarketAddress:  market.MarketAddress,
				FactoryAddress: market.FactoryAddress,
			})
		}
	}
	return markets, nil
}

func (store *fakeAnalyticsStore) UpdateAnalyticsMarketLifecycle(ctx context.Context, input repository.UpdateAnalyticsMarketLifecycleInput) error {
	store.lifecycle[input.MarketAddress] = input
	return nil
}

func (store *fakeAnalyticsStore) UpdateAnalyticsIndexerState(ctx context.Context, input repository.UpdateAnalyticsIndexerStateInput) error {
	store.state = input
	return nil
}

func (store *fakeAnalyticsStore) RebuildAnalyticsSummaryCache(ctx context.Context, factoryAddress string) (repository.AnalyticsSummary, error) {
	metrics := repository.AnalyticsMetrics{MarketContractsFound: int64(len(store.markets))}
	uniqueWallets := map[string]bool{}
	resolvedMarkets := map[string]bool{}
	cancelledMarkets := map[string]bool{}
	for _, event := range store.events {
		switch event.EventName {
		case MarketDeployedEvent:
			metrics.MarketsCreated++
		case PositionOpenedEvent:
			metrics.PositionEvents++
			metrics.TotalTrades++
			if event.Side == "YES" {
				metrics.YesPositionEvents++
			}
			if event.Side == "NO" {
				metrics.NoPositionEvents++
			}
			metrics.TestnetUSDCVolume = "child-volume"
		case MarketResolvedEvent:
			resolvedMarkets[event.MarketAddress] = true
		case MarketCancelledEvent:
			cancelledMarkets[event.MarketAddress] = true
		case PayoutClaimedEvent:
			metrics.ClaimEvents++
			metrics.PayoutsClaimed++
		case RefundClaimedEvent:
			metrics.ClaimEvents++
			metrics.RefundsClaimed++
		}
		if event.WalletAddress != "" {
			uniqueWallets[event.WalletAddress] = true
		}
	}
	metrics.UniqueWallets = int64(len(uniqueWallets))
	metrics.ResolvedMarkets = int64(len(resolvedMarkets))
	metrics.CancelledMarkets = int64(len(cancelledMarkets))
	if metrics.TestnetUSDCVolume == "" {
		metrics.TestnetUSDCVolume = "0"
	}
	summary := repository.AnalyticsSummary{
		Status:         repository.AnalyticsStatusOK,
		SourceStatus:   repository.AnalyticsSourceIndexed,
		FactoryAddress: factoryAddress,
		GeneratedAt:    time.Date(2026, 6, 28, 15, 0, 0, 0, time.UTC),
		LatestBlock:    sql.NullInt64{Int64: 49152802, Valid: true},
		LatestEventAt:  sql.NullTime{Time: time.Date(2026, 6, 28, 14, 53, 38, 0, time.UTC), Valid: true},
		Metrics:        metrics,
	}
	store.summary = summary
	return summary, nil
}

func TestBackfillDryRunDoesNotWrite(t *testing.T) {
	store := newFakeAnalyticsStore()
	client := &fakeLogsClient{pages: []LogsPage{{Items: []BlockscoutLog{testMarketDeployedLog("0xtx1", 1)}}}}

	result, err := NewBackfiller(client, store).Run(context.Background(), BackfillOptions{
		FactoryAddress: "0xFactory",
		DryRun:         true,
		PageLimit:      1,
	})
	if err != nil {
		t.Fatalf("dry-run backfill: %v", err)
	}
	if result.EventsParsed != 1 {
		t.Fatalf("expected one parsed event, got %d", result.EventsParsed)
	}
	if len(store.markets) != 0 || len(store.events) != 0 {
		t.Fatalf("dry-run should not write, got markets=%d events=%d", len(store.markets), len(store.events))
	}
}

func TestBackfillIdempotentEventInsertAndSummary(t *testing.T) {
	store := newFakeAnalyticsStore()
	log := testMarketDeployedLog("0xtx1", 2)
	client := &fakeLogsClient{pages: []LogsPage{{Items: []BlockscoutLog{log, log}}}}

	result, err := NewBackfiller(client, store).Run(context.Background(), BackfillOptions{
		FactoryAddress: "0xFactory",
		DryRun:         false,
		PageLimit:      1,
		ChainID:        DefaultChainID,
	})
	if err != nil {
		t.Fatalf("write backfill: %v", err)
	}
	if result.EventsParsed != 2 {
		t.Fatalf("expected two parsed events, got %d", result.EventsParsed)
	}
	if result.EventsInserted != 1 {
		t.Fatalf("expected one inserted event after duplicate, got %d", result.EventsInserted)
	}
	if len(store.markets) != 1 {
		t.Fatalf("expected one market upsert, got %d", len(store.markets))
	}
	if result.Summary.Metrics.MarketsCreated != 1 {
		t.Fatalf("expected rebuilt summary markets_created=1, got %#v", result.Summary.Metrics)
	}
	if store.state.LastIndexedBlock != 49152802 {
		t.Fatalf("expected indexer state latest block, got %d", store.state.LastIndexedBlock)
	}
}

func TestBackfillSkipsLogsBeforeFromBlock(t *testing.T) {
	store := newFakeAnalyticsStore()
	client := &fakeLogsClient{pages: []LogsPage{{Items: []BlockscoutLog{testMarketDeployedLog("0xtx1", 1)}}}}

	result, err := NewBackfiller(client, store).Run(context.Background(), BackfillOptions{
		FactoryAddress: "0xFactory",
		FromBlock:      49152803,
		DryRun:         false,
		PageLimit:      1,
	})
	if err != nil {
		t.Fatalf("backfill: %v", err)
	}
	if result.EventsParsed != 0 || len(store.events) != 0 {
		t.Fatalf("expected old log to be skipped, result=%#v events=%d", result, len(store.events))
	}
}

func TestBackfillChildEventsDryRunDoesNotWrite(t *testing.T) {
	store := newFakeAnalyticsStore()
	store.markets["0xMarket"] = repository.UpsertAnalyticsMarketInput{
		MarketAddress:  "0xMarket",
		FactoryAddress: "0xFactory",
	}
	client := &fakeLogsClient{pagesByAddress: map[string][]LogsPage{
		"0xFactory": {{Items: nil}},
		"0xMarket":  {{Items: []BlockscoutLog{testPositionOpenedLog("0xchild1", 1)}}},
	}}

	result, err := NewBackfiller(client, store).Run(context.Background(), BackfillOptions{
		FactoryAddress:      "0xFactory",
		DryRun:              true,
		PageLimit:           1,
		IncludeMarketEvents: true,
	})
	if err != nil {
		t.Fatalf("child dry-run backfill: %v", err)
	}
	if result.EventsParsed != 1 {
		t.Fatalf("expected one child event parsed, got %d", result.EventsParsed)
	}
	if len(store.events) != 0 || len(store.lifecycle) != 0 {
		t.Fatalf("dry-run should not write child events or lifecycle updates")
	}
}

func TestBackfillChildEventsIdempotentAndSummaryIncludesChildMetrics(t *testing.T) {
	store := newFakeAnalyticsStore()
	store.markets["0xMarket"] = repository.UpsertAnalyticsMarketInput{
		MarketAddress:  "0xMarket",
		FactoryAddress: "0xFactory",
	}
	position := testPositionOpenedLog("0xposition", 1)
	client := &fakeLogsClient{pagesByAddress: map[string][]LogsPage{
		"0xFactory": {{Items: nil}},
		"0xMarket": {{
			Items: []BlockscoutLog{
				position,
				position,
				testMarketResolvedLog("0xresolved", 2),
				testPayoutClaimedLog("0xpayout", 3),
				testRefundClaimedLog("0xrefund", 4),
			},
		}},
	}}

	result, err := NewBackfiller(client, store).Run(context.Background(), BackfillOptions{
		FactoryAddress:      "0xFactory",
		DryRun:              false,
		PageLimit:           1,
		IncludeMarketEvents: true,
	})
	if err != nil {
		t.Fatalf("child write backfill: %v", err)
	}
	if result.EventsParsed != 5 {
		t.Fatalf("expected five parsed child logs including duplicate, got %d", result.EventsParsed)
	}
	if result.EventsInserted != 4 {
		t.Fatalf("expected duplicate child event to be ignored, got inserted=%d", result.EventsInserted)
	}
	if store.lifecycle["0xMarket"].Status != "RESOLVED" || store.lifecycle["0xMarket"].WinningOutcome != "YES" {
		t.Fatalf("expected lifecycle update for resolved market, got %#v", store.lifecycle["0xMarket"])
	}
	if result.Summary.Metrics.TotalTrades != 1 ||
		result.Summary.Metrics.YesPositionEvents != 1 ||
		result.Summary.Metrics.ClaimEvents != 2 ||
		result.Summary.Metrics.ResolvedMarkets != 1 {
		t.Fatalf("expected child metrics in summary, got %#v", result.Summary.Metrics)
	}
}

func testMarketDeployedLog(txHash string, logIndex int) BlockscoutLog {
	return BlockscoutLog{
		BlockNumber:     49152802,
		BlockTimestamp:  "2026-06-28T14:53:38.000000Z",
		Index:           logIndex,
		TransactionHash: txHash,
		Raw:             json.RawMessage(`{"safe":"raw"}`),
		Decoded: &BlockscoutDecoded{
			MethodCall: "MarketDeployed(string indexed marketId, address indexed market, address indexed creator, address resolver, address collateralToken, uint256 closeTimestamp, string question)",
			Parameters: []BlockscoutParameter{
				{Name: "marketId", Value: "0xmarketidhash"},
				{Name: "market", Value: "0x09646deC03f5724C38BD486b0992A8CaF50Fcc59"},
				{Name: "creator", Value: "0xE2BB0d3445f5681994413879f5eF0802B4c2F624"},
				{Name: "resolver", Value: "0xE2BB0d3445f5681994413879f5eF0802B4c2F624"},
				{Name: "collateralToken", Value: "0x3600000000000000000000000000000000000000"},
				{Name: "closeTimestamp", Value: "1782658963"},
				{Name: "question", Value: "SignalArc test market"},
			},
		},
	}
}

func testPositionOpenedLog(txHash string, logIndex int) BlockscoutLog {
	return testChildLog(txHash, logIndex, "PositionOpened(address indexed user, uint8 indexed side, uint256 amount)", []BlockscoutParameter{
		{Name: "user", Value: "0xUser"},
		{Name: "side", Value: "1"},
		{Name: "amount", Value: "1000000"},
	})
}

func testMarketResolvedLog(txHash string, logIndex int) BlockscoutLog {
	return testChildLog(txHash, logIndex, "MarketResolved(uint8 winningOutcome)", []BlockscoutParameter{
		{Name: "winningOutcome", Value: "1"},
	})
}

func testPayoutClaimedLog(txHash string, logIndex int) BlockscoutLog {
	return testChildLog(txHash, logIndex, "PayoutClaimed(address indexed user, uint256 amount)", []BlockscoutParameter{
		{Name: "user", Value: "0xWinner"},
		{Name: "amount", Value: "2000000"},
	})
}

func testRefundClaimedLog(txHash string, logIndex int) BlockscoutLog {
	return testChildLog(txHash, logIndex, "RefundClaimed(address indexed user, uint256 amount)", []BlockscoutParameter{
		{Name: "user", Value: "0xRefunded"},
		{Name: "amount", Value: "3000000"},
	})
}

func testChildLog(txHash string, logIndex int, methodCall string, params []BlockscoutParameter) BlockscoutLog {
	return BlockscoutLog{
		BlockNumber:     49152803,
		BlockTimestamp:  "2026-06-28T14:54:38.000000Z",
		Index:           logIndex,
		TransactionHash: txHash,
		Raw:             json.RawMessage(`{"child":"raw"}`),
		Decoded: &BlockscoutDecoded{
			MethodCall: methodCall,
			Parameters: params,
		},
	}
}
