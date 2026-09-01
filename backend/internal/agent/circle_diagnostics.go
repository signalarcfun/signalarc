package agent

import (
	"context"
	"errors"
	"strings"
)

// CircleCLIError is a sanitized wrapper for Circle CLI failures. It carries:
//   - the operation name (for structured logs)
//   - a coarse error class (auth_required / unknown), classified by
//     ClassifyCircleErrorOutput against documented AUTH_REQUIRED markers
//   - a sanitized one-line summary safe to log (PII-redacted)
//   - the underlying public sentinel (e.g. ErrCircleAgentWalletBalanceFailed)
//   - optional ExecuteContext with non-sensitive metadata about the invocation
//
// Public HTTP error codes returned by handlers must remain unchanged. This
// type exists so handlers can errors.As() it and emit structured log details
// without leaking raw CLI stderr/stdout, email, OTP, request_id, or session
// tokens.
type CircleCLIError struct {
	Operation        string
	ErrorClass       string
	SanitizedSummary string
	Err              error
	// ExecCtx carries non-sensitive execution metadata for structured logging.
	// It is nil for non-execute operations (balance, list, etc.).
	ExecCtx *ExecuteContext
}

type circleDiagnosticError interface {
	ErrorClass() string
	SanitizedSummary() string
}

func (err *CircleCLIError) Error() string {
	if err == nil {
		return ""
	}
	if err.Err != nil {
		return err.Err.Error()
	}
	return "Circle CLI command failed"
}

func (err *CircleCLIError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

// CircleErrorClassFromError extracts the coarse Circle error class from any
// error that wraps a *CircleCLIError. It returns CircleErrorClassUnknown when
// no class is available.
func CircleErrorClassFromError(err error) string {
	if err == nil {
		return ""
	}
	var diagnosticErr circleDiagnosticError
	if errors.As(err, &diagnosticErr) && diagnosticErr != nil {
		if errorClass := strings.TrimSpace(diagnosticErr.ErrorClass()); errorClass != "" {
			return errorClass
		}
	}
	var cliErr *CircleCLIError
	if errors.As(err, &cliErr) && cliErr != nil {
		if strings.TrimSpace(cliErr.ErrorClass) != "" {
			return cliErr.ErrorClass
		}
	}
	return CircleErrorClassUnknown
}

// CircleErrorSummaryFromError extracts the sanitized summary, if any.
func CircleErrorSummaryFromError(err error) string {
	if err == nil {
		return ""
	}
	var diagnosticErr circleDiagnosticError
	if errors.As(err, &diagnosticErr) && diagnosticErr != nil {
		return diagnosticErr.SanitizedSummary()
	}
	var cliErr *CircleCLIError
	if errors.As(err, &cliErr) && cliErr != nil {
		return cliErr.SanitizedSummary
	}
	return ""
}

// CircleExecuteContextFromError extracts the ExecuteContext from any error
// that wraps a *CircleCLIError. Returns nil when no context is available.
func CircleExecuteContextFromError(err error) *ExecuteContext {
	if err == nil {
		return nil
	}
	var cliErr *CircleCLIError
	if errors.As(err, &cliErr) && cliErr != nil {
		return cliErr.ExecCtx
	}
	return nil
}

// AgentSessionLivenessState marks whether the Circle CLI agent session is
// usable on this backend instance for the given agent wallet address.
type AgentSessionLivenessState string

const (
	// AgentSessionLivenessLive means the Circle CLI returned a parseable
	// agent wallet list that contains the registered agent wallet address.
	AgentSessionLivenessLive AgentSessionLivenessState = "live"
	// AgentSessionLivenessAuthRequired means the CLI output indicated
	// AUTH_REQUIRED or that the registered agent wallet is not present in
	// the local Circle CLI agent wallet list on this instance.
	AgentSessionLivenessAuthRequired AgentSessionLivenessState = "auth_required"
	// AgentSessionLivenessUnknown means the CLI failed for a reason that
	// is not classified as AUTH_REQUIRED.
	AgentSessionLivenessUnknown AgentSessionLivenessState = "unknown"
)

// AgentSessionLivenessResult is what backend handlers receive when they
// perform a liveness probe through the Circle CLI wallet resolver. It is
// safe to surface to API callers; nothing here contains raw CLI stdout/
// stderr, email, OTP, request_id, or session tokens.
type AgentSessionLivenessResult struct {
	State      AgentSessionLivenessState
	ErrorClass string
	// Reason is a sanitized, user-facing message. It is populated when
	// State is not AgentSessionLivenessLive.
	Reason string
}

// CircleAgentSessionLivenessChecker is implemented by Circle CLI wallet
// resolvers that can probe local CLI session liveness. The API layer uses
// a type assertion so existing test doubles without this method continue
// to compile.
type CircleAgentSessionLivenessChecker interface {
	CheckAgentSessionLiveness(ctx context.Context, agentWalletAddress string) AgentSessionLivenessResult
}
