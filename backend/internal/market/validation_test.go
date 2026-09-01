package market

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestCreateMarketRequestAcceptsOptionalHTTPSCoverImageURL(t *testing.T) {
	coverImageURL := " https://example.com/market-cover.png "
	request := validCreateMarketRequest()
	request.CoverImageURL = &coverImageURL

	input, err := request.ToRepositoryInput(time.Now())
	if err != nil {
		t.Fatalf("expected valid cover image URL, got %v", err)
	}
	if !input.CoverImageURL.Valid {
		t.Fatal("expected cover image URL to be stored")
	}
	if input.CoverImageURL.String != strings.TrimSpace(coverImageURL) {
		t.Fatalf("expected trimmed cover image URL, got %q", input.CoverImageURL.String)
	}
}

func TestCreateMarketRequestAllowsMissingCoverImageURL(t *testing.T) {
	request := validCreateMarketRequest()

	input, err := request.ToRepositoryInput(time.Now())
	if err != nil {
		t.Fatalf("expected missing cover image URL to be valid, got %v", err)
	}
	if input.CoverImageURL.Valid {
		t.Fatalf("expected missing cover image URL to remain null, got %q", input.CoverImageURL.String)
	}
}

func TestCreateMarketRequestRejectsInvalidCoverImageURL(t *testing.T) {
	testCases := map[string]string{
		"http scheme":  "http://example.com/market-cover.png",
		"base64 data":  "data:image/png;base64,abc123",
		"missing host": "https:///market-cover.png",
		"not a URL":    "market-cover.png",
		"too long":     "https://example.com/" + strings.Repeat("a", 2049),
	}

	for name, coverImageURL := range testCases {
		coverImageURL := coverImageURL
		t.Run(name, func(t *testing.T) {
			request := validCreateMarketRequest()
			request.CoverImageURL = &coverImageURL

			if _, err := request.ToRepositoryInput(time.Now()); err == nil {
				t.Fatalf("expected cover image URL %q to be rejected", coverImageURL)
			}
		})
	}
}

func TestCreateMarketRequestRequiresDeploymentFields(t *testing.T) {
	testCases := map[string]func(*CreateMarketRequest){
		"id": func(request *CreateMarketRequest) {
			request.ID = ""
		},
		"market_contract_address": func(request *CreateMarketRequest) {
			request.MarketContractAddress = ""
		},
		"market_deployment_tx_hash": func(request *CreateMarketRequest) {
			request.MarketDeploymentTxHash = ""
		},
		"market_factory_address": func(request *CreateMarketRequest) {
			request.MarketFactoryAddress = ""
		},
		"resolver_address": func(request *CreateMarketRequest) {
			request.ResolverAddress = ""
		},
	}

	for name, mutate := range testCases {
		t.Run(name, func(t *testing.T) {
			request := validCreateMarketRequest()
			mutate(&request)

			if _, err := request.ToRepositoryInput(time.Now()); err == nil {
				t.Fatalf("expected missing %s to be rejected", name)
			}
		})
	}
}

func TestCreateMarketRequestRejectsInvalidDeploymentFields(t *testing.T) {
	testCases := map[string]func(*CreateMarketRequest){
		"id": func(request *CreateMarketRequest) {
			request.ID = "not-a-uuid"
		},
		"market_contract_address": func(request *CreateMarketRequest) {
			request.MarketContractAddress = "0x1234"
		},
		"market_deployment_tx_hash": func(request *CreateMarketRequest) {
			request.MarketDeploymentTxHash = "0x1234"
		},
		"market_factory_address": func(request *CreateMarketRequest) {
			request.MarketFactoryAddress = "0x1234"
		},
		"resolver_address": func(request *CreateMarketRequest) {
			request.ResolverAddress = "0x1234"
		},
	}

	for name, mutate := range testCases {
		t.Run(name, func(t *testing.T) {
			request := validCreateMarketRequest()
			mutate(&request)

			if _, err := request.ToRepositoryInput(time.Now()); err == nil {
				t.Fatalf("expected invalid %s to be rejected", name)
			}
		})
	}
}

func TestCreateMarketRequestRejectsServerOwnedLifecycleFields(t *testing.T) {
	payload := `{
		"id": "20000000-0000-4000-8000-000000000002",
		"creator_user_id": "10000000-0000-4000-8000-000000000001",
		"title": "Will SignalArc require deployed markets?",
		"chain": "Arc Testnet",
		"closes_at": "` + time.Now().Add(24*time.Hour).Format(time.RFC3339) + `",
		"market_contract_address": "0x1111111111111111111111111111111111111111",
		"market_deployment_tx_hash": "0x2222222222222222222222222222222222222222222222222222222222222222",
		"market_factory_address": "0x3333333333333333333333333333333333333333",
		"resolver_address": "0x4444444444444444444444444444444444444444",
		"status": "OPEN",
		"onchain_deployment_status": "DEPLOYED"
	}`

	var request CreateMarketRequest
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		t.Fatalf("expected valid JSON, got %v", err)
	}

	if _, err := request.ToRepositoryInput(time.Now()); err == nil {
		t.Fatal("expected server-owned lifecycle fields to be rejected")
	}
}

func TestCreateMarketRequestMapsDeployedRequestToRepositoryInput(t *testing.T) {
	request := validCreateMarketRequest()
	now := time.Now()

	input, err := request.ToRepositoryInput(now)
	if err != nil {
		t.Fatalf("expected valid deployed create request, got %v", err)
	}

	if input.ID != strings.ToLower(request.ID) {
		t.Fatalf("expected id %q, got %q", request.ID, input.ID)
	}
	if input.MarketContractAddress != strings.ToLower(request.MarketContractAddress) {
		t.Fatalf("expected market contract address to be normalized, got %q", input.MarketContractAddress)
	}
	if input.MarketDeploymentTxHash != strings.ToLower(request.MarketDeploymentTxHash) {
		t.Fatalf("expected deployment tx hash to be normalized, got %q", input.MarketDeploymentTxHash)
	}
	if input.MarketFactoryAddress != strings.ToLower(request.MarketFactoryAddress) {
		t.Fatalf("expected factory address to be normalized, got %q", input.MarketFactoryAddress)
	}
	if input.ResolverAddress != strings.ToLower(request.ResolverAddress) {
		t.Fatalf("expected resolver address to be normalized, got %q", input.ResolverAddress)
	}
}

func validCreateMarketRequest() CreateMarketRequest {
	return CreateMarketRequest{
		ID:                     "20000000-0000-4000-8000-000000000002",
		CreatorUserID:          "10000000-0000-4000-8000-000000000001",
		Title:                  "Will SignalArc support market images?",
		Chain:                  "Arc Testnet",
		ClosesAt:               time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		MarketContractAddress:  "0x1111111111111111111111111111111111111111",
		MarketDeploymentTxHash: "0x2222222222222222222222222222222222222222222222222222222222222222",
		MarketFactoryAddress:   "0x3333333333333333333333333333333333333333",
		ResolverAddress:        "0x4444444444444444444444444444444444444444",
	}
}
