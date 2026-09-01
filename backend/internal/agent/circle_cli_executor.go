package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os/exec"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type CommandRunner interface {
	Run(context.Context, string, []string) ([]byte, error)
}

type ExecCommandRunner struct{}

func (runner ExecCommandRunner) Run(ctx context.Context, name string, args []string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	// CombinedOutput captures both stdout and stderr so AUTH_REQUIRED markers
	// emitted on stderr can be classified by the caller. The classified error
	// surfaced to API callers stays generic; raw output is only used to set
	// a sanitized error class on the wrapped CircleCLIError.
	output, err := command.CombinedOutput()
	if err != nil {
		rawOutput := string(output)
		errorClass := ClassifyCircleErrorOutput(rawOutput, err.Error())
		sanitizedSummary := BuildExecuteSummary(rawOutput, err)
		exitStatus := ExtractExitStatus(err.Error())
		commandCategory := ClassifyExecuteCommandCategory(args)

		// Capture the sanitized process error when CLI output is empty,
		// so operators can distinguish process-level failures.
		processError := ""
		if strings.TrimSpace(rawOutput) == "" {
			processError = SanitizeExecuteOutput(err.Error())
		}

		return output, &CircleCLIError{
			Operation:        "circle_cli_exec",
			ErrorClass:       errorClass,
			SanitizedSummary: sanitizedSummary,
			Err:              errors.New("Circle CLI command failed"),
			ExecCtx: &ExecuteContext{
				CommandCategory: commandCategory,
				ExitStatus:      exitStatus,
				Chain:           extractChainFromArgs(args),
				RawOutputLen:    len(output),
				ProcessError:    processError,
			},
		}
	}
	return output, nil
}

type CircleCLIExecutorConfig struct {
	Enabled       bool
	CLIPath       string
	Chain         string
	Timeout       time.Duration
	AgentFactory  string
	CommandRunner CommandRunner
}

type CircleCLIExecutor struct {
	cfg CircleCLIExecutorConfig
}

func NewCircleCLIExecutor(cfg CircleCLIExecutorConfig) *CircleCLIExecutor {
	if strings.TrimSpace(cfg.CLIPath) == "" {
		cfg.CLIPath = "circle"
	}
	if strings.TrimSpace(cfg.Chain) == "" {
		cfg.Chain = ChainArcTestnet
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 120 * time.Second
	}
	if strings.TrimSpace(cfg.AgentFactory) == "" {
		cfg.AgentFactory = AgentFactoryAddress
	}
	if cfg.CommandRunner == nil {
		cfg.CommandRunner = ExecCommandRunner{}
	}
	return &CircleCLIExecutor{cfg: cfg}
}

func (executor *CircleCLIExecutor) ExecuteCreateMarket(ctx context.Context, intent Intent) (ExecutionResult, error) {
	if err := executor.validateIntent(intent, ActionCreateMarket); err != nil {
		return ExecutionResult{}, err
	}

	if err := validateCloseTimestampFresh(intent.CloseTimestamp); err != nil {
		return ExecutionResult{}, err
	}

	resolver := intent.AgentWalletAddress
	collateralToken := strings.TrimSpace(intent.CollateralToken)
	if collateralToken == "" {
		collateralToken = ArcTestnetUSDCAddress
	}

	txHash, err := executor.walletExecute(ctx,
		"createMarket(string,string,uint256,address,address)",
		[]string{intent.MarketID, intent.Question, intent.CloseTimestamp, resolver, collateralToken},
		intent.AgentWalletAddress,
		executor.cfg.AgentFactory,
	)
	if err != nil {
		return ExecutionResult{}, err
	}

	marketCount, err := executor.contractQueryUint256(ctx, "marketCount()", nil, executor.cfg.AgentFactory)
	if err != nil {
		return ExecutionResult{}, err
	}
	lastIndex, err := previousIndex(marketCount)
	if err != nil {
		return ExecutionResult{}, err
	}
	createdMarket, err := executor.contractQueryAddress(ctx, "allMarkets(uint256)", []string{lastIndex}, executor.cfg.AgentFactory)
	if err != nil {
		return ExecutionResult{}, err
	}
	isMarket, err := executor.contractQueryBool(ctx, "isMarket(address)", []string{createdMarket}, executor.cfg.AgentFactory)
	if err != nil {
		return ExecutionResult{}, err
	}

	return ExecutionResult{
		IntentID:            intent.ID,
		AgentID:             intent.AgentID,
		AgentWalletAddress:  intent.AgentWalletAddress,
		WalletProvider:      intent.WalletProvider,
		Action:              intent.Action,
		Status:              StatusExecuted,
		ExecutionMode:       ExecutionModeCircleAgentWalletCLI,
		Network:             NetworkArcTestnet,
		AgentFactoryAddress: executor.cfg.AgentFactory,
		BroadcastPerformed:  true,
		TransactionHash:     txHash,
		Readback: ExecutionReadback{
			MarketCount:   marketCount,
			CreatedMarket: createdMarket,
			IsMarket:      &isMarket,
		},
	}, nil
}

func (executor *CircleCLIExecutor) ExecuteBuyYes(ctx context.Context, intent Intent) (ExecutionResult, error) {
	return executor.executeBuy(ctx, intent, ActionBuyYes, "buyYes(uint256)", "yesPositions(address)", "totalYes()")
}

func (executor *CircleCLIExecutor) ExecuteBuyNo(ctx context.Context, intent Intent) (ExecutionResult, error) {
	return executor.executeBuy(ctx, intent, ActionBuyNo, "buyNo(uint256)", "noPositions(address)", "totalNo()")
}

func (executor *CircleCLIExecutor) ExecuteCloseMarket(ctx context.Context, intent Intent) (ExecutionResult, error) {
	return executor.executeLifecycle(ctx, intent, lifecycleActionSpec{
		action:    ActionCloseMarket,
		signature: "closeMarket()",
		readback:  readbackMarketState,
	})
}

func (executor *CircleCLIExecutor) ExecuteResolveMarket(ctx context.Context, intent Intent) (ExecutionResult, error) {
	if err := executor.validateIntent(intent, ActionResolveMarket); err != nil {
		return ExecutionResult{}, err
	}
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

func (executor *CircleCLIExecutor) ExecuteClaimPayout(ctx context.Context, intent Intent) (ExecutionResult, error) {
	return executor.executeLifecycle(ctx, intent, lifecycleActionSpec{
		action:    ActionClaimPayout,
		signature: "claimPayout()",
		readback:  readbackPayout,
	})
}

func (executor *CircleCLIExecutor) ExecuteCancelMarket(ctx context.Context, intent Intent) (ExecutionResult, error) {
	return executor.executeLifecycle(ctx, intent, lifecycleActionSpec{
		action:    ActionCancelMarket,
		signature: "cancelMarket()",
		readback:  readbackRefund,
	})
}

func (executor *CircleCLIExecutor) ExecuteClaimRefund(ctx context.Context, intent Intent) (ExecutionResult, error) {
	return executor.executeLifecycle(ctx, intent, lifecycleActionSpec{
		action:    ActionClaimRefund,
		signature: "claimRefund()",
		readback:  readbackRefund,
	})
}

func (executor *CircleCLIExecutor) executeBuy(ctx context.Context, intent Intent, expectedAction string, buySignature string, positionSignature string, totalSignature string) (ExecutionResult, error) {
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

	approveHash, err := executor.walletExecute(ctx,
		"approve(address,uint256)",
		[]string{intent.MarketContractAddress, amountBaseUnits},
		intent.AgentWalletAddress,
		ArcTestnetUSDCAddress,
	)
	if err != nil {
		return ExecutionResult{}, err
	}

	buyHash, err := executor.walletExecute(ctx,
		buySignature,
		[]string{amountBaseUnits},
		intent.AgentWalletAddress,
		intent.MarketContractAddress,
	)
	if err != nil {
		return ExecutionResult{}, err
	}

	positionValue, err := executor.contractQueryUint256(ctx, positionSignature, []string{intent.AgentWalletAddress}, intent.MarketContractAddress)
	if err != nil {
		return ExecutionResult{}, err
	}
	totalValue, err := executor.contractQueryUint256(ctx, totalSignature, nil, intent.MarketContractAddress)
	if err != nil {
		return ExecutionResult{}, err
	}
	totalCollateral, err := executor.contractQueryUint256(ctx, "totalCollateral()", nil, intent.MarketContractAddress)
	if err != nil {
		return ExecutionResult{}, err
	}
	usdcBalance, err := executor.contractQueryUint256(ctx, "balanceOf(address)", []string{intent.MarketContractAddress}, ArcTestnetUSDCAddress)
	if err != nil {
		return ExecutionResult{}, err
	}

	readback := ExecutionReadback{
		TotalCollateral: totalCollateral,
		USDCBalance:     usdcBalance,
	}
	if expectedAction == ActionBuyYes {
		readback.YesPositions = positionValue
		readback.TotalYes = totalValue
	} else {
		readback.NoPositions = positionValue
		readback.TotalNo = totalValue
	}

	return ExecutionResult{
		IntentID:               intent.ID,
		AgentID:                intent.AgentID,
		AgentWalletAddress:     intent.AgentWalletAddress,
		WalletProvider:         intent.WalletProvider,
		Action:                 intent.Action,
		Status:                 StatusExecuted,
		ExecutionMode:          ExecutionModeCircleAgentWalletCLI,
		Network:                NetworkArcTestnet,
		AgentFactoryAddress:    executor.cfg.AgentFactory,
		MarketContractAddress:  intent.MarketContractAddress,
		BroadcastPerformed:     true,
		ApproveTransactionHash: approveHash,
		TransactionHash:        buyHash,
		Readback:               readback,
	}, nil
}

type lifecycleReadbackKind string

const (
	readbackMarketState lifecycleReadbackKind = "market_state"
	readbackResolution  lifecycleReadbackKind = "resolution"
	readbackPayout      lifecycleReadbackKind = "payout"
	readbackRefund      lifecycleReadbackKind = "refund"
)

type lifecycleActionSpec struct {
	action    string
	signature string
	params    []string
	readback  lifecycleReadbackKind
}

func (executor *CircleCLIExecutor) executeLifecycle(ctx context.Context, intent Intent, spec lifecycleActionSpec) (ExecutionResult, error) {
	if err := executor.validateIntent(intent, spec.action); err != nil {
		return ExecutionResult{}, err
	}
	if strings.TrimSpace(intent.MarketContractAddress) == "" {
		return ExecutionResult{}, ErrIntentInvalid
	}

	txHash, err := executor.walletExecute(ctx,
		spec.signature,
		spec.params,
		intent.AgentWalletAddress,
		intent.MarketContractAddress,
	)
	if err != nil {
		return ExecutionResult{}, err
	}

	readback, err := executor.lifecycleReadback(ctx, intent, spec.readback)
	if err != nil {
		return ExecutionResult{}, err
	}

	return ExecutionResult{
		IntentID:              intent.ID,
		AgentID:               intent.AgentID,
		AgentWalletAddress:    intent.AgentWalletAddress,
		WalletProvider:        intent.WalletProvider,
		Action:                intent.Action,
		Status:                StatusExecuted,
		ExecutionMode:         ExecutionModeCircleAgentWalletCLI,
		Network:               NetworkArcTestnet,
		AgentFactoryAddress:   executor.cfg.AgentFactory,
		MarketContractAddress: intent.MarketContractAddress,
		BroadcastPerformed:    true,
		TransactionHash:       txHash,
		Readback:              readback,
	}, nil
}

func (executor *CircleCLIExecutor) lifecycleReadback(ctx context.Context, intent Intent, kind lifecycleReadbackKind) (ExecutionReadback, error) {
	readback := ExecutionReadback{}

	status, err := executor.contractQueryUint256(ctx, "status()", nil, intent.MarketContractAddress)
	if err != nil {
		return ExecutionReadback{}, err
	}
	readback.MarketStatus = status

	switch kind {
	case readbackMarketState:
		isOpen, err := executor.contractQueryBool(ctx, "isOpen()", nil, intent.MarketContractAddress)
		if err != nil {
			return ExecutionReadback{}, err
		}
		readback.IsOpen = &isOpen
	case readbackResolution, readbackPayout:
		winningOutcome, err := executor.contractQueryUint256(ctx, "winningOutcome()", nil, intent.MarketContractAddress)
		if err != nil {
			return ExecutionReadback{}, err
		}
		claimablePayout, err := executor.contractQueryUint256(ctx, "claimablePayout(address)", []string{intent.AgentWalletAddress}, intent.MarketContractAddress)
		if err != nil {
			return ExecutionReadback{}, err
		}
		hasClaimed, err := executor.contractQueryBool(ctx, "hasClaimed(address)", []string{intent.AgentWalletAddress}, intent.MarketContractAddress)
		if err != nil {
			return ExecutionReadback{}, err
		}
		usdcBalance, err := executor.contractQueryUint256(ctx, "balanceOf(address)", []string{intent.MarketContractAddress}, ArcTestnetUSDCAddress)
		if err != nil {
			return ExecutionReadback{}, err
		}
		readback.WinningOutcome = winningOutcome
		readback.ClaimablePayout = claimablePayout
		readback.HasClaimed = &hasClaimed
		readback.USDCBalance = usdcBalance
	case readbackRefund:
		claimableRefund, err := executor.contractQueryUint256(ctx, "claimableRefund(address)", []string{intent.AgentWalletAddress}, intent.MarketContractAddress)
		if err != nil {
			return ExecutionReadback{}, err
		}
		hasClaimed, err := executor.contractQueryBool(ctx, "hasClaimed(address)", []string{intent.AgentWalletAddress}, intent.MarketContractAddress)
		if err != nil {
			return ExecutionReadback{}, err
		}
		usdcBalance, err := executor.contractQueryUint256(ctx, "balanceOf(address)", []string{intent.MarketContractAddress}, ArcTestnetUSDCAddress)
		if err != nil {
			return ExecutionReadback{}, err
		}
		readback.ClaimableRefund = claimableRefund
		readback.HasClaimed = &hasClaimed
		readback.USDCBalance = usdcBalance
	default:
		return ExecutionReadback{}, ErrExecutionNotImplemented
	}

	return readback, nil
}

func (executor *CircleCLIExecutor) validateIntent(intent Intent, expectedAction string) error {
	if executor == nil {
		return ErrExecutionConfigInvalid
	}
	if !executor.cfg.Enabled {
		return ErrExecutionProviderDisabled
	}
	if intent.Status != StatusConfirmed {
		return ErrIntentNotConfirmed
	}
	if !intent.ValidationResult.Valid {
		return ErrIntentInvalid
	}
	if intent.Action != expectedAction {
		return ErrExecutionNotImplemented
	}
	if intent.WalletProvider != WalletProviderCircleAgentWallet {
		return ErrExecutionConfigInvalid
	}
	if !AgentWalletAllowsAction(AgentWallet{AllowedActions: intent.AllowedActions}, expectedAction) {
		return ErrIntentInvalid
	}
	if intent.AgentWalletAddress == "" {
		return ErrIntentInvalid
	}
	if executor.cfg.Chain != ChainArcTestnet {
		return ErrExecutionConfigInvalid
	}
	return nil
}

func normalizeOutcomeParam(outcome string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(outcome)) {
	case "1", "yes":
		return "1", nil
	case "2", "no":
		return "2", nil
	default:
		return "", ErrIntentInvalid
	}
}

const usdcDecimals = 6

func usdcAmountToBaseUnits(amount string) (string, error) {
	trimmed := strings.TrimSpace(amount)
	if trimmed == "" {
		return "", ErrIntentInvalid
	}

	parts := strings.Split(trimmed, ".")
	if len(parts) > 2 || parts[0] == "" {
		return "", ErrIntentInvalid
	}

	whole := parts[0]
	if !isDecimalDigits(whole) {
		return "", ErrIntentInvalid
	}

	fractional := ""
	if len(parts) == 2 {
		fractional = parts[1]
		if fractional == "" || len(fractional) > usdcDecimals || !isDecimalDigits(fractional) {
			return "", ErrIntentInvalid
		}
	}

	fractional += strings.Repeat("0", usdcDecimals-len(fractional))
	baseUnits := strings.TrimLeft(whole+fractional, "0")
	if baseUnits == "" {
		return "", ErrIntentInvalid
	}

	return baseUnits, nil
}

func isDecimalDigits(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func (executor *CircleCLIExecutor) walletExecute(ctx context.Context, signature string, params []string, walletAddress string, contractAddress string) (string, error) {
	args := []string{"wallet", "execute", signature}
	args = append(args, params...)
	args = append(args,
		"--address", walletAddress,
		"--contract", contractAddress,
		"--chain", executor.cfg.Chain,
		"--output", "json",
	)

	output, err := executor.run(ctx, args)
	if err != nil {
		// Enrich the CircleCLIError with execution context for diagnostics.
		var cliErr *CircleCLIError
		if errors.As(err, &cliErr) {
			if cliErr.ExecCtx == nil {
				cliErr.ExecCtx = &ExecuteContext{}
			}
			cliErr.ExecCtx.FunctionSignature = signature
			cliErr.ExecCtx.ContractAddress = RedactAddress(contractAddress)
			cliErr.ExecCtx.WalletAddress = RedactAddress(walletAddress)
			cliErr.ExecCtx.Chain = executor.cfg.Chain
			cliErr.ExecCtx.CommandCategory = "wallet_execute"
		}
		return "", err
	}
	hash, ok := findJSONValue(output, "transactionHash", "transaction_hash", "txHash", "tx_hash", "hash")
	if !ok || !looksLikeHash(hash) {
		return "", errors.New("Circle CLI transaction hash not found in JSON output")
	}
	return hash, nil
}

func (executor *CircleCLIExecutor) contractQuery(ctx context.Context, signature string, params []string, contractAddress string) (string, error) {
	args := []string{"contract", "query", signature}
	args = append(args, params...)
	args = append(args,
		"--contract", contractAddress,
		"--chain", executor.cfg.Chain,
		"--output", "json",
	)

	output, err := executor.run(ctx, args)
	if err != nil {
		// Enrich the CircleCLIError with query context for diagnostics.
		var cliErr *CircleCLIError
		if errors.As(err, &cliErr) {
			if cliErr.ExecCtx == nil {
				cliErr.ExecCtx = &ExecuteContext{}
			}
			cliErr.ExecCtx.FunctionSignature = signature
			cliErr.ExecCtx.ContractAddress = RedactAddress(contractAddress)
			cliErr.ExecCtx.Chain = executor.cfg.Chain
			cliErr.ExecCtx.CommandCategory = "contract_query"
		}
		return "", err
	}
	value, ok := findJSONValue(output, "result", "value", "output", "data", "returnValue")
	if !ok {
		return "", errors.New("Circle CLI query result not found in JSON output")
	}
	return value, nil
}

func (executor *CircleCLIExecutor) contractQueryUint256(ctx context.Context, signature string, params []string, contractAddress string) (string, error) {
	value, err := executor.contractQuery(ctx, signature, params, contractAddress)
	if err != nil {
		return "", err
	}
	return decodeUint256Scalar(value)
}

func (executor *CircleCLIExecutor) contractQueryAddress(ctx context.Context, signature string, params []string, contractAddress string) (string, error) {
	value, err := executor.contractQuery(ctx, signature, params, contractAddress)
	if err != nil {
		return "", err
	}
	return decodeAddressScalar(value)
}

func (executor *CircleCLIExecutor) contractQueryBool(ctx context.Context, signature string, params []string, contractAddress string) (bool, error) {
	value, err := executor.contractQuery(ctx, signature, params, contractAddress)
	if err != nil {
		return false, err
	}
	return decodeBoolScalar(value)
}

func (executor *CircleCLIExecutor) run(parent context.Context, args []string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, executor.cfg.Timeout)
	defer cancel()

	output, err := executor.cfg.CommandRunner.Run(ctx, executor.cfg.CLIPath, args)
	if err != nil {
		// Preserve the classified *CircleCLIError when the underlying
		// runner already produced one; otherwise wrap the generic error
		// with an unknown error class so callers can errors.As() it for
		// structured logging without leaking raw CLI output.
		var cliErr *CircleCLIError
		if errors.As(err, &cliErr) {
			return nil, cliErr
		}
		return nil, &CircleCLIError{
			Operation:        "circle_cli_exec",
			ErrorClass:       CircleErrorClassUnknown,
			SanitizedSummary: BuildExecuteSummary("", err),
			Err:              errors.New("Circle CLI command failed"),
			ExecCtx: &ExecuteContext{
				CommandCategory: ClassifyExecuteCommandCategory(args),
				ExitStatus:      ExtractExitStatus(err.Error()),
				Chain:           extractChainFromArgs(args),
				RawOutputLen:    0,
				ProcessError:    SanitizeExecuteOutput(err.Error()),
			},
		}
	}
	if len(output) == 0 {
		return nil, &CircleCLIError{
			Operation:        "circle_cli_exec",
			ErrorClass:       CircleErrorClassUnknown,
			SanitizedSummary: "Circle CLI command returned empty output",
			Err:              errors.New("Circle CLI command returned empty output"),
			ExecCtx: &ExecuteContext{
				CommandCategory: ClassifyExecuteCommandCategory(args),
				Chain:           extractChainFromArgs(args),
				RawOutputLen:    0,
			},
		}
	}
	return output, nil
}

func findJSONValue(output []byte, keys ...string) (string, bool) {
	var decoded any
	if err := json.Unmarshal(output, &decoded); err != nil {
		return "", false
	}
	return findValue(decoded, keys)
}

func findValue(value any, keys []string) (string, bool) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range keys {
			if child, ok := typed[key]; ok {
				if text, ok := scalarToString(child); ok {
					return text, true
				}
				if text, ok := findValue(child, keys); ok {
					return text, true
				}
			}
		}
		for _, child := range typed {
			if text, ok := findValue(child, keys); ok {
				return text, true
			}
		}
	case []any:
		if len(typed) == 1 {
			return scalarToString(typed[0])
		}
		for _, child := range typed {
			if text, ok := findValue(child, keys); ok {
				return text, true
			}
		}
	default:
		return scalarToString(typed)
	}
	return "", false
}

func scalarToString(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed), strings.TrimSpace(typed) != ""
	case float64:
		if typed == float64(int64(typed)) {
			return fmt.Sprintf("%.0f", typed), true
		}
		return fmt.Sprintf("%v", typed), true
	case bool:
		if typed {
			return "true", true
		}
		return "false", true
	default:
		return "", false
	}
}

func previousIndex(value string) (string, error) {
	decoded, err := decodeUint256Scalar(value)
	if err != nil {
		return "", err
	}
	count, ok := new(big.Int).SetString(decoded, 10)
	if !ok || count.Sign() <= 0 {
		return "", errors.New("invalid marketCount readback")
	}
	return new(big.Int).Sub(count, big.NewInt(1)).String(), nil
}

func decodeUint256Scalar(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("empty uint256 readback")
	}
	if strings.HasPrefix(value, "0x") || strings.HasPrefix(value, "0X") {
		integer, ok := new(big.Int).SetString(strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X"), 16)
		if !ok {
			return "", errors.New("invalid hex uint256 readback")
		}
		return integer.String(), nil
	}
	integer, ok := new(big.Int).SetString(value, 10)
	if !ok || integer.Sign() < 0 {
		return "", errors.New("invalid decimal uint256 readback")
	}
	return integer.String(), nil
}

func decodeAddressScalar(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("empty address readback")
	}
	if strings.HasPrefix(value, "0x") || strings.HasPrefix(value, "0X") {
		hexValue := strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X")
		if len(hexValue) == 64 {
			hexValue = hexValue[24:]
		}
		if len(hexValue) != 40 {
			return "", errors.New("invalid address readback")
		}
		return common.HexToAddress("0x" + hexValue).Hex(), nil
	}
	return "", errors.New("invalid address readback")
}

func decodeBoolScalar(value string) (bool, error) {
	value = strings.TrimSpace(value)
	if strings.EqualFold(value, "true") {
		return true, nil
	}
	if strings.EqualFold(value, "false") {
		return false, nil
	}
	decoded, err := decodeUint256Scalar(value)
	if err != nil {
		return false, err
	}
	switch decoded {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, errors.New("invalid bool readback")
	}
}

func looksLikeHash(value string) bool {
	value = strings.TrimSpace(value)
	return strings.HasPrefix(value, "0x") && len(value) >= 10
}

// closeTimestampSafetyMarginSeconds is the minimum number of seconds that
// closeTimestamp must be ahead of the current time at execution. This
// prevents submitting transactions that will certainly revert because the
// contract constructor enforces closeTimestamp > block.timestamp.
const closeTimestampSafetyMarginSeconds = 60

// validateCloseTimestampFresh checks that the close_timestamp value on a
// create_market intent is still sufficiently in the future at execution time.
// Returns ErrCreateMarketCloseTimestampStale if the timestamp is stale.
func validateCloseTimestampFresh(closeTimestamp string) error {
	closeTimestamp = strings.TrimSpace(closeTimestamp)
	if closeTimestamp == "" {
		return ErrIntentInvalid
	}
	ts, ok := new(big.Int).SetString(closeTimestamp, 10)
	if !ok || ts.Sign() <= 0 {
		return ErrIntentInvalid
	}
	nowPlusMargin := big.NewInt(time.Now().Unix() + closeTimestampSafetyMarginSeconds)
	if ts.Cmp(nowPlusMargin) <= 0 {
		return ErrCreateMarketCloseTimestampStale
	}
	return nil
}

// extractChainFromArgs extracts the --chain value from CLI args, if present.
func extractChainFromArgs(args []string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--chain" {
			return args[i+1]
		}
	}
	return ""
}
