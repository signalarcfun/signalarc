package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/circleapi"
)

type CircleAPIBalanceClient interface {
	GetWalletTokenBalances(context.Context, string) ([]circleapi.WalletTokenBalance, error)
}

type CircleAPIBalanceReaderConfig struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
	Client  CircleAPIBalanceClient
}

type CircleAPIBalanceReader struct {
	cfg    CircleAPIBalanceReaderConfig
	client CircleAPIBalanceClient
}

func NewCircleAPIBalanceReader(cfg CircleAPIBalanceReaderConfig) (*CircleAPIBalanceReader, error) {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 120 * time.Second
	}
	if cfg.Client == nil {
		client, err := circleapi.NewClient(circleapi.ClientConfig{
			APIKey:  cfg.APIKey,
			BaseURL: cfg.BaseURL,
			Timeout: cfg.Timeout,
		})
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrExecutionConfigInvalid, err)
		}
		cfg.Client = client
	}
	return &CircleAPIBalanceReader{cfg: cfg, client: cfg.Client}, nil
}

func (reader *CircleAPIBalanceReader) GetAgentWalletBalances(ctx context.Context, request CircleAgentWalletBalanceRequest) (CircleAgentWalletBalances, error) {
	if reader == nil || reader.client == nil {
		return CircleAgentWalletBalances{}, ErrCircleAgentWalletBalanceFailed
	}
	if strings.TrimSpace(request.WalletProvider) != WalletProviderCircleAgentWallet {
		return CircleAgentWalletBalances{}, ErrCircleAgentWalletBalanceFailed
	}
	if strings.TrimSpace(request.Chain) != ChainArcTestnet {
		return CircleAgentWalletBalances{}, ErrCircleAgentWalletBalanceFailed
	}
	walletID, err := circleWalletIDFromPolicyMetadata(request.PolicyMetadata)
	if err != nil {
		return CircleAgentWalletBalances{}, fmt.Errorf("%w: %v", ErrCircleAgentWalletBalanceFailed, err)
	}
	ctx, cancel := context.WithTimeout(ctx, reader.cfg.Timeout)
	defer cancel()

	tokenBalances, err := reader.client.GetWalletTokenBalances(ctx, walletID)
	if err != nil {
		return CircleAgentWalletBalances{}, err
	}
	balances := make([]any, 0, len(tokenBalances))
	for _, balance := range tokenBalances {
		balances = append(balances, balance)
	}
	return CircleAgentWalletBalances{Balances: balances}, nil
}
