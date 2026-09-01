package circleapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetWalletTokenBalances(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/w3s/wallets/wallet-1/balances" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("missing authorization header")
		}
		_, _ = w.Write([]byte(`{
			"data": {
				"tokenBalances": [
					{
						"token": {
							"id": "token-1",
							"name": "USDC",
							"symbol": "USDC",
							"decimals": 6,
							"address": "0x3600000000000000000000000000000000000000",
							"blockchain": "ARC-TESTNET"
						},
						"amount": "10.25"
					}
				]
			}
		}`))
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{APIKey: "test-key", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	balances, err := client.GetWalletTokenBalances(context.Background(), "wallet-1")
	if err != nil {
		t.Fatalf("get balances: %v", err)
	}
	if len(balances) != 1 {
		t.Fatalf("expected one balance, got %#v", balances)
	}
	if balances[0].Token.Symbol != "USDC" || balances[0].Token.Decimals != 6 || balances[0].Amount != "10.25" {
		t.Fatalf("unexpected balance: %#v", balances[0])
	}
}

func TestGetWalletTokenBalancesEmptyResponseReturnsEmptySlice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{APIKey: "test-key", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	balances, err := client.GetWalletTokenBalances(context.Background(), "wallet-1")
	if err != nil {
		t.Fatalf("get balances: %v", err)
	}
	if balances == nil {
		t.Fatal("expected non-nil empty balances")
	}
	if len(balances) != 0 {
		t.Fatalf("expected empty balances, got %#v", balances)
	}
}
