package circleapi

import "time"

const DefaultBaseURL = "https://api.circle.com"

type ClientConfig struct {
	APIKey       string
	BaseURL      string
	Timeout      time.Duration
	PollInterval time.Duration
	PollTimeout  time.Duration
}

type CreateContractExecutionTransactionInput struct {
	WalletID               string
	ContractAddress        string
	AbiFunctionSignature   string
	AbiParameters          []string
	EntitySecretCiphertext string
	FeeLevel               string
}

type CreateContractExecutionTransactionResponse struct {
	ID     string
	Status string
}

type Transaction struct {
	ID              string
	Status          string
	TransactionHash string
}

type WalletTokenBalance struct {
	Token  WalletToken `json:"token"`
	Amount string      `json:"amount"`
}

type WalletToken struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Symbol     string `json:"symbol,omitempty"`
	Decimals   int    `json:"decimals,omitempty"`
	Address    string `json:"address,omitempty"`
	Blockchain string `json:"blockchain,omitempty"`
}

type createContractExecutionTransactionRequest struct {
	IdempotencyKey         string   `json:"idempotencyKey"`
	WalletID               string   `json:"walletId"`
	ContractAddress        string   `json:"contractAddress"`
	AbiFunctionSignature   string   `json:"abiFunctionSignature"`
	AbiParameters          []string `json:"abiParameters,omitempty"`
	EntitySecretCiphertext string   `json:"entitySecretCiphertext"`
	FeeLevel               string   `json:"feeLevel,omitempty"`
}

type envelope[T any] struct {
	Data T `json:"data"`
}

type createTransactionData struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type transactionData struct {
	ID              string              `json:"id"`
	Status          string              `json:"status"`
	State           string              `json:"state"`
	TransactionHash string              `json:"transactionHash"`
	TxHash          string              `json:"txHash"`
	Transaction     *transactionDetails `json:"transaction"`
}

type transactionDetails struct {
	ID              string `json:"id"`
	Status          string `json:"status"`
	State           string `json:"state"`
	TransactionHash string `json:"transactionHash"`
	TxHash          string `json:"txHash"`
}

type walletBalancesData struct {
	TokenBalances []WalletTokenBalance `json:"tokenBalances"`
}

type entityPublicKeyData struct {
	PublicKey string `json:"publicKey"`
}
