package api

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wahyu241205/SignalArc/backend/internal/httpjson"
	"github.com/wahyu241205/SignalArc/backend/internal/repository"
)

const defaultAnalyticsFactoryAddress = "0x02555FC5EE3c53938f2F0356e963865503442A56"

type analyticsSummaryReader interface {
	GetSummary(ctx context.Context, cacheKey string, factoryAddress string) (repository.AnalyticsSummary, error)
}

type analyticsSummaryResponse struct {
	Status         string                      `json:"status"`
	SourceStatus   string                      `json:"source_status"`
	FactoryAddress string                      `json:"factory_address"`
	GeneratedAt    time.Time                   `json:"generated_at"`
	LatestEventAt  *time.Time                  `json:"latest_event_at"`
	LatestBlock    *int64                      `json:"latest_block"`
	Metrics        repository.AnalyticsMetrics `json:"metrics"`
}

func registerAnalyticsRoutes(router chi.Router, analyticsRepository analyticsSummaryReader) {
	router.Get("/analytics/summary", func(w http.ResponseWriter, r *http.Request) {
		factoryAddress := analyticsFactoryAddress()
		summary, err := analyticsRepository.GetSummary(r.Context(), repository.AnalyticsSummaryCacheKey, factoryAddress)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "analytics_summary_failed", "failed to load analytics summary")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, newAnalyticsSummaryResponse(summary, factoryAddress))
	})
}

func analyticsFactoryAddress() string {
	if value := os.Getenv("SIGNAL_ARC_MARKET_FACTORY_ADDRESS"); value != "" {
		return value
	}
	return defaultAnalyticsFactoryAddress
}

func newAnalyticsSummaryResponse(summary repository.AnalyticsSummary, fallbackFactoryAddress string) analyticsSummaryResponse {
	factoryAddress := summary.FactoryAddress
	if factoryAddress == "" {
		factoryAddress = fallbackFactoryAddress
	}

	status := summary.Status
	if status == "" {
		status = repository.AnalyticsStatusOK
	}

	sourceStatus := summary.SourceStatus
	if sourceStatus == "" {
		sourceStatus = repository.AnalyticsSourceNotIndexed
	}

	generatedAt := summary.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}

	var latestEventAt *time.Time
	if summary.LatestEventAt.Valid {
		value := summary.LatestEventAt.Time
		latestEventAt = &value
	}

	var latestBlock *int64
	if summary.LatestBlock.Valid {
		value := summary.LatestBlock.Int64
		latestBlock = &value
	}

	if summary.Metrics.TestnetUSDCVolume == "" {
		summary.Metrics.TestnetUSDCVolume = "0"
	}

	return analyticsSummaryResponse{
		Status:         status,
		SourceStatus:   sourceStatus,
		FactoryAddress: factoryAddress,
		GeneratedAt:    generatedAt,
		LatestEventAt:  latestEventAt,
		LatestBlock:    latestBlock,
		Metrics:        summary.Metrics,
	}
}
