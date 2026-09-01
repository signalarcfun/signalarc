package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	ErrCircleAgentWalletNotFound            = errors.New("Circle Agent Wallet not found")
	ErrCircleAgentWalletResolutionAmbiguous = errors.New("Circle Agent Wallet resolution ambiguous")
	ErrCircleAgentWalletResolutionFailed    = errors.New("Circle Agent Wallet resolution failed")
	ErrCircleAgentWalletBalanceFailed       = errors.New("Circle Agent Wallet balance lookup failed")
	// ErrCircleCLIAuthRequired is a classification sentinel returned (joined
	// alongside the existing public sentinels) when the Circle CLI command
	// output indicates the local CLI agent session is missing or expired.
	// Public error codes returned by HTTP handlers are intentionally not
	// changed; this sentinel is used only by handlers that want to log a
	// structured error_class or surface a non-misleading liveness status.
	ErrCircleCLIAuthRequired = errors.New("Circle CLI agent session not available")
)

const (
	// CircleErrorClassAuthRequired marks Circle CLI output that contains the
	// documented AUTH_REQUIRED marker or the equivalent "no agent session is
	// active" / "circle wallet login" guidance. Detection is text-based
	// because the Circle CLI does not currently document a stable structured
	// error code on stderr; matching is intentionally narrow and tolerant of
	// case so other failures fall back to CircleErrorClassUnknown.
	CircleErrorClassAuthRequired = "auth_required"
	// CircleErrorClassUnknown is the default error class for Circle CLI
	// failures that did not match the AUTH_REQUIRED markers above.
	CircleErrorClassUnknown = "unknown"
)

// circleAuthRequiredMarkers lists case-insensitive substrings that indicate
// the Circle CLI agent session is missing on the host. These match the
// production AUTH_REQUIRED error message snippet observed in Cloud Run logs.
// Behavior beyond these markers is unknown / not documented.
var circleAuthRequiredMarkers = []string{
	"AUTH_REQUIRED",
	"no agent session is active",
	"no local wallet matches",
	"circle wallet login",
}

// ClassifyCircleErrorOutput returns CircleErrorClassAuthRequired when any of
// the supplied texts (typically combined CLI stdout/stderr and the underlying
// error string) contains one of the documented AUTH_REQUIRED markers, and
// CircleErrorClassUnknown otherwise.
func ClassifyCircleErrorOutput(texts ...string) string {
	for _, text := range texts {
		lowered := strings.ToLower(text)
		for _, marker := range circleAuthRequiredMarkers {
			if strings.Contains(lowered, strings.ToLower(marker)) {
				return CircleErrorClassAuthRequired
			}
		}
	}
	return CircleErrorClassUnknown
}

var evmAddressPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

type CircleAgentWallet struct {
	Address string
	Chain   string
}

type CircleAgentWalletBalances struct {
	Balances []any
}

type CircleAgentWalletBalanceRequest struct {
	AgentID            string
	AgentWalletAddress string
	WalletProvider     string
	Chain              string
	PolicyMetadata     map[string]string
}

type CircleAgentWalletBalanceReader interface {
	GetAgentWalletBalances(context.Context, CircleAgentWalletBalanceRequest) (CircleAgentWalletBalances, error)
}

type CircleWalletResolver interface {
	ResolveAgentWallet(context.Context, string) (CircleAgentWallet, error)
	GetAgentWalletBalances(context.Context, string) (CircleAgentWalletBalances, error)
}

type CircleCLIWalletResolverConfig struct {
	CLIPath       string
	Chain         string
	Timeout       time.Duration
	CommandRunner EnvCommandRunner
}

type CircleCLIWalletResolver struct {
	cfg CircleCLIWalletResolverConfig
}

func NewCircleCLIWalletResolver(cfg CircleCLIWalletResolverConfig) *CircleCLIWalletResolver {
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
	return &CircleCLIWalletResolver{cfg: cfg}
}

func (resolver *CircleCLIWalletResolver) ResolveAgentWallet(parent context.Context, email string) (CircleAgentWallet, error) {
	if resolver == nil || resolver.cfg.Chain != ChainArcTestnet {
		return CircleAgentWallet{}, ErrCircleAgentWalletResolutionFailed
	}
	ctx, cancel := context.WithTimeout(parent, resolver.cfg.Timeout)
	defer cancel()

	args := []string{"wallet", "list", "--type", "agent", "--chain", ChainArcTestnet, "--output", "json"}
	output, err := resolver.cfg.CommandRunner.RunWithEnv(ctx, resolver.cfg.CLIPath, args, []string{"CIRCLE_ACCEPT_TERMS=1"})
	if err != nil {
		resolver.logResolutionFailure("circle_agent_wallet_list", string(output), err.Error(), email, "")
		return CircleAgentWallet{}, ErrCircleAgentWalletResolutionFailed
	}

	wallets, parseErr := parseCircleAgentWallets(output)
	if parseErr != nil {
		resolver.logResolutionFailure("circle_agent_wallet_list", string(output), parseErr.Error(), email, "")
		return CircleAgentWallet{}, ErrCircleAgentWalletResolutionFailed
	}
	switch len(wallets) {
	case 0:
		resolver.logResolutionFailure("circle_agent_wallet_list", string(output), "", email, "")
		return CircleAgentWallet{}, ErrCircleAgentWalletNotFound
	case 1:
		return wallets[0], nil
	default:
		resolver.logResolutionFailure("circle_agent_wallet_list", string(output), "multiple ARC-TESTNET agent wallets returned; exact onboarding match is unknown / not documented", email, "")
		return CircleAgentWallet{}, ErrCircleAgentWalletResolutionAmbiguous
	}
}

func (resolver *CircleCLIWalletResolver) GetAgentWalletBalances(parent context.Context, address string) (CircleAgentWalletBalances, error) {
	if resolver == nil || resolver.cfg.Chain != ChainArcTestnet || strings.TrimSpace(address) == "" {
		return CircleAgentWalletBalances{}, ErrCircleAgentWalletBalanceFailed
	}
	ctx, cancel := context.WithTimeout(parent, resolver.cfg.Timeout)
	defer cancel()

	address = strings.TrimSpace(address)
	args := []string{"wallet", "balance", "--address", address, "--chain", ChainArcTestnet, "--output", "json"}
	output, err := resolver.cfg.CommandRunner.RunWithEnv(ctx, resolver.cfg.CLIPath, args, []string{"CIRCLE_ACCEPT_TERMS=1"})
	if err != nil {
		resolver.logResolutionFailure("circle_agent_wallet_balance", string(output), err.Error(), "", address)
		return CircleAgentWalletBalances{}, classifyCircleCLIBalanceError(string(output), err.Error())
	}

	balances, parseErr := parseCircleAgentWalletBalances(output)
	if parseErr != nil {
		resolver.logResolutionFailure("circle_agent_wallet_balance", string(output), parseErr.Error(), "", address)
		return CircleAgentWalletBalances{}, classifyCircleCLIBalanceError(string(output), parseErr.Error())
	}
	return CircleAgentWalletBalances{Balances: balances}, nil
}

// CheckAgentSessionLiveness probes the local Circle CLI agent wallet list to
// determine whether the host filesystem state required to operate the
// registered agent wallet is reachable from this backend instance.
//
// SignalArc never persists Circle CLI session state. This helper exists so
// the API layer can avoid returning a misleading "active" status when the
// Cloud Run instance currently handling the request has no Circle CLI
// session. It is intentionally a read-only command and shares the same path
// already used by ResolveAgentWallet.
func (resolver *CircleCLIWalletResolver) CheckAgentSessionLiveness(parent context.Context, agentWalletAddress string) AgentSessionLivenessResult {
	if resolver == nil {
		return AgentSessionLivenessResult{
			State:      AgentSessionLivenessUnknown,
			ErrorClass: CircleErrorClassUnknown,
			Reason:     "Circle CLI wallet resolver is not configured on this backend instance",
		}
	}
	if resolver.cfg.Chain != ChainArcTestnet {
		return AgentSessionLivenessResult{
			State:      AgentSessionLivenessUnknown,
			ErrorClass: CircleErrorClassUnknown,
			Reason:     "Circle CLI chain is not ARC-TESTNET",
		}
	}
	address := strings.TrimSpace(agentWalletAddress)
	if address == "" {
		return AgentSessionLivenessResult{
			State:      AgentSessionLivenessUnknown,
			ErrorClass: CircleErrorClassUnknown,
			Reason:     "agent wallet address is required for liveness probe",
		}
	}

	ctx, cancel := context.WithTimeout(parent, resolver.cfg.Timeout)
	defer cancel()

	args := []string{"wallet", "list", "--type", "agent", "--chain", ChainArcTestnet, "--output", "json"}
	output, err := resolver.cfg.CommandRunner.RunWithEnv(ctx, resolver.cfg.CLIPath, args, []string{"CIRCLE_ACCEPT_TERMS=1"})
	if err != nil {
		errorClass := ClassifyCircleErrorOutput(string(output), err.Error())
		resolver.logResolutionFailure("circle_agent_session_liveness", string(output), err.Error(), "", address)
		if errorClass == CircleErrorClassAuthRequired {
			return AgentSessionLivenessResult{
				State:      AgentSessionLivenessAuthRequired,
				ErrorClass: CircleErrorClassAuthRequired,
				Reason:     "Circle CLI agent session is not active on this backend instance; OTP onboarding must be re-run",
			}
		}
		return AgentSessionLivenessResult{
			State:      AgentSessionLivenessUnknown,
			ErrorClass: CircleErrorClassUnknown,
			Reason:     "Circle CLI agent wallet liveness probe failed",
		}
	}

	wallets, parseErr := parseCircleAgentWallets(output)
	if parseErr != nil {
		errorClass := ClassifyCircleErrorOutput(string(output), parseErr.Error())
		resolver.logResolutionFailure("circle_agent_session_liveness", string(output), parseErr.Error(), "", address)
		if errorClass == CircleErrorClassAuthRequired {
			return AgentSessionLivenessResult{
				State:      AgentSessionLivenessAuthRequired,
				ErrorClass: CircleErrorClassAuthRequired,
				Reason:     "Circle CLI agent session is not active on this backend instance; OTP onboarding must be re-run",
			}
		}
		return AgentSessionLivenessResult{
			State:      AgentSessionLivenessUnknown,
			ErrorClass: CircleErrorClassUnknown,
			Reason:     "Circle CLI agent wallet liveness probe returned unparseable output",
		}
	}

	for _, wallet := range wallets {
		if strings.EqualFold(wallet.Address, address) {
			return AgentSessionLivenessResult{State: AgentSessionLivenessLive}
		}
	}

	resolver.logResolutionFailure("circle_agent_session_liveness", string(output), "registered agent wallet not present in local Circle CLI list", "", address)
	return AgentSessionLivenessResult{
		State:      AgentSessionLivenessAuthRequired,
		ErrorClass: CircleErrorClassAuthRequired,
		Reason:     "registered agent wallet is not present in the local Circle CLI agent wallet list on this backend instance",
	}
}

// classifyCircleCLIBalanceError preserves the existing public sentinel
// (ErrCircleAgentWalletBalanceFailed) while attaching a sanitized error
// class for structured logs. Handlers that wish to log more context can
// extract the class with CircleErrorClassFromError.
func classifyCircleCLIBalanceError(output string, errText string) error {
	return &CircleCLIError{
		Operation:        "circle_agent_wallet_balance",
		ErrorClass:       ClassifyCircleErrorOutput(output, errText),
		SanitizedSummary: sanitizeCircleOnboardingText(errText),
		Err:              ErrCircleAgentWalletBalanceFailed,
	}
}

func (resolver *CircleCLIWalletResolver) logResolutionFailure(operation string, output string, errText string, email string, address string) {
	sanitizedOutput := sanitizeCircleOnboardingText(output, email, address)
	sanitizedError := sanitizeCircleOnboardingText(errText, email, address)
	log.Error().
		Str("operation", operation).
		Str("output", sanitizedOutput).
		Str("error", sanitizedError).
		Msg("Circle CLI wallet read failed")
}

// extractJSONFromCLIOutput finds the first JSON object or array in raw CLI
// output that may be prefixed with non-JSON text such as Node.js deprecation
// warnings. It returns the byte slice starting at the first '{' or '[' that
// successfully parses as JSON. If no valid JSON envelope is found, it returns
// nil and an error.
func extractJSONFromCLIOutput(raw []byte) ([]byte, error) {
	// Fast path: if the output starts with JSON, return as-is.
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		var probe json.RawMessage
		if json.Unmarshal(trimmed, &probe) == nil {
			return trimmed, nil
		}
	}

	// Scan for each '{' or '[' and try to parse from that position.
	for i := 0; i < len(raw); i++ {
		if raw[i] != '{' && raw[i] != '[' {
			continue
		}
		candidate := raw[i:]
		var probe json.RawMessage
		if json.Unmarshal(candidate, &probe) == nil {
			return candidate, nil
		}
	}

	return nil, errors.New("no valid JSON object or array found in CLI output")
}

func parseCircleAgentWallets(output []byte) ([]CircleAgentWallet, error) {
	cleaned, extractErr := extractJSONFromCLIOutput(output)
	if extractErr != nil {
		return nil, extractErr
	}
	var decoded any
	if err := json.Unmarshal(cleaned, &decoded); err != nil {
		return nil, err
	}
	seen := map[string]CircleAgentWallet{}
	collectCircleAgentWallets(decoded, seen)
	wallets := make([]CircleAgentWallet, 0, len(seen))
	for _, wallet := range seen {
		wallets = append(wallets, wallet)
	}
	return wallets, nil
}

func collectCircleAgentWallets(value any, seen map[string]CircleAgentWallet) {
	switch typed := value.(type) {
	case map[string]any:
		if address := circleWalletAddressFromMap(typed); address != "" {
			chain := circleWalletChainFromMap(typed)
			if chain == "" || strings.EqualFold(chain, ChainArcTestnet) {
				seen[strings.ToLower(address)] = CircleAgentWallet{Address: address, Chain: ChainArcTestnet}
			}
		}
		for _, child := range typed {
			collectCircleAgentWallets(child, seen)
		}
	case []any:
		for _, child := range typed {
			collectCircleAgentWallets(child, seen)
		}
	}
}

func circleWalletAddressFromMap(value map[string]any) string {
	for _, key := range []string{"address", "walletAddress", "wallet_address"} {
		if text, ok := scalarToString(value[key]); ok && evmAddressPattern.MatchString(text) {
			return text
		}
	}
	return ""
}

func circleWalletChainFromMap(value map[string]any) string {
	for _, key := range []string{"chain", "network", "blockchain"} {
		if text, ok := scalarToString(value[key]); ok {
			return text
		}
	}
	return ""
}

func parseCircleAgentWalletBalances(output []byte) ([]any, error) {
	cleaned, extractErr := extractJSONFromCLIOutput(output)
	if extractErr != nil {
		return nil, extractErr
	}
	var decoded any
	if err := json.Unmarshal(cleaned, &decoded); err != nil {
		return nil, err
	}
	balances, ok := findBalancesArray(decoded)
	if !ok {
		return nil, errors.New("balances array not found")
	}
	return balances, nil
}

func findBalancesArray(value any) ([]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		if balances, ok := typed["balances"].([]any); ok {
			return balances, true
		}
		for _, child := range typed {
			if balances, ok := findBalancesArray(child); ok {
				return balances, true
			}
		}
	case []any:
		for _, child := range typed {
			if balances, ok := findBalancesArray(child); ok {
				return balances, true
			}
		}
	}
	return nil, false
}
