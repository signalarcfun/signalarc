package circleapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestPollTransactionTerminalStatuses(t *testing.T) {
	tests := []struct {
		status  string
		wantErr error
	}{
		{status: "COMPLETE"},
		{status: "COMPLETED"},
		{status: "CONFIRMED"},
		{status: "SUCCESS"},
		{status: "SUCCEEDED"},
		{status: "FAILED", wantErr: ErrTransactionFailed},
		{status: "CANCELLED", wantErr: ErrTransactionFailed},
		{status: "CANCELED", wantErr: ErrTransactionFailed},
		{status: "DENIED", wantErr: ErrTransactionFailed},
		{status: "REJECTED", wantErr: ErrTransactionFailed},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/w3s/transactions/tx_123" {
					t.Fatalf("unexpected path %s", r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer test-key" {
					t.Fatalf("missing authorization header")
				}
				_, _ = w.Write([]byte(`{"data":{"id":"tx_123","status":"` + tt.status + `","transactionHash":"0xabc"}}`))
			}))
			defer server.Close()

			client, err := NewClient(ClientConfig{
				APIKey:       "test-key",
				BaseURL:      server.URL,
				PollInterval: time.Millisecond,
				PollTimeout:  time.Second,
			})
			if err != nil {
				t.Fatalf("new client: %v", err)
			}
			tx, err := client.PollTransaction(context.Background(), "tx_123")
			if tt.wantErr == nil && err != nil {
				t.Fatalf("expected success, got %v", err)
			}
			if tt.wantErr != nil && err != tt.wantErr {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
			if tx.Status != tt.status {
				t.Fatalf("expected status %q, got %q", tt.status, tx.Status)
			}
		})
	}
}

func TestPollTransactionNestedCompleteState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/w3s/transactions/tx_nested" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":{"transaction":{"id":"tx_nested","state":"COMPLETE","txHash":"0x43ca28a4770cecb431f157aa20a4b7454b039519dedef9338e1490a4d385b264"}}}`))
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		APIKey:       "test-key",
		BaseURL:      server.URL,
		PollInterval: time.Millisecond,
		PollTimeout:  time.Second,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	tx, err := client.PollTransaction(context.Background(), "tx_nested")
	if err != nil {
		t.Fatalf("expected nested COMPLETE state to succeed, got %v", err)
	}
	if tx.ID != "tx_nested" || tx.Status != "COMPLETE" || tx.TransactionHash != "0x43ca28a4770cecb431f157aa20a4b7454b039519dedef9338e1490a4d385b264" {
		t.Fatalf("unexpected transaction: %#v", tx)
	}
}

func TestGetTransactionParsesNestedCircleTransactionShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/w3s/transactions/23b4c798-e506-57ed-8101-f26b3710a5b6" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":{"transaction":{"id":"23b4c798-e506-57ed-8101-f26b3710a5b6","state":"COMPLETE","txHash":"0x43ca28a4770cecb431f157aa20a4b7454b039519dedef9338e1490a4d385b264","operation":"CONTRACT_EXECUTION"}}}`))
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{APIKey: "test-key", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	tx, err := client.GetTransaction(context.Background(), "23b4c798-e506-57ed-8101-f26b3710a5b6")
	if err != nil {
		t.Fatalf("get transaction: %v", err)
	}
	expected := Transaction{
		ID:              "23b4c798-e506-57ed-8101-f26b3710a5b6",
		Status:          "COMPLETE",
		TransactionHash: "0x43ca28a4770cecb431f157aa20a4b7454b039519dedef9338e1490a4d385b264",
	}
	if tx != expected {
		t.Fatalf("unexpected transaction\nexpected: %#v\nactual:   %#v", expected, tx)
	}
}

func TestGetTransactionParsesFlatTransactionShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/w3s/transactions/tx_flat" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":{"id":"tx_flat","status":"COMPLETE","transactionHash":"0xabc"}}`))
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{APIKey: "test-key", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	tx, err := client.GetTransaction(context.Background(), "tx_flat")
	if err != nil {
		t.Fatalf("get transaction: %v", err)
	}
	expected := Transaction{ID: "tx_flat", Status: "COMPLETE", TransactionHash: "0xabc"}
	if tx != expected {
		t.Fatalf("unexpected transaction\nexpected: %#v\nactual:   %#v", expected, tx)
	}
}

func TestCreateContractExecutionTransactionPayload(t *testing.T) {
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/w3s/developer/transactions/contractExecution" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		buffer, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		body = string(buffer)
		_, _ = w.Write([]byte(`{"data":{"id":"tx_create","status":"INITIATED"}}`))
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{APIKey: "test-key", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	resp, err := client.CreateContractExecutionTransaction(context.Background(), CreateContractExecutionTransactionInput{
		WalletID:               "wallet-1",
		ContractAddress:        "0x3333333333333333333333333333333333333333",
		AbiFunctionSignature:   "buyYes(uint256)",
		AbiParameters:          []string{"1000000"},
		EntitySecretCiphertext: "test-ciphertext",
		FeeLevel:               "MEDIUM",
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	if resp.ID != "tx_create" {
		t.Fatalf("unexpected id %q", resp.ID)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	idempotencyKey, ok := got["idempotencyKey"].(string)
	if !ok || idempotencyKey == "" {
		t.Fatalf("expected idempotencyKey, got %#v", got["idempotencyKey"])
	}
	delete(got, "idempotencyKey")
	expected := map[string]any{
		"walletId":               "wallet-1",
		"contractAddress":        "0x3333333333333333333333333333333333333333",
		"abiFunctionSignature":   "buyYes(uint256)",
		"abiParameters":          []any{"1000000"},
		"entitySecretCiphertext": "test-ciphertext",
		"feeLevel":               "MEDIUM",
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected payload shape\nexpected: %#v\nactual:   %#v", expected, got)
	}
	if _, ok := got["fee"]; ok {
		t.Fatalf("REST payload must use flat feeLevel, not nested fee: %#v", got)
	}
}
