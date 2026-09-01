# SignalArc Agent Instructions

## Project Identity

SignalArc is an Arc-native prediction market infrastructure platform.

It enables creators, institutions, developers, and AI agents to launch, trade, resolve, and analyze USDC-settled event markets with transparent settlement and real-time probability signals.

SignalArc is not a Polymarket clone. It is an API-first infrastructure layer for prediction markets, market intelligence, resolver workflows, and USDC settlement on Arc.

## Architecture

This repository is a modular monorepo.

- `apps/web` contains the Next.js frontend.
- `backend` contains the Go backend API.
- `contracts` will contain Solidity/Foundry smart contracts.
- `docs` will contain public documentation and API references.

## Frontend Rules

The frontend is UI-only.

Allowed responsibilities:

- Landing page
- Market list UI
- Market detail UI
- Trading panel UI
- User portfolio UI
- Creator dashboard UI
- Resolver dashboard UI
- Intelligence dashboard UI
- Wallet connection UI
- API calls to the Go backend

Not allowed in frontend:

- Core market logic
- Settlement logic
- Resolution logic
- Circle API secrets
- Database access
- Private keys
- Server-side payment orchestration

## Backend Rules

The Go backend owns all business logic.

Backend responsibilities:

- User and wallet mapping
- Market creation
- Market validation
- Trading intent validation
- Position tracking
- Resolution workflow
- Settlement state tracking
- Circle webhook handling
- Arc transaction coordination
- Agent API
- API key management
- Audit logs

## Domain Model

Primary domains:

- user
- wallet
- market
- trade
- position
- liquidity
- resolution
- settlement
- oracle
- circle
- arc
- webhook
- agent
- audit

Avoid old xbtpay/payment-gateway terminology unless referring to historical migration notes.

Do not use these domains for new SignalArc code:

- invoice
- merchant
- payment link
- payroll
- batch payment

## Official Documentation Requirement

For Arc and Circle behavior, use only official documentation as source of truth:

- Arc documentation: https://docs.arc.io/
- Arc prediction market blog: https://www.arc.io/blog/build-institutional-grade-prediction-markets-on-arc-arc-blueprints
- Circle developer documentation: https://developers.circle.com/
- Circle grants: https://www.circle.com/grant

If behavior is not documented, mark it as unknown / not documented.

Do not invent Arc, Circle, USDC, wallet, CCTP, Gateway, Paymaster, or settlement behavior.

## Technical Stack

Frontend:

- Next.js
- TypeScript
- Tailwind CSS
- shadcn/ui
- Wagmi
- Viem
- RainbowKit
- TanStack Query
- Zod
- React Hook Form

Backend:

- Go
- Chi router
- pgx
- sqlc
- PostgreSQL
- golang-migrate
- zerolog
- validator

Contracts:

- Solidity
- Foundry
- OpenZeppelin

Core infrastructure:

- Arc
- USDC
- Circle Developer Platform

## Package Manager

Use pnpm for frontend/workspace package management.

Do not add npm lockfiles.

Allowed lockfile:

- pnpm-lock.yaml

Not allowed:

- package-lock.json
- yarn.lock

## Generated Files

Do not commit generated or local dependency files.

Never commit:

- node_modules
- .next
- out
- dist
- .env
- private keys
- API keys
- local database files

## Development Discipline

Prefer small, reviewable commits.

Do not add multiple unrelated systems in one change.

Do not introduce mock integrations that look like real production integrations.

Use placeholders only when clearly named as placeholders.

When implementing external integrations, separate interface, config, and implementation.

