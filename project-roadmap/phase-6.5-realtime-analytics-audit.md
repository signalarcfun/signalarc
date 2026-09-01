# Phase 6.5 Realtime Analytics Audit

## Scope

Phase 6.5A audits whether the existing public analytics page can move from a static historical snapshot to near-realtime Arc Testnet analytics using the existing deployed contracts and Arcscan APIs.

Hard boundaries for this phase:
- No smart contract changes.
- No ABI changes, except later documentation of existing verified ABI usage.
- No contract deployment.
- No explorer API keys in frontend code.
- No production writes yet.
- Audit and implementation plan only.

Primary active factory under review:
- Address: `0x02555FC5EE3c53938f2F0356e963865503442A56`
- Explorer: `https://testnet.arcscan.app/address/0x02555FC5EE3c53938f2F0356e963865503442A56`

## Executive Finding

Near-realtime analytics is feasible without changing contracts, but it should be implemented as a backend-owned indexer/cache, not as direct browser queries to Arcscan.

Arcscan exposes enough public API surface to ingest:
- factory transactions
- factory logs
- decoded `MarketDeployed` logs for the verified active factory
- child market logs, including decoded `MarketCreated` and `PositionOpened` events

The backend should own explorer credentials, rate limiting, pagination, ABI decoding fallback, persistence, cache refresh, and public analytics response shaping. The frontend analytics page should consume only a backend analytics summary endpoint.

## Current Analytics Data Flow

Current frontend flow:
- `apps/web/src/app/analytics/page.tsx` renders `AnalyticsShell`.
- `apps/web/src/modules/analytics/components/analytics-shell.tsx` renders the entire dashboard.
- `apps/web/src/modules/analytics/analytics-utils.ts` contains all metrics, links, top markets, status badges, proof points, and limitations as static TypeScript constants.
- `apps/web/src/modules/analytics/types.ts` defines display-only types for these static constants.

Current backend flow:
- No analytics read endpoint was found.
- No analytics indexer, repository, migration, cache table, or scheduled ingestion worker was found.
- Existing backend endpoints provide market lists, market detail, Arc contract metadata, agent intent/execution APIs, and status APIs.
- `backend/internal/api/arc_handlers.go` exposes `/arc/contract`, including explorer URL and the configured `SIGNAL_ARC_MARKET_FACTORY_ADDRESS`.
- `backend/internal/repository/markets.go` stores deployed market metadata for backend-created markets, including `market_contract_address`, `market_deployment_tx_hash`, `market_factory_address`, and `resolver_address`.

Current public page state:
- The page presents a historical "Legacy Factory Snapshot" for `0x837e09E8D7806E0e7b740b798173756315E51206`.
- It also links to the active factory `0x02555FC5EE3c53938f2F0356e963865503442A56`, but the displayed metrics are not computed from that active factory at request time.

## Current Freshness Limitation

The current analytics page is not a 24h live query, a cached backend query, or an Arcscan query. It is a static compile-time snapshot embedded in frontend source.

Freshness limits:
- Metrics do not update when new markets are created.
- Metrics do not update when trades, claims, refunds, cancels, closes, or resolutions occur.
- Top markets do not update from chain activity.
- Unique wallet counts and USDC volume are historical constants.
- Active factory data and legacy snapshot data are mixed in page copy and links.

## Contract and Event Map

### Active Factory

The active factory address is verified by Arcscan as `SignalArcMarketFactory` and emits:

`MarketDeployed(string indexed marketId, address indexed market, address indexed creator, address resolver, address collateralToken, uint256 closeTimestamp, string question)`

Repo sources:
- `contracts/src/SignalArcMarketFactory.sol`
- `apps/web/src/lib/contracts/abis/index.ts`

Analytics usage:
- Total created markets: count `MarketDeployed` events or read `marketCount()`.
- Market address discovery: `market` indexed event parameter or `allMarkets(index)`.
- Unique creators: indexed `creator`.
- Per-market metadata: `marketId`, `resolver`, `collateralToken`, `closeTimestamp`, `question`.
- Latest activity: latest successful `MarketDeployed` block/timestamp/transaction.

### Market Contract

The active factory deploys `SignalArcMarket`, which emits:

`MarketCreated(string question, uint256 closeTimestamp, address resolver)`
- Constructor event. Useful for verification, but factory `MarketDeployed` is better for global market discovery.

`MarketClosed()`
- Counts closed markets if needed, but the requested dashboard metric is open/resolved/cancelled. Open should be derived from current status/time or status reads, not from this event alone.

`MarketCancelled()`
- Cancelled market count.

`MarketResolved(Outcome winningOutcome)`
- Resolved market count and YES/NO outcome distribution.

`PositionOpened(address indexed user, Outcome indexed side, uint256 amount)`
- Total trades: count events.
- YES position events: count events where `side == 1`.
- NO position events: count events where `side == 2`.
- Unique wallets: distinct indexed `user` across position, claim, and refund events depending on desired definition.
- USDC volume: sum `amount`, normalized by USDC decimals. Arc Testnet USDC address in repo docs and runtime config is `0x3600000000000000000000000000000000000000`.

`PayoutClaimed(address indexed user, uint256 amount)`
- Claim events for resolved markets.
- Claimed payout volume.
- Claiming wallets.

`RefundClaimed(address indexed user, uint256 amount)`
- Refund events for cancelled markets.
- Refunded volume.
- Refunding wallets.

Repo source:
- `contracts/src/SignalArcMarket.sol`

### Agent Contract Variant

The repo also has `SignalArcAgentMarketFactory` and `SignalArcAgentMarket` under `contracts/src/agent`.

Those contracts emit `AgentMarketDeployed`, `AgentPositionOpened`, `AgentPayoutClaimed`, and `AgentRefundClaimed`. Backend Circle CLI execution code references this agent ABI, while the audited active factory address on Arcscan is verified as the non-agent `SignalArcMarketFactory` and emits `MarketDeployed`.

Phase 6.5B should explicitly choose the active analytics contract family per configured factory address:
- For `0x02555FC5EE3c53938f2F0356e963865503442A56`, index `MarketDeployed` and `SignalArcMarket` events.
- Keep the agent-event map documented for future factories only if an agent factory address is configured and verified.

## Metric Derivation Map

| Metric | Primary source | Derivation |
| --- | --- | --- |
| Total created markets | Factory `MarketDeployed` logs, plus `marketCount()` reconciliation | Count unique deployed market addresses |
| Open markets | Per-market `status()`, `isOpen()`, and `closeTimestamp` | Count markets with status `Open` and current time before close timestamp |
| Resolved markets | `MarketResolved` logs and/or `status()` | Count markets with resolved event or current status `Resolved` |
| Cancelled markets | `MarketCancelled` logs and/or `status()` | Count markets with cancelled event or current status `Cancelled` |
| Total trades | `PositionOpened` logs | Count all position events |
| YES position events | `PositionOpened` logs | Count side `Yes` / enum value `1` |
| NO position events | `PositionOpened` logs | Count side `No` / enum value `2` |
| Unique wallets | `PositionOpened.user`, optionally claim/refund users | Distinct lowercase wallet addresses |
| USDC volume | `PositionOpened.amount` | Sum base units and normalize by token decimals |
| Claim events | `PayoutClaimed` and `RefundClaimed` logs | Count both classes, with separate payout/refund counts |

## Explorer/API Capability Map

Observed Arcscan behavior on 2026-06-28:

| Capability | Endpoint shape | Result |
| --- | --- | --- |
| Factory transactions | `/api?module=account&action=txlist&address=...` | Returned successful `createMarket` transactions for the active factory. |
| Factory logs | `/api?module=logs&action=getLogs&address=...&fromBlock=0&toBlock=latest` | Returned factory logs, including `MarketDeployed` topic and timestamps. |
| Decoded factory logs | `/api/v2/addresses/{address}/logs` | Returned decoded `MarketDeployed(...)` records and `next_page_params`. |
| Decoded factory transactions | `/api/v2/addresses/{address}/transactions` | Returned decoded `createMarket(...)` input and pagination params. |
| Child market logs | `/api/v2/addresses/{market}/logs` | Returned decoded `MarketCreated(...)` and `PositionOpened(...)` logs for a newly deployed market. |
| Etherscan proxy module | `/api?module=proxy&action=eth_blockNumber` | Returned `Unknown module`; do not rely on Etherscan proxy API on Arcscan. |

Recommended ingestion preference:
1. Use Blockscout v2 logs endpoint for decoded logs and cursor pagination.
2. Keep Etherscan-style `logs/getLogs` as a fallback for raw logs where v2 decoding or pagination fails.
3. Use Arc JSON-RPC directly for read reconciliation (`marketCount`, `allMarkets`, `status`, `isOpen`, `totalCollateral`) if backend RPC config is available.

The `contracts/.env` file contains a `BLOCKSCOUT_API_KEY` variable name, but the value must remain local-only and server-side. No frontend environment variable should expose it.

## Recommended Backend Architecture

Add a backend analytics package that owns all explorer interaction and serves a public read model.

Recommended components:
- `backend/internal/analytics/types.go`: canonical event, market, and summary DTOs.
- `backend/internal/analytics/abi.go`: existing verified ABI fragments for decoding only.
- `backend/internal/analytics/arcscan_client.go`: Blockscout/Arcscan client with server-side API key injection, timeout, retry, and pagination.
- `backend/internal/analytics/indexer.go`: idempotent incremental indexer.
- `backend/internal/analytics/aggregator.go`: derives dashboard metrics from indexed events and readbacks.
- `backend/internal/repository/analytics.go`: persistence layer.
- `backend/internal/api/analytics_handlers.go`: `GET /analytics/summary` and optional `GET /analytics/markets`.
- `backend/internal/config/config.go`: server-side config for Arcscan base URL, API key, active factory address, polling interval, and from-block.

Frontend later:
- Replace static constants with a backend fetch layer.
- Keep the page presentation in `AnalyticsShell`.
- Add loading/error/empty states already present in the module.
- Do not call Arcscan from `apps/web`.

## Proposed DB/Cache Schema

Use durable tables for event ingestion and one summary cache table for fast frontend reads.

Suggested migration: `backend/migrations/0000XX_create_analytics_index.up.sql`

Tables:

`analytics_indexer_state`
- `source TEXT PRIMARY KEY`
- `factory_address TEXT NOT NULL`
- `last_indexed_block BIGINT NOT NULL`
- `last_indexed_log_key TEXT`
- `last_success_at TIMESTAMPTZ`
- `last_error TEXT`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`

`analytics_markets`
- `market_address TEXT PRIMARY KEY`
- `factory_address TEXT NOT NULL`
- `market_id_hash TEXT`
- `creator_address TEXT`
- `resolver_address TEXT`
- `collateral_token_address TEXT`
- `question TEXT`
- `close_timestamp TIMESTAMPTZ`
- `deployment_tx_hash TEXT`
- `deployment_block BIGINT`
- `deployment_timestamp TIMESTAMPTZ`
- `status TEXT`
- `winning_outcome TEXT`
- `total_yes NUMERIC`
- `total_no NUMERIC`
- `total_collateral NUMERIC`
- `last_indexed_block BIGINT`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`

`analytics_events`
- `chain_id INTEGER NOT NULL`
- `contract_address TEXT NOT NULL`
- `market_address TEXT`
- `factory_address TEXT`
- `event_name TEXT NOT NULL`
- `transaction_hash TEXT NOT NULL`
- `block_number BIGINT NOT NULL`
- `log_index INTEGER NOT NULL`
- `block_timestamp TIMESTAMPTZ`
- `wallet_address TEXT`
- `side TEXT`
- `amount_base_units NUMERIC`
- `raw JSONB NOT NULL`
- primary key `(chain_id, transaction_hash, log_index)`

`analytics_summary_cache`
- `cache_key TEXT PRIMARY KEY`
- `factory_address TEXT NOT NULL`
- `payload JSONB NOT NULL`
- `latest_block BIGINT`
- `latest_event_at TIMESTAMPTZ`
- `generated_at TIMESTAMPTZ NOT NULL DEFAULT now()`

Indexes:
- `analytics_events (factory_address, block_number, log_index)`
- `analytics_events (market_address, event_name)`
- `analytics_events (wallet_address)`
- `analytics_markets (factory_address, status)`

## Refresh Interval Recommendation

Recommended defaults:
- Poll interval: 30 seconds for the active factory and recently active markets.
- Backoff: 2 minutes after rate-limit or transient explorer failures.
- Full reconciliation: every 10 to 15 minutes, read `marketCount()` and selected market state from RPC to catch missed logs or status drift.
- Historical backfill: batch by block ranges or v2 pagination until the factory deployment block is fully indexed.
- Confirmation lag: process only blocks at least 2 to 5 blocks behind the latest observed block if reorg safety is desired. For Arc Testnet dashboard freshness, 2 blocks is likely enough, but exact finality should be treated as unknown unless documented by Arc.

Public frontend cache:
- Backend response can be cached for 15 to 30 seconds.
- Include `generated_at`, `latest_event_at`, `latest_block`, and `source_status`.

## Risks and Rate Limits

Risks:
- Arcscan rate limits are not documented in repo. Treat them as unknown until key-specific limits are verified.
- Blockscout v2 pagination must be handled correctly through `next_page_params`.
- Etherscan-style `getLogs` can return large result sets; use block windows and event topics.
- Factory logs alone do not prove current open/resolved/cancelled state; use market logs plus periodic contract reads.
- Indexed `marketId` is a string, so event topics store a hash rather than the original string in raw logs. Decoded v2 logs currently recover input values when the contract is verified, but the backend should not depend exclusively on decoded explorer output.
- Unique wallets are addresses, not unique people.
- USDC volume is testnet collateral movement, not production or mainnet trading volume.
- Circle Agent Wallet attribution is not visible from chain alone; the chain shows transaction senders and market users, not Circle session/user identity.
- The active factory page and current static analytics page use different factories; implementation must avoid mixing active and legacy metrics.

Security requirements:
- `BLOCKSCOUT_API_KEY` stays backend-only.
- No API keys or private RPC URLs in `NEXT_PUBLIC_*`.
- Do not store raw Circle CLI session output or secrets in analytics events.
- Do not treat analytics indexing as authorization, settlement, or production write proof.

## Implementation Subphases

### Phase 6.5B - Backend Indexer Skeleton and Backfill

Goal:
- Add server-side analytics config, Arcscan client, event DTOs, migrations, and a CLI/manual backfill entry point.

Files expected to change:
- `backend/internal/config/config.go`
- `backend/internal/analytics/types.go`
- `backend/internal/analytics/abi.go`
- `backend/internal/analytics/arcscan_client.go`
- `backend/internal/analytics/indexer.go`
- `backend/internal/repository/analytics.go`
- `backend/migrations/0000XX_create_analytics_index.up.sql`
- `backend/migrations/0000XX_create_analytics_index.down.sql`
- `backend/internal/analytics/*_test.go`
- `backend/.env.example`

No frontend changes required in 6.5B.

### Phase 6.5C - Analytics Aggregation API

Goal:
- Add a backend read API that returns the dashboard summary from indexed/cached data.

Files expected to change:
- `backend/internal/analytics/aggregator.go`
- `backend/internal/api/analytics_handlers.go`
- `backend/internal/api/router.go` or the existing route registration file
- `backend/internal/api/analytics_handlers_test.go`
- `docs/API.md`
- `project-roadmap/signalarc-custom-gpt-openapi.json` only if agent clients need this read-only endpoint

Endpoint recommendation:
- `GET /analytics/summary`

Response should include:
- factory address
- generated timestamp
- latest indexed block
- latest event timestamp
- metrics
- lifecycle counts
- top markets
- source status and freshness warning

### Phase 6.5D - Frontend Realtime Analytics Page

Goal:
- Replace static metrics with backend analytics data while preserving the current analytics presentation.

Files expected to change:
- `apps/web/src/modules/analytics/types.ts`
- `apps/web/src/modules/analytics/analytics-utils.ts`
- `apps/web/src/modules/analytics/components/analytics-shell.tsx`
- `apps/web/src/app/analytics/page.tsx` only if server-side fetch metadata is needed
- optional new file: `apps/web/src/modules/analytics/api.ts`
- `apps/web/src/modules/analytics/components/analytics-loading-skeleton.tsx`
- `apps/web/src/modules/analytics/components/analytics-error-state.tsx`
- `apps/web/.env.example` only for backend API base URL, not Arcscan

Frontend must not change:
- No Arcscan API key.
- No direct Arcscan log polling.
- No contract writes.
- No ABI mutation.

## Validation Plan for Later Phases

Backend:
- Unit tests for decoding factory and market events.
- Unit tests for pagination and rate-limit backoff.
- Repository tests for idempotent event upserts.
- Aggregator tests for metric derivation.
- Manual backfill dry run against active factory without writing production data.

Frontend:
- Build and lint after 6.5D.
- Browser check for loading, error, empty, and populated states.
- Verify displayed factory address matches `0x02555FC5EE3c53938f2F0356e963865503442A56`.
- Verify page shows freshness metadata and testnet limitations.

## Phase 6.5A Conclusion

Upgrade path is clear:
1. Keep existing contracts and ABIs unchanged.
2. Build a backend-only Arcscan/Blockscout event indexer.
3. Persist normalized factory and market events.
4. Aggregate into a cached summary.
5. Point the analytics page at the backend summary endpoint.

Do not implement the page as client-side Arcscan polling. The API key, retry behavior, pagination state, and event reconciliation belong in the backend.
