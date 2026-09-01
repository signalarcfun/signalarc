package agent

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/circleapi"
)

type fakeCircleAPIClient struct {
	inputs []circleapi.CreateContractExecutionTransactionInput
}

type fakeEntitySecretCiphertextProvider struct {
	ciphertexts []string
	calls       int
}

func (client *fakeCircleAPIClient) CreateContractExecutionTransaction(_ context.Context, input circleapi.CreateContractExecutionTransactionInput) (circleapi.CreateContractExecutionTransactionResponse, error) {
	client.inputs = append(client.inputs, input)
	return circleapi.CreateContractExecutionTransactionResponse{ID: "circle_tx"}, nil
}

func (client *fakeCircleAPIClient) GetTransaction(context.Context, string) (circleapi.Transaction, error) {
	return circleapi.Transaction{}, nil
}

func (client *fakeCircleAPIClient) PollTransaction(context.Context, string) (circleapi.Transaction, error) {
	return circleapi.Transaction{ID: "circle_tx", Status: "COMPLETE", TransactionHash: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, nil
}

func (provider *fakeEntitySecretCiphertextProvider) Ciphertext(context.Context) (string, error) {
	provider.calls++
	if len(provider.ciphertexts) == 0 {
		return "test-ciphertext", nil
	}
	ciphertext := provider.ciphertexts[0]
	provider.ciphertexts = provider.ciphertexts[1:]
	return ciphertext, nil
}

func TestCircleAPIExecutorMissingWalletIDFailsClosed(t *testing.T) {
	executor := newTestCircleAPIExecutor(&fakeCircleAPIClient{})
	intent := confirmedIntent(ActionBuyYes)
	intent.PolicyMetadata = nil

	_, err := executor.ExecuteBuyYes(context.Background(), intent)
	if !errors.Is(err, ErrCircleWalletIDMissing) {
		t.Fatalf("expected ErrCircleWalletIDMissing, got %v", err)
	}
}

func TestCircleAPIExecutorBuyYesBuildsApproveAndBuyPayloads(t *testing.T) {
	client := &fakeCircleAPIClient{}
	provider := &fakeEntitySecretCiphertextProvider{ciphertexts: []string{"approve-ciphertext", "buy-ciphertext"}}
	executor := newTestCircleAPIExecutorWithProvider(client, provider)
	intent := confirmedIntent(ActionBuyYes)
	intent.Amount = "1.5"
	intent.PolicyMetadata = map[string]string{"circle_wallet_id": "wallet-1"}

	result, err := executor.ExecuteBuyYes(context.Background(), intent)
	if err != nil {
		t.Fatalf("execute buy yes: %v", err)
	}
	if result.ExecutionMode != ExecutionModeCircleDeveloperWalletAPI {
		t.Fatalf("expected api execution mode, got %q", result.ExecutionMode)
	}
	if len(client.inputs) != 2 {
		t.Fatalf("expected approve and buy payloads, got %d", len(client.inputs))
	}
	assertCircleInput(t, client.inputs[0], circleapi.CreateContractExecutionTransactionInput{
		WalletID:               "wallet-1",
		ContractAddress:        ArcTestnetUSDCAddress,
		AbiFunctionSignature:   "approve(address,uint256)",
		AbiParameters:          []string{"0x3333333333333333333333333333333333333333", "1500000"},
		EntitySecretCiphertext: "approve-ciphertext",
		FeeLevel:               "MEDIUM",
	})
	assertCircleInput(t, client.inputs[1], circleapi.CreateContractExecutionTransactionInput{
		WalletID:               "wallet-1",
		ContractAddress:        "0x3333333333333333333333333333333333333333",
		AbiFunctionSignature:   "buyYes(uint256)",
		AbiParameters:          []string{"1500000"},
		EntitySecretCiphertext: "buy-ciphertext",
		FeeLevel:               "MEDIUM",
	})
	if provider.calls != 2 {
		t.Fatalf("expected provider to be called once per contract execution, got %d", provider.calls)
	}
}

func TestCircleAPIExecutorActionToABIMapping(t *testing.T) {
	tests := []struct {
		action    string
		execute   func(context.Context, *CircleAPIExecutor, Intent) (ExecutionResult, error)
		signature string
		params    []string
		contract  string
	}{
		{
			action: ActionCreateMarket,
			execute: func(ctx context.Context, e *CircleAPIExecutor, intent Intent) (ExecutionResult, error) {
				intent.Resolver = "0x2222222222222222222222222222222222222222"
				return e.ExecuteCreateMarket(ctx, intent)
			},
			signature: "createMarket(string,string,uint256,address,address)",
			contract:  AgentFactoryAddress,
		},
		{action: ActionCloseMarket, execute: func(ctx context.Context, e *CircleAPIExecutor, intent Intent) (ExecutionResult, error) {
			return e.ExecuteCloseMarket(ctx, intent)
		}, signature: "closeMarket()", contract: "0x3333333333333333333333333333333333333333"},
		{action: ActionResolveMarket, execute: func(ctx context.Context, e *CircleAPIExecutor, intent Intent) (ExecutionResult, error) {
			intent.Outcome = "yes"
			return e.ExecuteResolveMarket(ctx, intent)
		}, signature: "resolve(uint8)", params: []string{"1"}, contract: "0x3333333333333333333333333333333333333333"},
		{action: ActionClaimPayout, execute: func(ctx context.Context, e *CircleAPIExecutor, intent Intent) (ExecutionResult, error) {
			return e.ExecuteClaimPayout(ctx, intent)
		}, signature: "claimPayout()", contract: "0x3333333333333333333333333333333333333333"},
		{action: ActionCancelMarket, execute: func(ctx context.Context, e *CircleAPIExecutor, intent Intent) (ExecutionResult, error) {
			return e.ExecuteCancelMarket(ctx, intent)
		}, signature: "cancelMarket()", contract: "0x3333333333333333333333333333333333333333"},
		{action: ActionClaimRefund, execute: func(ctx context.Context, e *CircleAPIExecutor, intent Intent) (ExecutionResult, error) {
			return e.ExecuteClaimRefund(ctx, intent)
		}, signature: "claimRefund()", contract: "0x3333333333333333333333333333333333333333"},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			client := &fakeCircleAPIClient{}
			executor := newTestCircleAPIExecutor(client)
			intent := confirmedIntent(tt.action)
			intent.PolicyMetadata = map[string]string{"circle_wallet_id": "wallet-1"}
			intent.CloseTimestamp = futureTimestamp()
			_, err := tt.execute(context.Background(), executor, intent)
			if err != nil {
				t.Fatalf("execute: %v", err)
			}
			if len(client.inputs) == 0 {
				t.Fatal("expected Circle API payload")
			}
			got := client.inputs[len(client.inputs)-1]
			if got.ContractAddress != tt.contract || got.AbiFunctionSignature != tt.signature {
				t.Fatalf("unexpected payload %#v", got)
			}
			if tt.action == ActionCreateMarket {
				tt.params = []string{"market-1", "Will SignalArc execute through Circle?", intent.CloseTimestamp, "0x2222222222222222222222222222222222222222", ArcTestnetUSDCAddress}
			}
			if tt.params != nil && !reflect.DeepEqual(got.AbiParameters, tt.params) {
				t.Fatalf("expected params %#v, got %#v", tt.params, got.AbiParameters)
			}
		})
	}
}

func newTestCircleAPIExecutor(client *fakeCircleAPIClient) *CircleAPIExecutor {
	return newTestCircleAPIExecutorWithProvider(client, &fakeEntitySecretCiphertextProvider{})
}

func newTestCircleAPIExecutorWithProvider(client *fakeCircleAPIClient, provider circleapi.EntitySecretCiphertextProvider) *CircleAPIExecutor {
	executor, err := NewCircleAPIExecutor(CircleAPIExecutorConfig{
		Enabled:                        true,
		EntitySecretCiphertextProvider: provider,
		Timeout:                        time.Second,
		AgentFactory:                   AgentFactoryAddress,
		Client:                         client,
	})
	if err != nil {
		panic(err)
	}
	return executor
}

func assertCircleInput(t *testing.T, actual circleapi.CreateContractExecutionTransactionInput, expected circleapi.CreateContractExecutionTransactionInput) {
	t.Helper()
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %#v, got %#v", expected, actual)
	}
}
