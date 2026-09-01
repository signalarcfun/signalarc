package agent

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const agentFactoryABIJSON = `[
	{"type":"function","name":"createMarket","inputs":[{"name":"marketId","type":"string"},{"name":"question","type":"string"},{"name":"closeTimestamp","type":"uint256"},{"name":"resolver","type":"address"},{"name":"collateralToken","type":"address"}],"outputs":[{"name":"market","type":"address"}],"stateMutability":"nonpayable"},
	{"type":"function","name":"marketCount","inputs":[],"outputs":[{"name":"","type":"uint256"}],"stateMutability":"view"},
	{"type":"function","name":"isMarket","inputs":[{"name":"","type":"address"}],"outputs":[{"name":"","type":"bool"}],"stateMutability":"view"},
	{"type":"event","name":"AgentMarketDeployed","inputs":[{"name":"marketId","type":"string","indexed":true},{"name":"market","type":"address","indexed":true},{"name":"admin","type":"address","indexed":true},{"name":"resolver","type":"address","indexed":false},{"name":"collateralToken","type":"address","indexed":false},{"name":"closeTimestamp","type":"uint256","indexed":false},{"name":"question","type":"string","indexed":false}],"anonymous":false}
]`

const agentMarketABIJSON = `[
	{"type":"function","name":"buyYes","inputs":[{"name":"amount","type":"uint256"}],"outputs":[],"stateMutability":"nonpayable"},
	{"type":"function","name":"buyNo","inputs":[{"name":"amount","type":"uint256"}],"outputs":[],"stateMutability":"nonpayable"},
	{"type":"function","name":"yesPositions","inputs":[{"name":"user","type":"address"}],"outputs":[{"name":"","type":"uint256"}],"stateMutability":"view"},
	{"type":"function","name":"noPositions","inputs":[{"name":"user","type":"address"}],"outputs":[{"name":"","type":"uint256"}],"stateMutability":"view"},
	{"type":"function","name":"totalYes","inputs":[],"outputs":[{"name":"","type":"uint256"}],"stateMutability":"view"},
	{"type":"function","name":"totalNo","inputs":[],"outputs":[{"name":"","type":"uint256"}],"stateMutability":"view"},
	{"type":"function","name":"totalCollateral","inputs":[],"outputs":[{"name":"","type":"uint256"}],"stateMutability":"view"}
]`

const erc20ABIJSON = `[
	{"type":"function","name":"approve","inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable"},
	{"type":"function","name":"balanceOf","inputs":[{"name":"account","type":"address"}],"outputs":[{"name":"","type":"uint256"}],"stateMutability":"view"},
	{"type":"function","name":"allowance","inputs":[{"name":"owner","type":"address"},{"name":"spender","type":"address"}],"outputs":[{"name":"","type":"uint256"}],"stateMutability":"view"}
]`

const ArcTestnetUSDCAddress = "0x3600000000000000000000000000000000000000"

var (
	ErrExecutionNotImplemented          = errors.New("agent execution action is not implemented")
	ErrIntentNotConfirmed               = errors.New("agent intent is not confirmed")
	ErrExecutionConfigInvalid           = errors.New("agent execution config is invalid")
	ErrExecutionProviderDisabled        = errors.New("agent execution provider is disabled")
	ErrCreateMarketCloseTimestampStale  = errors.New("close_timestamp must be in the future before execution")
)

type Executor interface {
	ExecuteCreateMarket(context.Context, Intent) (ExecutionResult, error)
	ExecuteBuyYes(context.Context, Intent) (ExecutionResult, error)
	ExecuteBuyNo(context.Context, Intent) (ExecutionResult, error)
	ExecuteCloseMarket(context.Context, Intent) (ExecutionResult, error)
	ExecuteResolveMarket(context.Context, Intent) (ExecutionResult, error)
	ExecuteClaimPayout(context.Context, Intent) (ExecutionResult, error)
	ExecuteCancelMarket(context.Context, Intent) (ExecutionResult, error)
	ExecuteClaimRefund(context.Context, Intent) (ExecutionResult, error)
}

type ExecutionResult struct {
	IntentID               string
	AgentID                string
	AgentWalletAddress     string
	WalletProvider         string
	Action                 string
	Status                 string
	ExecutionMode          string
	Network                string
	AgentFactoryAddress    string
	MarketContractAddress  string
	BroadcastPerformed     bool
	ApproveTransactionHash string
	TransactionHash        string
	Readback               ExecutionReadback
}

type ExecutionReadback struct {
	MarketCount     string
	CreatedMarket   string
	IsMarket        *bool
	MarketStatus    string
	WinningOutcome  string
	YesPositions    string
	NoPositions     string
	TotalYes        string
	TotalNo         string
	TotalCollateral string
	ClaimablePayout string
	ClaimableRefund string
	HasClaimed      *bool
	IsOpen          *bool
	USDCBalance     string
	USDCAllowance   string
}

type ArcExecutorConfig struct {
	RPCURL             string
	ExecutorPrivateKey string
	AgentFactory       string
}

type ArcExecutor struct {
	cfg ArcExecutorConfig
}

func NewArcExecutorFromEnv() (*ArcExecutor, error) {
	cfg := ArcExecutorConfig{
		RPCURL:             strings.TrimSpace(os.Getenv("ARC_TESTNET_RPC_URL")),
		ExecutorPrivateKey: strings.TrimSpace(os.Getenv("AGENT_EXECUTOR_PRIVATE_KEY")),
		AgentFactory:       strings.TrimSpace(os.Getenv("AGENT_FACTORY_ADDRESS")),
	}
	if cfg.AgentFactory == "" {
		cfg.AgentFactory = AgentFactoryAddress
	}

	executor := &ArcExecutor{cfg: cfg}
	if err := executor.validateConfig(); err != nil {
		return nil, err
	}

	return executor, nil
}

func (executor *ArcExecutor) ExecuteCreateMarket(ctx context.Context, intent Intent) (ExecutionResult, error) {
	if intent.Status != StatusConfirmed {
		return ExecutionResult{}, ErrIntentNotConfirmed
	}
	if !intent.ValidationResult.Valid {
		return ExecutionResult{}, ErrIntentInvalid
	}
	if intent.Action != ActionCreateMarket {
		return ExecutionResult{}, ErrExecutionNotImplemented
	}
	if err := executor.validateConfig(); err != nil {
		return ExecutionResult{}, err
	}

	parsedABI, err := abi.JSON(strings.NewReader(agentFactoryABIJSON))
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("parse agent factory abi: %w", err)
	}

	client, err := ethclient.DialContext(ctx, executor.cfg.RPCURL)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("connect arc testnet rpc: %w", err)
	}
	defer client.Close()

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(executor.cfg.ExecutorPrivateKey, "0x"))
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("%w: AGENT_EXECUTOR_PRIVATE_KEY is invalid", ErrExecutionConfigInvalid)
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("read chain id: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("create transactor: %w", err)
	}
	auth.Context = ctx

	closeTimestamp, ok := new(big.Int).SetString(intent.CloseTimestamp, 10)
	if !ok {
		return ExecutionResult{}, fmt.Errorf("%w: close_timestamp is invalid", ErrIntentInvalid)
	}

	factoryAddress := common.HexToAddress(executor.cfg.AgentFactory)
	factory := bind.NewBoundContract(factoryAddress, parsedABI, client, client, client)
	tx, err := factory.Transact(auth, "createMarket", intent.MarketID, intent.Question, closeTimestamp, common.HexToAddress(intent.Resolver), common.HexToAddress(intent.CollateralToken))
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("broadcast createMarket: %w", err)
	}

	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("wait for createMarket receipt: %w", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return ExecutionResult{}, fmt.Errorf("createMarket transaction failed with receipt status %d", receipt.Status)
	}

	createdMarket := createdMarketFromReceipt(parsedABI, factoryAddress, receipt)
	marketCount, err := readMarketCount(ctx, client, parsedABI, factoryAddress)
	if err != nil {
		return ExecutionResult{}, err
	}

	readback := ExecutionReadback{
		MarketCount: marketCount.String(),
	}
	if createdMarket != (common.Address{}) {
		readback.CreatedMarket = createdMarket.Hex()
		isMarket, err := readIsMarket(ctx, client, parsedABI, factoryAddress, createdMarket)
		if err != nil {
			return ExecutionResult{}, err
		}
		readback.IsMarket = &isMarket
	}

	return ExecutionResult{
		IntentID:            intent.ID,
		AgentID:             intent.AgentID,
		AgentWalletAddress:  intent.AgentWalletAddress,
		WalletProvider:      intent.WalletProvider,
		Action:              intent.Action,
		Status:              StatusExecuted,
		ExecutionMode:       ExecutionModeAgentContract,
		Network:             NetworkArcTestnet,
		AgentFactoryAddress: factoryAddress.Hex(),
		BroadcastPerformed:  true,
		TransactionHash:     tx.Hash().Hex(),
		Readback:            readback,
	}, nil
}

func (executor *ArcExecutor) ExecuteBuyYes(ctx context.Context, intent Intent) (ExecutionResult, error) {
	return executor.executeBuyPosition(ctx, intent, ActionBuyYes, "buyYes", "yesPositions", "totalYes")
}

func (executor *ArcExecutor) ExecuteBuyNo(ctx context.Context, intent Intent) (ExecutionResult, error) {
	return executor.executeBuyPosition(ctx, intent, ActionBuyNo, "buyNo", "noPositions", "totalNo")
}

func (executor *ArcExecutor) ExecuteCloseMarket(context.Context, Intent) (ExecutionResult, error) {
	return ExecutionResult{}, ErrExecutionNotImplemented
}

func (executor *ArcExecutor) ExecuteResolveMarket(context.Context, Intent) (ExecutionResult, error) {
	return ExecutionResult{}, ErrExecutionNotImplemented
}

func (executor *ArcExecutor) ExecuteClaimPayout(context.Context, Intent) (ExecutionResult, error) {
	return ExecutionResult{}, ErrExecutionNotImplemented
}

func (executor *ArcExecutor) ExecuteCancelMarket(context.Context, Intent) (ExecutionResult, error) {
	return ExecutionResult{}, ErrExecutionNotImplemented
}

func (executor *ArcExecutor) ExecuteClaimRefund(context.Context, Intent) (ExecutionResult, error) {
	return ExecutionResult{}, ErrExecutionNotImplemented
}

func (executor *ArcExecutor) executeBuyPosition(ctx context.Context, intent Intent, expectedAction string, buyMethod string, positionMethod string, totalMethod string) (ExecutionResult, error) {
	if intent.Status != StatusConfirmed {
		return ExecutionResult{}, ErrIntentNotConfirmed
	}
	if !intent.ValidationResult.Valid {
		return ExecutionResult{}, ErrIntentInvalid
	}
	if intent.Action != expectedAction {
		return ExecutionResult{}, ErrExecutionNotImplemented
	}
	if err := executor.validateConfig(); err != nil {
		return ExecutionResult{}, err
	}
	if intent.UserWallet == "" || intent.MarketContractAddress == "" || intent.Amount == "" {
		return ExecutionResult{}, ErrIntentInvalid
	}

	amount, ok := new(big.Int).SetString(intent.Amount, 10)
	if !ok || amount.Sign() <= 0 {
		return ExecutionResult{}, fmt.Errorf("%w: amount must be a positive integer base-unit value", ErrIntentInvalid)
	}

	marketABI, err := abi.JSON(strings.NewReader(agentMarketABIJSON))
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("parse agent market abi: %w", err)
	}
	tokenABI, err := abi.JSON(strings.NewReader(erc20ABIJSON))
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("parse erc20 abi: %w", err)
	}

	client, err := ethclient.DialContext(ctx, executor.cfg.RPCURL)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("connect arc testnet rpc: %w", err)
	}
	defer client.Close()

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(executor.cfg.ExecutorPrivateKey, "0x"))
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("%w: AGENT_EXECUTOR_PRIVATE_KEY is invalid", ErrExecutionConfigInvalid)
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("read chain id: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("create transactor: %w", err)
	}
	auth.Context = ctx
	executorAddress := auth.From

	marketAddress := common.HexToAddress(intent.MarketContractAddress)
	tokenAddress := common.HexToAddress(ArcTestnetUSDCAddress)
	token := bind.NewBoundContract(tokenAddress, tokenABI, client, client, client)
	approveTx, err := token.Transact(auth, "approve", marketAddress, amount)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("broadcast USDC approve: %w", err)
	}
	approveReceipt, err := bind.WaitMined(ctx, client, approveTx)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("wait for USDC approve receipt: %w", err)
	}
	if approveReceipt.Status != types.ReceiptStatusSuccessful {
		return ExecutionResult{}, fmt.Errorf("USDC approve transaction failed with receipt status %d", approveReceipt.Status)
	}

	market := bind.NewBoundContract(marketAddress, marketABI, client, client, client)
	buyTx, err := market.Transact(auth, buyMethod, amount)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("broadcast %s: %w", buyMethod, err)
	}
	buyReceipt, err := bind.WaitMined(ctx, client, buyTx)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("wait for %s receipt: %w", buyMethod, err)
	}
	if buyReceipt.Status != types.ReceiptStatusSuccessful {
		return ExecutionResult{}, fmt.Errorf("%s transaction failed with receipt status %d", buyMethod, buyReceipt.Status)
	}

	positions, err := readUint256(ctx, client, marketABI, marketAddress, positionMethod, executorAddress)
	if err != nil {
		return ExecutionResult{}, err
	}
	total, err := readUint256(ctx, client, marketABI, marketAddress, totalMethod)
	if err != nil {
		return ExecutionResult{}, err
	}
	totalCollateral, err := readUint256(ctx, client, marketABI, marketAddress, "totalCollateral")
	if err != nil {
		return ExecutionResult{}, err
	}
	usdcBalance, err := readUint256(ctx, client, tokenABI, tokenAddress, "balanceOf", marketAddress)
	if err != nil {
		return ExecutionResult{}, err
	}
	usdcAllowance, err := readUint256(ctx, client, tokenABI, tokenAddress, "allowance", executorAddress, marketAddress)
	if err != nil {
		return ExecutionResult{}, err
	}

	readback := ExecutionReadback{
		TotalCollateral: totalCollateral.String(),
		USDCBalance:     usdcBalance.String(),
		USDCAllowance:   usdcAllowance.String(),
	}
	if expectedAction == ActionBuyYes {
		readback.YesPositions = positions.String()
		readback.TotalYes = total.String()
	} else {
		readback.NoPositions = positions.String()
		readback.TotalNo = total.String()
	}

	return ExecutionResult{
		IntentID:               intent.ID,
		AgentID:                intent.AgentID,
		AgentWalletAddress:     intent.AgentWalletAddress,
		WalletProvider:         intent.WalletProvider,
		Action:                 intent.Action,
		Status:                 StatusExecuted,
		ExecutionMode:          ExecutionModeAgentContract,
		Network:                NetworkArcTestnet,
		AgentFactoryAddress:    common.HexToAddress(executor.cfg.AgentFactory).Hex(),
		MarketContractAddress:  marketAddress.Hex(),
		BroadcastPerformed:     true,
		ApproveTransactionHash: approveTx.Hash().Hex(),
		TransactionHash:        buyTx.Hash().Hex(),
		Readback:               readback,
	}, nil
}

func (executor *ArcExecutor) validateConfig() error {
	if executor == nil {
		return ErrExecutionConfigInvalid
	}
	if executor.cfg.RPCURL == "" {
		return fmt.Errorf("%w: ARC_TESTNET_RPC_URL is required", ErrExecutionConfigInvalid)
	}
	if executor.cfg.ExecutorPrivateKey == "" {
		return fmt.Errorf("%w: AGENT_EXECUTOR_PRIVATE_KEY is required", ErrExecutionConfigInvalid)
	}
	if !common.IsHexAddress(executor.cfg.AgentFactory) {
		return fmt.Errorf("%w: AGENT_FACTORY_ADDRESS is invalid", ErrExecutionConfigInvalid)
	}
	return nil
}

func createdMarketFromReceipt(parsedABI abi.ABI, factoryAddress common.Address, receipt *types.Receipt) common.Address {
	event := parsedABI.Events["AgentMarketDeployed"]
	for _, log := range receipt.Logs {
		if log.Address != factoryAddress || len(log.Topics) < 3 || log.Topics[0] != event.ID {
			continue
		}
		return common.BytesToAddress(log.Topics[2].Bytes())
	}
	return common.Address{}
}

func readMarketCount(ctx context.Context, client *ethclient.Client, parsedABI abi.ABI, factoryAddress common.Address) (*big.Int, error) {
	data, err := parsedABI.Pack("marketCount")
	if err != nil {
		return nil, fmt.Errorf("pack marketCount call: %w", err)
	}

	output, err := client.CallContract(ctx, ethereum.CallMsg{To: &factoryAddress, Data: data}, nil)
	if err != nil {
		return nil, fmt.Errorf("read marketCount: %w", err)
	}

	values, err := parsedABI.Unpack("marketCount", output)
	if err != nil {
		return nil, fmt.Errorf("unpack marketCount: %w", err)
	}
	if len(values) != 1 {
		return nil, errors.New("marketCount returned unexpected value count")
	}

	marketCount, ok := values[0].(*big.Int)
	if !ok {
		return nil, errors.New("marketCount returned unexpected value type")
	}

	return marketCount, nil
}

func readIsMarket(ctx context.Context, client *ethclient.Client, parsedABI abi.ABI, factoryAddress common.Address, market common.Address) (bool, error) {
	data, err := parsedABI.Pack("isMarket", market)
	if err != nil {
		return false, fmt.Errorf("pack isMarket call: %w", err)
	}

	output, err := client.CallContract(ctx, ethereum.CallMsg{To: &factoryAddress, Data: data}, nil)
	if err != nil {
		return false, fmt.Errorf("read isMarket: %w", err)
	}

	values, err := parsedABI.Unpack("isMarket", output)
	if err != nil {
		return false, fmt.Errorf("unpack isMarket: %w", err)
	}
	if len(values) != 1 {
		return false, errors.New("isMarket returned unexpected value count")
	}

	isMarket, ok := values[0].(bool)
	if !ok {
		return false, errors.New("isMarket returned unexpected value type")
	}

	return isMarket, nil
}

func readUint256(ctx context.Context, client *ethclient.Client, parsedABI abi.ABI, contractAddress common.Address, method string, args ...any) (*big.Int, error) {
	data, err := parsedABI.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf("pack %s call: %w", method, err)
	}

	output, err := client.CallContract(ctx, ethereum.CallMsg{To: &contractAddress, Data: data}, nil)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", method, err)
	}

	values, err := parsedABI.Unpack(method, output)
	if err != nil {
		return nil, fmt.Errorf("unpack %s: %w", method, err)
	}
	if len(values) != 1 {
		return nil, fmt.Errorf("%s returned unexpected value count", method)
	}

	value, ok := values[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("%s returned unexpected value type", method)
	}

	return value, nil
}
