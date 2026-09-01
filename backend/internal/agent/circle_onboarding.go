package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	ErrCircleOnboardingDisabled                = errors.New("Circle Agent Wallet OTP onboarding start is disabled")
	ErrCircleOnboardingRequestIDNotDocumented  = errors.New("Circle CLI OTP start request_id field is unknown / not documented")
	ErrCircleOnboardingRequestIDMissing        = errors.New("Circle CLI OTP start request_id not found")
	ErrCircleOnboardingRequestIDEmpty          = errors.New("Circle CLI OTP start request_id is empty")
	ErrCircleOnboardingRequestIDNotAvailable   = errors.New("Circle OTP request is not available; backend restart requires onboarding restart")
	ErrCircleOnboardingCommandFailed           = errors.New("Circle CLI OTP start command failed")
	ErrCircleOnboardingCommandReturnedNoOutput = errors.New("Circle CLI OTP start command returned empty output")
	ErrCircleOnboardingUnsupportedChain        = errors.New("Circle Agent Wallet onboarding chain must be ARC-TESTNET")
	ErrCircleOnboardingEmailRequired           = errors.New("user_email is required for Circle Agent Wallet OTP onboarding")
	ErrCircleOnboardingOTPRequired             = errors.New("otp is required for Circle Agent Wallet OTP verification")
)

type CircleOTPStartResult struct {
	RequestID string
	ExpiresAt time.Time
}

type CircleOnboardingRunner interface {
	StartOTP(context.Context, string) (CircleOTPStartResult, error)
	VerifyOTP(context.Context, string, string) error
}

type CircleOnboardingStarter struct {
	Enabled      bool
	Runner       CircleOnboardingRunner
	RequestStore *CircleOTPRequestStore
}

func (starter CircleOnboardingStarter) StartOTP(ctx context.Context, onboardingID string, email string) (CircleOTPStartResult, string, bool, error) {
	if !starter.Enabled {
		return CircleOTPStartResult{}, "", false, ErrCircleOnboardingDisabled
	}
	if starter.Runner == nil {
		return CircleOTPStartResult{}, "", true, ErrCircleOnboardingCommandFailed
	}

	result, err := starter.Runner.StartOTP(ctx, email)
	if err != nil {
		return CircleOTPStartResult{}, "", true, err
	}
	requestID := strings.TrimSpace(result.RequestID)
	if requestID == "" {
		return CircleOTPStartResult{}, "", true, ErrCircleOnboardingRequestIDEmpty
	}
	if result.ExpiresAt.IsZero() {
		result.ExpiresAt = time.Now().UTC().Add(10 * time.Minute)
	}
	result.RequestID = requestID
	requestIDHash := HashCircleRequestID(requestID)
	if starter.RequestStore != nil {
		starter.RequestStore.Save(onboardingID, requestID)
	}
	return result, requestIDHash, true, nil
}

func (starter CircleOnboardingStarter) VerifyOTP(ctx context.Context, onboardingID string, otp string) (bool, error) {
	if !starter.Enabled {
		return false, ErrCircleOnboardingDisabled
	}
	if starter.Runner == nil {
		return true, ErrCircleOnboardingCommandFailed
	}
	otp = strings.TrimSpace(otp)
	if otp == "" {
		return true, ErrCircleOnboardingOTPRequired
	}
	requestID, ok := starter.RequestStore.Get(onboardingID)
	if !ok {
		return true, ErrCircleOnboardingRequestIDNotAvailable
	}
	if err := starter.Runner.VerifyOTP(ctx, requestID, otp); err != nil {
		return true, err
	}
	return true, nil
}

type CircleOTPRequestStore struct {
	mu                      sync.RWMutex
	requestIDByOnboardingID map[string]string
}

func NewCircleOTPRequestStore() *CircleOTPRequestStore {
	return &CircleOTPRequestStore{requestIDByOnboardingID: map[string]string{}}
}

func (store *CircleOTPRequestStore) Save(onboardingID string, requestID string) {
	if store == nil {
		return
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	store.requestIDByOnboardingID[strings.TrimSpace(onboardingID)] = strings.TrimSpace(requestID)
}

func (store *CircleOTPRequestStore) Get(onboardingID string) (string, bool) {
	if store == nil {
		return "", false
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	requestID, ok := store.requestIDByOnboardingID[strings.TrimSpace(onboardingID)]
	return requestID, ok
}

func (store *CircleOTPRequestStore) Delete(onboardingID string) {
	if store == nil {
		return
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.requestIDByOnboardingID, strings.TrimSpace(onboardingID))
}

func HashCircleRequestID(requestID string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(requestID)))
	return hex.EncodeToString(sum[:])
}

type CircleCLIOnboardingRunnerConfig struct {
	CLIPath       string
	Chain         string
	Timeout       time.Duration
	CommandRunner EnvCommandRunner
}

type EnvCommandRunner interface {
	RunWithEnv(context.Context, string, []string, []string) ([]byte, error)
}

type ExecEnvCommandRunner struct{}

func (runner ExecEnvCommandRunner) RunWithEnv(ctx context.Context, name string, args []string, env []string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Env = append(os.Environ(), env...)
	output, err := command.CombinedOutput()
	if err != nil {
		return output, ErrCircleOnboardingCommandFailed
	}
	return output, nil
}

type CircleOnboardingCommandError struct {
	Operation       string
	SanitizedOutput string
	Err             error
}

func (err *CircleOnboardingCommandError) Error() string {
	if err == nil {
		return ErrCircleOnboardingCommandFailed.Error()
	}
	if strings.TrimSpace(err.SanitizedOutput) == "" {
		return ErrCircleOnboardingCommandFailed.Error()
	}
	return ErrCircleOnboardingCommandFailed.Error() + ": " + err.SanitizedOutput
}

func (err *CircleOnboardingCommandError) Unwrap() error {
	if err == nil || err.Err == nil {
		return ErrCircleOnboardingCommandFailed
	}
	return err.Err
}

var (
	circleRequestIDLinePattern    = regexp.MustCompile(`(?im)\brequest[\s_-]*id\b\s*[:=]\s*([^\s,"'{}]+)`)
	circleRequestIDCommandPattern = regexp.MustCompile(`(?im)--request\s+([^\s,"'{}]+)`)
	circleOTPCommandPattern       = regexp.MustCompile(`(?im)--otp\s+([^\s,"'{}]+)`)
)

type CircleCLIOnboardingRunner struct {
	cfg CircleCLIOnboardingRunnerConfig
}

func NewCircleCLIOnboardingRunner(cfg CircleCLIOnboardingRunnerConfig) *CircleCLIOnboardingRunner {
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
	return &CircleCLIOnboardingRunner{cfg: cfg}
}

func (runner *CircleCLIOnboardingRunner) StartOTP(parent context.Context, email string) (CircleOTPStartResult, error) {
	if runner == nil {
		return CircleOTPStartResult{}, ErrCircleOnboardingCommandFailed
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return CircleOTPStartResult{}, ErrCircleOnboardingEmailRequired
	}
	if runner.cfg.Chain != ChainArcTestnet {
		return CircleOTPStartResult{}, ErrCircleOnboardingUnsupportedChain
	}

	ctx, cancel := context.WithTimeout(parent, runner.cfg.Timeout)
	defer cancel()

	args := []string{"wallet", "login", email, "--init", "--type", "agent", "--testnet", "--output", "json"}
	output, err := runner.cfg.CommandRunner.RunWithEnv(ctx, runner.cfg.CLIPath, args, []string{"CIRCLE_ACCEPT_TERMS=1"})
	if len(output) == 0 {
		if err != nil {
			return CircleOTPStartResult{}, runner.circleOnboardingStartError(err, output, email)
		}
		runner.logCircleOTPStartFailure("", "", email)
		return CircleOTPStartResult{}, ErrCircleOnboardingCommandReturnedNoOutput
	}

	requestID, ok := findCircleOTPRequestID(output)
	if err != nil {
		if ok && strings.TrimSpace(requestID) != "" {
			runner.logCircleOTPStartFailure(string(output), err.Error(), email)
			return CircleOTPStartResult{
				RequestID: strings.TrimSpace(requestID),
				ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
			}, nil
		}
		return CircleOTPStartResult{}, runner.circleOnboardingStartError(err, output, email)
	}
	if !ok {
		runner.logCircleOTPStartFailure(string(output), "", email)
		return CircleOTPStartResult{}, ErrCircleOnboardingRequestIDNotDocumented
	}
	if strings.TrimSpace(requestID) == "" {
		runner.logCircleOTPStartFailure(string(output), "", email)
		return CircleOTPStartResult{}, ErrCircleOnboardingRequestIDMissing
	}

	return CircleOTPStartResult{
		RequestID: strings.TrimSpace(requestID),
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
	}, nil
}

func (runner *CircleCLIOnboardingRunner) circleOnboardingStartError(err error, output []byte, email string) error {
	sanitizedOutput := runner.logCircleOTPStartFailure(string(output), err.Error(), email)
	return &CircleOnboardingCommandError{
		Operation:       "circle_otp_start",
		SanitizedOutput: sanitizedOutput,
		Err:             ErrCircleOnboardingCommandFailed,
	}
}

func (runner *CircleCLIOnboardingRunner) logCircleOTPStartFailure(output string, errText string, email string) string {
	requestID, _ := findCircleOTPRequestID([]byte(output))
	sanitizedOutput := sanitizeCircleOnboardingText(output, requestID, email)
	sanitizedError := sanitizeCircleOnboardingText(errText, requestID, email)
	log.Error().
		Str("operation", "circle_otp_start").
		Str("output", sanitizedOutput).
		Str("error", sanitizedError).
		Msg("Circle CLI OTP start failed")
	return sanitizedOutput
}

func (runner *CircleCLIOnboardingRunner) VerifyOTP(parent context.Context, requestID string, otp string) error {
	if runner == nil {
		return ErrCircleOnboardingCommandFailed
	}
	requestID = strings.TrimSpace(requestID)
	otp = strings.TrimSpace(otp)
	if requestID == "" {
		return ErrCircleOnboardingRequestIDEmpty
	}
	if otp == "" {
		return ErrCircleOnboardingOTPRequired
	}
	if runner.cfg.Chain != ChainArcTestnet {
		return ErrCircleOnboardingUnsupportedChain
	}

	ctx, cancel := context.WithTimeout(parent, runner.cfg.Timeout)
	defer cancel()

	args := []string{"wallet", "login", "--request", requestID, "--otp", otp}
	output, err := runner.cfg.CommandRunner.RunWithEnv(ctx, runner.cfg.CLIPath, args, []string{"CIRCLE_ACCEPT_TERMS=1"})
	if err != nil {
		sanitizedOutput := sanitizeCircleOnboardingText(string(output), requestID, otp)
		sanitizedError := sanitizeCircleOnboardingText(err.Error(), requestID, otp)
		log.Error().
			Str("operation", "circle_otp_verify").
			Str("output", sanitizedOutput).
			Str("error", sanitizedError).
			Msg("Circle CLI OTP verify failed")
		return &CircleOnboardingCommandError{
			Operation:       "circle_otp_verify",
			SanitizedOutput: sanitizedOutput,
			Err:             ErrCircleOnboardingCommandFailed,
		}
	}
	return nil
}

func sanitizeCircleOnboardingText(value string, secrets ...string) string {
	sanitized := value
	for _, secret := range secrets {
		secret = strings.TrimSpace(secret)
		if secret == "" {
			continue
		}
		sanitized = strings.ReplaceAll(sanitized, secret, "[redacted]")
	}
	sanitized = redactCirclePatternCapture(sanitized, circleRequestIDCommandPattern)
	sanitized = redactCirclePatternCapture(sanitized, circleRequestIDLinePattern)
	sanitized = redactCirclePatternCapture(sanitized, circleOTPCommandPattern)
	return strings.TrimSpace(sanitized)
}

func redactCirclePatternCapture(value string, pattern *regexp.Regexp) string {
	return pattern.ReplaceAllStringFunc(value, func(match string) string {
		parts := pattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		return strings.Replace(match, parts[1], "[redacted]", 1)
	})
}

func findCircleOTPRequestID(output []byte) (string, bool) {
	var decoded any
	if err := json.Unmarshal(output, &decoded); err == nil {
		// Circle documents that --init returns a request ID, but the exact JSON
		// field name is unknown / not documented in the official pages reviewed.
		if requestID, ok := findCircleOTPRequestIDField(decoded); ok {
			return requestID, true
		}
		if requestID, ok := findCircleOTPRequestIDInJSONStrings(decoded); ok {
			return requestID, true
		}
	}
	return findCircleOTPRequestIDFromText(string(output))
}

func findCircleOTPRequestIDField(value any) (string, bool) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"request_id", "requestId", "requestID", "id"} {
			if child, ok := typed[key]; ok {
				if text, ok := scalarToString(child); ok {
					return text, true
				}
			}
		}
		for _, child := range typed {
			if text, ok := findCircleOTPRequestIDField(child); ok {
				return text, true
			}
		}
	case []any:
		for _, child := range typed {
			if text, ok := findCircleOTPRequestIDField(child); ok {
				return text, true
			}
		}
	}
	return "", false
}

func findCircleOTPRequestIDInJSONStrings(value any) (string, bool) {
	switch typed := value.(type) {
	case map[string]any:
		for _, child := range typed {
			if requestID, ok := findCircleOTPRequestIDInJSONStrings(child); ok {
				return requestID, true
			}
		}
	case []any:
		for _, child := range typed {
			if requestID, ok := findCircleOTPRequestIDInJSONStrings(child); ok {
				return requestID, true
			}
		}
	case string:
		return findCircleOTPRequestIDFromText(typed)
	}
	return "", false
}

func findCircleOTPRequestIDFromText(output string) (string, bool) {
	output = strings.TrimSpace(output)
	if output == "" {
		return "", false
	}
	if match := circleRequestIDCommandPattern.FindStringSubmatch(output); len(match) == 2 {
		return strings.TrimSpace(match[1]), true
	}
	if match := circleRequestIDLinePattern.FindStringSubmatch(output); len(match) == 2 {
		return strings.TrimSpace(match[1]), true
	}
	if !strings.ContainsAny(output, " \t\r\n") && !strings.ContainsAny(output, "{}[]:=") {
		return output, true
	}
	return "", false
}
