package analytics

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestArcscanClientFetchAddressLogsWithPaginationAndAPIKey(t *testing.T) {
	const apiKey = "secret-blockscout-key"
	var sawAPIKey bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/addresses/0xFactory/logs" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("apikey") == apiKey {
			sawAPIKey = true
		}
		if r.URL.Query().Get("block_number") != "123" {
			t.Fatalf("expected page param block_number=123, got query %s", r.URL.RawQuery)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{
					"block_number":     124,
					"block_timestamp":  "2026-06-28T14:53:38.000000Z",
					"index":            7,
					"transaction_hash": "0xabc",
					"topics":           []string{"0xtopic"},
					"decoded": map[string]any{
						"method_call": "MarketDeployed(string indexed marketId, address indexed market, address indexed creator, address resolver, address collateralToken, uint256 closeTimestamp, string question)",
						"parameters":  []map[string]any{},
					},
				},
			},
			"next_page_params": map[string]any{
				"block_number": 125,
				"index":        8,
			},
		})
	}))
	defer server.Close()

	client := NewArcscanClient(ArcscanClientConfig{BaseURL: server.URL, APIKey: apiKey})
	page, err := client.FetchAddressLogs(context.Background(), "0xFactory", map[string]string{"block_number": "123"})
	if err != nil {
		t.Fatalf("fetch logs: %v", err)
	}
	if !sawAPIKey {
		t.Fatal("expected API key to be sent to Blockscout endpoint")
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected one item, got %d", len(page.Items))
	}
	if len(page.Items[0].Raw) == 0 {
		t.Fatal("expected raw log JSON to be preserved")
	}
	if page.NextPageParams["block_number"] != "125" || page.NextPageParams["index"] != "8" {
		t.Fatalf("unexpected next page params %#v", page.NextPageParams)
	}
}

func TestArcscanClientReturnsClearStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewArcscanClient(ArcscanClientConfig{BaseURL: server.URL})
	if _, err := client.FetchAddressLogs(context.Background(), "0xFactory", nil); err == nil {
		t.Fatal("expected status error")
	}
}
