package agent

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestSanitizeExecuteOutputRedactsEmail(t *testing.T) {
	input := "Error: failed for user@example.com on ARC-TESTNET"
	result := SanitizeExecuteOutput(input)
	if strings.Contains(result, "user@example.com") {
		t.Fatalf("email not redacted: %q", result)
	}
	if !strings.Contains(result, "[email-redacted]") {
		t.Fatalf("expected [email-redacted] placeholder, got %q", result)
	}
}

func TestSanitizeExecuteOutputRedactsOTP(t *testing.T) {
	input := "circle wallet verify --otp 123456 --request abc123"
	result := SanitizeExecuteOutput(input)
	if strings.Contains(result, "123456") {
		t.Fatalf("OTP not redacted: %q", result)
	}
}

func TestSanitizeExecuteOutputRedactsRequestID(t *testing.T) {
	input := "request_id: 8f61f53b-4b61-4e75-adff-5b6338234a15"
	result := SanitizeExecuteOutput(input)
	if strings.Contains(result, "8f61f53b-4b61-4e75-adff-5b6338234a15") {
		t.Fatalf("request ID not redacted: %q", result)
	}
}

func TestSanitizeExecuteOutputRedactsBearerToken(t *testing.T) {
	input := "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.abc123"
	result := SanitizeExecuteOutput(input)
	if strings.Contains(result, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9") {
		t.Fatalf("bearer token not redacted: %q", result)
	}
}

func TestSanitizeExecuteOutputRedactsPrivateKey(t *testing.T) {
	input := "-----BEGIN EC PRIVATE KEY-----\nMHQCAQEEIBkg...\n-----END EC PRIVATE KEY-----"
	result := SanitizeExecuteOutput(input)
	if strings.Contains(result, "MHQCAQEEIBkg") {
		t.Fatalf("private key not redacted: %q", result)
	}
	if !strings.Contains(result, "[key-redacted]") {
		t.Fatalf("expected [key-redacted] placeholder, got %q", result)
	}
}

func TestSanitizeExecuteOutputTruncatesLongOutput(t *testing.T) {
	// Generate output longer than maxSanitizedSummaryLen
	long := strings.Repeat("error details from contract revert ", 100)
	result := SanitizeExecuteOutput(long)
	if len(result) > maxSanitizedSummaryLen+10 { // +10 for "..." suffix
		t.Fatalf("output not truncated: length %d", len(result))
	}
	if !strings.HasSuffix(result, "...") {
		t.Fatalf("expected truncation marker, got suffix %q", result[len(result)-10:])
	}
}

func TestSanitizeExecuteOutputPreservesTransactionHash(t *testing.T) {
	txHash := "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	input := "Error: transaction " + txHash + " reverted"
	result := SanitizeExecuteOutput(input)
	if !strings.Contains(result, txHash) {
		t.Fatalf("transaction hash was stripped: %q", result)
	}
}

func TestSanitizeExecuteOutputRedactsBareHexSecret(t *testing.T) {
	bareHex := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	input := "data " + bareHex + " leaked"
	result := SanitizeExecuteOutput(input)
	if strings.Contains(result, bareHex) {
		t.Fatalf("bare hex secret not redacted: %q", result)
	}
	if !strings.Contains(result, "[hex-redacted]") {
		t.Fatalf("expected [hex-redacted] placeholder, got %q", result)
	}
}

func TestSanitizeExecuteOutputPreservesContractRevertMessage(t *testing.T) {
	input := `Error: execution reverted: ERC20: transfer amount exceeds balance`
	result := SanitizeExecuteOutput(input)
	if !strings.Contains(result, "execution reverted") {
		t.Fatalf("revert message lost: %q", result)
	}
	if !strings.Contains(result, "transfer amount exceeds balance") {
		t.Fatalf("revert reason lost: %q", result)
	}
}

func TestSanitizeExecuteOutputReturnsEmptyForBlankInput(t *testing.T) {
	if result := SanitizeExecuteOutput(""); result != "" {
		t.Fatalf("expected empty, got %q", result)
	}
	if result := SanitizeExecuteOutput("   "); result != "" {
		t.Fatalf("expected empty for whitespace, got %q", result)
	}
}

func TestRedactAddressPartiallyRedacts(t *testing.T) {
	address := "0x69aE770e8b2F96297101FeC4dc123B3801dA7d80"
	result := RedactAddress(address)
	if result != "0x69aE...7d80" {
		t.Fatalf("unexpected redacted address: %q", result)
	}
}

func TestRedactAddressReturnsEmptyForInvalid(t *testing.T) {
	if result := RedactAddress(""); result != "" {
		t.Fatalf("expected empty for empty input, got %q", result)
	}
	if result := RedactAddress("short"); result != "" {
		t.Fatalf("expected empty for short input, got %q", result)
	}
	if result := RedactAddress("not-a-hex-address-at-all"); result != "" {
		t.Fatalf("expected empty for non-hex, got %q", result)
	}
}

func TestClassifyExecuteCommandCategory(t *testing.T) {
	cases := []struct {
		args     []string
		expected string
	}{
		{[]string{"wallet", "execute", "buyYes(uint256)"}, "wallet_execute"},
		{[]string{"contract", "query", "status()"}, "contract_query"},
		{[]string{"wallet", "balance", "--address", "0x123"}, "wallet_balance"},
		{[]string{"wallet", "list", "--type", "agent"}, "wallet_list"},
		{[]string{"unknown"}, "unknown"},
		{[]string{}, "unknown"},
	}
	for _, tc := range cases {
		result := ClassifyExecuteCommandCategory(tc.args)
		if result != tc.expected {
			t.Fatalf("args %v: expected %q, got %q", tc.args, tc.expected, result)
		}
	}
}

func TestExtractExitStatus(t *testing.T) {
	if status := ExtractExitStatus("exit status 1"); status != "1" {
		t.Fatalf("expected 1, got %q", status)
	}
	if status := ExtractExitStatus("exit status 127"); status != "127" {
		t.Fatalf("expected 127, got %q", status)
	}
	if status := ExtractExitStatus("some other error"); status != "" {
		t.Fatalf("expected empty, got %q", status)
	}
}

func TestCircleExecuteContextFromError(t *testing.T) {
	execCtx := &ExecuteContext{
		Action:            "buy_yes",
		FunctionSignature: "buyYes(uint256)",
		ContractAddress:   "0x3333...3333",
		WalletAddress:     "0x9999...9999",
		Chain:             "ARC-TESTNET",
		CommandCategory:   "wallet_execute",
		ExitStatus:        "1",
	}
	err := &CircleCLIError{
		Operation:        "circle_cli_exec",
		ErrorClass:       CircleErrorClassUnknown,
		SanitizedSummary: "execution reverted",
		Err:              errors.New("Circle CLI command failed"),
		ExecCtx:          execCtx,
	}

	extracted := CircleExecuteContextFromError(err)
	if extracted == nil {
		t.Fatal("expected non-nil ExecuteContext")
	}
	if extracted.Action != "buy_yes" {
		t.Fatalf("expected action buy_yes, got %q", extracted.Action)
	}
	if extracted.FunctionSignature != "buyYes(uint256)" {
		t.Fatalf("expected function signature, got %q", extracted.FunctionSignature)
	}
	if extracted.ExitStatus != "1" {
		t.Fatalf("expected exit status 1, got %q", extracted.ExitStatus)
	}
}

func TestCircleExecuteContextFromErrorReturnsNilForNonCLIError(t *testing.T) {
	err := errors.New("some other error")
	if ctx := CircleExecuteContextFromError(err); ctx != nil {
		t.Fatalf("expected nil for non-CLI error, got %#v", ctx)
	}
}

func TestCircleExecuteContextFromErrorReturnsNilForNil(t *testing.T) {
	if ctx := CircleExecuteContextFromError(nil); ctx != nil {
		t.Fatalf("expected nil for nil error, got %#v", ctx)
	}
}

func TestExecuteFailurePreservesPublicBehavior(t *testing.T) {
	// Simulate a Circle CLI execute failure with stderr output containing
	// a contract revert. Verify that:
	// 1. The public error message stays generic
	// 2. The error class is extractable
	// 3. The sanitized summary retains diagnostic info
	// 4. Raw CLI output is not in the public error string
	runner := &fakeCommandRunnerWithOutput{
		output: []byte(`Error: execution reverted: ERC20: transfer amount exceeds balance
at ContractRunner.execute (node_modules/@circle/cli/dist/runner.js:42:11)`),
		err: errors.New("exit status 1"),
	}
	executor := NewCircleCLIExecutor(CircleCLIExecutorConfig{
		Enabled:       true,
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		Timeout:       1,
		AgentFactory:  AgentFactoryAddress,
		CommandRunner: runner,
	})

	_, err := executor.ExecuteBuyYes(context.Background(), confirmedIntent(ActionBuyYes))
	if err == nil {
		t.Fatal("expected error")
	}

	// Public error message must stay generic
	if err.Error() != "Circle CLI command failed" {
		t.Fatalf("public error message changed: %q", err.Error())
	}

	// Error class must be extractable
	errorClass := CircleErrorClassFromError(err)
	if errorClass != CircleErrorClassUnknown {
		t.Fatalf("expected unknown class for revert, got %q", errorClass)
	}

	// Sanitized summary must retain diagnostic info
	summary := CircleErrorSummaryFromError(err)
	if !strings.Contains(summary, "execution reverted") {
		t.Fatalf("sanitized summary lost revert info: %q", summary)
	}
	if !strings.Contains(summary, "transfer amount exceeds balance") {
		t.Fatalf("sanitized summary lost revert reason: %q", summary)
	}

	// ExecuteContext must be populated
	execCtx := CircleExecuteContextFromError(err)
	if execCtx == nil {
		t.Fatal("expected ExecuteContext on execute failure")
	}
	if execCtx.FunctionSignature != "approve(address,uint256)" {
		t.Fatalf("expected approve signature (first call), got %q", execCtx.FunctionSignature)
	}
	if execCtx.CommandCategory != "wallet_execute" {
		t.Fatalf("expected wallet_execute category, got %q", execCtx.CommandCategory)
	}
	if execCtx.Chain != ChainArcTestnet {
		t.Fatalf("expected ARC-TESTNET chain, got %q", execCtx.Chain)
	}
	if execCtx.ExitStatus != "1" {
		t.Fatalf("expected exit status 1, got %q", execCtx.ExitStatus)
	}
}

func TestExecuteFailureAuthRequiredClassificationStillWorks(t *testing.T) {
	runner := &fakeCommandRunnerWithOutput{
		output: []byte("Error: AUTH_REQUIRED\nNo local wallet matches 0xa9914bca9123ba0079be8c968f632c0db6400fe7 on ARC-TESTNET, and no agent session is active.\nRun `circle wallet login user@example.com --type agent`"),
		err:    errors.New("exit status 1"),
	}
	executor := NewCircleCLIExecutor(CircleCLIExecutorConfig{
		Enabled:       true,
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		Timeout:       1,
		AgentFactory:  AgentFactoryAddress,
		CommandRunner: runner,
	})

	_, err := executor.ExecuteBuyYes(context.Background(), confirmedIntent(ActionBuyYes))
	if err == nil {
		t.Fatal("expected error")
	}

	// AUTH_REQUIRED classification must still work
	errorClass := CircleErrorClassFromError(err)
	if errorClass != CircleErrorClassAuthRequired {
		t.Fatalf("expected auth_required class, got %q", errorClass)
	}

	// Email must be redacted in sanitized summary
	summary := CircleErrorSummaryFromError(err)
	if strings.Contains(summary, "user@example.com") {
		t.Fatalf("email not redacted in summary: %q", summary)
	}

	// Public error message must stay generic
	if err.Error() != "Circle CLI command failed" {
		t.Fatalf("public error message changed: %q", err.Error())
	}
}

func TestSanitizedCLISummaryIsRetainedForLogs(t *testing.T) {
	// Verify that when Circle CLI returns a meaningful error (contract revert,
	// gas estimation failure, etc.), the sanitized summary preserves enough
	// diagnostic information for log analysis.
	cases := []struct {
		name     string
		output   string
		contains string
	}{
		{
			name:     "contract_revert",
			output:   "Error: execution reverted: Market is not open",
			contains: "Market is not open",
		},
		{
			name:     "gas_estimation",
			output:   "Error: cannot estimate gas; transaction may fail or may require manual gas limit",
			contains: "cannot estimate gas",
		},
		{
			name:     "insufficient_funds",
			output:   "Error: insufficient funds for gas * price + value",
			contains: "insufficient funds",
		},
		{
			name:     "nonce_too_low",
			output:   "Error: nonce too low: next nonce 5, tx nonce 3",
			contains: "nonce too low",
		},
		{
			name:     "function_selector",
			output:   "Error: execution reverted: unrecognized function selector 0xdeadbeef",
			contains: "unrecognized function selector",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeExecuteOutput(tc.output)
			if !strings.Contains(result, tc.contains) {
				t.Fatalf("diagnostic info lost: expected %q in %q", tc.contains, result)
			}
		})
	}
}

func TestSensitiveValuesAreRedacted(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		banned string
	}{
		{
			name:   "email_in_error",
			input:  "failed for admin@signalarc.com",
			banned: "admin@signalarc.com",
		},
		{
			name:   "bearer_token",
			input:  "Bearer eyJhbGciOiJSUzI1NiJ9.payload.signature",
			banned: "eyJhbGciOiJSUzI1NiJ9",
		},
		{
			name:   "otp_in_command",
			input:  "circle wallet verify --otp 847291",
			banned: "847291",
		},
		{
			name:   "request_id_in_output",
			input:  "request_id: secret-req-id-12345",
			banned: "secret-req-id-12345",
		},
		{
			name:   "token_assignment",
			input:  "token=sk_live_abc123def456",
			banned: "sk_live_abc123def456",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeExecuteOutput(tc.input)
			if strings.Contains(result, tc.banned) {
				t.Fatalf("sensitive value not redacted: %q still in %q", tc.banned, result)
			}
		})
	}
}

func TestBuildExecuteSummaryPrefersOutputOverProcessError(t *testing.T) {
	summary := BuildExecuteSummary(
		"Error: execution reverted: Market is not open",
		errors.New("exit status 1"),
	)
	if !strings.Contains(summary, "Market is not open") {
		t.Fatalf("expected CLI output in summary, got %q", summary)
	}
	if strings.Contains(summary, "process_error") {
		t.Fatalf("process_error prefix should not appear when CLI output is available: %q", summary)
	}
}

func TestBuildExecuteSummaryFallsBackToProcessError(t *testing.T) {
	summary := BuildExecuteSummary("", errors.New("signal: killed"))
	if !strings.Contains(summary, "process_error:") {
		t.Fatalf("expected process_error prefix, got %q", summary)
	}
	if !strings.Contains(summary, "signal: killed") {
		t.Fatalf("expected process error content, got %q", summary)
	}
}

func TestBuildExecuteSummaryEmptyOutputAndNilError(t *testing.T) {
	summary := BuildExecuteSummary("", nil)
	if summary != "Circle CLI command failed with no output" {
		t.Fatalf("expected fallback message, got %q", summary)
	}
}

func TestBuildExecuteSummaryRedactsProcessError(t *testing.T) {
	summary := BuildExecuteSummary("", errors.New("exec failed for user@example.com: exit status 1"))
	if strings.Contains(summary, "user@example.com") {
		t.Fatalf("email not redacted in process error summary: %q", summary)
	}
	if !strings.Contains(summary, "[email-redacted]") {
		t.Fatalf("expected redacted placeholder in summary: %q", summary)
	}
}

func TestExecuteFailureEmptyOutputCapturesProcessError(t *testing.T) {
	// Simulate a Circle CLI failure where the process crashes with no output.
	runner := &fakeCommandRunnerWithOutput{
		output: []byte{},
		err:    errors.New("signal: killed"),
	}
	executor := NewCircleCLIExecutor(CircleCLIExecutorConfig{
		Enabled:       true,
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		Timeout:       1,
		AgentFactory:  AgentFactoryAddress,
		CommandRunner: runner,
	})

	_, err := executor.ExecuteBuyYes(context.Background(), confirmedIntent(ActionBuyYes))
	if err == nil {
		t.Fatal("expected error")
	}

	// Public error message must stay generic
	if err.Error() != "Circle CLI command failed" {
		t.Fatalf("public error message changed: %q", err.Error())
	}

	// ExecuteContext must carry process error and raw output len
	execCtx := CircleExecuteContextFromError(err)
	if execCtx == nil {
		t.Fatal("expected ExecuteContext")
	}
	if execCtx.RawOutputLen != 0 {
		t.Fatalf("expected raw output len 0, got %d", execCtx.RawOutputLen)
	}
	if execCtx.ProcessError == "" {
		t.Fatal("expected process error to be populated when output is empty")
	}
	if !strings.Contains(execCtx.ProcessError, "signal: killed") {
		t.Fatalf("expected process error to contain signal info, got %q", execCtx.ProcessError)
	}

	// Summary should include process error info
	summary := CircleErrorSummaryFromError(err)
	if !strings.Contains(summary, "process_error") {
		t.Fatalf("expected process_error in summary when output is empty, got %q", summary)
	}
}

func TestExecuteFailureWithOutputCapturesRawOutputLen(t *testing.T) {
	cliOutput := "Error: execution reverted: ERC20: transfer amount exceeds balance"
	runner := &fakeCommandRunnerWithOutput{
		output: []byte(cliOutput),
		err:    errors.New("exit status 1"),
	}
	executor := NewCircleCLIExecutor(CircleCLIExecutorConfig{
		Enabled:       true,
		CLIPath:       "circle",
		Chain:         ChainArcTestnet,
		Timeout:       1,
		AgentFactory:  AgentFactoryAddress,
		CommandRunner: runner,
	})

	_, err := executor.ExecuteBuyYes(context.Background(), confirmedIntent(ActionBuyYes))
	if err == nil {
		t.Fatal("expected error")
	}

	execCtx := CircleExecuteContextFromError(err)
	if execCtx == nil {
		t.Fatal("expected ExecuteContext")
	}
	if execCtx.RawOutputLen != len(cliOutput) {
		t.Fatalf("expected raw output len %d, got %d", len(cliOutput), execCtx.RawOutputLen)
	}
	// ProcessError should be empty when CLI output is available
	if execCtx.ProcessError != "" {
		t.Fatalf("expected empty process error when CLI output is available, got %q", execCtx.ProcessError)
	}
}

// fakeCommandRunnerWithOutput is a test double that always returns the same
// output and error, simulating a Circle CLI failure with specific stderr.
type fakeCommandRunnerWithOutput struct {
	output []byte
	err    error
}

func (r *fakeCommandRunnerWithOutput) Run(_ context.Context, _ string, args []string) ([]byte, error) {
	if r.err != nil {
		rawOutput := string(r.output)
		errorClass := ClassifyCircleErrorOutput(rawOutput, r.err.Error())
		sanitizedSummary := BuildExecuteSummary(rawOutput, r.err)
		processError := ""
		if strings.TrimSpace(rawOutput) == "" {
			processError = SanitizeExecuteOutput(r.err.Error())
		}
		return r.output, &CircleCLIError{
			Operation:        "circle_cli_exec",
			ErrorClass:       errorClass,
			SanitizedSummary: sanitizedSummary,
			Err:              errors.New("Circle CLI command failed"),
			ExecCtx: &ExecuteContext{
				CommandCategory: ClassifyExecuteCommandCategory(args),
				ExitStatus:      ExtractExitStatus(r.err.Error()),
				Chain:           extractChainFromArgs(args),
				RawOutputLen:    len(r.output),
				ProcessError:    processError,
			},
		}
	}
	return r.output, nil
}
