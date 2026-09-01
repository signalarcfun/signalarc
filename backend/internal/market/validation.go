package market

import (
	"database/sql"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/repository"
	"github.com/wahyu241205/SignalArc/backend/internal/validation"
)

func (request CreateMarketRequest) ToRepositoryInput(now time.Time) (repository.CreateMarketInput, error) {
	if request.hasForbiddenKeys {
		return repository.CreateMarketInput{}, errors.New("market lifecycle fields are server-owned")
	}

	id := strings.TrimSpace(request.ID)
	if id == "" || !validation.IsUUIDShape(id) {
		return repository.CreateMarketInput{}, errors.New("id is required")
	}

	creatorUserID := strings.TrimSpace(request.CreatorUserID)
	if creatorUserID == "" || !validation.IsUUIDShape(creatorUserID) {
		return repository.CreateMarketInput{}, errors.New("creator_user_id is required")
	}

	title := strings.TrimSpace(request.Title)
	if title == "" {
		return repository.CreateMarketInput{}, errors.New("title is required")
	}

	chain := strings.TrimSpace(request.Chain)
	if chain == "" {
		return repository.CreateMarketInput{}, errors.New("chain is required")
	}

	if strings.TrimSpace(request.ClosesAt) == "" {
		return repository.CreateMarketInput{}, errors.New("closes_at is required")
	}
	closesAt, err := time.Parse(time.RFC3339, strings.TrimSpace(request.ClosesAt))
	if err != nil {
		return repository.CreateMarketInput{}, err
	}
	if !closesAt.After(now) {
		return repository.CreateMarketInput{}, errors.New("closes_at must be in the future")
	}

	opensAt := sql.NullTime{}
	if request.OpensAt != nil && strings.TrimSpace(*request.OpensAt) != "" {
		parsedOpensAt, err := time.Parse(time.RFC3339, strings.TrimSpace(*request.OpensAt))
		if err != nil {
			return repository.CreateMarketInput{}, err
		}
		if !parsedOpensAt.Before(closesAt) {
			return repository.CreateMarketInput{}, errors.New("opens_at must be before closes_at")
		}
		opensAt = sql.NullTime{Time: parsedOpensAt, Valid: true}
	}

	coverImageURL, err := optionalHTTPSURL(request.CoverImageURL)
	if err != nil {
		return repository.CreateMarketInput{}, err
	}

	deploymentInput, err := request.deploymentInput()
	if err != nil {
		return repository.CreateMarketInput{}, err
	}

	return repository.CreateMarketInput{
		ID:                     strings.ToLower(id),
		CreatorUserID:          creatorUserID,
		Title:                  title,
		Description:            validation.OptionalString(request.Description),
		Category:               validation.OptionalString(request.Category),
		CoverImageURL:          coverImageURL,
		Status:                 newMarketStatus(now, opensAt),
		OutcomeYesLabel:        validation.DefaultString(request.OutcomeYesLabel, "YES"),
		OutcomeNoLabel:         validation.DefaultString(request.OutcomeNoLabel, "NO"),
		CollateralAsset:        validation.DefaultString(request.CollateralAsset, "USDC"),
		Chain:                  chain,
		ResolutionSource:       validation.OptionalString(request.ResolutionSource),
		OpensAt:                opensAt,
		ClosesAt:               closesAt,
		MarketContractAddress:  deploymentInput.MarketContractAddress,
		MarketDeploymentTxHash: deploymentInput.MarketDeploymentTxHash,
		MarketFactoryAddress:   deploymentInput.MarketFactoryAddress,
		ResolverAddress:        deploymentInput.ResolverAddress,
	}, nil
}

func optionalHTTPSURL(value *string) (sql.NullString, error) {
	if value == nil {
		return sql.NullString{}, nil
	}

	trimmedValue := strings.TrimSpace(*value)
	if trimmedValue == "" {
		return sql.NullString{}, nil
	}
	if len(trimmedValue) > 2048 {
		return sql.NullString{}, errors.New("cover_image_url must be at most 2048 characters")
	}

	parsedURL, err := url.ParseRequestURI(trimmedValue)
	if err != nil || parsedURL.Scheme != "https" || parsedURL.Host == "" {
		return sql.NullString{}, errors.New("cover_image_url must be a valid HTTPS URL")
	}

	return sql.NullString{String: trimmedValue, Valid: true}, nil
}

func newMarketStatus(now time.Time, opensAt sql.NullTime) string {
	if opensAt.Valid && opensAt.Time.After(now) {
		return "DRAFT"
	}

	return "OPEN"
}

func (request AttachMarketContractRequest) ToRepositoryInput() (repository.AttachMarketContractInput, error) {
	return deploymentFieldsToRepositoryInput(
		request.MarketContractAddress,
		request.MarketDeploymentTxHash,
		request.MarketFactoryAddress,
		request.ResolverAddress,
	)
}

func (request CreateMarketRequest) deploymentInput() (repository.AttachMarketContractInput, error) {
	return deploymentFieldsToRepositoryInput(
		request.MarketContractAddress,
		request.MarketDeploymentTxHash,
		request.MarketFactoryAddress,
		request.ResolverAddress,
	)
}

func deploymentFieldsToRepositoryInput(
	contractAddress string,
	deploymentTxHash string,
	factoryAddress string,
	resolverAddressValue string,
) (repository.AttachMarketContractInput, error) {
	marketContractAddress := strings.TrimSpace(contractAddress)
	marketDeploymentTxHash := strings.TrimSpace(deploymentTxHash)
	marketFactoryAddress := strings.TrimSpace(factoryAddress)
	resolverAddress := strings.TrimSpace(resolverAddressValue)

	if !validation.IsEVMAddressShape(marketContractAddress) {
		return repository.AttachMarketContractInput{}, errors.New("market_contract_address must be an EVM address")
	}
	if !validation.IsEVMTxHashShape(marketDeploymentTxHash) {
		return repository.AttachMarketContractInput{}, errors.New("market_deployment_tx_hash must be an EVM transaction hash")
	}
	if !validation.IsEVMAddressShape(marketFactoryAddress) {
		return repository.AttachMarketContractInput{}, errors.New("market_factory_address must be an EVM address")
	}
	if !validation.IsEVMAddressShape(resolverAddress) {
		return repository.AttachMarketContractInput{}, errors.New("resolver_address must be an EVM address")
	}

	return repository.AttachMarketContractInput{
		MarketContractAddress:  strings.ToLower(marketContractAddress),
		MarketDeploymentTxHash: strings.ToLower(marketDeploymentTxHash),
		MarketFactoryAddress:   strings.ToLower(marketFactoryAddress),
		ResolverAddress:        strings.ToLower(resolverAddress),
	}, nil
}
