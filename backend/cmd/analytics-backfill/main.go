package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wahyu241205/SignalArc/backend/internal/analytics"
	"github.com/wahyu241205/SignalArc/backend/internal/database"
	"github.com/wahyu241205/SignalArc/backend/internal/repository"
)

const activeFactoryAddress = "0x02555FC5EE3c53938f2F0356e963865503442A56"

type commandOutput struct {
	Status              string                       `json:"status"`
	Mode                string                       `json:"mode"`
	FactoryAddress      string                       `json:"factory_address"`
	PagesFetched        int                          `json:"pages_fetched"`
	LogsSeen            int                          `json:"logs_seen"`
	EventsParsed        int                          `json:"events_parsed"`
	EventsInserted      int                          `json:"events_inserted"`
	MarketsUpserted     int                          `json:"markets_upserted"`
	IncludeMarketEvents bool                         `json:"include_market_events"`
	LatestBlock         *int64                       `json:"latest_block"`
	LatestEventAt       *time.Time                   `json:"latest_event_at"`
	Summary             *repository.AnalyticsSummary `json:"summary,omitempty"`
}

func main() {
	var (
		dryRun              = flag.Bool("dry-run", true, "fetch and parse logs without writing to the database")
		factoryAddress      = flag.String("factory", activeFactoryAddress, "SignalArc factory contract address")
		fromBlock           = flag.Int64("from-block", 0, "minimum block number to ingest")
		pageLimit           = flag.Int("page-limit", 1, "maximum Blockscout v2 pages to fetch; set 0 for all pages")
		baseURL             = flag.String("base-url", analytics.DefaultArcscanBaseURL, "Arcscan/Blockscout base URL")
		timeoutSeconds      = flag.Int("timeout-seconds", 15, "HTTP timeout in seconds")
		includeMarketEvents = flag.Bool("include-market-events", false, "also fetch logs for market contracts already discovered in analytics_markets")
	)
	flag.Parse()

	ctx := context.Background()
	apiKey := os.Getenv("BLOCKSCOUT_API_KEY")
	if apiKey == "" {
		apiKey = readBlockscoutAPIKey("../contracts/.env")
	}
	if apiKey == "" {
		apiKey = readBlockscoutAPIKey("contracts/.env")
	}

	client := analytics.NewArcscanClient(analytics.ArcscanClientConfig{
		BaseURL: *baseURL,
		APIKey:  apiKey,
		Timeout: time.Duration(*timeoutSeconds) * time.Second,
	})

	var store analytics.Store
	var db *database.DB
	if !*dryRun || *includeMarketEvents {
		databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
		if databaseURL == "" {
			log.Fatal().Msg("DATABASE_URL is required when dry-run=false or include-market-events=true")
		}
		connected, err := database.Connect(ctx, databaseURL)
		if err != nil {
			log.Fatal().Err(err).Msg("connect database")
		}
		db = connected
		defer db.Close()
		store = repository.NewAnalyticsRepository(db)
	}

	result, err := analytics.NewBackfiller(client, store).Run(ctx, analytics.BackfillOptions{
		FactoryAddress:      strings.TrimSpace(*factoryAddress),
		FromBlock:           *fromBlock,
		PageLimit:           *pageLimit,
		DryRun:              *dryRun,
		ChainID:             analytics.DefaultChainID,
		IncludeMarketEvents: *includeMarketEvents,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("analytics backfill failed")
	}

	if err := json.NewEncoder(os.Stdout).Encode(newCommandOutput(result)); err != nil {
		log.Fatal().Err(err).Msg("encode analytics backfill output")
	}
}

func newCommandOutput(result analytics.BackfillResult) commandOutput {
	mode := "write"
	if result.DryRun {
		mode = "dry_run"
	}

	var latestBlock *int64
	if result.LatestBlock.Valid {
		value := result.LatestBlock.Int64
		latestBlock = &value
	}

	var latestEventAt *time.Time
	if result.LatestEventAt.Valid {
		value := result.LatestEventAt.Time
		latestEventAt = &value
	}

	var summary *repository.AnalyticsSummary
	if !result.DryRun {
		summary = &result.Summary
	}

	return commandOutput{
		Status:              "ok",
		Mode:                mode,
		FactoryAddress:      result.FactoryAddress,
		PagesFetched:        result.PagesFetched,
		LogsSeen:            result.LogsSeen,
		EventsParsed:        result.EventsParsed,
		EventsInserted:      result.EventsInserted,
		MarketsUpserted:     result.MarketsUpserted,
		IncludeMarketEvents: result.IncludeMarketEvents,
		LatestBlock:         latestBlock,
		LatestEventAt:       latestEventAt,
		Summary:             summary,
	}
}

func readBlockscoutAPIKey(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) != "BLOCKSCOUT_API_KEY" {
			continue
		}
		return strings.Trim(strings.TrimSpace(value), `"'`)
	}
	return ""
}

func init() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage of analytics-backfill:\n")
		flag.PrintDefaults()
	}
}
