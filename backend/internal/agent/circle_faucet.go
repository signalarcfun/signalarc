package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// FaucetTokenUSDC is the only token requested by the SignalArc faucet helper.
// SignalArc never accepts arbitrary token names from API callers; this is fixed
// to USDC on ARC-TESTNET because the documented Circle CLI testnet faucet
// targets the testnet USDC asset on Arc Testnet.
const FaucetTokenUSDC = "usdc"

// FaucetStatusRequested marks that SignalArc forwarded a faucet request to the
// provider. SignalArc does not claim provider-side faucet success; callers must
// rely on the provider response embedded under result.
const FaucetStatusRequested = "requested"

// ErrCircleAgentWalletFaucetFailed is the only error the faucet runner returns
// to the API layer. Callers must not surface its underlying CLI/runner output
// because that output may contain Circle credential paths, request IDs, or
// other secret-like material.
var ErrCircleAgentWalletFaucetFailed = errors.New("Circle Agent Wallet faucet request failed")

// ErrCircleAgentWalletFaucetNotConfigured signals that the SignalArc faucet
// helper is not enabled in this runtime. The API layer maps this to HTTP 501
// circle_agent_wallet_faucet_not_configured.
var ErrCircleAgentWalletFaucetNotConfigured = errors.New("Circle Agent Wallet faucet is not configured")

// CircleAgentWalletFaucetResult is the parsed result of a Circle CLI testnet
// faucet command. Either JSON or Message will be populated, never both. JSON
// holds parsed CLI JSON output (which may be prefixed by Node deprecation
// warnings; the parser strips that prefix). Message holds sanitized text-only
// success output when the CLI returned non-JSON text on success.
type CircleAgentWalletFaucetResult struct {
	JSON    any
	Message string
}

// CircleAgentWalletFaucet is the read-only faucet capability exposed to
// SignalArc HTTP handlers. Implementations must only target the registered
// agent wallet address provided by the backend, must only target ARC-TESTNET,
// and must not perform transfers, swaps, contract execution, or mainnet
// funding.
type CircleAgentWalletFaucet interface {
	RequestFaucet(context.Context, string) (CircleAgentWalletFaucetResult, error)
}

// CircleCLIFaucetRunnerConfig configures the disabled-by-default Circle CLI
// faucet runner. Enabled must be true for the runner to actually invoke the
// Circle CLI. Chain must be ARC-TESTNET; the runner refuses any other chain
// because Circle only exposes a documented testnet faucet on ARC-TESTNET.
type CircleCLIFaucetRunnerConfig struct {
	Enabled       bool
	CLIPath       string
	Chain         string
	Timeout       time.Duration
	CommandRunner EnvCommandRunner
}

// CircleCLIFaucetRunner runs `circle wallet fund --address <addr> --chain
// ARC-TESTNET --token usdc --output json` and parses the warning-prefixed CLI
// output the same way the wallet list/balance helpers do.
//
// The runner deliberately does not pass --amount, --method, --open, --export,
// transfer, swap, execute, or any mainnet funding option. The token is fixed
// to usdc and the chain is fixed to ARC-TESTNET so external callers can never
// extend the command shape through the API.
type CircleCLIFaucetRunner struct {
	cfg CircleCLIFaucetRunnerConfig
}

// NewCircleCLIFaucetRunner returns a CircleCLIFaucetRunner with safe defaults
// (CLI path "circle", chain ARC-TESTNET, 120s timeout, real exec runner).
func NewCircleCLIFaucetRunner(cfg CircleCLIFaucetRunnerConfig) *CircleCLIFaucetRunner {
	if strings.TrimSpace(cfg.CLIPath) == "" {
		cfg.CLIPath = "circle"
	}
	if strings.TrimSpace(cfg.Chain) == "" {
		cfg.Chain = ChainArcTestnet
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 120 * time.Second
	}
	if cfg.CommandRunner == nil {
		cfg.CommandRunner = ExecEnvCommandRunner{}
	}
	return &CircleCLIFaucetRunner{cfg: cfg}
}

// Enabled reports whether the runner will actually call the Circle CLI.
func (runner *CircleCLIFaucetRunner) IsEnabled() bool {
	return runner != nil && runner.cfg.Enabled
}

// RequestFaucet runs the Circle CLI testnet faucet command for the registered
// agent wallet address. It accepts only ARC-TESTNET and only the USDC token
// and never sees any caller-supplied recipient.
func (runner *CircleCLIFaucetRunner) RequestFaucet(parent context.Context, address string) (CircleAgentWalletFaucetResult, error) {
	if runner == nil {
		return CircleAgentWalletFaucetResult{}, ErrCircleAgentWalletFaucetNotConfigured
	}
	if !runner.cfg.Enabled {
		return CircleAgentWalletFaucetResult{}, ErrCircleAgentWalletFaucetNotConfigured
	}
	if runner.cfg.Chain != ChainArcTestnet {
		return CircleAgentWalletFaucetResult{}, ErrCircleAgentWalletFaucetFailed
	}
	address = strings.TrimSpace(address)
	if address == "" {
		return CircleAgentWalletFaucetResult{}, ErrCircleAgentWalletFaucetFailed
	}

	ctx, cancel := context.WithTimeout(parent, runner.cfg.Timeout)
	defer cancel()

	args := []string{
		"wallet", "fund",
		"--address", address,
		"--chain", ChainArcTestnet,
		"--token", FaucetTokenUSDC,
		"--output", "json",
	}
	output, err := runner.cfg.CommandRunner.RunWithEnv(ctx, runner.cfg.CLIPath, args, []string{"CIRCLE_ACCEPT_TERMS=1"})
	if err != nil {
		runner.logFaucetFailure(string(output), err.Error(), address)
		return CircleAgentWalletFaucetResult{}, ErrCircleAgentWalletFaucetFailed
	}

	result, parseErr := parseCircleFaucetOutput(output)
	if parseErr != nil {
		runner.logFaucetFailure(string(output), parseErr.Error(), address)
		return CircleAgentWalletFaucetResult{}, ErrCircleAgentWalletFaucetFailed
	}
	return result, nil
}

func (runner *CircleCLIFaucetRunner) logFaucetFailure(output string, errText string, address string) {
	sanitizedOutput := sanitizeCircleOnboardingText(output, address)
	sanitizedError := sanitizeCircleOnboardingText(errText, address)
	log.Error().
		Str("operation", "circle_agent_wallet_faucet").
		Str("output", sanitizedOutput).
		Str("error", sanitizedError).
		Msg("Circle CLI faucet request failed")
}

// parseCircleFaucetOutput parses Circle CLI faucet output. It first tries to
// extract a JSON object/array (skipping any Node deprecation warning prefix);
// if successful, it returns the parsed JSON under JSON. Otherwise, if the CLI
// produced any non-empty text, it returns sanitized text under Message. Empty
// output is treated as a parse failure so callers translate it to a generic
// faucet failure error.
func parseCircleFaucetOutput(output []byte) (CircleAgentWalletFaucetResult, error) {
	if cleaned, extractErr := extractJSONFromCLIOutput(output); extractErr == nil {
		var decoded any
		if jsonErr := json.Unmarshal(cleaned, &decoded); jsonErr == nil {
			return CircleAgentWalletFaucetResult{JSON: decoded}, nil
		}
	}
	text := strings.TrimSpace(string(output))
	if text == "" {
		return CircleAgentWalletFaucetResult{}, errors.New("circle faucet output is empty")
	}
	// Text-only success: Circle credential paths, request IDs, OTP material, and
	// emails should never appear in faucet success output, but redact the same
	// patterns the OTP sanitizer covers so any unexpected secret-like substring
	// is removed before SignalArc surfaces it to the API caller.
	return CircleAgentWalletFaucetResult{Message: sanitizeCircleOnboardingText(text)}, nil
}
