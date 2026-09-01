package circleapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) CreateContractExecutionTransaction(ctx context.Context, input CreateContractExecutionTransactionInput) (CreateContractExecutionTransactionResponse, error) {
	if strings.TrimSpace(input.WalletID) == "" || strings.TrimSpace(input.ContractAddress) == "" || strings.TrimSpace(input.AbiFunctionSignature) == "" || strings.TrimSpace(input.EntitySecretCiphertext) == "" {
		return CreateContractExecutionTransactionResponse{}, ErrConfigInvalid
	}
	idempotencyKey, err := newIdempotencyKey()
	if err != nil {
		return CreateContractExecutionTransactionResponse{}, err
	}
	payload := createContractExecutionTransactionRequest{
		IdempotencyKey:         idempotencyKey,
		WalletID:               strings.TrimSpace(input.WalletID),
		ContractAddress:        strings.TrimSpace(input.ContractAddress),
		AbiFunctionSignature:   strings.TrimSpace(input.AbiFunctionSignature),
		AbiParameters:          append([]string{}, input.AbiParameters...),
		EntitySecretCiphertext: strings.TrimSpace(input.EntitySecretCiphertext),
		FeeLevel:               strings.TrimSpace(input.FeeLevel),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return CreateContractExecutionTransactionResponse{}, err
	}
	var decoded envelope[createTransactionData]
	if err := c.do(ctx, http.MethodPost, "/v1/w3s/developer/transactions/contractExecution", bytes.NewReader(body), &decoded); err != nil {
		return CreateContractExecutionTransactionResponse{}, err
	}
	if decoded.Data.ID == "" {
		return CreateContractExecutionTransactionResponse{}, fmt.Errorf("circle contract execution response missing transaction id")
	}
	return CreateContractExecutionTransactionResponse{ID: decoded.Data.ID, Status: decoded.Data.Status}, nil
}

func (c *Client) GetTransaction(ctx context.Context, id string) (Transaction, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Transaction{}, ErrConfigInvalid
	}
	var decoded envelope[transactionData]
	if err := c.do(ctx, http.MethodGet, "/v1/w3s/transactions/"+id, nil, &decoded); err != nil {
		return Transaction{}, err
	}
	data := decoded.Data
	if data.Transaction != nil {
		data.ID = firstNonEmpty(data.Transaction.ID, data.ID)
		data.Status = firstNonEmpty(data.Transaction.Status, data.Transaction.State, data.Status, data.State)
		data.TransactionHash = firstNonEmpty(data.Transaction.TransactionHash, data.Transaction.TxHash, data.TransactionHash, data.TxHash)
	}
	txHash := firstNonEmpty(data.TransactionHash, data.TxHash)
	return Transaction{
		ID:              firstNonEmpty(data.ID, id),
		Status:          firstNonEmpty(data.Status, data.State),
		TransactionHash: txHash,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
