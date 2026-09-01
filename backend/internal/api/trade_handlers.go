package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/wahyu241205/SignalArc/backend/internal/httpjson"
	"github.com/wahyu241205/SignalArc/backend/internal/repository"
	"github.com/wahyu241205/SignalArc/backend/internal/trade"
)

func registerTradeRoutes(router chi.Router, tradesRepository *repository.TradesRepository, marketsRepository *repository.MarketsRepository) {
	router.Post("/trade-intents", func(w http.ResponseWriter, r *http.Request) {
		var request trade.CreateTradeIntentRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON request body")
			return
		}

		input, err := request.ToRepositoryInput()
		if err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_trade_intent", "invalid trade intent")
			return
		}

		market, err := marketsRepository.GetMarketByID(r.Context(), input.MarketID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httpjson.WriteError(w, http.StatusNotFound, "market_not_found", "market not found")
				return
			}

			httpjson.WriteError(w, http.StatusInternalServerError, "trade_intent_create_failed", "failed to create trade intent")
			return
		}

		if market.Status != "OPEN" {
			httpjson.WriteError(w, http.StatusBadRequest, "market_not_open", "market is not open")
			return
		}

		trade, err := tradesRepository.CreateTradeIntent(r.Context(), input)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "trade_intent_create_failed", "failed to create trade intent")
			return
		}

		httpjson.WriteJSON(w, http.StatusCreated, map[string]any{
			"trade": newTradeResponse(trade),
			"execution": map[string]string{
				"status": "not_executed",
				"reason": "onchain execution is not implemented in Phase 3 backend MVP",
			},
		})
	})
}
