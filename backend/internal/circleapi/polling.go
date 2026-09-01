package circleapi

import (
	"context"
	"errors"
	"strings"
	"time"
)

func (c *Client) PollTransaction(ctx context.Context, id string) (Transaction, error) {
	if c == nil {
		return Transaction{}, ErrConfigInvalid
	}
	ctx, cancel := context.WithTimeout(ctx, c.pollTimeout)
	defer cancel()

	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()

	last := Transaction{ID: strings.TrimSpace(id)}
	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return Transaction{}, newPollingTimeoutError(last)
			}
			return Transaction{}, ctx.Err()
		default:
		}

		tx, err := c.GetTransaction(ctx, id)
		if err != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return Transaction{}, newPollingTimeoutError(last)
			}
			return Transaction{}, err
		}
		last = tx
		switch normalizedTransactionStatus(tx.Status) {
		case transactionStatusComplete:
			return tx, nil
		case transactionStatusFailed:
			return tx, ErrTransactionFailed
		}

		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return Transaction{}, newPollingTimeoutError(last)
			}
			return Transaction{}, ctx.Err()
		case <-ticker.C:
		}
	}
}

type transactionTerminalStatus string

const (
	transactionStatusPending  transactionTerminalStatus = "pending"
	transactionStatusComplete transactionTerminalStatus = "complete"
	transactionStatusFailed   transactionTerminalStatus = "failed"
)

func normalizedTransactionStatus(status string) transactionTerminalStatus {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "COMPLETE", "COMPLETED", "CONFIRMED", "SUCCESS", "SUCCEEDED":
		return transactionStatusComplete
	case "FAILED", "CANCELLED", "CANCELED", "DENIED", "REJECTED":
		return transactionStatusFailed
	default:
		return transactionStatusPending
	}
}
