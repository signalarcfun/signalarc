# SignalArc

SignalArc is an Arc-native prediction market infrastructure platform for USDC-settled event markets, market intelligence, resolver workflows, and agent-readable APIs.

The project is designed as API-first Web3 infrastructure for creators, institutions, developers, and AI agents. It is not a production deployment yet, and it does not claim mainnet readiness, audit completion, custody approval, or compliance approval.

## Current Status

| Area | Status |
| --- | --- |
| Local MVP | Frontend/backend local MVP has been verified in the current project state. |
| Frontend | Next.js app in `apps/web`; runs manually with `pnpm dev:web`. |
| Backend | Go API in `backend`; local backend and PostgreSQL run through Docker Compose. |
| Contracts | Foundry/Solidity prototype in `contracts`; `SignalArcMarket` is deployed on Arc Testnet. |
| Wallet flow | Current working tree includes a browser-wallet Arc Testnet transaction flow for USDC approval and `openPosition`. |
| Public domain | `signalarc.fun` is the planned deployment domain. |
| Production | Production deployment and DNS are not completed. |

## What SignalArc Provides

- USDC-settled event market infrastructure.
- Arc Testnet smart contract prototype for binary market positions.
- Browser-wallet testnet transaction flow where implemented by the frontend.
- Market creation, market detail, position, resolution, settlement, and intelligence API surfaces.
- Agent-readable market APIs for market discovery and reporting.
- Documentation for local development, API usage, architecture, deployment planning, and security boundaries.

## Architecture

SignalArc is a modular monorepo:

| Path | Purpose |
| --- | --- |
| `apps/web` | Next.js frontend, TypeScript UI, wallet connection, API calls, and current Arc Testnet browser-wallet trade prototype. |
| `backend` | Go/Chi API, validation, repository access, local contract metadata, and agent-readable market endpoints. |
| `contracts` | Foundry smart contracts, Solidity prototype, contract tests, and Arc Testnet deployment notes. |
| `docs` | Public-facing documentation suite. |

Local backend infrastructure uses PostgreSQL and the Go backend container through Docker Compose. The frontend runs separately during development.

## Documentation

- [REST API](./docs/API.md)
- [Agent API](./docs/AGENT_API.md)
- [Architecture](./docs/ARCHITECTURE.md)
- [Local Development](./docs/LOCAL_DEVELOPMENT.md)
- [Onchain Integration](./docs/ONCHAIN_INTEGRATION.md)
- [Deployment Plan](./docs/DEPLOYMENT_PLAN.md)
- [Security and Secrets](./docs/SECURITY_AND_SECRETS.md)

## Local Development Quick Start

Start the local backend stack:

```bash
docker compose up -d signalarc-postgres signalarc-backend
```

Check backend health:

```bash
curl http://localhost:4000/health
curl http://localhost:4000/readyz
```

Start the frontend:

```bash
pnpm dev:web
```

For full setup, environment variables, validation commands, database access, and troubleshooting, see [docs/LOCAL_DEVELOPMENT.md](./docs/LOCAL_DEVELOPMENT.md).

## Planned Public Surfaces

| Surface | Planned URL | Status |
| --- | --- | --- |
| Frontend | https://signalarc.fun | Planned deployment target; not live yet. |
| API | https://api.signalarc.fun | Planned deployment target; not live yet. |
| Docs | https://docs.signalarc.fun | Planned deployment target; not live yet. |

## Security Notice

- Never commit `.env`, `.env.local`, private keys, seed phrases, API keys, database credentials, RPC secrets, or WalletConnect project IDs that are meant to stay local or deployment-only.
- Current contract and wallet behavior is Arc Testnet prototype behavior only.
- No production custody is claimed.
- No production settlement approval is claimed.
- No audit or compliance approval is claimed.
- Backend must not sign user transactions; browser wallets sign user testnet transactions where implemented.

## License

MIT. See [LICENSE](./LICENSE).
