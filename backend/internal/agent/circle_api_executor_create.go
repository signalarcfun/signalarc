package agent

import (
	"context"
	"strings"
)

func (executor *CircleAPIExecutor) ExecuteCreateMarket(ctx context.Context, intent Intent) (ExecutionResult, error) {
	if err := executor.validateIntent(intent, ActionCreateMarket); err != nil {
		return ExecutionResult{}, err
	}
	if err := validateCloseTimestampFresh(intent.CloseTimestamp); err != nil {
		return ExecutionResult{}, err
	}

	resolver := strings.TrimSpace(intent.Resolver)
	if resolver == "" {
		return ExecutionResult{}, ErrIntentInvalid
	}
	collateralToken := strings.TrimSpace(intent.CollateralToken)
	if collateralToken == "" {
		collateralToken = ArcTestnetUSDCAddress
	}

	txHash, err := executor.executeContract(ctx, intent, executor.cfg.AgentFactory, "createMarket(string,string,uint256,address,address)", []string{intent.MarketID, intent.Question, intent.CloseTimestamp, resolver, collateralToken})
	if err != nil {
		return ExecutionResult{}, err
	}

	result := baseResult(intent, ExecutionModeCircleDeveloperWalletAPI, executor.cfg.AgentFactory)
	result.TransactionHash = txHash
	result.Readback = executor.createMarketReadback(ctx, txHash)
	return result, nil
}
