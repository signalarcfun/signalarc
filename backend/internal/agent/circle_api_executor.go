package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/circleapi"
)

var ErrCircleWalletIDMissing = errors.New("circle_wallet_id policy metadata is required")

type CircleAPIClient interface {
	CreateContractExecutionTransaction(context.Context, circleapi.CreateContractExecutionTransactionInput) (circleapi.CreateContractExecutionTransactionResponse, error)
	GetTransaction(context.Context, string) (circleapi.Transaction, error)
	PollTransaction(context.Context, string) (circleapi.Transaction, error)
}

type CircleAPIExecutorConfig struct {
	Enabled                         bool
	APIKey                          string
	StaticDevEntitySecretCiphertext string
	EntitySecretCiphertextProvider  circleapi.EntitySecretCiphertextProvider
	BaseURL                         string
	Timeout                         time.Duration
	AgentFactory                    string
	RPCURL                          string
	Client                          CircleAPIClient
}

type CircleAPIExecutor struct {
	cfg                     CircleAPIExecutorConfig
	client                  CircleAPIClient
	entitySecretCiphertexts circleapi.EntitySecretCiphertextProvider
}

func NewCircleAPIExecutor(cfg CircleAPIExecutorConfig) (*CircleAPIExecutor, error) {
	if strings.TrimSpace(cfg.AgentFactory) == "" {
		cfg.AgentFactory = AgentFactoryAddress
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 120 * time.Second
	}
	if strings.TrimSpace(cfg.RPCURL) == "" {
		cfg.RPCURL = strings.TrimSpace(os.Getenv("ARC_TESTNET_RPC_URL"))
	}
	if cfg.EntitySecretCiphertextProvider == nil && strings.TrimSpace(cfg.StaticDevEntitySecretCiphertext) != "" {
		cfg.EntitySecretCiphertextProvider = circleapi.NewEnvEntitySecretCiphertextProvider(cfg.StaticDevEntitySecretCiphertext)
	}
	if cfg.Client == nil {
		client, err := circleapi.NewClient(circleapi.ClientConfig{
			APIKey:      cfg.APIKey,
			BaseURL:     cfg.BaseURL,
			Timeout:     cfg.Timeout,
			PollTimeout: cfg.Timeout,
		})
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrExecutionConfigInvalid, err)
		}
		cfg.Client = client
	}
	return &CircleAPIExecutor{cfg: cfg, client: cfg.Client, entitySecretCiphertexts: cfg.EntitySecretCiphertextProvider}, nil
}

func NewCircleAPIExecutorFromEnv(enabled bool, timeout time.Duration, agentFactory string) (*CircleAPIExecutor, error) {
	apiKey := os.Getenv("CIRCLE_API_KEY")
	baseURL := os.Getenv("CIRCLE_API_BASE_URL")
	var provider circleapi.EntitySecretCiphertextProvider
	if rawEntitySecret := strings.TrimSpace(os.Getenv("CIRCLE_ENTITY_SECRET")); rawEntitySecret != "" {
		rawProvider, err := circleapi.NewRawEntitySecretCiphertextProvider(circleapi.RawEntitySecretCiphertextProviderConfig{
			APIKey:          apiKey,
			BaseURL:         baseURL,
			RawEntitySecret: rawEntitySecret,
			Timeout:         timeout,
		})
		if err != nil {
			return nil, fmt.Errorf("%w: circle entity secret ciphertext provider is invalid", ErrExecutionConfigInvalid)
		}
		provider = rawProvider
	}
	return NewCircleAPIExecutor(CircleAPIExecutorConfig{
		Enabled:                         enabled,
		APIKey:                          apiKey,
		StaticDevEntitySecretCiphertext: os.Getenv("CIRCLE_ENTITY_SECRET_CIPHERTEXT"),
		EntitySecretCiphertextProvider:  provider,
		BaseURL:                         baseURL,
		Timeout:                         timeout,
		AgentFactory:                    agentFactory,
	})
}

func (executor *CircleAPIExecutor) executeContract(ctx context.Context, intent Intent, contractAddress string, signature string, params []string) (string, error) {
	if executor == nil || executor.client == nil {
		return "", ErrExecutionConfigInvalid
	}
	walletID, err := circleWalletIDFromPolicyMetadata(intent.PolicyMetadata)
	if err != nil {
		return "", err
	}
	if executor.entitySecretCiphertexts == nil {
		return "", fmt.Errorf("%w: circle entity secret ciphertext provider is required", ErrExecutionConfigInvalid)
	}
	ctx, cancel := context.WithTimeout(ctx, executor.cfg.Timeout)
	defer cancel()
	entitySecretCiphertext, err := executor.entitySecretCiphertexts.Ciphertext(ctx)
	if err != nil || strings.TrimSpace(entitySecretCiphertext) == "" {
		return "", fmt.Errorf("%w: circle entity secret ciphertext provider is required", ErrExecutionConfigInvalid)
	}

	created, err := executor.client.CreateContractExecutionTransaction(ctx, circleapi.CreateContractExecutionTransactionInput{
		WalletID:               walletID,
		ContractAddress:        contractAddress,
		AbiFunctionSignature:   signature,
		AbiParameters:          params,
		EntitySecretCiphertext: entitySecretCiphertext,
		FeeLevel:               "MEDIUM",
	})
	if err != nil {
		return "", err
	}
	tx, err := executor.client.PollTransaction(ctx, created.ID)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(tx.TransactionHash) == "" {
		return "", errors.New("circle transaction completed without transaction hash")
	}
	return tx.TransactionHash, nil
}

func (executor *CircleAPIExecutor) validateIntent(intent Intent, expectedAction string) error {
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
	if strings.TrimSpace(intent.AgentWalletAddress) == "" {
		return ErrIntentInvalid
	}
	return nil
}

func circleWalletIDFromPolicyMetadata(metadata map[string]string) (string, error) {
	if metadata == nil {
		return "", ErrCircleWalletIDMissing
	}
	walletID := strings.TrimSpace(metadata["circle_wallet_id"])
	if walletID == "" {
		return "", ErrCircleWalletIDMissing
	}
	return walletID, nil
}

func baseResult(intent Intent, mode string, agentFactory string) ExecutionResult {
	return ExecutionResult{
		IntentID:            intent.ID,
		AgentID:             intent.AgentID,
		AgentWalletAddress:  intent.AgentWalletAddress,
		WalletProvider:      intent.WalletProvider,
		Action:              intent.Action,
		Status:              StatusExecuted,
		ExecutionMode:       mode,
		Network:             NetworkArcTestnet,
		AgentFactoryAddress: agentFactory,
		BroadcastPerformed:  true,
	}
}
