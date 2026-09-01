package circleapi

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"strings"
	"sync"
	"time"
)

var ErrEntitySecretCiphertextUnavailable = errors.New("circle entity secret ciphertext is unavailable")

type EntitySecretCiphertextProvider interface {
	Ciphertext(context.Context) (string, error)
}

// EnvEntitySecretCiphertextProvider is a temporary static/dev fallback for
// local REST integration work. Production must generate a unique
// entitySecretCiphertext for every Circle API request from the raw entity
// secret and Circle entity public key, or use an approved secret provider that
// guarantees per-request ciphertext uniqueness.
type EnvEntitySecretCiphertextProvider struct {
	ciphertext string
}

func NewEnvEntitySecretCiphertextProvider(ciphertext string) EnvEntitySecretCiphertextProvider {
	return EnvEntitySecretCiphertextProvider{ciphertext: strings.TrimSpace(ciphertext)}
}

func (provider EnvEntitySecretCiphertextProvider) Ciphertext(context.Context) (string, error) {
	if provider.ciphertext == "" {
		return "", ErrEntitySecretCiphertextUnavailable
	}
	return provider.ciphertext, nil
}

type EntityPublicKeyClient interface {
	GetEntityPublicKey(context.Context) (string, error)
}

type RawEntitySecretCiphertextProviderConfig struct {
	APIKey          string
	BaseURL         string
	RawEntitySecret string
	Timeout         time.Duration
	Client          EntityPublicKeyClient
}

// RawEntitySecretCiphertextProvider is the production Circle REST provider.
// CIRCLE_ENTITY_SECRET is the raw entity secret and must be supplied only from
// a secret manager or protected environment. The provider caches Circle's
// entity public key, then generates a fresh entitySecretCiphertext for each
// Circle API request using RSA-OAEP with SHA-256. Ciphertexts are never cached
// or reused.
type RawEntitySecretCiphertextProvider struct {
	client          EntityPublicKeyClient
	rawEntitySecret []byte

	mu        sync.RWMutex
	publicKey *rsa.PublicKey
}

func NewRawEntitySecretCiphertextProvider(cfg RawEntitySecretCiphertextProviderConfig) (*RawEntitySecretCiphertextProvider, error) {
	rawEntitySecret, err := decodeRawEntitySecret(cfg.RawEntitySecret)
	if err != nil {
		return nil, ErrEntitySecretCiphertextUnavailable
	}
	client := cfg.Client
	if client == nil {
		circleClient, err := NewClient(ClientConfig{
			APIKey:  cfg.APIKey,
			BaseURL: cfg.BaseURL,
			Timeout: cfg.Timeout,
		})
		if err != nil {
			return nil, err
		}
		client = circleClient
	}
	return &RawEntitySecretCiphertextProvider{
		client:          client,
		rawEntitySecret: rawEntitySecret,
	}, nil
}

func (provider *RawEntitySecretCiphertextProvider) Ciphertext(ctx context.Context) (string, error) {
	if provider == nil {
		return "", ErrEntitySecretCiphertextUnavailable
	}
	publicKey, err := provider.cachedPublicKey(ctx)
	if err != nil {
		return "", err
	}
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, provider.rawEntitySecret, nil)
	if err != nil {
		return "", ErrEntitySecretCiphertextUnavailable
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (provider *RawEntitySecretCiphertextProvider) cachedPublicKey(ctx context.Context) (*rsa.PublicKey, error) {
	provider.mu.RLock()
	publicKey := provider.publicKey
	provider.mu.RUnlock()
	if publicKey != nil {
		return publicKey, nil
	}

	provider.mu.Lock()
	defer provider.mu.Unlock()
	if provider.publicKey != nil {
		return provider.publicKey, nil
	}
	rawPublicKey, err := provider.client.GetEntityPublicKey(ctx)
	if err != nil {
		return nil, err
	}
	parsed, err := parseRSAPublicKey(rawPublicKey)
	if err != nil {
		return nil, ErrEntitySecretCiphertextUnavailable
	}
	provider.publicKey = parsed
	return parsed, nil
}

func decodeRawEntitySecret(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X")
	if value == "" {
		return nil, ErrEntitySecretCiphertextUnavailable
	}
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) == 0 {
		return nil, ErrEntitySecretCiphertextUnavailable
	}
	return decoded, nil
}

func parseRSAPublicKey(value string) (*rsa.PublicKey, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, ErrEntitySecretCiphertextUnavailable
	}
	block, _ := pem.Decode([]byte(value))
	var der []byte
	if block != nil {
		der = block.Bytes
	} else {
		normalized := strings.Join(strings.Fields(value), "")
		decoded, err := base64.StdEncoding.DecodeString(normalized)
		if err != nil {
			return nil, err
		}
		der = decoded
	}
	parsed, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, err
	}
	publicKey, ok := parsed.(*rsa.PublicKey)
	if !ok {
		return nil, ErrEntitySecretCiphertextUnavailable
	}
	return publicKey, nil
}
