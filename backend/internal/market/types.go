package market

import "encoding/json"

type CreateMarketRequest struct {
	ID                     string  `json:"id"`
	CreatorUserID          string  `json:"creator_user_id"`
	Title                  string  `json:"title"`
	Description            *string `json:"description"`
	Category               *string `json:"category"`
	CoverImageURL          *string `json:"cover_image_url"`
	OutcomeYesLabel        *string `json:"outcome_yes_label"`
	OutcomeNoLabel         *string `json:"outcome_no_label"`
	CollateralAsset        *string `json:"collateral_asset"`
	Chain                  string  `json:"chain"`
	ResolutionSource       *string `json:"resolution_source"`
	OpensAt                *string `json:"opens_at"`
	ClosesAt               string  `json:"closes_at"`
	MarketContractAddress  string  `json:"market_contract_address"`
	MarketDeploymentTxHash string  `json:"market_deployment_tx_hash"`
	MarketFactoryAddress   string  `json:"market_factory_address"`
	ResolverAddress        string  `json:"resolver_address"`
	hasForbiddenKeys       bool
}

type AttachMarketContractRequest struct {
	MarketContractAddress  string `json:"market_contract_address"`
	MarketDeploymentTxHash string `json:"market_deployment_tx_hash"`
	MarketFactoryAddress   string `json:"market_factory_address"`
	ResolverAddress        string `json:"resolver_address"`
}

func (request *CreateMarketRequest) UnmarshalJSON(data []byte) error {
	type createMarketRequestAlias CreateMarketRequest
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var alias createMarketRequestAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	_, hasStatus := raw["status"]
	_, hasWinningOutcome := raw["winning_outcome"]
	_, hasResolvedAt := raw["resolved_at"]
	_, hasSettledAt := raw["settled_at"]
	_, hasOnchainDeploymentStatus := raw["onchain_deployment_status"]

	*request = CreateMarketRequest(alias)
	request.hasForbiddenKeys = hasStatus ||
		hasWinningOutcome ||
		hasResolvedAt ||
		hasSettledAt ||
		hasOnchainDeploymentStatus
	return nil
}
