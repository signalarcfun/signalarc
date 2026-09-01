# Phase 6.5D Child Market Event Ingestion

## Scope

Phase 6.5D extends the manual analytics ingestion path to fetch decoded logs from market contracts already discovered in `analytics_markets`.

This phase does not:
- modify smart contracts
- modify contract ABIs
- modify frontend files
- expose `BLOCKSCOUT_API_KEY` to frontend code or public API responses
- add scheduled or autonomous polling
- change trading, agent, market creation, lifecycle, or settlement behavior

## Implemented Event Coverage

When `analytics-backfill` runs with `-include-market-events=true`, it reads discovered market addresses for the selected factory and ingests decoded child market events:

- `PositionOpened(address indexed user, Outcome indexed side, uint256 amount)`
- `MarketResolved(Outcome winningOutcome)`
- `MarketCancelled()`
- `PayoutClaimed(address indexed user, uint256 amount)`
- `RefundClaimed(address indexed user, uint256 amount)`

Unknown or unrelated decoded logs are ignored safely.

Normalization:
- YES/NO sides and outcomes are normalized to `YES` / `NO`.
- Amounts are stored as base-unit strings compatible with the `NUMERIC` `amount_base_units` column.
- Raw Blockscout log JSON is preserved in `analytics_events.raw`.

## Manual Command

Safe defaults remain:
- `-dry-run=true`
- `-page-limit=1`
- `-include-market-events=false`

Dry-run with child market events enabled requires `DATABASE_URL` because the command must read already discovered market addresses from `analytics_markets`, but it does not write when `-dry-run=true`.

Example:

```bash
DATABASE_URL="postgres://<user>:<password>@127.0.0.1:15433/signalarc?sslmode=disable" \
go run ./cmd/analytics-backfill \
  -dry-run=true \
  -include-market-events=true \
  -page-limit 1
```

Local write run:

```bash
DATABASE_URL="postgres://<user>:<password>@127.0.0.1:15433/signalarc?sslmode=disable" \
go run ./cmd/analytics-backfill \
  -dry-run=false \
  -include-market-events=true \
  -page-limit 1
```

## Summary Behavior

After a successful local write run with market events:
- `analytics_events` includes factory and child market events idempotently.
- `analytics_markets.status` updates to `RESOLVED` or `CANCELLED` when those events are observed.
- `analytics_markets.winning_outcome` updates to `YES` or `NO` when `MarketResolved` is observed.
- `analytics_summary_cache` is rebuilt.
- `GET /analytics/summary` can reflect non-zero trade, YES/NO position, claim, payout, refund, resolved, cancelled, unique wallet, and testnet USDC volume metrics when child events exist.

## Validation Notes

Expected local validation:
- `cd backend && go test -count=1 ./...`
- `git diff --check`
- `docker compose up -d --build signalarc-backend`
- `docker compose run --rm --entrypoint analytics-backfill signalarc-backend --help`
- local dry-run with `-include-market-events=true`
- local DB-only write run with `-include-market-events=true`
- local `GET /analytics/summary` after write run

## Next Phase Boundary

Scheduled polling is still not implemented. Any autonomous ingestion loop should be a separate phase with explicit rate-limit, retry, observability, and production rollout controls.
