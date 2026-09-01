CREATE TABLE analytics_indexer_state (
    source TEXT PRIMARY KEY,
    factory_address TEXT NOT NULL,
    last_indexed_block BIGINT NOT NULL DEFAULT 0,
    last_indexed_log_key TEXT,
    last_success_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_analytics_indexer_state_source_non_empty CHECK (
        length(btrim(source)) > 0
    ),
    CONSTRAINT chk_analytics_indexer_state_factory_non_empty CHECK (
        length(btrim(factory_address)) > 0
    ),
    CONSTRAINT chk_analytics_indexer_state_last_block_non_negative CHECK (
        last_indexed_block >= 0
    )
);

CREATE TABLE analytics_markets (
    market_address TEXT PRIMARY KEY,
    factory_address TEXT NOT NULL,
    market_id_hash TEXT,
    creator_address TEXT,
    resolver_address TEXT,
    collateral_token_address TEXT,
    question TEXT,
    close_timestamp TIMESTAMPTZ,
    deployment_tx_hash TEXT,
    deployment_block BIGINT,
    deployment_timestamp TIMESTAMPTZ,
    status TEXT,
    winning_outcome TEXT,
    total_yes NUMERIC NOT NULL DEFAULT 0,
    total_no NUMERIC NOT NULL DEFAULT 0,
    total_collateral NUMERIC NOT NULL DEFAULT 0,
    last_indexed_block BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_analytics_markets_market_address_non_empty CHECK (
        length(btrim(market_address)) > 0
    ),
    CONSTRAINT chk_analytics_markets_factory_non_empty CHECK (
        length(btrim(factory_address)) > 0
    ),
    CONSTRAINT chk_analytics_markets_status CHECK (
        status IS NULL OR status IN ('OPEN', 'CLOSED', 'RESOLVED', 'CANCELLED')
    ),
    CONSTRAINT chk_analytics_markets_winning_outcome CHECK (
        winning_outcome IS NULL OR winning_outcome IN ('YES', 'NO')
    ),
    CONSTRAINT chk_analytics_markets_totals_non_negative CHECK (
        total_yes >= 0 AND total_no >= 0 AND total_collateral >= 0
    )
);

CREATE INDEX idx_analytics_markets_factory_status ON analytics_markets (factory_address, status);
CREATE INDEX idx_analytics_markets_deployment_block ON analytics_markets (deployment_block);

CREATE TABLE analytics_events (
    chain_id INTEGER NOT NULL,
    contract_address TEXT NOT NULL,
    market_address TEXT,
    factory_address TEXT,
    event_name TEXT NOT NULL,
    transaction_hash TEXT NOT NULL,
    block_number BIGINT NOT NULL,
    log_index INTEGER NOT NULL,
    block_timestamp TIMESTAMPTZ,
    wallet_address TEXT,
    side TEXT,
    amount_base_units NUMERIC NOT NULL DEFAULT 0,
    raw JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (chain_id, transaction_hash, log_index),

    CONSTRAINT chk_analytics_events_contract_address_non_empty CHECK (
        length(btrim(contract_address)) > 0
    ),
    CONSTRAINT chk_analytics_events_event_name_non_empty CHECK (
        length(btrim(event_name)) > 0
    ),
    CONSTRAINT chk_analytics_events_tx_hash_non_empty CHECK (
        length(btrim(transaction_hash)) > 0
    ),
    CONSTRAINT chk_analytics_events_block_non_negative CHECK (
        block_number >= 0
    ),
    CONSTRAINT chk_analytics_events_log_index_non_negative CHECK (
        log_index >= 0
    ),
    CONSTRAINT chk_analytics_events_side CHECK (
        side IS NULL OR side IN ('YES', 'NO')
    ),
    CONSTRAINT chk_analytics_events_amount_non_negative CHECK (
        amount_base_units >= 0
    )
);

CREATE INDEX idx_analytics_events_factory_block ON analytics_events (factory_address, block_number, log_index);
CREATE INDEX idx_analytics_events_market_event ON analytics_events (market_address, event_name);
CREATE INDEX idx_analytics_events_wallet ON analytics_events (wallet_address)
    WHERE wallet_address IS NOT NULL;
CREATE INDEX idx_analytics_events_event_name ON analytics_events (event_name);

CREATE TABLE analytics_summary_cache (
    cache_key TEXT PRIMARY KEY,
    factory_address TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    latest_block BIGINT,
    latest_event_at TIMESTAMPTZ,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_analytics_summary_cache_key_non_empty CHECK (
        length(btrim(cache_key)) > 0
    ),
    CONSTRAINT chk_analytics_summary_cache_factory_non_empty CHECK (
        length(btrim(factory_address)) > 0
    ),
    CONSTRAINT chk_analytics_summary_cache_latest_block_non_negative CHECK (
        latest_block IS NULL OR latest_block >= 0
    )
);

CREATE INDEX idx_analytics_summary_cache_factory ON analytics_summary_cache (factory_address);
