# Grant Readiness

SignalArc is positioned as Arc-native prediction market infrastructure for USDC-settled event markets, market intelligence, resolver workflow, and agent-readable APIs.

This document separates implemented and live work from testnet prototype behavior, planned work, and unknown / not documented behavior. It documents current readiness honestly; it does not claim full grant submission completeness.

## Live Surfaces

| Surface | URL | Status |
| --- | --- | --- |
| Frontend | https://signalarc.fun | Live on Vercel. |
| Backend API | https://api.signalarc.fun | Live on GCP Cloud Run service `signalarc-backend-api`. |
| Production database | GCP Cloud SQL PostgreSQL, schema version 18. | Live. |
| Custom GPT | Preconfigured to call https://api.signalarc.fun. | Live. |
| Docs | https://docs.signalarc.fun | Documentation deployment target. |

ngrok URLs are local development conveniences only and are never used by judges or end users.

## Implemented

| Area | Current state |
| --- | --- |
| Frontend MVP | Live Next.js app at https://signalarc.fun. Local development still supported via `pnpm dev:web`. |
| Backend API | Live Go/Chi API at https://api.signalarc.fun implementing health, readiness, schema validation, markets, trade intents, positions, resolutions, settlements, agent markets, and Arc contract metadata. |
| Agent surface | Live endpoints for onboarding start, OTP verify, onboarding lookup, session, wallet, balance, ARC-TESTNET faucet, and market intent preview/confirm/execute. |
| Custom GPT | Preconfigured against the live API. End users and judges do not need to import OpenAPI manually. |
| Local backend stack | Docker Compose runs PostgreSQL and backend. |
| Database schema | Production schema migrated to version 18 on GCP Cloud SQL. |
| Backend container | Includes Node/npm and the global Circle CLI (`@circle-fin/cli`) so ARC-TESTNET agent flows run inside the container image. |
| Wallet frontend | RainbowKit/Wagmi/Viem external wallet UI. |
| Arc Testnet contracts | `SignalArcMarket`, `SignalArcAgentMarket`, and `SignalArcAgentMarketFactory` deployed on Arc Testnet. |

## Available Capabilities

| Capability | Status |
| --- | --- |
| Health | Available. |
| Onboarding | Available. |
| OTP verify | Available. |
| Session | Available. |
| Wallet | Available. |
| Balance | Available. |
| ARC-TESTNET faucet | Available. |
| Market intent lifecycle (preview / confirm / execute) | Available. |

## Out of Scope

| Capability | Status |
| --- | --- |
| Arbitrary transfer | Not available. Out of scope. |
| Withdraw / deposit | Not available. Out of scope. |
| Logout / agent session management | Not available. Out of scope. |
| Mainnet funding | Not available. Out of scope. |
| Arc mainnet deployment | Not deployed. Out of scope. |
| API key enforcement, paid access, autonomous trading, production SLA | Not implemented. |

## Testnet Prototype

| Area | Current state |
| --- | --- |
| Contracts | Arc Testnet `SignalArcMarket`, `SignalArcAgentMarket`, and `SignalArcAgentMarketFactory` prototypes. |
| Browser transaction flow | Frontend includes Arc Testnet USDC approval and `openPosition`. |
| Agent transaction flow | Backend Circle Agent Wallet provider handles agent-driven create/buy/close/resolve/claim/cancel/refund on Arc Testnet through the Circle CLI. |
| Explorer links | Frontend generates Arcscan transaction links. |
| Settlement | Prototype contracts support claim/refund paths in Solidity, but production settlement is not approved. |

## Circle Relevance

| Area | Status |
| --- | --- |
| USDC collateral | Prototype contracts use a USDC-like ERC20 collateral model and the Arc Testnet USDC interface address. |
| Circle Agent Wallet | Live in the agent flow on Arc Testnet through the Circle CLI bundled in the backend container. |
| Circle Developer Platform | Used for ARC-TESTNET agent wallet operations only. |
| Circle Agents | Used through the Custom GPT preconfigured to call the live API. |
| Circle API keys | Must not be committed or exposed. |

## Grant Submission Surface

For grant judges, the recommended testing surface is:

1. The published SignalArc GPT Agent (preconfigured to https://api.signalarc.fun).
2. The live frontend at https://signalarc.fun.
3. The live API at https://api.signalarc.fun.

Judges do not need to import OpenAPI manually. The end-to-end testing flow is documented in [Agent API](./AGENT_API.md#judge--user-testing-guide).

## Remaining Gaps

- Demo video.
- Product README and technical README completeness for grant submission packaging.
- Production monitoring/logging baseline documentation.
- Risk/compliance disclaimer for grant submission.
- Audit status (none claimed).
- Deeper Circle integration features beyond ARC-TESTNET agent flows.

## Unknown / Not Documented

- Arc mainnet deployment path for SignalArc: unknown / not documented in this repository.
- Production custody model: unknown / not documented.
- Production settlement approval: unknown / not documented.
- Compliance approval: unknown / not documented.

## Readiness Assessment

SignalArc has a live frontend, a live backend, a live ARC-TESTNET agent flow accessible through a published Custom GPT, an Arc Testnet contract prototype, and a clear API-first product direction. Remaining work is largely grant submission packaging and production monitoring/audit material, not core product implementation.
