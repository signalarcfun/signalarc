package agent

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

type fakeCommandRunner struct {
	outputs [][]byte
	errAt   int
	calls   []fakeCommandCall
}

type fakeCommandCall struct {
	name string
	args []string
}

func (runner *fakeCommandRunner) Run(_ context.Context, name string, args []string) ([]byte, error) {
	runner.calls = append(runner.calls, fakeCommandCall{name: name, args: append([]string{}, args...)})
	if runner.errAt > 0 && len(runner.calls) == runner.errAt {
		return nil, errors.New("raw sensitive CLI error")
	}
	if len(runner.outputs) == 0 {
		return nil, errors.New("missing fake output")
	}
	output := runner.outputs[0]
	runner.outputs = runner.outputs[1:]
	return output, nil
}

func TestCircleCLIExecutorCreateMarketBuildsCommandsAndReadback(t *testing.T) {
	runner := &fakeCommandRunner{outputs: [][]byte{
		[]byte(`{"transactionHash":"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`),
		jsonResult(abiUint256("9")),
		jsonResult(abiAddress("0xabcf081e456c1a11106def590666a07b76d456f8")),
		jsonResult(abiUint256("1")),
	}}
	executor := newTestCircleCLIExecutor(runner, true)
	intent := confirmedIntent(ActionCreateMarket)
	intent.Resolver = "0x2222222222222222222222222222222222222222"

	result, err := executor.ExecuteCreateMarket(context.Background(), intent)
	if err != nil {
		t.Fatalf("execute create market: %v", err)
	}

	// Verify the CLI args structure. The close_timestamp is dynamic so we
	// check the surrounding args and verify the timestamp position separately.
	firstCall := runner.calls[0].args
	expectedPrefix := []string{
		"wallet", "execute", "createMarket(string,string,uint256,address,address)",
		"market-1", "Will SignalArc execute through Circle?",
	}
	for i, expected := range expectedPrefix {
		if firstCall[i] != expected {
			t.Fatalf("arg[%d] expected %q, got %q", i, expected, firstCall[i])
		}
	}
	// Position 5 is the dynamic close_timestamp; verify it parses as a positive integer.
	if _, ok := new(big.Int).SetString(firstCall[5], 10); !ok {
		t.Fatalf("expected close_timestamp arg to be a decimal integer, got %q", firstCall[5])
	}
	expectedSuffix := []string{
		"0x9999999999999999999999999999999999999999", ArcTestnetUSDCAddress,
		"--address", "0x9999999999999999999999999999999999999999",
		"--contract", AgentFactoryAddress,
		"--chain", ChainArcTestnet,
		"--output", "json",
	}
	suffixStart := 6
	for i, expected := range expectedSuffix {
		if firstCall[suffixStart+i] != expected {
			t.Fatalf("arg[%d] expected %q, got %q", suffixStart+i, expected, firstCall[suffixStart+i])
		}
	}
	assertArgs(t, runner.calls[2].args, []string{
		"contract", "query", "allMarkets(uint256)", "8",
		"--contract", AgentFactoryAddress,
		"--chain", ChainArcTestnet,
		"--output", "json",
	})
	if result.ExecutionMode != ExecutionModeCircleAgentWalletCLI {
		t.Fatalf("expected circle execution mode, got %q", result.ExecutionMode)
	}
	if result.TransactionHash != "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("unexpected tx hash %q", result.TransactionHash)
	}
	if result.Readback.MarketCount != "9" || result.Readback.CreatedMarket != "0xabcF081E456c1A11106DeF590666A07B76d456f8" {
		t.Fatalf("unexpected readback %#v", result.Readback)
	}
	if result.Readback.IsMarket == nil || !*result.Readback.IsMarket {
		t.Fatalf("expected is market true, got %#v", result.Readback.IsMarket)
	}
}

func TestCircleCLIExecutorBuyYesBuildsApproveAndBuyCommands(t *testing.T) {
	runner := &fakeCommandRunner{outputs: buyOutputs("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "1000000", "1000000", "1000000", "1000000")}
	executor := newTestCircleCLIExecutor(runner, true)
	intent := confirmedIntent(ActionBuyYes)
	intent.Amount = "1"

	result, err := executor.ExecuteBuyYes(context.Background(), intent)
	if err != nil {
		t.Fatalf("execute buy yes: %v", err)
	}

	assertArgs(t, runner.calls[0].args, []string{
		"wallet", "execute", "approve(address,uint256)",
		"0x3333333333333333333333333333333333333333", "1000000",
		"--address", "0x9999999999999999999999999999999999999999",
		"--contract", ArcTestnetUSDCAddress,
		"--chain", ChainArcTestnet,
		"--output", "json",
	})
	assertArgs(t, runner.calls[1].args, []string{
		"wallet", "execute", "buyYes(uint256)", "1000000",
		"--address", "0x9999999999999999999999999999999999999999",
		"--contract", "0x3333333333333333333333333333333333333333",
		"--chain", ChainArcTestnet,
		"--output", "json",
	})
	if result.ApproveTransactionHash != "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" || result.TransactionHash != "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" {
		t.Fatalf("unexpected hashes %#v", result)
	}
	if result.Readback.YesPositions != "1000000" || result.Readback.TotalYes != "1000000" {
		t.Fatalf("unexpected readback %#v", result.Readback)
	}
}

func TestCircleCLIExecutorBuyNoBuildsApproveAndBuyCommands(t *testing.T) {
	runner := &fakeCommandRunner{outputs: buyOutputs("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "1500000", "1500000", "1500000", "1500000")}
	executor := newTestCircleCLIExecutor(runner, true)
	intent := confirmedIntent(ActionBuyNo)
	intent.Amount = "1.5"

	result, err := executor.ExecuteBuyNo(context.Background(), intent)
	if err != nil {
		t.Fatalf("execute buy no: %v", err)
	}

	assertArgs(t, runner.calls[0].args, []string{
		"wallet", "execute", "approve(address,uint256)",
		"0x3333333333333333333333333333333333333333", "1500000",
		"--address", "0x9999999999999999999999999999999999999999",
		"--contract", ArcTestnetUSDCAddress,
		"--chain", ChainArcTestnet,
		"--output", "json",
	})
	assertArgs(t, runner.calls[1].args, []string{
		"wallet", "execute", "buyNo(uint256)", "1500000",
		"--address", "0x9999999999999999999999999999999999999999",
		"--contract", "0x3333333333333333333333333333333333333333",
		"--chain", ChainArcTestnet,
		"--output", "json",
	})
	if result.Readback.NoPositions != "1500000" || result.Readback.TotalNo != "1500000" {
		t.Fatalf("unexpected readback %#v", result.Readback)
	}
}

func TestCircleCLIExecutorBuyRejectsUSDCAmountBeyondSixDecimals(t *testing.T) {
	runner := &fakeCommandRunner{}
	executor := newTestCircleCLIExecutor(runner, true)
	intent := confirmedIntent(ActionBuyYes)
	intent.Amount = "1.0000001"

	_, err := executor.ExecuteBuyYes(context.Background(), intent)
	if !errors.Is(err, ErrIntentInvalid) {
		t.Fatalf("expected ErrIntentInvalid, got %v", err)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("expected no Circle CLI calls for invalid precision, got %d", len(runner.calls))
	}
}

func TestCircleCLIExecutorDisabledFailsClosed(t *testing.T) {
	executor := newTestCircleCLIExecutor(&fakeCommandRunner{}, false)

	_, err := executor.ExecuteCreateMarket(context.Background(), confirmedIntent(ActionCreateMarket))
	if !errors.Is(err, ErrExecutionProviderDisabled) {
		t.Fatalf("expected disabled error, got %v", err)
	}
}

func TestCircleCLIExecutorDisallowedActionFailsClosed(t *testing.T) {
	executor := newTestCircleCLIExecutor(&fakeCommandRunner{}, true)
	intent := confirmedIntent(ActionBuyYes)
	intent.AllowedActions = []string{ActionBuyNo}

	_, err := executor.ExecuteBuyYes(context.Background(), intent)
	if !errors.Is(err, ErrIntentInvalid) {
		t.Fatalf("expected invalid intent, got %v", err)
	}
}

func TestCircleCLIExecutorCloseMarketBuildsCommandAndReadback(t *testing.T) {
	runner := &fakeCommandRunner{outputs: [][]byte{
		[]byte(`{"transactionHash":"0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"}`),
		jsonResult(abiUint256("1")),
		jsonResult(abiUint256("0")),
	}}
	executor := newTestCircleCLIExecutor(runner, true)

	result, err := executor.ExecuteCloseMarket(context.Background(), confirmedIntent(ActionCloseMarket))
	if err != nil {
		t.Fatalf("execute close market: %v", err)
	}

	assertArgs(t, runner.calls[0].args, []string{
		"wallet", "execute", "closeMarket()",
		"--address", "0x9999999999999999999999999999999999999999",
		"--contract", "0x3333333333333333333333333333333333333333",
		"--chain", ChainArcTestnet,
		"--output", "json",
	})
	if result.TransactionHash != "0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc" {
		t.Fatalf("unexpected tx hash %q", result.TransactionHash)
	}
	if result.Readback.MarketStatus != "1" || result.Readback.IsOpen == nil || *result.Readback.IsOpen {
		t.Fatalf("unexpected close readback %#v", result.Readback)
	}
}

func TestCircleCLIExecutorResolveMarketBuildsCommandAndReadback(t *testing.T) {
	runner := &fakeCommandRunner{outputs: lifecyclePayoutOutputs(
		"0xdddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
		"2",
		"1",
		"1000000",
		false,
		"2000000",
	)}
	executor := newTestCircleCLIExecutor(runner, true)
	intent := confirmedIntent(ActionResolveMarket)
	intent.Outcome = "yes"

	result, err := executor.ExecuteResolveMarket(context.Background(), intent)
	if err != nil {
		t.Fatalf("execute resolve market: %v", err)
	}

	assertArgs(t, runner.calls[0].args, []string{
		"wallet", "execute", "resolve(uint8)", "1",
		"--address", "0x9999999999999999999999999999999999999999",
		"--contract", "0x3333333333333333333333333333333333333333",
		"--chain", ChainArcTestnet,
		"--output", "json",
	})
	if result.Readback.MarketStatus != "2" || result.Readback.WinningOutcome != "1" || result.Readback.ClaimablePayout != "1000000" {
		t.Fatalf("unexpected resolve readback %#v", result.Readback)
	}
	if result.Readback.HasClaimed == nil || *result.Readback.HasClaimed {
		t.Fatalf("expected has claimed false, got %#v", result.Readback.HasClaimed)
	}
}

func TestCircleCLIExecutorClaimPayoutBuildsCommandAndReadback(t *testing.T) {
	runner := &fakeCommandRunner{outputs: lifecyclePayoutOutputs(
		"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		"2",
		"1",
		"1000000",
		true,
		"0",
	)}
	executor := newTestCircleCLIExecutor(runner, true)

	result, err := executor.ExecuteClaimPayout(context.Background(), confirmedIntent(ActionClaimPayout))
	if err != nil {
		t.Fatalf("execute claim payout: %v", err)
	}

	assertArgs(t, runner.calls[0].args, []string{
		"wallet", "execute", "claimPayout()",
		"--address", "0x9999999999999999999999999999999999999999",
		"--contract", "0x3333333333333333333333333333333333333333",
		"--chain", ChainArcTestnet,
		"--output", "json",
	})
	if result.Readback.HasClaimed == nil || !*result.Readback.HasClaimed || result.Readback.USDCBalance != "0" {
		t.Fatalf("unexpected payout readback %#v", result.Readback)
	}
}

func TestCircleCLIExecutorCancelMarketBuildsCommandAndReadback(t *testing.T) {
	runner := &fakeCommandRunner{outputs: lifecycleRefundOutputs(
		"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		"3",
		"2000000",
		false,
		"2000000",
	)}
	executor := newTestCircleCLIExecutor(runner, true)

	result, err := executor.ExecuteCancelMarket(context.Background(), confirmedIntent(ActionCancelMarket))
	if err != nil {
		t.Fatalf("execute cancel market: %v", err)
	}

	assertArgs(t, runner.calls[0].args, []string{
		"wallet", "execute", "cancelMarket()",
		"--address", "0x9999999999999999999999999999999999999999",
		"--contract", "0x3333333333333333333333333333333333333333",
		"--chain", ChainArcTestnet,
		"--output", "json",
	})
	if result.Readback.MarketStatus != "3" || result.Readback.ClaimableRefund != "2000000" {
		t.Fatalf("unexpected cancel readback %#v", result.Readback)
	}
}

func TestCircleCLIExecutorClaimRefundBuildsCommandAndReadback(t *testing.T) {
	runner := &fakeCommandRunner{outputs: lifecycleRefundOutputs(
		"0x1111111111111111111111111111111111111111111111111111111111111111",
		"3",
		"2000000",
		true,
		"0",
	)}
	executor := newTestCircleCLIExecutor(runner, true)

	result, err := executor.ExecuteClaimRefund(context.Background(), confirmedIntent(ActionClaimRefund))
	if err != nil {
		t.Fatalf("execute claim refund: %v", err)
	}

	assertArgs(t, runner.calls[0].args, []string{
		"wallet", "execute", "claimRefund()",
		"--address", "0x9999999999999999999999999999999999999999",
		"--contract", "0x3333333333333333333333333333333333333333",
		"--chain", ChainArcTestnet,
		"--output", "json",
	})
	if result.Readback.HasClaimed == nil || !*result.Readback.HasClaimed || result.Readback.USDCBalance != "0" {
		t.Fatalf("unexpected refund readback %#v", result.Readback)
	}
}

func TestCircleCLIExecutorLifecycleActionFailsWhenMissingMarketContractAddress(t *testing.T) {
	executor := newTestCircleCLIExecutor(&fakeCommandRunner{}, true)
	intent := confirmedIntent(ActionCloseMarket)
	intent.MarketContractAddress = ""

	_, err := executor.ExecuteCloseMarket(context.Background(), intent)
	if !errors.Is(err, ErrIntentInvalid) {
		t.Fatalf("expected invalid intent, got %v", err)
	}
}

func TestCircleCLIExecutorLifecycleActionFailsWhenNotAllowed(t *testing.T) {
	executor := newTestCircleCLIExecutor(&fakeCommandRunner{}, true)
	intent := confirmedIntent(ActionCancelMarket)
	intent.AllowedActions = []string{ActionBuyYes}

	_, err := executor.ExecuteCancelMarket(context.Background(), intent)
	if !errors.Is(err, ErrIntentInvalid) {
		t.Fatalf("expected invalid intent, got %v", err)
	}
}

func TestCircleCLIExecutorLifecycleDisabledFailsClosed(t *testing.T) {
	executor := newTestCircleCLIExecutor(&fakeCommandRunner{}, false)

	_, err := executor.ExecuteClaimRefund(context.Background(), confirmedIntent(ActionClaimRefund))
	if !errors.Is(err, ErrExecutionProviderDisabled) {
		t.Fatalf("expected disabled error, got %v", err)
	}
}

func TestCircleCLIExecutorLifecycleReadbackErrorIsSanitized(t *testing.T) {
	runner := &fakeCommandRunner{
		outputs: [][]byte{[]byte(`{"transactionHash":"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`)},
		errAt:   2,
	}
	executor := newTestCircleCLIExecutor(runner, true)

	_, err := executor.ExecuteCloseMarket(context.Background(), confirmedIntent(ActionCloseMarket))
	if err == nil {
		t.Fatal("expected readback error")
	}
	if err.Error() == "raw sensitive CLI error" {
		t.Fatal("expected sanitized error")
	}
}

func TestCircleCLIExecutorLifecycleDoesNotUsePrivateKeyFallback(t *testing.T) {
	runner := &fakeCommandRunner{outputs: lifecycleRefundOutputs(
		"0x2222222222222222222222222222222222222222222222222222222222222222",
		"3",
		"2000000",
		true,
		"0",
	)}
	executor := newTestCircleCLIExecutor(runner, true)

	_, err := executor.ExecuteClaimRefund(context.Background(), confirmedIntent(ActionClaimRefund))
	if err != nil {
		t.Fatalf("execute claim refund: %v", err)
	}

	for _, call := range runner.calls {
		if call.name != "circle" {
			t.Fatalf("expected circle command only, got %q", call.name)
		}
		joined := strings.Join(call.args, " ")
		if strings.Contains(strings.ToLower(joined), "private") || strings.Contains(joined, "AGENT_EXECUTOR_PRIVATE_KEY") {
			t.Fatalf("unexpected private key fallback in args: %#v", call.args)
		}
	}
}

func TestCircleCLIExecutorReadbackErrorIsSanitized(t *testing.T) {
	runner := &fakeCommandRunner{
		outputs: [][]byte{[]byte(`{"transactionHash":"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`)},
		errAt:   2,
	}
	executor := newTestCircleCLIExecutor(runner, true)

	_, err := executor.ExecuteCreateMarket(context.Background(), confirmedIntent(ActionCreateMarket))
	if err == nil {
		t.Fatal("expected readback error")
	}
	if err.Error() == "raw sensitive CLI error" {
		t.Fatal("expected sanitized error")
	}
}

func TestCircleCLIExecutorDecodesABIHexScalars(t *testing.T) {
	uintValue, err := decodeUint256Scalar("0x0000000000000000000000000000000000000000000000000000000000000008")
	if err != nil {
		t.Fatalf("decode uint256: %v", err)
	}
	if uintValue != "8" {
		t.Fatalf("expected 8, got %q", uintValue)
	}

	index, err := previousIndex("0x0000000000000000000000000000000000000000000000000000000000000008")
	if err != nil {
		t.Fatalf("previous index: %v", err)
	}
	if index != "7" {
		t.Fatalf("expected previous index 7, got %q", index)
	}

	address, err := decodeAddressScalar("0x000000000000000000000000abcf081e456c1a11106def590666a07b76d456f8")
	if err != nil {
		t.Fatalf("decode address: %v", err)
	}
	if address != "0xabcF081E456c1A11106DeF590666A07B76d456f8" {
		t.Fatalf("unexpected address %q", address)
	}

	boolean, err := decodeBoolScalar("0x0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Fatalf("decode bool: %v", err)
	}
	if !boolean {
		t.Fatal("expected true")
	}
}

func TestCircleCLIExecutorCreateMarketStaleCloseTimestampReturnsError(t *testing.T) {
	runner := &fakeCommandRunner{outputs: [][]byte{
		[]byte(`{"transactionHash":"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`),
	}}
	executor := newTestCircleCLIExecutor(runner, true)
	intent := confirmedIntent(ActionCreateMarket)
	// Set close_timestamp to a value in the past
	intent.CloseTimestamp = pastTimestamp()

	_, err := executor.ExecuteCreateMarket(context.Background(), intent)
	if !errors.Is(err, ErrCreateMarketCloseTimestampStale) {
		t.Fatalf("expected ErrCreateMarketCloseTimestampStale, got %v", err)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("expected no CLI calls for stale timestamp, got %d", len(runner.calls))
	}
}

func TestCircleCLIExecutorCreateMarketFutureCloseTimestampProceeds(t *testing.T) {
	runner := &fakeCommandRunner{outputs: [][]byte{
		[]byte(`{"transactionHash":"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`),
		jsonResult(abiUint256("9")),
		jsonResult(abiAddress("0xabcf081e456c1a11106def590666a07b76d456f8")),
		jsonResult(abiUint256("1")),
	}}
	executor := newTestCircleCLIExecutor(runner, true)
	intent := confirmedIntent(ActionCreateMarket)
	// Set close_timestamp 2 hours in the future (well beyond the 60s margin)
	intent.CloseTimestamp = strconv.FormatInt(time.Now().Add(2*time.Hour).Unix(), 10)

	result, err := executor.ExecuteCreateMarket(context.Background(), intent)
	if err != nil {
		t.Fatalf("expected success for future timestamp, got %v", err)
	}
	if len(runner.calls) == 0 {
		t.Fatal("expected CLI calls for future timestamp")
	}
	if result.Status != StatusExecuted {
		t.Fatalf("expected executed status, got %q", result.Status)
	}
}

func TestCircleCLIExecutorBuyYesNotAffectedByCloseTimestampGuard(t *testing.T) {
	runner := &fakeCommandRunner{outputs: buyOutputs(
		"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"1000000", "1000000", "1000000", "1000000",
	)}
	executor := newTestCircleCLIExecutor(runner, true)
	intent := confirmedIntent(ActionBuyYes)
	// Even with a stale close_timestamp, buy_yes should not be affected
	intent.CloseTimestamp = pastTimestamp()

	result, err := executor.ExecuteBuyYes(context.Background(), intent)
	if err != nil {
		t.Fatalf("expected buy_yes to succeed regardless of close_timestamp, got %v", err)
	}
	if result.Status != StatusExecuted {
		t.Fatalf("expected executed status, got %q", result.Status)
	}
}

func newTestCircleCLIExecutor(runner *fakeCommandRunner, enabled bool) *CircleCLIExecutor {
	return NewCircleCLIExecutor(CircleCLIExecutorConfig{
		Enabled:       enabled,
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		Timeout:       time.Second,
		AgentFactory:  AgentFactoryAddress,
		CommandRunner: runner,
	})
}

func confirmedIntent(action string) Intent {
	return Intent{
		ID:                    "agent_intent_test",
		AgentID:               "agent_test",
		AgentWalletAddress:    "0x9999999999999999999999999999999999999999",
		WalletProvider:        WalletProviderCircleAgentWallet,
		AllowedActions:        []string{ActionCreateMarket, ActionBuyYes, ActionBuyNo, ActionCloseMarket, ActionResolveMarket, ActionClaimPayout, ActionCancelMarket, ActionClaimRefund},
		Action:                action,
		Status:                StatusConfirmed,
		UserWallet:            "0x1111111111111111111111111111111111111111",
		MarketID:              "market-1",
		MarketContractAddress: "0x3333333333333333333333333333333333333333",
		Amount:                "1000000",
		Resolver:              "",
		CollateralToken:       ArcTestnetUSDCAddress,
		CloseTimestamp:        futureTimestamp(),
		Question:              "Will SignalArc execute through Circle?",
		ValidationResult:      ValidationResult{Valid: true, Errors: []string{}},
	}
}

// futureTimestamp returns a Unix-seconds string 1 hour in the future.
func futureTimestamp() string {
	return strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10)
}

// pastTimestamp returns a Unix-seconds string 1 hour in the past.
func pastTimestamp() string {
	return strconv.FormatInt(time.Now().Add(-time.Hour).Unix(), 10)
}

func lifecyclePayoutOutputs(txHash string, status string, winningOutcome string, claimablePayout string, hasClaimed bool, balance string) [][]byte {
	return [][]byte{
		[]byte(`{"transactionHash":"` + txHash + `"}`),
		jsonResult(abiUint256(status)),
		jsonResult(abiUint256(winningOutcome)),
		jsonResult(abiUint256(claimablePayout)),
		jsonResult(abiBool(hasClaimed)),
		jsonResult(abiUint256(balance)),
	}
}

func lifecycleRefundOutputs(txHash string, status string, claimableRefund string, hasClaimed bool, balance string) [][]byte {
	return [][]byte{
		[]byte(`{"transactionHash":"` + txHash + `"}`),
		jsonResult(abiUint256(status)),
		jsonResult(abiUint256(claimableRefund)),
		jsonResult(abiBool(hasClaimed)),
		jsonResult(abiUint256(balance)),
	}
}

func buyOutputs(approveHash string, buyHash string, position string, total string, collateral string, balance string) [][]byte {
	return [][]byte{
		[]byte(`{"transactionHash":"` + approveHash + `"}`),
		[]byte(`{"transactionHash":"` + buyHash + `"}`),
		jsonResult(abiUint256(position)),
		jsonResult(abiUint256(total)),
		jsonResult(abiUint256(collateral)),
		jsonResult(abiUint256(balance)),
	}
}

func jsonResult(value string) []byte {
	return []byte(`{"result":"` + value + `"}`)
}

func abiUint256(decimal string) string {
	value, ok := new(big.Int).SetString(decimal, 10)
	if !ok {
		panic("invalid test decimal")
	}
	return "0x" + fmt.Sprintf("%064x", value)
}

func abiAddress(address string) string {
	trimmed := strings.TrimPrefix(strings.TrimPrefix(address, "0x"), "0X")
	return "0x" + strings.Repeat("0", 64-len(trimmed)) + trimmed
}

func abiBool(value bool) string {
	if value {
		return abiUint256("1")
	}
	return abiUint256("0")
}

func assertArgs(t *testing.T, actual []string, expected []string) {
	t.Helper()
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected args %#v, got %#v", expected, actual)
	}
}
