# Phase 6.5E Frontend Analytics Integration

## Scope

Phase 6.5E wires the existing `/analytics` page to the backend-owned analytics summary endpoint:

- `GET /analytics/summary`

This phase does not:
- modify smart contracts
- modify contract ABIs
- expose `BLOCKSCOUT_API_KEY` or explorer secrets to frontend code
- add direct Arcscan or Blockscout frontend calls
- add scheduled polling
- change backend ingestion behavior
- deploy, push, or create a PR

## Frontend Data Flow

The analytics page now loads the public backend summary through a typed frontend module:

- `apps/web/src/modules/analytics/api.ts`
- `apps/web/src/modules/analytics/use-analytics-summary.ts`

The module calls the existing frontend API helper in `apps/web/src/lib/api.ts`, so browser code only talks to the SignalArc backend base URL configured by `NEXT_PUBLIC_API_BASE_URL` or the local fallback `http://localhost:4000`.

No explorer API key or server-side ingestion configuration is referenced by the frontend.

## Render Behavior

When the backend returns an indexed or cached summary:
- executive analytics metric cards render backend metrics
- the hero provenance card shows backend-cache provenance
- freshness metadata shows `source_status`, `generated_at`, `latest_event_at`, `latest_block`, and `factory_address`
- active factory links point at the factory address returned by the backend

When the backend is loading, unavailable, or not indexed:
- the page keeps the historical static analytics snapshot visible
- loading, error, or empty state messaging explains the fallback
- no runtime exception is surfaced to the user for missing analytics cache data

## Precision Handling

`metrics.testnet_usdc_volume` is treated as a base-unit string and formatted with string arithmetic in:

- `apps/web/src/modules/analytics/analytics-utils.ts`

This avoids converting potentially large base-unit values through JavaScript floating point numbers.

## Files Changed

- `apps/web/src/modules/analytics/api.ts`
- `apps/web/src/modules/analytics/use-analytics-summary.ts`
- `apps/web/src/modules/analytics/types.ts`
- `apps/web/src/modules/analytics/analytics-utils.ts`
- `apps/web/src/modules/analytics/components/analytics-shell.tsx`
- `apps/web/src/modules/analytics/index.ts`
- `project-roadmap/phase-6.5e-frontend-analytics-integration.md`

## Validation Plan

Expected local validation:

- `pnpm lint`
- `pnpm build`
- `pnpm dev:web`
- local browser check of `/analytics`
- confirm browser traffic only requests the backend `/analytics/summary` endpoint for analytics data
- confirm the rendered page shows backend metrics from local `/analytics/summary` when the backend is running and indexed

## Next Phase Boundary

This phase keeps the frontend read-only. Any refresh controls, polling, richer market leaderboards, or production ingestion scheduling should be planned as separate follow-up work.
