# Deployment Topology

This document describes the live deployment topology for SignalArc. Operational commands, secrets, and credentials are intentionally excluded.

## Current Deployment Status

| Item | Status |
| --- | --- |
| Production frontend | Live at `https://signalarc.fun` on Vercel. |
| Production backend | Live at `https://api.signalarc.fun` on GCP Cloud Run service `signalarc-backend-api`. |
| Production database | GCP Cloud SQL PostgreSQL, migrated to version 18. |
| Backend container image | Includes Node/npm and the global `@circle-fin/cli` binary. `signalarc-api` and `circle` are both on `PATH`. |
| DNS | `signalarc.fun` and `api.signalarc.fun` are configured. |
| Custom GPT integration | Live and pointed at `https://api.signalarc.fun`. |
| Arc Testnet contract | Deployed prototype reference exists. |
| Arc mainnet | Not deployed and not approved. |

## Live Domains

| Domain | Target |
| --- | --- |
| `https://signalarc.fun` | Vercel frontend. |
| `https://api.signalarc.fun` | GCP Cloud Run service `signalarc-backend-api`. |
| `https://docs.signalarc.fun` | Documentation deployment target. |

ngrok URLs are local development conveniences only. Production traffic always goes through `https://signalarc.fun` and `https://api.signalarc.fun`.

## Live Architecture

```text
Browser / Custom GPT
        |
        v
https://signalarc.fun  (Vercel frontend)
        |
        v
https://api.signalarc.fun  (Cloud Run signalarc-backend-api)
        |
        +--> Cloud SQL PostgreSQL (schema version 18)
        |
        +--> Circle CLI (@circle-fin/cli) inside the container
        |
        +--> Arc Testnet RPC
```

## Frontend

- Platform: Vercel.
- Primary domain: `signalarc.fun`.
- Optional alias: `www.signalarc.fun`.
- Build is driven from the monorepo `apps/web` workspace.
- Public env vars (no secrets) are configured in Vercel project settings.

## Backend

- Platform: GCP Cloud Run.
- Service name: `signalarc-backend-api`.
- Domain: `api.signalarc.fun`.
- Runtime: Docker container built from the Go backend.
- Container layout:
  - `/usr/local/bin/signalarc-api` — Go API binary.
  - `/usr/local/bin/circle` — Circle CLI installed globally via `npm install -g @circle-fin/cli`.
- Production CORS allows the production frontend origin `https://signalarc.fun` and local development origins.

## Database

- Provider: GCP Cloud SQL (PostgreSQL).
- Schema migrations are applied through the approved migration workflow.
- Current production migration version: 18.
- Migration catalog state is not edited manually.

## DNS

| Record | Target |
| --- | --- |
| `signalarc.fun` | Vercel frontend target. |
| `www.signalarc.fun` | Optional Vercel frontend alias. |
| `api.signalarc.fun` | Cloud Run custom domain mapping. |
| `docs.signalarc.fun` | Documentation hosting target. |

## Environment Variables

Frontend public variables:

- `NEXT_PUBLIC_API_BASE_URL=https://api.signalarc.fun`
- `NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID`

Backend variables (set in Cloud Run, not in source control):

- `APP_ENV`
- `APP_PORT`
- `DATABASE_URL`
- `CIRCLE_CLI_PATH`
- `CIRCLE_AGENT_WALLET_CHAIN`
- `CIRCLE_AGENT_WALLET_TIMEOUT_SECONDS`
- `CIRCLE_AGENT_WALLET_EXECUTION_ENABLED`
- `CIRCLE_AGENT_ONBOARDING_OTP_START_ENABLED`
- `CIRCLE_AGENT_WALLET_FAUCET_ENABLED`

Secrets (private keys, API keys, OTPs, Circle request IDs, Circle session files) are never stored in source control, the container image, or the database.

## Pre-Deploy Checklist

- Local Docker backend health passes.
- Local backend readiness passes.
- Local schema validation passes.
- Frontend lint passes.
- Frontend type check passes.
- Frontend production build passes.
- Wallet flow is tested on Arc Testnet.
- Backend tests pass via `go test ./...`.
- No secrets are present in git.
- Production CORS is configured for `https://signalarc.fun`.
- DNS targets are known and verified.
- Rollback plan is documented.
- Monitoring/logging baseline is defined.

## Out of Scope

- Arc mainnet deployment.
- Production custody or production settlement approval.
- Audit claim.
- Arbitrary transfer, withdraw/deposit, agent logout endpoints, and mainnet funding actions.
- Operational deploy commands and provider credentials, which live outside this document.
