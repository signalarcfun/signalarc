package trade

import (
	"errors"
	"math/big"
	"strings"

	"github.com/wahyu241205/SignalArc/backend/internal/repository"
	"github.com/wahyu241205/SignalArc/backend/internal/validation"
)

func (request CreateTradeIntentRequest) ToRepositoryInput() (repository.CreateTradeIntentInput, error) {
	userID := strings.TrimSpace(request.UserID)
	marketID := strings.TrimSpace(request.MarketID)
	outcome := strings.TrimSpace(request.Outcome)
	side := strings.TrimSpace(request.Side)
	quantityValue := strings.TrimSpace(request.Quantity)
	priceValue := strings.TrimSpace(request.Price)

	if userID == "" || !validation.IsUUIDShape(userID) {
		return repository.CreateTradeIntentInput{}, errors.New("user_id is required")
	}
	if marketID == "" || !validation.IsUUIDShape(marketID) {
		return repository.CreateTradeIntentInput{}, errors.New("market_id is required")
	}
	if outcome != "YES" && outcome != "NO" {
		return repository.CreateTradeIntentInput{}, errors.New("outcome must be YES or NO")
	}
	if side != "BUY" && side != "SELL" {
		return repository.CreateTradeIntentInput{}, errors.New("side must be BUY or SELL")
	}

	quantity, ok := validation.ParseDecimal(quantityValue)
	if !ok || quantity.Sign() <= 0 {
		return repository.CreateTradeIntentInput{}, errors.New("quantity must be greater than zero")
	}

	price, ok := validation.ParseDecimal(priceValue)
	if !ok || price.Sign() < 0 || price.Cmp(big.NewRat(1, 1)) > 0 {
		return repository.CreateTradeIntentInput{}, errors.New("price must be between zero and one")
	}

	collateral, ok := validation.DecimalString(new(big.Rat).Mul(quantity, price))
	if !ok {
		return repository.CreateTradeIntentInput{}, errors.New("decimal requires more than 18 fractional digits")
	}

	return repository.CreateTradeIntentInput{
		UserID:           userID,
		MarketID:         marketID,
		Outcome:          outcome,
		Side:             side,
		Quantity:         quantityValue,
		Price:            priceValue,
		CollateralAmount: collateral,
	}, nil
}
