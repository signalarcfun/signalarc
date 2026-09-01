package walletauth

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const ChallengeLifetime = 5 * time.Minute

var ErrInvalidAddress = errors.New("wallet address is invalid")
var ErrInvalidSignature = errors.New("wallet signature is invalid")

func NormalizeAddress(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if !common.IsHexAddress(trimmed) {
		return "", ErrInvalidAddress
	}

	return strings.ToLower(common.HexToAddress(trimmed).Hex()), nil
}

func NewNonce() (string, error) {
	var bytes [32]byte
	if _, err := cryptorand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate wallet auth nonce: %w", err)
	}

	return hex.EncodeToString(bytes[:]), nil
}

func ChallengeMessage(address, nonce string, issuedAt, expiresAt time.Time) string {
	return fmt.Sprintf(
		"SignalArc wallet authentication\nDomain: signalarc.fun\nAddress: %s\nNonce: %s\nIssued At: %s\nExpires At: %s",
		address,
		nonce,
		issuedAt.UTC().Format(time.RFC3339),
		expiresAt.UTC().Format(time.RFC3339),
	)
}

func VerifyPersonalSign(address, message, signature string) error {
	normalizedAddress, err := NormalizeAddress(address)
	if err != nil {
		return err
	}

	signatureBytes, err := hex.DecodeString(strings.TrimPrefix(strings.TrimSpace(signature), "0x"))
	if err != nil || len(signatureBytes) != crypto.SignatureLength {
		return ErrInvalidSignature
	}
	if signatureBytes[64] == 27 || signatureBytes[64] == 28 {
		signatureBytes[64] -= 27
	}
	if signatureBytes[64] != 0 && signatureBytes[64] != 1 {
		return ErrInvalidSignature
	}

	publicKey, err := crypto.SigToPub(accounts.TextHash([]byte(message)), signatureBytes)
	if err != nil {
		return ErrInvalidSignature
	}
	if strings.ToLower(crypto.PubkeyToAddress(*publicKey).Hex()) != normalizedAddress {
		return ErrInvalidSignature
	}

	return nil
}
