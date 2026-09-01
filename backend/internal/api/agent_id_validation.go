package api

import (
	"regexp"
	"strings"
)

// agentIDPattern enforces the documented SignalArc multi-user agent_id shape.
// The pattern intentionally allows existing live shapes such as
// agent_adenhusen65_live_002 and the recommended chatgpt suffix shape
// agent_<slug>_chatgpt_<suffix> while rejecting whitespace, slashes, and
// special characters that have caused unstable Custom GPT onboarding.
var agentIDPattern = regexp.MustCompile(`^agent_[a-z0-9](?:[a-z0-9_-]*[a-z0-9])?$`)

// agentIDMinimumLength is the minimum length enforced for agent_id values.
// "agent_" prefix plus at least 4 identifying characters keeps the shape
// recognizable and rejects placeholder shapes such as agent_a or agent_x.
const agentIDMinimumLength = 10

// genericAgentIDBlocklist is the explicit list of generic agent_id values
// that have been observed to collide across SignalArc Custom GPT users.
// Keep entries lowercase; matching is case-insensitive.
var genericAgentIDBlocklist = map[string]struct{}{
	"signalarc-gpt-agent": {},
	"signalarc_gpt_agent": {},
	"agent_desi_001":      {},
	"default":             {},
	"defaultagent":        {},
	"default_agent":       {},
	"test":                {},
	"testagent":           {},
	"test_agent":          {},
	"demo":                {},
	"demoagent":           {},
	"demo_agent":          {},
	"user":                {},
	"useragent":           {},
	"user_agent":          {},
	"agent":               {},
	"chatgpt":             {},
	"chatgpt_agent":       {},
}

// validateAgentID returns the trimmed agent_id and an ordered list of
// validation error strings. The empty error slice means the id is acceptable.
//
// Validation rules:
//
//  1. agent_id is required and must be a non-empty string.
//  2. agent_id must not be one of the documented generic placeholder values
//     such as "signalarc-gpt-agent", "default", "test", "demo", "agent",
//     "agent_desi_001", or any case-insensitive equivalent.
//  3. agent_id must match the SignalArc shape: lowercase letters, digits,
//     underscores, and hyphens, starting with "agent_" and at least
//     agentIDMinimumLength characters long.
//
// validateAgentID does not change behavior of any existing non-agent_id field
// validation. Callers should append the returned error strings into their
// existing details array.
func validateAgentID(rawAgentID string) (string, []string) {
	trimmed := strings.TrimSpace(rawAgentID)
	if trimmed == "" {
		return "", []string{"agent_id is required"}
	}

	lower := strings.ToLower(trimmed)
	if _, blocked := genericAgentIDBlocklist[lower]; blocked {
		return trimmed, []string{
			"agent_id is a generic placeholder; use agent_<slug>_chatgpt_<suffix> or another unique value",
		}
	}

	if len(trimmed) < agentIDMinimumLength {
		return trimmed, []string{
			"agent_id must be at least 10 characters long and start with agent_",
		}
	}

	if !agentIDPattern.MatchString(strings.ToLower(trimmed)) {
		return trimmed, []string{
			"agent_id must match agent_<slug> using lowercase letters, digits, underscore, or hyphen",
		}
	}

	return trimmed, nil
}
