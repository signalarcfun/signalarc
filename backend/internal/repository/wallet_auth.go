package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/wahyu241205/SignalArc/backend/internal/database"
)

var ErrWalletAuthChallengeUnavailable = errors.New("wallet authentication challenge is unavailable")

type WalletAuthChallenge struct {
	ID            string
	WalletAddress string
	Nonce         string
	Message       string
	ExpiresAt     time.Time
}

type WalletAuthSession struct {
	UserID        string
	WalletAddress string
	ExpiresAt     time.Time
}

type WalletAuthRepository struct {
	db *database.DB
}

func NewWalletAuthRepository(db *database.DB) *WalletAuthRepository {
	return &WalletAuthRepository{db: db}
}

func (r *WalletAuthRepository) CreateChallenge(ctx context.Context, address, nonce, message string, expiresAt time.Time) (WalletAuthChallenge, error) {
	var challenge WalletAuthChallenge
	err := r.db.QueryRow(ctx, `
		INSERT INTO wallet_auth_challenges (wallet_address, nonce, message, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, wallet_address, nonce, message, expires_at
	`, address, nonce, message, expiresAt).Scan(
		&challenge.ID,
		&challenge.WalletAddress,
		&challenge.Nonce,
		&challenge.Message,
		&challenge.ExpiresAt,
	)
	return challenge, err
}

func (r *WalletAuthRepository) GetUsableChallenge(ctx context.Context, id string) (WalletAuthChallenge, error) {
	var challenge WalletAuthChallenge
	err := r.db.QueryRow(ctx, `
		SELECT id::text, wallet_address, nonce, message, expires_at
		FROM wallet_auth_challenges
		WHERE id = $1 AND consumed_at IS NULL AND expires_at > now()
	`, id).Scan(
		&challenge.ID,
		&challenge.WalletAddress,
		&challenge.Nonce,
		&challenge.Message,
		&challenge.ExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return WalletAuthChallenge{}, ErrWalletAuthChallengeUnavailable
	}
	return challenge, err
}

func (r *WalletAuthRepository) ConsumeChallengeAndCreateSession(ctx context.Context, challengeID, address, tokenHash string, expiresAt time.Time) (WalletAuthSession, error) {
	var session WalletAuthSession
	err := r.db.QueryRow(ctx, `
		WITH consumed AS (
			UPDATE wallet_auth_challenges
			SET consumed_at = now()
			WHERE id = $1
				AND wallet_address = $2
				AND consumed_at IS NULL
				AND expires_at > now()
			RETURNING wallet_address
		),
		existing_wallet AS (
			SELECT user_id, address
			FROM wallets
			WHERE lower(address) = (SELECT wallet_address FROM consumed)
		),
			created_user AS (
			INSERT INTO users (external_id)
			SELECT 'wallet:' || wallet_address
			FROM consumed
			WHERE NOT EXISTS (SELECT 1 FROM existing_wallet)
			ON CONFLICT (external_id) DO UPDATE SET updated_at = now()
			RETURNING id
		),
		created_wallet AS (
			INSERT INTO wallets (user_id, provider, address, chain, is_primary)
			SELECT (SELECT id FROM created_user),
				'wallet_signature', wallet_address, 'ARC-TESTNET', true
			FROM consumed
			WHERE NOT EXISTS (SELECT 1 FROM existing_wallet)
			ON CONFLICT (address) DO UPDATE SET updated_at = now()
			RETURNING user_id, address
		),
		bound_wallet AS (
			SELECT user_id, lower(address) AS address FROM existing_wallet
			UNION ALL
			SELECT user_id, address FROM created_wallet
		)
		INSERT INTO wallet_auth_sessions (user_id, wallet_address, token_hash, expires_at)
		SELECT user_id, address, $3, $4
		FROM bound_wallet
		RETURNING user_id::text, wallet_address, expires_at
	`, challengeID, address, tokenHash, expiresAt).Scan(
		&session.UserID,
		&session.WalletAddress,
		&session.ExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return WalletAuthSession{}, ErrWalletAuthChallengeUnavailable
	}
	return session, err
}

func (r *WalletAuthRepository) GetActiveSessionByTokenHash(ctx context.Context, tokenHash string) (WalletAuthSession, error) {
	var session WalletAuthSession
	err := r.db.QueryRow(ctx, `
		SELECT user_id::text, wallet_address, expires_at
		FROM wallet_auth_sessions
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()
	`, tokenHash).Scan(&session.UserID, &session.WalletAddress, &session.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
		return WalletAuthSession{}, ErrWalletAuthChallengeUnavailable
	}
	return session, err
}
