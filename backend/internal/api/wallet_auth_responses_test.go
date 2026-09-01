package api

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/repository"
)

func TestWalletAuthChallengeResponseUsesFrontendContractKeys(t *testing.T) {
	expiresAt := time.Date(2026, time.September, 1, 12, 0, 0, 0, time.UTC)
	response := newWalletAuthChallengeResponse(repository.WalletAuthChallenge{
		ID:            "challenge-id",
		WalletAddress: "0x1111111111111111111111111111111111111111",
		Nonce:         "nonce",
		Message:       "Sign this challenge",
		ExpiresAt:     expiresAt,
	})

	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal wallet auth challenge response: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(encoded, &body); err != nil {
		t.Fatalf("decode wallet auth challenge response: %v", err)
	}

	for _, key := range []string{"id", "wallet_address", "nonce", "message", "expires_at"} {
		if _, ok := body[key]; !ok {
			t.Errorf("expected response key %q", key)
		}
	}
	for _, key := range []string{"ID", "WalletAddress", "Nonce", "Message", "ExpiresAt"} {
		if _, ok := body[key]; ok {
			t.Errorf("unexpected repository field key %q", key)
		}
	}
}
