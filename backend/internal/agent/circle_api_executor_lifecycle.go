package agent

import (
	"context"
	"strings"
)

func (executor *CircleAPIExecutor) ExecuteCloseMarket(ctx context.Context, intent Intent) (ExecutionResult, error) {
	return executor.executeLifecycle(ctx, intent, lifecycleActionSpec{
		action:    ActionCloseMarket,
		signature: "closeMarket()",
		readback:  readbackMarketState,
	})
}

func (executor *CircleAPIExecutor) ExecuteResolveMarket(ctx context.Context, intent Intent) (ExecutionResult, error) {
	outcome, err := normalizeOutcomeParam(intent.Outcome)
	if err != nil {
		return ExecutionResult{}, err
	}
	return executor.executeLifecycle(ctx, intent, lifecycleActionSpec{
		action:    ActionResolveMarket,
		signature: "resolve(uint8)",
		params:    []string{outcome},
		readback:  readbackResolution,
	})
}

func (executor *CircleAPIExecutor) ExecuteClaimPayout(ctx context.Context, intent Intent) (ExecutionResult, error) {
	return executor.executeLifecycle(ctx, intent, lifecycleActionSpec{
		action:    ActionClaimPayout,
		signature: "claimPayout()",
		readback:  readbackPayout,
	})
}

func (executor *CircleAPIExecutor) ExecuteCancelMarket(ctx context.Context, intent Intent) (ExecutionResult, error) {
	return executor.executeLifecycle(ctx, intent, lifecycleActionSpec{
		action:    ActionCancelMarket,
		signature: "cancelMarket()",
		readback:  readbackRefund,
	})
}

func (executor *CircleAPIExecutor) ExecuteClaimRefund(ctx context.Context, intent Intent) (ExecutionResult, error) {
	return executor.executeLifecycle(ctx, intent, lifecycleActionSpec{
		action:    ActionClaimRefund,
		signature: "claimRefund()",
		readback:  readbackRefund,
	})
}

func (executor *CircleAPIExecutor) executeLifecycle(ctx context.Context, intent Intent, spec lifecycleActionSpec) (ExecutionResult, error) {
	if err := executor.validateIntent(intent, spec.action); err != nil {
		return ExecutionResult{}, err
	}
	if strings.TrimSpace(intent.MarketContractAddress) == "" {
		return ExecutionResult{}, ErrIntentInvalid
	}
	txHash, err := executor.executeContract(ctx, intent, intent.MarketContractAddress, spec.signature, spec.params)
	if err != nil {
		return ExecutionResult{}, err
	}

	result := baseResult(intent, ExecutionModeCircleDeveloperWalletAPI, executor.cfg.AgentFactory)
	result.MarketContractAddress = intent.MarketContractAddress
	result.TransactionHash = txHash
	result.Readback = executor.lifecycleReadbackRPC(ctx, intent, spec.readback)
	return result, nil
}
