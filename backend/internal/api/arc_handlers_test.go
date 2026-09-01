package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestArcContractReturnsConfiguredFactoryAddress(t *testing.T) {
	const factoryAddress = "0x837e09E8D7806E0e7b740b798173756315E51206"
	t.Setenv("SIGNAL_ARC_MARKET_FACTORY_ADDRESS", factoryAddress)

	router := chi.NewRouter()
	registerArcRoutes(router)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/arc/contract", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var body arcContractResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.SignalArcMarketFactory != factoryAddress {
		t.Fatalf("expected factory address %q, got %q", factoryAddress, body.SignalArcMarketFactory)
	}
	if body.Network != "Arc Testnet" {
		t.Fatalf("expected existing network field to remain unchanged, got %q", body.Network)
	}
}

func TestArcContractReturnsEmptyFactoryAddressWhenUnset(t *testing.T) {
	t.Setenv("SIGNAL_ARC_MARKET_FACTORY_ADDRESS", "")

	router := chi.NewRouter()
	registerArcRoutes(router)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/arc/contract", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var body arcContractResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.SignalArcMarketFactory != "" {
		t.Fatalf("expected empty factory address when env is unset, got %q", body.SignalArcMarketFactory)
	}
}
