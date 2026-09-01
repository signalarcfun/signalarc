# SignalArc Documentation

SignalArc is an Arc-native prediction market infrastructure platform for USDC-settled event markets, market intelligence, resolver workflows, and agent-readable APIs. It is an API-first Web3 application stack for creators, institutions, developers, and AI agents.

SignalArc is not positioned as a clone of any existing prediction market product. The repository currently contains a modular local MVP, an Arc Testnet contract prototype, and deployment planning artifacts. Production deployment is not completed.

## Current Status

| Area | Status |
| --- | --- |
| Frontend | Next.js app runs manually with `pnpm dev:web`; current working tree includes RainbowKit/Wagmi/Viem wallet UI and Arc Testnet browser transaction wiring. |
| Backend API | Go/Chi API runs locally with PostgreSQL through Docker. |
| Database | PostgreSQL schema migrations are complete through version 13 in local development state. |
| Smart contract | `SignalArcMarket` is deployed on Arc Testnet as a prototype reference. |
| Circle integration | Circle SDK/API integration is not implemented in the current repository. |
| Production deployment | Not completed. DNS is not configured. |

## Planned Public URLs

These domains are purchased or planned deployment targets. They are not live deployment claims.

| Surface | Planned URL | Status |
| --- | --- | --- |
| Frontend | https://signalarc.fun | Planned target; live deployment is not completed. |
| API | https://api.signalarc.fun | Planned target; live deployment is not completed. |
| Docs | https://docs.signalarc.fun | Planned target; live deployment is not completed. |

## Documentation Map

- [API.md](./API.md): implemented REST API reference.
- [AGENT_API.md](./AGENT_API.md): agent-readable API reference and limitations.
- [ARCHITECTURE.md](./ARCHITECTURE.md): system architecture and deployment model.
- [LOCAL_DEVELOPMENT.md](./LOCAL_DEVELOPMENT.md): local setup, commands, and troubleshooting.
- [FRONTEND_WALLET.md](./FRONTEND_WALLET.md): external wallet integration and Arc Testnet trade flow.
- [ONCHAIN_INTEGRATION.md](./ONCHAIN_INTEGRATION.md): Arc Testnet contract integration details.
- [DEPLOYMENT_PLAN.md](./DEPLOYMENT_PLAN.md): planned production deployment path.
- [SECURITY_AND_SECRETS.md](./SECURITY_AND_SECRETS.md): security and secret handling policy.
- [GRANT_READINESS.md](./GRANT_READINESS.md): Circle/Arc grant readiness status.
- [privacy.mdx](./privacy.mdx): privacy policy for the SignalArc Arc Testnet prototype and Custom GPT Action.

## Implemented Capabilities

- Local backend and PostgreSQL can run through Docker Compose.
- Frontend can run manually with the monorepo `pnpm dev:web` script.
- Backend exposes health, readiness, schema validation, market, trade-intent, position, resolution, settlement, agent-market, and Arc contract metadata endpoints.
- Arc Testnet contract reference is available through the backend and frontend constants.
- Current frontend working tree includes external wallet connection and Arc Testnet transaction flow for USDC approval plus `openPosition`.
- Current frontend wallet execution is Arc Testnet prototype behavior only.

## Current Limitations

- No production deployment is live.
- DNS is not configured.
- No Arc mainnet deployment is present.
- No production custody or production settlement approval is claimed.
- No audit is claimed.
- No Circle SDK/API integration is implemented unless added in a later change.
- No API authentication, API key enforcement, paid access, autonomous trading, or production SLA is implemented.
