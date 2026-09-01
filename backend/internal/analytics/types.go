package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/repository"
)

const (
	DefaultArcscanBaseURL = "https://testnet.arcscan.app"
	DefaultChainID        = 5042002
	MarketDeployedEvent   = "MarketDeployed"
	PositionOpenedEvent   = "PositionOpened"
	MarketResolvedEvent   = "MarketResolved"
	MarketCancelledEvent  = "MarketCancelled"
	PayoutClaimedEvent    = "PayoutClaimed"
	RefundClaimedEvent    = "RefundClaimed"
	IndexerSourceFactory  = "arcscan_factory_market_deployed"
	IndexerSourceMarkets  = "arcscan_market_events"
)

type BlockscoutLog struct {
	BlockNumber      int64              `json:"block_number"`
	BlockTimestamp   string             `json:"block_timestamp"`
	Data             string             `json:"data"`
	Decoded          *BlockscoutDecoded `json:"decoded"`
	Index            int                `json:"index"`
	Topics           []string           `json:"topics"`
	TransactionHash  string             `json:"transaction_hash"`
	Raw              json.RawMessage    `json:"-"`
	AdditionalFields map[string]any     `json:"-"`
}

type BlockscoutDecoded struct {
	MethodCall string                `json:"method_call"`
	MethodID   string                `json:"method_id"`
	Parameters []BlockscoutParameter `json:"parameters"`
}

type BlockscoutParameter struct {
	Indexed bool   `json:"indexed"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Value   string `json:"value"`
}

type LogsPage struct {
	Items          []BlockscoutLog
	NextPageParams map[string]string
}

type MarketDeployed struct {
	FactoryAddress         string
	MarketIDHash           string
	MarketAddress          string
	CreatorAddress         string
	ResolverAddress        string
	CollateralTokenAddress string
	CloseTimestamp         string
	Question               string
	TransactionHash        string
	BlockNumber            int64
	LogIndex               int
	BlockTimestamp         time.Time
	Raw                    json.RawMessage
}

type MarketEvent struct {
	FactoryAddress  string
	MarketAddress   string
	EventName       string
	TransactionHash string
	BlockNumber     int64
	LogIndex        int
	BlockTimestamp  time.Time
	WalletAddress   string
	Side            string
	WinningOutcome  string
	AmountBaseUnits string
	Status          string
	Raw             json.RawMessage
}

type Client interface {
	FetchAddressLogs(context.Context, string, map[string]string) (LogsPage, error)
}

type Store interface {
	UpsertAnalyticsMarket(context.Context, repository.UpsertAnalyticsMarketInput) error
	InsertAnalyticsEvent(context.Context, repository.InsertAnalyticsEventInput) (bool, error)
	ListAnalyticsMarketsByFactory(context.Context, string) ([]repository.AnalyticsMarketContract, error)
	UpdateAnalyticsMarketLifecycle(context.Context, repository.UpdateAnalyticsMarketLifecycleInput) error
	UpdateAnalyticsIndexerState(context.Context, repository.UpdateAnalyticsIndexerStateInput) error
	RebuildAnalyticsSummaryCache(context.Context, string) (repository.AnalyticsSummary, error)
}

type BackfillOptions struct {
	FactoryAddress      string
	FromBlock           int64
	PageLimit           int
	DryRun              bool
	ChainID             int
	IncludeMarketEvents bool
}

type BackfillResult struct {
	FactoryAddress      string
	DryRun              bool
	IncludeMarketEvents bool
	PagesFetched        int
	LogsSeen            int
	EventsParsed        int
	EventsInserted      int
	MarketsUpserted     int
	LatestBlock         sql.NullInt64
	LatestEventAt       sql.NullTime
	Summary             repository.AnalyticsSummary
}
