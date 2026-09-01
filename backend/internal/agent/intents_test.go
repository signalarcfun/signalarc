package agent

import "testing"

func TestStoreCreateIntentPolicyMetadataPropagation(t *testing.T) {
	store := NewStore()

	walletMetadata := map[string]string{
		" circle_wallet_id ": " wallet-from-registration ",
		"max_trade_amount":   " 10 ",
		"empty_value":        " ",
		" ":                  "empty_key",
	}
	registered, err := store.RegisterAgentWallet(AgentWallet{
		AgentID:            "agent_test",
		UserWallet:         "0x1111111111111111111111111111111111111111",
		AgentWalletAddress: "0x9999999999999999999999999999999999999999",
		WalletProvider:     WalletProviderCircleAgentWallet,
		Chain:              ChainArcTestnet,
		AllowedActions:     []string{ActionBuyYes},
		Status:             WalletStatusActive,
		PolicyMetadata:     walletMetadata,
	})
	if err != nil {
		t.Fatalf("register wallet: %v", err)
	}
	walletMetadata["circle_wallet_id"] = "mutated-registration-map"
	walletMetadata["max_trade_amount"] = "99"
	if registered.PolicyMetadata["circle_wallet_id"] != "wallet-from-registration" {
		t.Fatalf("registered wallet metadata aliased input: %#v", registered.PolicyMetadata)
	}

	registeredWallet, err := store.GetAgentWallet("agent_test")
	if err != nil {
		t.Fatalf("get wallet: %v", err)
	}
	if registeredWallet.PolicyMetadata["circle_wallet_id"] != "wallet-from-registration" {
		t.Fatalf("stored wallet metadata aliased input: %#v", registeredWallet.PolicyMetadata)
	}
	if _, ok := registeredWallet.PolicyMetadata["empty_value"]; ok {
		t.Fatalf("expected empty metadata values to be dropped: %#v", registeredWallet.PolicyMetadata)
	}

	explicitMetadata := map[string]string{
		" circle_wallet_id ": " wallet-from-intent ",
		"route":              " explicit ",
	}
	explicitIntent, err := store.CreateIntent(baseBuyYesIntentInput(CreateIntentInput{
		AgentID:        "agent_test",
		PolicyMetadata: explicitMetadata,
	}))
	if err != nil {
		t.Fatalf("create explicit intent: %v", err)
	}
	explicitMetadata["circle_wallet_id"] = "mutated-intent-map"
	if explicitIntent.PolicyMetadata["circle_wallet_id"] != "wallet-from-intent" {
		t.Fatalf("explicit intent metadata not preserved: %#v", explicitIntent.PolicyMetadata)
	}
	if explicitIntent.PolicyMetadata["route"] != "explicit" {
		t.Fatalf("explicit intent metadata not normalized: %#v", explicitIntent.PolicyMetadata)
	}

	storedExplicitIntent, err := store.GetIntent(explicitIntent.ID)
	if err != nil {
		t.Fatalf("get explicit intent: %v", err)
	}
	if storedExplicitIntent.PolicyMetadata["circle_wallet_id"] != "wallet-from-intent" {
		t.Fatalf("stored explicit intent metadata aliased input: %#v", storedExplicitIntent.PolicyMetadata)
	}

	hydratedIntent, err := store.CreateIntent(baseBuyYesIntentInput(CreateIntentInput{
		AgentID: "agent_test",
	}))
	if err != nil {
		t.Fatalf("create hydrated intent: %v", err)
	}
	if hydratedIntent.PolicyMetadata["circle_wallet_id"] != "wallet-from-registration" {
		t.Fatalf("expected wallet metadata hydration, got %#v", hydratedIntent.PolicyMetadata)
	}
	if hydratedIntent.PolicyMetadata["max_trade_amount"] != "10" {
		t.Fatalf("expected normalized wallet metadata hydration, got %#v", hydratedIntent.PolicyMetadata)
	}
}

func baseBuyYesIntentInput(overrides CreateIntentInput) CreateIntentInput {
	input := CreateIntentInput{
		AgentID:               "agent_test",
		Action:                ActionBuyYes,
		UserWallet:            "0x1111111111111111111111111111111111111111",
		MarketID:              "market-1",
		MarketContractAddress: "0x3333333333333333333333333333333333333333",
		Amount:                "1",
	}
	if overrides.AgentID != "" {
		input.AgentID = overrides.AgentID
	}
	if overrides.PolicyMetadata != nil {
		input.PolicyMetadata = overrides.PolicyMetadata
	}
	return input
}
