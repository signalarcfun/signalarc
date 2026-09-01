package agent

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func (executor *CircleAPIExecutor) createMarketReadback(ctx context.Context, txHash string) ExecutionReadback {
	readback := ExecutionReadback{}
	client := executor.readbackClient(ctx)
	if client == nil {
		return readback
	}
	defer client.Close()

	factoryABI, err := abi.JSON(strings.NewReader(agentFactoryABIJSON))
	if err != nil {
		return readback
	}
	factory := common.HexToAddress(executor.cfg.AgentFactory)
	if count, err := rpcCallUint256(ctx, client, factoryABI, factory, "marketCount"); err == nil {
		readback.MarketCount = count
	}
	receipt, err := client.TransactionReceipt(ctx, common.HexToHash(txHash))
	if err == nil && receipt != nil {
		createdMarket := createdMarketFromReceipt(factoryABI, factory, receipt)
		if createdMarket != (common.Address{}) {
			readback.CreatedMarket = createdMarket.Hex()
			if isMarket, err := rpcCallBool(ctx, client, factoryABI, factory, "isMarket", createdMarket); err == nil {
				readback.IsMarket = &isMarket
			}
		}
	}
	return readback
}

func (executor *CircleAPIExecutor) buyReadback(ctx context.Context, intent Intent, action string, positionSignature string, totalSignature string) ExecutionReadback {
	readback := ExecutionReadback{}
	client := executor.readbackClient(ctx)
	if client == nil {
		return readback
	}
	defer client.Close()

	marketABI, err := abi.JSON(strings.NewReader(agentMarketABIJSON))
	if err != nil {
		return readback
	}
	erc20ABI, err := abi.JSON(strings.NewReader(erc20ABIJSON))
	if err != nil {
		return readback
	}
	market := common.HexToAddress(intent.MarketContractAddress)
	agentWallet := common.HexToAddress(intent.AgentWalletAddress)
	if position, err := rpcCallUint256BySignature(ctx, client, marketABI, market, positionSignature, agentWallet); err == nil {
		if action == ActionBuyYes {
			readback.YesPositions = position
		} else {
			readback.NoPositions = position
		}
	}
	if total, err := rpcCallUint256BySignature(ctx, client, marketABI, market, totalSignature); err == nil {
		if action == ActionBuyYes {
			readback.TotalYes = total
		} else {
			readback.TotalNo = total
		}
	}
	if totalCollateral, err := rpcCallUint256(ctx, client, marketABI, market, "totalCollateral"); err == nil {
		readback.TotalCollateral = totalCollateral
	}
	usdc := common.HexToAddress(ArcTestnetUSDCAddress)
	if balance, err := rpcCallUint256(ctx, client, erc20ABI, usdc, "balanceOf", market); err == nil {
		readback.USDCBalance = balance
	}
	return readback
}

func (executor *CircleAPIExecutor) lifecycleReadbackRPC(ctx context.Context, intent Intent, kind lifecycleReadbackKind) ExecutionReadback {
	readback := ExecutionReadback{}
	client := executor.readbackClient(ctx)
	if client == nil {
		return readback
	}
	defer client.Close()

	marketABI, err := abi.JSON(strings.NewReader(agentMarketLifecycleABIJSON))
	if err != nil {
		return readback
	}
	erc20ABI, err := abi.JSON(strings.NewReader(erc20ABIJSON))
	if err != nil {
		return readback
	}
	market := common.HexToAddress(intent.MarketContractAddress)
	agentWallet := common.HexToAddress(intent.AgentWalletAddress)
	if status, err := rpcCallUint256(ctx, client, marketABI, market, "status"); err == nil {
		readback.MarketStatus = status
	}
	switch kind {
	case readbackMarketState:
		if isOpen, err := rpcCallBool(ctx, client, marketABI, market, "isOpen"); err == nil {
			readback.IsOpen = &isOpen
		}
	case readbackResolution, readbackPayout:
		if winning, err := rpcCallUint256(ctx, client, marketABI, market, "winningOutcome"); err == nil {
			readback.WinningOutcome = winning
		}
		if claimable, err := rpcCallUint256(ctx, client, marketABI, market, "claimablePayout", agentWallet); err == nil {
			readback.ClaimablePayout = claimable
		}
		if hasClaimed, err := rpcCallBool(ctx, client, marketABI, market, "hasClaimed", agentWallet); err == nil {
			readback.HasClaimed = &hasClaimed
		}
		executor.readUSDCBalance(ctx, client, erc20ABI, market, &readback)
	case readbackRefund:
		if claimable, err := rpcCallUint256(ctx, client, marketABI, market, "claimableRefund", agentWallet); err == nil {
			readback.ClaimableRefund = claimable
		}
		if hasClaimed, err := rpcCallBool(ctx, client, marketABI, market, "hasClaimed", agentWallet); err == nil {
			readback.HasClaimed = &hasClaimed
		}
		executor.readUSDCBalance(ctx, client, erc20ABI, market, &readback)
	}
	return readback
}

func (executor *CircleAPIExecutor) readUSDCBalance(ctx context.Context, client *ethclient.Client, parsedABI abi.ABI, market common.Address, readback *ExecutionReadback) {
	if balance, err := rpcCallUint256(ctx, client, parsedABI, common.HexToAddress(ArcTestnetUSDCAddress), "balanceOf", market); err == nil {
		readback.USDCBalance = balance
	}
}

func (executor *CircleAPIExecutor) readbackClient(ctx context.Context) *ethclient.Client {
	rpcURL := strings.TrimSpace(executor.cfg.RPCURL)
	if rpcURL == "" {
		return nil
	}
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil
	}
	return client
}

func rpcCallUint256(ctx context.Context, client *ethclient.Client, parsedABI abi.ABI, contract common.Address, method string, args ...any) (string, error) {
	output, err := rpcCall(ctx, client, parsedABI, contract, method, args...)
	if err != nil {
		return "", err
	}
	values, err := parsedABI.Unpack(method, output)
	if err != nil {
		return "", err
	}
	if len(values) == 0 {
		return "", ErrIntentInvalid
	}
	value, ok := values[0].(*big.Int)
	if !ok {
		return "", ErrIntentInvalid
	}
	return value.String(), nil
}

func rpcCallBool(ctx context.Context, client *ethclient.Client, parsedABI abi.ABI, contract common.Address, method string, args ...any) (bool, error) {
	output, err := rpcCall(ctx, client, parsedABI, contract, method, args...)
	if err != nil {
		return false, err
	}
	values, err := parsedABI.Unpack(method, output)
	if err != nil {
		return false, err
	}
	if len(values) == 0 {
		return false, ErrIntentInvalid
	}
	value, ok := values[0].(bool)
	if !ok {
		return false, ErrIntentInvalid
	}
	return value, nil
}

func rpcCallUint256BySignature(ctx context.Context, client *ethclient.Client, parsedABI abi.ABI, contract common.Address, signature string, args ...any) (string, error) {
	name := strings.TrimSuffix(signature, "()")
	name = strings.TrimSuffix(name, "(address)")
	return rpcCallUint256(ctx, client, parsedABI, contract, name, args...)
}

func rpcCall(ctx context.Context, client *ethclient.Client, parsedABI abi.ABI, contract common.Address, method string, args ...any) ([]byte, error) {
	data, err := parsedABI.Pack(method, args...)
	if err != nil {
		return nil, err
	}
	return client.CallContract(ctx, ethereum.CallMsg{To: &contract, Data: data}, nil)
}

const agentMarketLifecycleABIJSON = `[
	{"type":"function","name":"status","inputs":[],"outputs":[{"name":"","type":"uint8"}],"stateMutability":"view"},
	{"type":"function","name":"isOpen","inputs":[],"outputs":[{"name":"","type":"bool"}],"stateMutability":"view"},
	{"type":"function","name":"winningOutcome","inputs":[],"outputs":[{"name":"","type":"uint8"}],"stateMutability":"view"},
	{"type":"function","name":"claimablePayout","inputs":[{"name":"user","type":"address"}],"outputs":[{"name":"","type":"uint256"}],"stateMutability":"view"},
	{"type":"function","name":"claimableRefund","inputs":[{"name":"user","type":"address"}],"outputs":[{"name":"","type":"uint256"}],"stateMutability":"view"},
	{"type":"function","name":"hasClaimed","inputs":[{"name":"user","type":"address"}],"outputs":[{"name":"","type":"bool"}],"stateMutability":"view"}
]`
