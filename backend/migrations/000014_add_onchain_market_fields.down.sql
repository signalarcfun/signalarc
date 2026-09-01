ALTER TABLE markets
    DROP CONSTRAINT IF EXISTS chk_markets_onchain_deployment_status,
    DROP COLUMN IF EXISTS onchain_deployment_status,
    DROP COLUMN IF EXISTS resolver_address,
    DROP COLUMN IF EXISTS market_factory_address,
    DROP COLUMN IF EXISTS market_deployment_tx_hash,
    DROP COLUMN IF EXISTS market_contract_address;
