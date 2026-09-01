package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/wahyu241205/SignalArc/backend/internal/httpjson"
	"github.com/wahyu241205/SignalArc/backend/internal/market"
	"github.com/wahyu241205/SignalArc/backend/internal/repository"
	"github.com/wahyu241205/SignalArc/backend/internal/validation"
)

func registerMarketRoutes(router chi.Router, marketsRepository *repository.MarketsRepository) {
	router.Get("/markets", func(w http.ResponseWriter, r *http.Request) {
		markets, err := marketsRepository.ListMarkets(r.Context(), defaultMarketsLimit)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "markets_list_failed", "failed to list markets")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"markets": newMarketResponses(markets),
		})
	})

	router.Get("/agent/markets", func(w http.ResponseWriter, r *http.Request) {
		markets, err := marketsRepository.ListMarkets(r.Context(), defaultMarketsLimit)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "markets_list_failed", "failed to list markets")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"markets": newAgentMarketResponses(markets),
		})
	})

	router.Post("/markets", func(w http.ResponseWriter, r *http.Request) {
		var request market.CreateMarketRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON request body")
			return
		}

		input, err := request.ToRepositoryInput(time.Now())
		if err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_market_request", "invalid market request")
			return
		}

		market, err := marketsRepository.CreateMarket(r.Context(), input)
		if validation.IsForeignKeyViolation(err) {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_creator_user", "creator user is invalid")
			return
		}
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "market_create_failed", "failed to create market")
			return
		}

		httpjson.WriteJSON(w, http.StatusCreated, map[string]any{
			"market": newMarketResponse(market),
		})
	})

	router.Patch("/markets/{id}/contract", func(w http.ResponseWriter, r *http.Request) {
		marketID := chi.URLParam(r, "id")
		if !validation.IsUUIDShape(marketID) {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_market_id", "market id is invalid")
			return
		}

		var request market.AttachMarketContractRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON request body")
			return
		}

		input, err := request.ToRepositoryInput()
		if err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_contract_attachment", "invalid contract attachment")
			return
		}

		market, err := marketsRepository.AttachMarketContract(r.Context(), marketID, input)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httpjson.WriteError(w, http.StatusNotFound, "market_not_found", "market not found")
				return
			}
			if validation.IsUniqueViolation(err) {
				httpjson.WriteError(w, http.StatusConflict, "market_contract_already_attached", "market contract address is already attached")
				return
			}

			httpjson.WriteError(w, http.StatusInternalServerError, "market_contract_attach_failed", "failed to attach market contract")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"market": newMarketResponse(market),
		})
	})

	router.Get("/markets/{id}", func(w http.ResponseWriter, r *http.Request) {
		marketID := chi.URLParam(r, "id")
		market, err := marketsRepository.GetMarketByID(r.Context(), marketID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httpjson.WriteError(w, http.StatusNotFound, "market_not_found", "market not found")
				return
			}

			httpjson.WriteError(w, http.StatusInternalServerError, "market_get_failed", "failed to get market")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"market": newMarketResponse(market),
		})
	})
}
