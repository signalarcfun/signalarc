ALTER TABLE markets
    ADD COLUMN market_contract_address TEXT UNIQUE,
    ADD COLUMN market_deployment_tx_hash TEXT,
    ADD COLUMN market_factory_address TEXT,
    ADD COLUMN resolver_address TEXT,
    ADD COLUMN onchain_deployment_status TEXT NOT NULL DEFAULT 'NOT_DEPLOYED';

ALTER TABLE markets
    ADD CONSTRAINT chk_markets_onchain_deployment_status CHECK (
        onchain_deployment_status IN ('NOT_DEPLOYED', 'DEPLOYED', 'FAILED')
    );

