package circleapi

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) GetWalletTokenBalances(ctx context.Context, walletID string) ([]WalletTokenBalance, error) {
	walletID = strings.TrimSpace(walletID)
	if walletID == "" {
		return nil, ErrConfigInvalid
	}
	var decoded envelope[walletBalancesData]
	path := "/v1/w3s/wallets/" + url.PathEscape(walletID) + "/balances"
	if err := c.do(ctx, http.MethodGet, path, nil, &decoded); err != nil {
		return nil, err
	}
	if decoded.Data.TokenBalances == nil {
		return []WalletTokenBalance{}, nil
	}
	return decoded.Data.TokenBalances, nil
}
