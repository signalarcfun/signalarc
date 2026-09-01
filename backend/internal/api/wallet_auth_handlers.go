package api

import (
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wahyu241205/SignalArc/backend/internal/httpjson"
	"github.com/wahyu241205/SignalArc/backend/internal/repository"
	"github.com/wahyu241205/SignalArc/backend/internal/walletauth"
)

const walletAuthSessionLifetime = 24 * time.Hour

type walletAuthChallengeRequest struct {
	WalletAddress string `json:"wallet_address"`
	Address       string `json:"address"`
}

type walletAuthVerifyRequest struct {
	ChallengeID string `json:"challenge_id"`
	Address     string `json:"address"`
	Signature   string `json:"signature"`
}

func (request walletAuthChallengeRequest) normalizedAddress() (string, error) {
	walletAddress := strings.TrimSpace(request.WalletAddress)
	addressAlias := strings.TrimSpace(request.Address)

	if walletAddress == "" {
		return walletauth.NormalizeAddress(addressAlias)
	}

	normalizedWalletAddress, err := walletauth.NormalizeAddress(walletAddress)
	if err != nil || addressAlias == "" {
		return normalizedWalletAddress, err
	}
	normalizedAddressAlias, err := walletauth.NormalizeAddress(addressAlias)
	if err != nil || normalizedAddressAlias != normalizedWalletAddress {
		return "", walletauth.ErrInvalidAddress
	}

	return normalizedWalletAddress, nil
}

func registerWalletAuthRoutes(router chi.Router, store *repository.WalletAuthRepository) {
	router.Post("/auth/wallet/challenge", func(w http.ResponseWriter, r *http.Request) {
		var request walletAuthChallengeRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON request body")
			return
		}

		address, err := request.normalizedAddress()
		if err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_wallet_address", "wallet address is invalid")
			return
		}
		nonce, err := walletauth.NewNonce()
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "wallet_auth_unavailable", "wallet authentication is unavailable")
			return
		}

		now := time.Now().UTC()
		challenge, err := store.CreateChallenge(r.Context(), address, nonce, walletauth.ChallengeMessage(address, nonce, now, now.Add(walletauth.ChallengeLifetime)), now.Add(walletauth.ChallengeLifetime))
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "wallet_auth_challenge_failed", "wallet authentication challenge could not be created")
			return
		}

		httpjson.WriteJSON(w, http.StatusCreated, map[string]any{
			"challenge": newWalletAuthChallengeResponse(challenge),
		})
	})

	router.Post("/auth/wallet/verify", func(w http.ResponseWriter, r *http.Request) {
		var request walletAuthVerifyRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON request body")
			return
		}

		address, err := walletauth.NormalizeAddress(request.Address)
		if err != nil || strings.TrimSpace(request.ChallengeID) == "" {
			httpjson.WriteError(w, http.StatusBadRequest, "invalid_wallet_auth_request", "wallet authentication request is invalid")
			return
		}
		challenge, err := store.GetUsableChallenge(r.Context(), strings.TrimSpace(request.ChallengeID))
		if errors.Is(err, repository.ErrWalletAuthChallengeUnavailable) {
			httpjson.WriteError(w, http.StatusUnauthorized, "wallet_auth_challenge_invalid", "wallet authentication challenge is invalid or expired")
			return
		}
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "wallet_auth_challenge_failed", "wallet authentication challenge could not be read")
			return
		}
		if challenge.WalletAddress != address || walletauth.VerifyPersonalSign(address, challenge.Message, request.Signature) != nil {
			httpjson.WriteError(w, http.StatusUnauthorized, "wallet_signature_invalid", "wallet signature is invalid")
			return
		}

		var tokenBytes [32]byte
		if _, err := cryptorand.Read(tokenBytes[:]); err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "wallet_auth_unavailable", "wallet authentication is unavailable")
			return
		}
		token := hex.EncodeToString(tokenBytes[:])
		tokenHash := sha256.Sum256([]byte(token))
		session, err := store.ConsumeChallengeAndCreateSession(r.Context(), challenge.ID, address, hex.EncodeToString(tokenHash[:]), time.Now().UTC().Add(walletAuthSessionLifetime))
		if errors.Is(err, repository.ErrWalletAuthChallengeUnavailable) {
			httpjson.WriteError(w, http.StatusUnauthorized, "wallet_auth_challenge_invalid", "wallet authentication challenge is invalid or expired")
			return
		}
		if err != nil {
			httpjson.WriteError(w, http.StatusInternalServerError, "wallet_auth_session_failed", "wallet authentication session could not be created")
			return
		}

		httpjson.WriteJSON(w, http.StatusCreated, map[string]any{
			"session": map[string]any{
				"token":          token,
				"user_id":        session.UserID,
				"wallet_address": session.WalletAddress,
				"expires_at":     session.ExpiresAt,
			},
		})
	})
}
