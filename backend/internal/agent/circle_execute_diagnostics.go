package agent

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// maxSanitizedSummaryLen is the maximum byte length for sanitized CLI output
// summaries stored in structured log fields. This prevents unbounded log
// entries from large CLI stderr/stdout.
const maxSanitizedSummaryLen = 512

// ExecuteContext carries non-sensitive metadata about a Circle CLI execute
// invocation. It is attached to CircleCLIError so that structured logging at
// the handler layer can emit actionable diagnostics without exposing raw CLI
// output, secrets, or PII to API callers.
type ExecuteContext struct {
	// Action is the intent action (e.g. "create_market", "buy_yes").
	Action string
	// FunctionSignature is the Solidity function signature invoked
	// (e.g. "buyYes(uint256)").
	FunctionSignature string
	// ContractAddress is the target contract, partially redacted.
	ContractAddress string
	// WalletAddress is the agent wallet, partially redacted.
	WalletAddress string
	// Chain is the target chain (e.g. "ARC-TESTNET").
	Chain string
	// CommandCategory is the CLI command category ("wallet_execute" or
	// "contract_query") without exposing the full argument list.
	CommandCategory string
	// ExitStatus is the process exit code string when available.
	ExitStatus string
	// RawOutputLen is the byte length of the raw CLI combined output before
	// sanitization. Zero means the CLI produced no stdout/stderr at all,
	// which distinguishes "empty output" from "output was sanitized away".
	RawOutputLen int
	// ProcessError is the sanitized Go-level process error (e.g.
	// "signal: killed", "exec: \"circle\": executable file not found").
	// It is populated only when the raw CLI output is empty or unhelpful,
	// so the operator can distinguish process-level failures from
	// application-level CLI errors.
	ProcessError string
}

// SanitizeExecuteOutput produces a truncated, PII-redacted summary of Circle
// CLI combined stdout/stderr suitable for structured logging. It:
//   - Strips email-like patterns
//   - Strips OTP codes
//   - Strips Circle request IDs
//   - Strips bearer/auth tokens
//   - Strips private key material
//   - Collapses whitespace
//   - Truncates to maxSanitizedSummaryLen bytes on a clean UTF-8 boundary
//
// The result is safe to include in zerolog fields but must NOT be returned
// to API callers.
func SanitizeExecuteOutput(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}

	sanitized := raw

	// Redact email addresses
	sanitized = emailPattern.ReplaceAllString(sanitized, "[email-redacted]")

	// Redact OTP codes (6-digit sequences that look like OTPs in context)
	sanitized = redactCirclePatternCapture(sanitized, circleOTPCommandPattern)

	// Redact Circle request IDs
	sanitized = redactCirclePatternCapture(sanitized, circleRequestIDCommandPattern)
	sanitized = redactCirclePatternCapture(sanitized, circleRequestIDLinePattern)

	// Redact bearer/auth tokens
	sanitized = bearerTokenPattern.ReplaceAllString(sanitized, "Bearer [token-redacted]")
	sanitized = authTokenPattern.ReplaceAllString(sanitized, "${1}[token-redacted]")

	// Redact private key material
	sanitized = privateKeyPattern.ReplaceAllString(sanitized, "[key-redacted]")

	// Redact hex strings that look like secrets (64+ hex chars not preceded by 0x
	// in a transaction hash context)
	sanitized = longHexSecretPattern.ReplaceAllStringFunc(sanitized, func(match string) string {
		// Preserve transaction hashes (0x + 64 hex) and ABI-encoded values
		if strings.HasPrefix(match, "0x") || strings.HasPrefix(match, "0X") {
			return match
		}
		return "[hex-redacted]"
	})

	// Collapse whitespace
	sanitized = multiWhitespacePattern.ReplaceAllString(sanitized, " ")
	sanitized = strings.TrimSpace(sanitized)

	// Truncate to safe length on UTF-8 boundary
	sanitized = truncateUTF8(sanitized, maxSanitizedSummaryLen)

	return sanitized
}

// RedactAddress partially redacts an EVM address for logging.
// "0x1234567890abcdef1234567890abcdef12345678" becomes "0x1234...5678".
// Returns empty string for invalid input.
func RedactAddress(address string) string {
	address = strings.TrimSpace(address)
	if len(address) < 10 {
		return ""
	}
	if !strings.HasPrefix(address, "0x") && !strings.HasPrefix(address, "0X") {
		return ""
	}
	// Show first 6 chars (0x + 4 hex) and last 4 hex chars
	return address[:6] + "..." + address[len(address)-4:]
}

// ClassifyExecuteCommandCategory returns a safe command category string
// from CLI args without exposing the full argument list.
func ClassifyExecuteCommandCategory(args []string) string {
	if len(args) < 2 {
		return "unknown"
	}
	switch {
	case args[0] == "wallet" && args[1] == "execute":
		return "wallet_execute"
	case args[0] == "contract" && args[1] == "query":
		return "contract_query"
	case args[0] == "wallet" && args[1] == "balance":
		return "wallet_balance"
	case args[0] == "wallet" && args[1] == "list":
		return "wallet_list"
	default:
		return args[0] + "_" + args[1]
	}
}

// ExtractExitStatus extracts a process exit status from an error string.
// Returns the exit status portion or empty string if not found.
func ExtractExitStatus(errText string) string {
	if match := exitStatusPattern.FindStringSubmatch(errText); len(match) == 2 {
		return match[1]
	}
	return ""
}

// Regex patterns for sanitization.
var (
	emailPattern           = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	bearerTokenPattern     = regexp.MustCompile(`(?i)Bearer\s+[A-Za-z0-9\-._~+/]+=*`)
	authTokenPattern       = regexp.MustCompile(`(?i)(token|secret|key|password|authorization)\s*[:=]\s*[^\s,"'{}]+`)
	privateKeyPattern      = regexp.MustCompile(`(?i)(-----BEGIN[^-]*PRIVATE KEY-----[\s\S]*?-----END[^-]*PRIVATE KEY-----)`)
	longHexSecretPattern   = regexp.MustCompile(`(?:0[xX])?[0-9a-fA-F]{64,}\b`)
	multiWhitespacePattern = regexp.MustCompile(`\s+`)
	exitStatusPattern      = regexp.MustCompile(`exit status (\d+)`)
)

// BuildExecuteSummary constructs a sanitized diagnostic summary from CLI
// output and the process-level error. It prefers the CLI output when
// available, but falls back to the process error when the CLI produced no
// useful output. This prevents the generic "Circle CLI command failed"
// summary when the process error itself carries actionable information
// (e.g. "signal: killed", "exec: not found", "context deadline exceeded").
func BuildExecuteSummary(rawOutput string, processErr error) string {
	sanitized := SanitizeExecuteOutput(rawOutput)
	if sanitized != "" {
		return sanitized
	}
	// CLI produced no useful output; try the process error.
	if processErr != nil {
		errText := processErr.Error()
		// Sanitize the error text too (it could contain paths or env info).
		sanitizedErr := SanitizeExecuteOutput(errText)
		if sanitizedErr != "" {
			return "process_error: " + sanitizedErr
		}
	}
	return "Circle CLI command failed with no output"
}

// truncateUTF8 truncates a string to at most maxBytes bytes, ensuring the
// result is valid UTF-8 (does not split a multi-byte character).
func truncateUTF8(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	// Walk backwards from maxBytes to find a valid UTF-8 boundary
	truncated := s[:maxBytes]
	for !utf8.ValidString(truncated) && len(truncated) > 0 {
		truncated = truncated[:len(truncated)-1]
	}
	if len(truncated) < len(s) {
		truncated += "..."
	}
	return truncated
}
