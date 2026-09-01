package trade

import "testing"

func TestCreateTradeIntentRequestToRepositoryInput(t *testing.T) {
	request := validCreateTradeIntentRequest()
	request.Quantity = " 10.5 "
	request.Price = " 0.25 "

	input, err := request.ToRepositoryInput()
	if err != nil {
		t.Fatalf("expected valid trade intent, got %v", err)
	}

	if input.UserID != request.UserID {
		t.Fatalf("expected user id %q, got %q", request.UserID, input.UserID)
	}
	if input.MarketID != request.MarketID {
		t.Fatalf("expected market id %q, got %q", request.MarketID, input.MarketID)
	}
	if input.Outcome != "YES" {
		t.Fatalf("expected YES outcome, got %q", input.Outcome)
	}
	if input.Side != "BUY" {
		t.Fatalf("expected BUY side, got %q", input.Side)
	}
	if input.Quantity != "10.5" {
		t.Fatalf("expected trimmed quantity, got %q", input.Quantity)
	}
	if input.Price != "0.25" {
		t.Fatalf("expected trimmed price, got %q", input.Price)
	}
	if input.CollateralAmount != "2.625" {
		t.Fatalf("expected collateral 2.625, got %q", input.CollateralAmount)
	}
}

func TestCreateTradeIntentRequestRejectsInvalidInputs(t *testing.T) {
	testCases := map[string]CreateTradeIntentRequest{
		"missing user id": {
			UserID:   "",
			MarketID: "20000000-0000-4000-8000-000000000001",
			Outcome:  "YES",
			Side:     "BUY",
			Quantity: "1",
			Price:    "0.5",
		},
		"invalid user id": {
			UserID:   "not-a-uuid",
			MarketID: "20000000-0000-4000-8000-000000000001",
			Outcome:  "YES",
			Side:     "BUY",
			Quantity: "1",
			Price:    "0.5",
		},
		"missing market id": {
			UserID:   "10000000-0000-4000-8000-000000000001",
			MarketID: "",
			Outcome:  "YES",
			Side:     "BUY",
			Quantity: "1",
			Price:    "0.5",
		},
		"invalid outcome": {
			UserID:   "10000000-0000-4000-8000-000000000001",
			MarketID: "20000000-0000-4000-8000-000000000001",
			Outcome:  "MAYBE",
			Side:     "BUY",
			Quantity: "1",
			Price:    "0.5",
		},
		"invalid side": {
			UserID:   "10000000-0000-4000-8000-000000000001",
			MarketID: "20000000-0000-4000-8000-000000000001",
			Outcome:  "YES",
			Side:     "HOLD",
			Quantity: "1",
			Price:    "0.5",
		},
		"zero quantity": {
			UserID:   "10000000-0000-4000-8000-000000000001",
			MarketID: "20000000-0000-4000-8000-000000000001",
			Outcome:  "YES",
			Side:     "BUY",
			Quantity: "0",
			Price:    "0.5",
		},
		"invalid quantity": {
			UserID:   "10000000-0000-4000-8000-000000000001",
			MarketID: "20000000-0000-4000-8000-000000000001",
			Outcome:  "YES",
			Side:     "BUY",
			Quantity: "1.0000000000000000001",
			Price:    "0.5",
		},
		"price greater than one": {
			UserID:   "10000000-0000-4000-8000-000000000001",
			MarketID: "20000000-0000-4000-8000-000000000001",
			Outcome:  "YES",
			Side:     "BUY",
			Quantity: "1",
			Price:    "1.01",
		},
		"invalid price": {
			UserID:   "10000000-0000-4000-8000-000000000001",
			MarketID: "20000000-0000-4000-8000-000000000001",
			Outcome:  "YES",
			Side:     "BUY",
			Quantity: "1",
			Price:    "abc",
		},
		"too many collateral fractional digits": {
			UserID:   "10000000-0000-4000-8000-000000000001",
			MarketID: "20000000-0000-4000-8000-000000000001",
			Outcome:  "YES",
			Side:     "BUY",
			Quantity: "0.333333333333333333",
			Price:    "0.333333333333333333",
		},
	}

	for name, request := range testCases {
		request := request
		t.Run(name, func(t *testing.T) {
			if _, err := request.ToRepositoryInput(); err == nil {
				t.Fatal("expected invalid trade intent to be rejected")
			}
		})
	}
}

func validCreateTradeIntentRequest() CreateTradeIntentRequest {
	return CreateTradeIntentRequest{
		UserID:   "10000000-0000-4000-8000-000000000001",
		MarketID: "20000000-0000-4000-8000-000000000001",
		Outcome:  "YES",
		Side:     "BUY",
		Quantity: "1",
		Price:    "0.5",
	}
}
