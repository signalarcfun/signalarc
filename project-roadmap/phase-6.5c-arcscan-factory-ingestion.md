# Phase 6.5C Arcscan Factory Ingestion

## Scope

Phase 6.5C adds manual backend ingestion for active factory `MarketDeployed` events only.

This phase does not:
- modify smart contracts
- modify contract ABIs
- modify frontend files
- add autonomous or scheduled production polling
- ingest child market events
- change trading, agent, market creation, lifecycle, or settlement behavior
- expose `BLOCKSCOUT_API_KEY` to frontend code or public API responses

## Implemented Backend Pieces

New package:
- `backend/internal/analytics`

It provides:
- Blockscout v2 address logs client for `GET /api/v2/addresses/{address}/logs`
- optional server-side `BLOCKSCOUT_API_KEY` query support
- configurable Arcscan base URL, defaulting to `https://testnet.arcscan.app`
- HTTP timeout handling
- `next_page_params` pagination support
- decoded `MarketDeployed(...)` parser
- manual backfill service with dry-run and write modes

Repository additions:
- upsert `analytics_markets` from factory `MarketDeployed`
- insert `analytics_events` idempotently on `(chain_id, transaction_hash, log_index)`
- update `analytics_indexer_state`
- rebuild `analytics_summary_cache` from indexed analytics rows

Manual command:
- `backend/cmd/analytics-backfill`

Docker runtime availability:
- The backend Docker image builds and copies both `signalarc-api` and `analytics-backfill`.
- The image entrypoint remains `signalarc-api`, while manual backfills/jobs can invoke `analytics-backfill` explicitly with an override entrypoint.

Safe default:
- `-dry-run=true`
- `-page-limit=1`

Example dry run from `backend/`:

```bash
go run ./cmd/analytics-backfill \
  -dry-run=true \
  -factory 0x02555FC5EE3c53938f2F0356e963865503442A56 \
  -page-limit 1
```

Example local write run, only after local DB migrations are at v21:

```bash
DATABASE_URL="postgres://<user>:<password>@127.0.0.1:15433/signalarc?sslmode=disable" \
go run ./cmd/analytics-backfill \
  -dry-run=false \
  -factory 0x02555FC5EE3c53938f2F0356e963865503442A56 \
  -page-limit 1
```

The command reads `BLOCKSCOUT_API_KEY` from the process environment. For local convenience only, if the variable is absent it attempts to read only the `BLOCKSCOUT_API_KEY` line from `../contracts/.env` or `contracts/.env`; it does not load or print the rest of that file.

## Summary Behavior

After a successful local write run:
- `analytics_events.event_name = 'MarketDeployed'`
- `analytics_markets` contains discovered market contract rows
- `analytics_summary_cache` is rebuilt
- `GET /analytics/summary` should report non-zero:
  - `metrics.markets_created`
  - `metrics.market_contracts_found`
  - `latest_block`
  - `latest_event_at`

Because this phase ingests only factory events, trade and lifecycle metrics remain zero until later child-market event ingestion.

## Validation Notes

Expected local validation:
- `cd backend && go test -count=1 ./...`
- `git diff --check`
- `go run ./cmd/analytics-backfill -dry-run=true -page-limit 1`
- local DB-only write run after v21 migrations are applied
- local `GET /analytics/summary` after write run
- local Docker backend rebuild if Docker is available

## Next Phase Boundary

Phase 6.5D or a later ingestion phase can add child market event ingestion for:
- `PositionOpened`
- `MarketResolved`
- `MarketCancelled`
- `PayoutClaimed`
- `RefundClaimed`

That phase should keep the same manual/idempotent ingestion discipline until scheduled polling is explicitly approved.
