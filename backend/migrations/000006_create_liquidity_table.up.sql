CREATE TABLE liquidity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    market_id UUID NOT NULL REFERENCES markets(id) ON DELETE CASCADE,
    collateral_asset TEXT NOT NULL DEFAULT 'USDC',
    total_collateral_amount NUMERIC(36, 18) NOT NULL DEFAULT 0,
    available_collateral_amount NUMERIC(36, 18) NOT NULL DEFAULT 0,
    reserved_collateral_amount NUMERIC(36, 18) NOT NULL DEFAULT 0,
    fee_pool_amount NUMERIC(36, 18) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_liquidity_total_collateral_amount CHECK (total_collateral_amount >= 0),
    CONSTRAINT chk_liquidity_available_collateral_amount CHECK (available_collateral_amount >= 0),
    CONSTRAINT chk_liquidity_reserved_collateral_amount CHECK (reserved_collateral_amount >= 0),
    CONSTRAINT chk_liquidity_fee_pool_amount CHECK (fee_pool_amount >= 0),
    CONSTRAINT chk_liquidity_collateral_allocation CHECK (
        available_collateral_amount + reserved_collateral_amount <= total_collateral_amount
    ),
    CONSTRAINT uq_liquidity_market_id UNIQUE (market_id)
);

CREATE INDEX idx_liquidity_collateral_asset ON liquidity (collateral_asset);
CREATE INDEX idx_liquidity_updated_at ON liquidity (updated_at);
