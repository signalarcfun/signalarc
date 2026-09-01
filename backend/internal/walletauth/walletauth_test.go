package walletauth

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestVerifyPersonalSignAcceptsOwningWallet(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	address := crypto.PubkeyToAddress(key.PublicKey).Hex()
	message := ChallengeMessage(address, "nonce", time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 1, 0, 5, 0, 0, time.UTC))
	signature, err := crypto.Sign(accounts.TextHash([]byte(message)), key)
	if err != nil {
		t.Fatalf("sign message: %v", err)
	}

	if err := VerifyPersonalSign(address, message, "0x"+hex.EncodeToString(signature)); err != nil {
		t.Fatalf("verify signature: %v", err)
	}
}

func TestVerifyPersonalSignRejectsWrongWalletAndMalformedSignature(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	message := "SignalArc wallet authentication"
	signature, err := crypto.Sign(accounts.TextHash([]byte(message)), key)
	if err != nil {
		t.Fatalf("sign message: %v", err)
	}

	if err := VerifyPersonalSign("0x0000000000000000000000000000000000000001", message, "0x"+hex.EncodeToString(signature)); err != ErrInvalidSignature {
		t.Fatalf("expected invalid signature for another wallet, got %v", err)
	}
	if err := VerifyPersonalSign(crypto.PubkeyToAddress(key.PublicKey).Hex(), message, "0xnot-a-signature"); err != ErrInvalidSignature {
		t.Fatalf("expected malformed signature error, got %v", err)
	}
}

func TestNormalizeAddress(t *testing.T) {
	got, err := NormalizeAddress(" 0xB592d99cb3c98b77777d6288e5E5782Ac2Ce919a ")
	if err != nil {
		t.Fatalf("normalize address: %v", err)
	}
	if want := "0xb592d99cb3c98b77777d6288e5e5782ac2ce919a"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
