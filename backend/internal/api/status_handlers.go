package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/wahyu241205/SignalArc/backend/internal/database"
	"github.com/wahyu241205/SignalArc/backend/internal/httpjson"
	"github.com/wahyu241205/SignalArc/backend/internal/repository"
)

func registerStatusRoutes(router chi.Router, db *database.DB) {
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	router.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(r.Context()); err != nil {
			httpjson.WriteError(w, http.StatusServiceUnavailable, "database_unavailable", "database is not reachable")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	router.Get("/schema/validate", func(w http.ResponseWriter, r *http.Request) {
		result, err := db.ValidateSchema(r.Context())
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "schema_validation_failed", "schema validation query failed")
			return
		}

		statusCode := http.StatusOK
		if result.Status != "ok" {
			statusCode = http.StatusServiceUnavailable
		}

		httpjson.WriteJSON(w, statusCode, result)
	})
}

func registerPositionRoutes(router chi.Router, positionsRepository *repository.PositionsRepository) {
	router.Get("/users/{user_id}/positions", func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "user_id")
		positions, err := positionsRepository.ListPositionsByUserID(r.Context(), userID, defaultListLimit)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "positions_list_failed", "failed to list positions")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"positions": newPositionResponses(positions),
		})
	})

	router.Get("/markets/{market_id}/positions", func(w http.ResponseWriter, r *http.Request) {
		marketID := chi.URLParam(r, "market_id")
		positions, err := positionsRepository.ListPositionsByMarketID(r.Context(), marketID, defaultListLimit)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "positions_list_failed", "failed to list positions")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"positions": newPositionResponses(positions),
		})
	})
}

func registerResolutionRoutes(router chi.Router, resolutionsRepository *repository.ResolutionsRepository) {
	router.Get("/markets/{market_id}/resolution", func(w http.ResponseWriter, r *http.Request) {
		marketID := chi.URLParam(r, "market_id")
		resolution, err := resolutionsRepository.GetResolutionByMarketID(r.Context(), marketID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httpjson.WriteJSON(w, http.StatusOK, map[string]any{
					"resolution": nil,
				})
				return
			}

			httpjson.WriteError(w, http.StatusInternalServerError, "resolution_get_failed", "failed to get resolution")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"resolution": newResolutionResponse(resolution),
		})
	})
}

func registerSettlementRoutes(router chi.Router, settlementsRepository *repository.SettlementsRepository) {
	router.Get("/users/{user_id}/settlements", func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "user_id")
		settlements, err := settlementsRepository.ListSettlementsByUserID(r.Context(), userID, defaultListLimit)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "settlements_list_failed", "failed to list settlements")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"settlements": newSettlementResponses(settlements),
		})
	})

	router.Get("/markets/{market_id}/settlements", func(w http.ResponseWriter, r *http.Request) {
		marketID := chi.URLParam(r, "market_id")
		settlements, err := settlementsRepository.ListSettlementsByMarketID(r.Context(), marketID, defaultListLimit)
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "settlements_list_failed", "failed to list settlements")
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]any{
			"settlements": newSettlementResponses(settlements),
		})
	})
}
