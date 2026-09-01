package main

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/analytics"
	"github.com/wahyu241205/SignalArc/backend/internal/repository"
)

func TestCommandOutputDoesNotLeakAPIKey(t *testing.T) {
	const sensitiveValue = "redacted-test-value"
	result := analytics.BackfillResult{
		FactoryAddress: activeFactoryAddress,
		DryRun:         true,
		PagesFetched:   1,
		LogsSeen:       2,
		EventsParsed:   2,
		LatestBlock:    sql.NullInt64{Int64: 49152802, Valid: true},
		LatestEventAt:  sql.NullTime{Time: time.Date(2026, 6, 28, 14, 53, 38, 0, time.UTC), Valid: true},
		Summary: repository.AnalyticsSummary{
			Status:       repository.AnalyticsStatusOK,
			SourceStatus: sensitiveValue,
		},
	}

	output, err := json.Marshal(newCommandOutput(result))
	if err != nil {
		t.Fatalf("marshal command output: %v", err)
	}
	text := string(output)
	if strings.Contains(text, sensitiveValue) {
		t.Fatalf("dry-run output leaked secret-like summary value: %s", text)
	}
	if strings.Contains(strings.ToLower(text), "blockscout_api_key") {
		t.Fatalf("output leaked API key name: %s", text)
	}
}
