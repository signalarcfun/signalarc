package api

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/wahyu241205/SignalArc/backend/internal/httpjson"
)

type arcContractResponse struct {
	Network                string `json:"network"`
	ChainID                int    `json:"chain_id"`
	SignalArcMarket        string `json:"signal_arc_market"`
	SignalArcMarketFactory string `json:"signal_arc_market_factory"`
	USDCErc20Interface     string `json:"usdc_erc20_interface"`
	Explorer               string `json:"explorer"`
	Prototype              bool   `json:"prototype"`
	ProductionApproved     bool   `json:"production_approved"`
	Status                 string `json:"status"`
}

func registerArcRoutes(router chi.Router) {
	router.Get("/arc/contract", func(w http.ResponseWriter, r *http.Request) {
		httpjson.WriteJSON(w, http.StatusOK, arcContractResponse{
			Network:                "Arc Testnet",
			ChainID:                5042002,
			SignalArcMarket:        "0xf4ccc11A9e24fb996679F946C23C04AFd2797F26",
			SignalArcMarketFactory: os.Getenv("SIGNAL_ARC_MARKET_FACTORY_ADDRESS"),
			USDCErc20Interface:     "0x3600000000000000000000000000000000000000",
			Explorer:               "https://testnet.arcscan.app",
			Prototype:              true,
			ProductionApproved:     false,
			Status:                 "prototype_testnet_reference",
		})
	})
}
