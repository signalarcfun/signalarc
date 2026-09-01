package trade

type CreateTradeIntentRequest struct {
	UserID   string `json:"user_id"`
	MarketID string `json:"market_id"`
	Outcome  string `json:"outcome"`
	Side     string `json:"side"`
	Quantity string `json:"quantity"`
	Price    string `json:"price"`
}
