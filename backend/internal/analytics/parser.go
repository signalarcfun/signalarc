package analytics

import (
	"fmt"
	"strings"
	"time"
)

func ParseMarketDeployed(factoryAddress string, log BlockscoutLog) (MarketDeployed, bool, error) {
	if log.Decoded == nil {
		return MarketDeployed{}, false, nil
	}
	if !strings.HasPrefix(log.Decoded.MethodCall, "MarketDeployed(") {
		return MarketDeployed{}, false, nil
	}

	params := decodedParamMap(log.Decoded.Parameters)
	event := MarketDeployed{
		FactoryAddress:         factoryAddress,
		MarketIDHash:           params["marketId"],
		MarketAddress:          params["market"],
		CreatorAddress:         params["creator"],
		ResolverAddress:        params["resolver"],
		CollateralTokenAddress: params["collateralToken"],
		CloseTimestamp:         params["closeTimestamp"],
		Question:               params["question"],
		TransactionHash:        log.TransactionHash,
		BlockNumber:            log.BlockNumber,
		LogIndex:               log.Index,
		Raw:                    log.Raw,
	}

	if event.MarketAddress == "" {
		return MarketDeployed{}, true, fmt.Errorf("MarketDeployed log missing market parameter")
	}
	if event.TransactionHash == "" {
		return MarketDeployed{}, true, fmt.Errorf("MarketDeployed log missing transaction hash")
	}
	if log.BlockTimestamp != "" {
		timestamp, err := time.Parse(time.RFC3339Nano, log.BlockTimestamp)
		if err != nil {
			return MarketDeployed{}, true, fmt.Errorf("parse MarketDeployed block timestamp: %w", err)
		}
		event.BlockTimestamp = timestamp
	}

	return event, true, nil
}

func ParseMarketEvent(factoryAddress string, marketAddress string, log BlockscoutLog) (MarketEvent, bool, error) {
	if log.Decoded == nil {
		return MarketEvent{}, false, nil
	}

	timestamp, err := parseBlockTimestamp(log.BlockTimestamp)
	if err != nil {
		return MarketEvent{}, true, err
	}
	params := decodedParamMap(log.Decoded.Parameters)
	event := MarketEvent{
		FactoryAddress:  factoryAddress,
		MarketAddress:   marketAddress,
		TransactionHash: log.TransactionHash,
		BlockNumber:     log.BlockNumber,
		LogIndex:        log.Index,
		BlockTimestamp:  timestamp,
		Raw:             log.Raw,
	}

	switch {
	case strings.HasPrefix(log.Decoded.MethodCall, "PositionOpened("):
		side, err := normalizeOutcome(params["side"])
		if err != nil {
			return MarketEvent{}, true, fmt.Errorf("parse PositionOpened side: %w", err)
		}
		event.EventName = PositionOpenedEvent
		event.WalletAddress = params["user"]
		event.Side = side
		event.AmountBaseUnits = params["amount"]
	case strings.HasPrefix(log.Decoded.MethodCall, "MarketResolved("):
		outcome, err := normalizeOutcome(params["winningOutcome"])
		if err != nil {
			return MarketEvent{}, true, fmt.Errorf("parse MarketResolved winning outcome: %w", err)
		}
		event.EventName = MarketResolvedEvent
		event.WinningOutcome = outcome
		event.Status = "RESOLVED"
	case strings.HasPrefix(log.Decoded.MethodCall, "MarketCancelled("):
		event.EventName = MarketCancelledEvent
		event.Status = "CANCELLED"
	case strings.HasPrefix(log.Decoded.MethodCall, "PayoutClaimed("):
		event.EventName = PayoutClaimedEvent
		event.WalletAddress = params["user"]
		event.AmountBaseUnits = params["amount"]
	case strings.HasPrefix(log.Decoded.MethodCall, "RefundClaimed("):
		event.EventName = RefundClaimedEvent
		event.WalletAddress = params["user"]
		event.AmountBaseUnits = params["amount"]
	default:
		return MarketEvent{}, false, nil
	}

	if event.TransactionHash == "" {
		return MarketEvent{}, true, fmt.Errorf("%s log missing transaction hash", event.EventName)
	}
	if event.AmountBaseUnits == "" {
		event.AmountBaseUnits = "0"
	}

	return event, true, nil
}

func decodedParamMap(params []BlockscoutParameter) map[string]string {
	mapped := make(map[string]string, len(params))
	for _, param := range params {
		mapped[param.Name] = param.Value
	}
	return mapped
}

func parseBlockTimestamp(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	timestamp, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse block timestamp: %w", err)
	}
	return timestamp, nil
}

func normalizeOutcome(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "yes":
		return "YES", nil
	case "2", "no":
		return "NO", nil
	default:
		return "", fmt.Errorf("unknown outcome %q", value)
	}
}
