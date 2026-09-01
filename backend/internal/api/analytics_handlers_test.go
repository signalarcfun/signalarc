package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wahyu241205/SignalArc/backend/internal/repository"
)

type fakeAnalyticsSummaryReader struct {
	summary        repository.AnalyticsSummary
	cacheKey       string
	factoryAddress string
}

func (reader *fakeAnalyticsSummaryReader) GetSummary(ctx context.Context, cacheKey string, factoryAddress string) (repository.AnalyticsSummary, error) {
	reader.cacheKey = cacheKey
	reader.factoryAddress = factoryAddress
	return reader.summary, nil
}

func TestAnalyticsSummaryReturnsEmptyNotIndexedResponse(t *testing.T) {
	const factoryAddress = "0x02555FC5EE3c53938f2F0356e963865503442A56"
	t.Setenv("SIGNAL_ARC_MARKET_FACTORY_ADDRESS", factoryAddress)

	reader := &fakeAnalyticsSummaryReader{
		summary: repository.AnalyticsSummary{
			Status:         repository.AnalyticsStatusOK,
			SourceStatus:   repository.AnalyticsSourceNotIndexed,
			FactoryAddress: factoryAddress,
			GeneratedAt:    time.Date(2026, 6, 28, 15, 0, 0, 0, time.UTC),
		},
	}
	router := chi.NewRouter()
	registerAnalyticsRoutes(router, reader)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/analytics/summary", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var body analyticsSummaryResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Status != repository.AnalyticsStatusOK {
		t.Fatalf("expected ok status, got %q", body.Status)
	}
	if body.SourceStatus != repository.AnalyticsSourceNotIndexed {
		t.Fatalf("expected not_indexed source status, got %q", body.SourceStatus)
	}
	if body.FactoryAddress != factoryAddress {
		t.Fatalf("expected factory address %q, got %q", factoryAddress, body.FactoryAddress)
	}
	if body.LatestBlock != nil {
		t.Fatalf("expected nil latest block for empty analytics, got %d", *body.LatestBlock)
	}
	if body.LatestEventAt != nil {
		t.Fatalf("expected nil latest event time for empty analytics, got %s", body.LatestEventAt)
	}
	if body.Metrics.MarketsCreated != 0 ||
		body.Metrics.TotalTrades != 0 ||
		body.Metrics.UniqueWallets != 0 ||
		body.Metrics.TestnetUSDCVolume != "0" {
		t.Fatalf("expected zero/default metrics, got %#v", body.Metrics)
	}
	if reader.cacheKey != repository.AnalyticsSummaryCacheKey {
		t.Fatalf("expected cache key %q, got %q", repository.AnalyticsSummaryCacheKey, reader.cacheKey)
	}
	if reader.factoryAddress != factoryAddress {
		t.Fatalf("expected repository factory address %q, got %q", factoryAddress, reader.factoryAddress)
	}
}

func TestAnalyticsSummaryReturnsCachedSummaryResponse(t *testing.T) {
	const factoryAddress = "0x02555FC5EE3c53938f2F0356e963865503442A56"
	generatedAt := time.Date(2026, 6, 28, 15, 10, 0, 0, time.UTC)
	latestEventAt := time.Date(2026, 6, 28, 15, 9, 0, 0, time.UTC)
	reader := &fakeAnalyticsSummaryReader{
		summary: repository.AnalyticsSummary{
			Status:         repository.AnalyticsStatusOK,
			SourceStatus:   repository.AnalyticsSourceCached,
			FactoryAddress: factoryAddress,
			GeneratedAt:    generatedAt,
			LatestEventAt:  sql.NullTime{Time: latestEventAt, Valid: true},
			LatestBlock:    sql.NullInt64{Int64: 49152802, Valid: true},
			Metrics: repository.AnalyticsMetrics{
				MarketsCreated:       12,
				MarketContractsFound: 12,
				TotalTrades:          34,
				PositionEvents:       34,
				YesPositionEvents:    20,
				NoPositionEvents:     14,
				UniqueWallets:        9,
				TestnetUSDCVolume:    "123000000",
				ResolvedMarkets:      3,
				CancelledMarkets:     2,
				ClaimEvents:          5,
				PayoutsClaimed:       4,
				RefundsClaimed:       1,
			},
		},
	}
	router := chi.NewRouter()
	registerAnalyticsRoutes(router, reader)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/analytics/summary", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var body analyticsSummaryResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.SourceStatus != repository.AnalyticsSourceCached {
		t.Fatalf("expected cached source status, got %q", body.SourceStatus)
	}
	if body.GeneratedAt != generatedAt {
		t.Fatalf("expected generated_at %s, got %s", generatedAt, body.GeneratedAt)
	}
	if body.LatestBlock == nil || *body.LatestBlock != 49152802 {
		t.Fatalf("expected latest block 49152802, got %#v", body.LatestBlock)
	}
	if body.LatestEventAt == nil || !body.LatestEventAt.Equal(latestEventAt) {
		t.Fatalf("expected latest event %s, got %#v", latestEventAt, body.LatestEventAt)
	}
	if body.Metrics.TotalTrades != 34 || body.Metrics.PayoutsClaimed != 4 || body.Metrics.RefundsClaimed != 1 {
		t.Fatalf("expected cached metrics, got %#v", body.Metrics)
	}
}

func TestAnalyticsSummaryDoesNotLeakExplorerSecrets(t *testing.T) {
	const secret = "blockscout-secret-value"
	t.Setenv("BLOCKSCOUT_API_KEY", secret)

	reader := &fakeAnalyticsSummaryReader{
		summary: repository.AnalyticsSummary{
			Status:         repository.AnalyticsStatusOK,
			SourceStatus:   repository.AnalyticsSourceNotIndexed,
			FactoryAddress: defaultAnalyticsFactoryAddress,
			GeneratedAt:    time.Date(2026, 6, 28, 15, 0, 0, 0, time.UTC),
		},
	}
	router := chi.NewRouter()
	registerAnalyticsRoutes(router, reader)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/analytics/summary", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	body := response.Body.String()
	if strings.Contains(body, secret) {
		t.Fatalf("analytics response leaked explorer secret: %s", body)
	}
	if strings.Contains(strings.ToLower(body), "blockscout_api_key") {
		t.Fatalf("analytics response leaked explorer key name: %s", body)
	}
}

func TestAnalyticsSummaryRouteRegistration(t *testing.T) {
	reader := &fakeAnalyticsSummaryReader{
		summary: repository.AnalyticsSummary{
			Status:         repository.AnalyticsStatusOK,
			SourceStatus:   repository.AnalyticsSourceNotIndexed,
			FactoryAddress: defaultAnalyticsFactoryAddress,
			GeneratedAt:    time.Date(2026, 6, 28, 15, 0, 0, 0, time.UTC),
		},
	}
	router := chi.NewRouter()
	registerAnalyticsRoutes(router, reader)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/analytics/summary", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected registered analytics route to return %d, got %d", http.StatusOK, response.Code)
	}
}
