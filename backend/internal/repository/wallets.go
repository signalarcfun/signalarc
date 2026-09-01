package repository

import (
	"context"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/database"
)

type Wallet struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Provider  string    `json:"provider"`
	Address   string    `json:"address"`
	Chain     string    `json:"chain"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WalletsRepository struct {
	db *database.DB
}

func NewWalletsRepository(db *database.DB) *WalletsRepository {
	return &WalletsRepository{db: db}
}

func (r *WalletsRepository) GetWalletByID(ctx context.Context, id string) (Wallet, error) {
	var wallet Wallet
	err := r.db.QueryRow(ctx, `
		SELECT id::text, user_id::text, provider, address, chain, is_primary, created_at, updated_at
		FROM wallets
		WHERE id = $1
	`, id).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Provider,
		&wallet.Address,
		&wallet.Chain,
		&wallet.IsPrimary,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	return wallet, err
}

func (r *WalletsRepository) ListWalletsByUserID(ctx context.Context, userID string, limit int) ([]Wallet, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, user_id::text, provider, address, chain, is_primary, created_at, updated_at
		FROM wallets
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	wallets := []Wallet{}
	for rows.Next() {
		var wallet Wallet
		if err := rows.Scan(
			&wallet.ID,
			&wallet.UserID,
			&wallet.Provider,
			&wallet.Address,
			&wallet.Chain,
			&wallet.IsPrimary,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		); err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return wallets, nil
}
