package api

import (
	"encoding/json"
	"testing"
)

func TestWalletAuthChallengeRequestAcceptsWalletAddress(t *testing.T) {
	const payload = `{"wallet_address":"0xb592d99cb3c98b77777d6288e5e5782ac2ce919a"}`

	var request walletAuthChallengeRequest
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		t.Fatalf("decode request: %v", err)
	}

	got, err := request.normalizedAddress()
	if err != nil {
		t.Fatalf("normalize exact production address: %v", err)
	}
	if want := "0xb592d99cb3c98b77777d6288e5e5782ac2ce919a"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestWalletAuthChallengeRequestRetainsAddressAlias(t *testing.T) {
	const payload = `{"address":"0xB592D99CB3C98b77777D6288e5E5782AC2ce919a"}`

	var request walletAuthChallengeRequest
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		t.Fatalf("decode request: %v", err)
	}

	got, err := request.normalizedAddress()
	if err != nil {
		t.Fatalf("normalize address alias: %v", err)
	}
	if want := "0xb592d99cb3c98b77777d6288e5e5782ac2ce919a"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestWalletAuthChallengeRequestRejectsConflictingAddressFields(t *testing.T) {
	const payload = `{
		"wallet_address":"0xb592d99cb3c98b77777d6288e5e5782ac2ce919a",
		"address":"0x1111111111111111111111111111111111111111"
	}`

	var request walletAuthChallengeRequest
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		t.Fatalf("decode request: %v", err)
	}

	if _, err := request.normalizedAddress(); err == nil {
		t.Fatal("expected conflicting address fields to be rejected")
	}
}
