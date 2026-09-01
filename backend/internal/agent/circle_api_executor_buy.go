package agent

import (
	"context"
	"strings"
)

func (executor *CircleAPIExecutor) ExecuteBuyYes(ctx context.Context, intent Intent) (ExecutionResult, error) {
	return executor.executeBuy(ctx, intent, ActionBuyYes, "buyYes(uint256)", "yesPositions(address)", "totalYes()")
}

func (executor *CircleAPIExecutor) ExecuteBuyNo(ctx context.Context, intent Intent) (ExecutionResult, error) {
	return executor.executeBuy(ctx, intent, ActionBuyNo, "buyNo(uint256)", "noPositions(address)", "totalNo()")
}

func (executor *CircleAPIExecutor) executeBuy(ctx context.Context, intent Intent, expectedAction string, buySignature string, positionSignature string, totalSignature string) (ExecutionResult, error) {
	if err := executor.validateIntent(intent, expectedAction); err != nil {
		return ExecutionResult{}, err
	}
	if strings.TrimSpace(intent.MarketContractAddress) == "" || strings.TrimSpace(intent.Amount) == "" {
		return ExecutionResult{}, ErrIntentInvalid
	}
	amountBaseUnits, err := usdcAmountToBaseUnits(intent.Amount)
	if err != nil {
		return ExecutionResult{}, err
	}

	approveHash, err := executor.executeContract(ctx, intent, ArcTestnetUSDCAddress, "approve(address,uint256)", []string{intent.MarketContractAddress, amountBaseUnits})
	if err != nil {
		return ExecutionResult{}, err
	}
	buyHash, err := executor.executeContract(ctx, intent, intent.MarketContractAddress, buySignature, []string{amountBaseUnits})
	if err != nil {
		return ExecutionResult{}, err
	}

	readback := executor.buyReadback(ctx, intent, expectedAction, positionSignature, totalSignature)
	result := baseResult(intent, ExecutionModeCircleDeveloperWalletAPI, executor.cfg.AgentFactory)
	result.MarketContractAddress = intent.MarketContractAddress
	result.ApproveTransactionHash = approveHash
	result.TransactionHash = buyHash
	result.Readback = readback
	return result, nil
}
