package circleapi

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestRawEntitySecretCiphertextProviderFetchesPublicKeyAndEncrypts(t *testing.T) {
	privateKey := newTestRSAKey(t)
	publicKeyPEM := encodePublicKeyPEM(t, &privateKey.PublicKey)
	rawSecret := []byte("0123456789abcdef0123456789abcdef")
	rawSecretHex := hex.EncodeToString(rawSecret)
	var publicKeyRequests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/w3s/config/entity/publicKey" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-api-key" {
			t.Fatalf("unexpected authorization header")
		}
		atomic.AddInt32(&publicKeyRequests, 1)
		_ = json.NewEncoder(w).Encode(envelope[entityPublicKeyData]{
			Data: entityPublicKeyData{PublicKey: publicKeyPEM},
		})
	}))
	defer server.Close()

	provider, err := NewRawEntitySecretCiphertextProvider(RawEntitySecretCiphertextProviderConfig{
		APIKey:          "test-api-key",
		BaseURL:         server.URL,
		RawEntitySecret: rawSecretHex,
	})
	if err != nil {
		t.Fatalf("expected provider, got error %v", err)
	}

	ciphertext, err := provider.Ciphertext(context.Background())
	if err != nil {
		t.Fatalf("expected ciphertext, got error %v", err)
	}
	if strings.TrimSpace(ciphertext) == "" {
		t.Fatal("expected non-empty ciphertext")
	}
	if strings.Contains(ciphertext, rawSecretHex) {
		t.Fatal("ciphertext exposed raw entity secret")
	}
	decrypted := decryptTestCiphertext(t, privateKey, ciphertext)
	if string(decrypted) != string(rawSecret) {
		t.Fatal("ciphertext did not decrypt to the raw entity secret")
	}
	if atomic.LoadInt32(&publicKeyRequests) != 1 {
		t.Fatalf("expected one public key request, got %d", publicKeyRequests)
	}
}

func TestRawEntitySecretCiphertextProviderEncryptsPerRequest(t *testing.T) {
	privateKey := newTestRSAKey(t)
	publicKeyPEM := encodePublicKeyPEM(t, &privateKey.PublicKey)
	rawSecret := []byte("0123456789abcdef0123456789abcdef")
	provider, err := NewRawEntitySecretCiphertextProvider(RawEntitySecretCiphertextProviderConfig{
		RawEntitySecret: hex.EncodeToString(rawSecret),
		Client:          staticEntityPublicKeyClient{publicKey: publicKeyPEM},
	})
	if err != nil {
		t.Fatalf("expected provider, got error %v", err)
	}

	first, err := provider.Ciphertext(context.Background())
	if err != nil {
		t.Fatalf("expected first ciphertext, got error %v", err)
	}
	second, err := provider.Ciphertext(context.Background())
	if err != nil {
		t.Fatalf("expected second ciphertext, got error %v", err)
	}
	if first == second {
		t.Fatal("expected distinct ciphertext values across requests")
	}
	if string(decryptTestCiphertext(t, privateKey, first)) != string(rawSecret) {
		t.Fatal("first ciphertext did not decrypt to raw entity secret")
	}
	if string(decryptTestCiphertext(t, privateKey, second)) != string(rawSecret) {
		t.Fatal("second ciphertext did not decrypt to raw entity secret")
	}
}

func TestRawEntitySecretCiphertextProviderDoesNotReturnRawEntitySecret(t *testing.T) {
	privateKey := newTestRSAKey(t)
	publicKeyPEM := encodePublicKeyPEM(t, &privateKey.PublicKey)
	rawSecretHex := "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"
	provider, err := NewRawEntitySecretCiphertextProvider(RawEntitySecretCiphertextProviderConfig{
		RawEntitySecret: rawSecretHex,
		Client:          staticEntityPublicKeyClient{publicKey: publicKeyPEM},
	})
	if err != nil {
		t.Fatalf("expected provider, got error %v", err)
	}

	ciphertext, err := provider.Ciphertext(context.Background())
	if err != nil {
		t.Fatalf("expected ciphertext, got error %v", err)
	}
	if ciphertext == rawSecretHex || strings.Contains(ciphertext, rawSecretHex) {
		t.Fatal("provider returned raw entity secret material")
	}
}

type staticEntityPublicKeyClient struct {
	publicKey string
}

func (client staticEntityPublicKeyClient) GetEntityPublicKey(context.Context) (string, error) {
	return client.publicKey, nil
}

func newTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	return privateKey
}

func encodePublicKeyPEM(t *testing.T, publicKey *rsa.PublicKey) string {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
}

func decryptTestCiphertext(t *testing.T, privateKey *rsa.PrivateKey, ciphertext string) []byte {
	t.Helper()
	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		t.Fatalf("ciphertext was not base64: %v", err)
	}
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, decoded, nil)
	if err != nil {
		t.Fatalf("ciphertext could not be decrypted: %v", err)
	}
	return plaintext
}
