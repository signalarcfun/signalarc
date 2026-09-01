# Phase 6.5B Backend Analytics Cache

## Scope

Phase 6.5B adds the backend-owned persistence and public read endpoint needed before realtime analytics indexing is implemented.

This phase does not:
- modify smart contracts
- modify contract ABIs
- poll Arcscan
- add a scheduler
- expose `BLOCKSCOUT_API_KEY` to the frontend
- change the frontend analytics page
- write or assume production analytics data

## Implemented Foundation

Backend migration `000021_create_analytics_cache` adds:
- `analytics_indexer_state`
- `analytics_markets`
- `analytics_events`
- `analytics_summary_cache`

`analytics_events` uses primary key `(chain_id, transaction_hash, log_index)` so later indexers can safely upsert idempotently.

Backend repository support adds:
- latest summary cache reads by cache key
- summary cache upsert support
- aggregate fallback from `analytics_events` and `analytics_markets`
- empty/default fallback semantics when no indexed analytics data exists

Backend API adds:
- `GET /analytics/summary`

The endpoint returns `200 OK` with zero/default metrics and `source_status: "not_indexed"` when no cache or indexed data exists.

## Public Response Shape

The response includes:
- `status`
- `source_status`
- `factory_address`
- `generated_at`
- `latest_event_at`
- `latest_block`
- `metrics.markets_created`
- `metrics.market_contracts_found`
- `metrics.total_trades`
- `metrics.position_events`
- `metrics.yes_position_events`
- `metrics.no_position_events`
- `metrics.unique_wallets`
- `metrics.testnet_usdc_volume`
- `metrics.resolved_markets`
- `metrics.cancelled_markets`
- `metrics.claim_events`
- `metrics.payouts_claimed`
- `metrics.refunds_claimed`

## Files Changed

Backend:
- `backend/migrations/000021_create_analytics_cache.up.sql`
- `backend/migrations/000021_create_analytics_cache.down.sql`
- `backend/internal/repository/analytics.go`
- `backend/internal/api/analytics_handlers.go`
- `backend/internal/api/analytics_handlers_test.go`
- `backend/internal/api/router.go`
- `backend/internal/database/database.go`

Docs:
- `project-roadmap/phase-6.5b-backend-cache.md`

## Next Phase Boundary

Phase 6.5C should add the Arcscan/Blockscout ingestion worker or manual backfill path.

That later phase should:
- keep explorer keys backend-only
- implement pagination and rate-limit handling
- decode existing verified factory and market events only
- populate the cache tables added here
- preserve the frontend static analytics page until the frontend integration phase
