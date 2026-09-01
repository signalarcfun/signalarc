# Security and Secrets

SignalArc must keep secrets out of source control and out of public documentation. This repository currently documents local/testnet prototype behavior only and does not claim production custody, production settlement, audit completion, or compliance approval.

## Never Commit

- `.env`
- `.env.local`
- Private keys
- Seed phrases
- Circle API keys
- Blockscout API keys
- WalletConnect project IDs when project policy keeps them local or deployment-only
- Production database credentials
- RPC secrets
- Local database files
- `node_modules`
- `.next`
- `out`
- `dist`

## NEXT_PUBLIC Variables

`NEXT_PUBLIC_*` variables are exposed to browser JavaScript.

Safe examples:

- Public API base URL.
- WalletConnect project ID when intended for browser use and configured through deployment settings.

Never put secrets in `NEXT_PUBLIC_*`, including private keys, Circle API keys, database URLs, RPC secrets, or server-only credentials.

## Backend Secrets

Local development database values may appear in Docker Compose for the local database. Production secrets must be configured outside source control through the deployment platform or a secret manager.

Backend production secrets must not be written to Markdown files, committed examples, logs, frontend code, or generated artifacts.

## Wallet Model

- External wallets sign user transactions in the browser.
- Backend must not sign user transactions.
- No private key belongs in frontend or backend source code.
- Circle embedded wallet is not implemented.
- Circle SDK/API integration is not implemented in the current repository.

## Smart Contract Status

- Current contract is an Arc Testnet prototype.
- It is unaudited.
- It is not production custody.
- It is not production settlement.
- It does not prove mainnet readiness.

## Circle

- Circle API keys must never be placed in frontend code.
- Circle API keys must never be placed in public docs.
- Circle SDK/API integration is not implemented unless explicitly added in a future change.
- Circle behavior that is not implemented or not documented by official Circle sources is unknown / not documented.

## Deployment

- DNS/live deployment is not approved or completed yet.
- Production environment variables must be configured outside source control.
- CORS must be reviewed before production deployment.
- Production database credentials must use deployment secret management.
