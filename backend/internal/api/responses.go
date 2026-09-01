package api

import (
	"database/sql"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/repository"
)

type tradeResponse struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	MarketID         string    `json:"market_id"`
	Outcome          string    `json:"outcome"`
	Side             string    `json:"side"`
	Quantity         string    `json:"quantity"`
	Price            string    `json:"price"`
	CollateralAmount string    `json:"collateral_amount"`
	FeeAmount        string    `json:"fee_amount"`
	Status           string    `json:"status"`
	TxHash           *string   `json:"tx_hash"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type positionResponse struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	MarketID          string    `json:"market_id"`
	Outcome           string    `json:"outcome"`
	Quantity          string    `json:"quantity"`
	AverageEntryPrice string    `json:"average_entry_price"`
	RealizedPnL       string    `json:"realized_pnl"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type resolutionResponse struct {
	ID                string     `json:"id"`
	MarketID          string     `json:"market_id"`
	WinningOutcome    *string    `json:"winning_outcome"`
	Status            string     `json:"status"`
	ResolverType      *string    `json:"resolver_type"`
	EvidenceReference *string    `json:"evidence_reference"`
	ResolvedAt        *time.Time `json:"resolved_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type settlementResponse struct {
	ID           string     `json:"id"`
	MarketID     string     `json:"market_id"`
	UserID       *string    `json:"user_id"`
	ResolutionID *string    `json:"resolution_id"`
	Outcome      *string    `json:"outcome"`
	Amount       string     `json:"amount"`
	Status       string     `json:"status"`
	TxHash       *string    `json:"tx_hash"`
	SettledAt    *time.Time `json:"settled_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type marketResponse struct {
	ID                      string     `json:"id"`
	CreatorUserID           string     `json:"creator_user_id"`
	Title                   string     `json:"title"`
	Description             *string    `json:"description"`
	Category                *string    `json:"category"`
	CoverImageURL           *string    `json:"cover_image_url"`
	Status                  string     `json:"status"`
	OutcomeYesLabel         string     `json:"outcome_yes_label"`
	OutcomeNoLabel          string     `json:"outcome_no_label"`
	CollateralAsset         string     `json:"collateral_asset"`
	Chain                   string     `json:"chain"`
	ResolutionSource        *string    `json:"resolution_source"`
	OpensAt                 *time.Time `json:"opens_at"`
	ClosesAt                time.Time  `json:"closes_at"`
	ResolvedAt              *time.Time `json:"resolved_at"`
	SettledAt               *time.Time `json:"settled_at"`
	WinningOutcome          *string    `json:"winning_outcome"`
	MarketContractAddress   *string    `json:"market_contract_address"`
	MarketDeploymentTxHash  *string    `json:"market_deployment_tx_hash"`
	MarketFactoryAddress    *string    `json:"market_factory_address"`
	ResolverAddress         *string    `json:"resolver_address"`
	OnchainDeploymentStatus string     `json:"onchain_deployment_status"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

type agentMarketResponse struct {
	ID                    string    `json:"id"`
	Title                 string    `json:"title"`
	Status                string    `json:"status"`
	Category              *string   `json:"category"`
	CoverImageURL         *string   `json:"cover_image_url"`
	CollateralAsset       string    `json:"collateral_asset"`
	Chain                 string    `json:"chain"`
	ClosesAt              time.Time `json:"closes_at"`
	MarketContractAddress *string   `json:"market_contract_address"`
	ResolutionSource      *string   `json:"resolution_source"`
}

func newTradeResponse(trade repository.Trade) tradeResponse {
	return tradeResponse{
		ID:               trade.ID,
		UserID:           trade.UserID,
		MarketID:         trade.MarketID,
		Outcome:          trade.Outcome,
		Side:             trade.Side,
		Quantity:         trade.Quantity,
		Price:            trade.Price,
		CollateralAmount: trade.CollateralAmount,
		FeeAmount:        trade.FeeAmount,
		Status:           trade.Status,
		TxHash:           nullStringPtr(trade.TxHash),
		CreatedAt:        trade.CreatedAt,
		UpdatedAt:        trade.UpdatedAt,
	}
}

func newPositionResponses(positions []repository.Position) []positionResponse {
	responses := make([]positionResponse, 0, len(positions))
	for _, position := range positions {
		responses = append(responses, positionResponse{
			ID:                position.ID,
			UserID:            position.UserID,
			MarketID:          position.MarketID,
			Outcome:           position.Outcome,
			Quantity:          position.Quantity,
			AverageEntryPrice: position.AverageEntryPrice,
			RealizedPnL:       position.RealizedPnL,
			CreatedAt:         position.CreatedAt,
			UpdatedAt:         position.UpdatedAt,
		})
	}

	return responses
}

func newResolutionResponse(resolution repository.Resolution) resolutionResponse {
	return resolutionResponse{
		ID:                resolution.ID,
		MarketID:          resolution.MarketID,
		WinningOutcome:    nullStringPtr(resolution.WinningOutcome),
		Status:            resolution.Status,
		ResolverType:      nullStringPtr(resolution.ResolverType),
		EvidenceReference: nullStringPtr(resolution.EvidenceReference),
		ResolvedAt:        nullTimePtr(resolution.ResolvedAt),
		CreatedAt:         resolution.CreatedAt,
		UpdatedAt:         resolution.UpdatedAt,
	}
}

func newSettlementResponses(settlements []repository.Settlement) []settlementResponse {
	responses := make([]settlementResponse, 0, len(settlements))
	for _, settlement := range settlements {
		responses = append(responses, settlementResponse{
			ID:           settlement.ID,
			MarketID:     settlement.MarketID,
			UserID:       nullStringPtr(settlement.UserID),
			ResolutionID: nullStringPtr(settlement.ResolutionID),
			Outcome:      nullStringPtr(settlement.Outcome),
			Amount:       settlement.Amount,
			Status:       settlement.Status,
			TxHash:       nullStringPtr(settlement.TxHash),
			SettledAt:    nullTimePtr(settlement.SettledAt),
			CreatedAt:    settlement.CreatedAt,
			UpdatedAt:    settlement.UpdatedAt,
		})
	}

	return responses
}

func newAgentMarketResponses(markets []repository.Market) []agentMarketResponse {
	responses := make([]agentMarketResponse, 0, len(markets))
	for _, market := range markets {
		responses = append(responses, agentMarketResponse{
			ID:                    market.ID,
			Title:                 market.Title,
			Status:                market.Status,
			Category:              nullStringPtr(market.Category),
			CoverImageURL:         nullStringPtr(market.CoverImageURL),
			CollateralAsset:       market.CollateralAsset,
			Chain:                 market.Chain,
			ClosesAt:              market.ClosesAt,
			MarketContractAddress: nullStringPtr(market.MarketContractAddress),
			ResolutionSource:      nullStringPtr(market.ResolutionSource),
		})
	}

	return responses
}

func newMarketResponses(markets []repository.Market) []marketResponse {
	responses := make([]marketResponse, 0, len(markets))
	for _, market := range markets {
		responses = append(responses, newMarketResponse(market))
	}

	return responses
}

func newMarketResponse(market repository.Market) marketResponse {
	return marketResponse{
		ID:                      market.ID,
		CreatorUserID:           market.CreatorUserID,
		Title:                   market.Title,
		Description:             nullStringPtr(market.Description),
		Category:                nullStringPtr(market.Category),
		CoverImageURL:           nullStringPtr(market.CoverImageURL),
		Status:                  market.Status,
		OutcomeYesLabel:         market.OutcomeYesLabel,
		OutcomeNoLabel:          market.OutcomeNoLabel,
		CollateralAsset:         market.CollateralAsset,
		Chain:                   market.Chain,
		ResolutionSource:        nullStringPtr(market.ResolutionSource),
		OpensAt:                 nullTimePtr(market.OpensAt),
		ClosesAt:                market.ClosesAt,
		ResolvedAt:              nullTimePtr(market.ResolvedAt),
		SettledAt:               nullTimePtr(market.SettledAt),
		WinningOutcome:          nullStringPtr(market.WinningOutcome),
		MarketContractAddress:   nullStringPtr(market.MarketContractAddress),
		MarketDeploymentTxHash:  nullStringPtr(market.MarketDeploymentTxHash),
		MarketFactoryAddress:    nullStringPtr(market.MarketFactoryAddress),
		ResolverAddress:         nullStringPtr(market.ResolverAddress),
		OnchainDeploymentStatus: market.OnchainDeploymentStatus,
		CreatedAt:               market.CreatedAt,
		UpdatedAt:               market.UpdatedAt,
	}
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}
