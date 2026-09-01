# Local Development

This guide is for running SignalArc locally. Localhost URLs are intentionally used here for development only.

## Prerequisites

- Docker and Docker Compose.
- pnpm, matching the workspace package manager.
- Node.js compatible with Next.js 16; Node.js 20 or newer is recommended for this repository.
- Go, if running the backend outside Docker.
- Browser wallet extension for wallet testing.
- WalletConnect Project ID for WalletConnect support.

## Start Local Backend Stack

```bash
docker compose up -d signalarc-postgres signalarc-backend
```

The local backend listens on `localhost:4000`. PostgreSQL is exposed on `localhost:15433`.

## Check Backend

```bash
docker compose ps
curl http://localhost:4000/health
curl http://localhost:4000/readyz
curl http://localhost:4000/markets
curl http://localhost:4000/arc/contract
```

Expected health response:

```json
{"status":"ok"}
```

## Start Frontend

```bash
pnpm dev:web
```

Open http://localhost:3000.

## Environment Variables

Create local frontend env values in `apps/web/.env.local`.

```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:4000
NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID=replace_with_real_project_id
```

Never commit `.env.local`.

Backend local examples are documented in `backend/.env.example`. Do not print or commit real secret values.

## Database

| Access path | Host | Port |
| --- | --- | --- |
| Host machine | `localhost` | `15433` |
| Docker network | `signalarc-postgres` | `5432` |

The Docker Compose development database uses local-only credentials suitable for development, not production.

## Validation

```bash
pnpm --dir apps/web lint
pnpm --dir apps/web exec tsc --noEmit
pnpm --dir apps/web build
```

Backend validation outside Docker can use:

```bash
cd backend
go test ./...
```

## Troubleshooting

### Port 3000 Already In Use

Stop the stale Next.js process or run the frontend on another development port.

### Backend Port 4000 Unavailable

Check container state and logs:

```bash
docker compose ps
docker compose logs signalarc-backend
```

### Database Not Ready

Check PostgreSQL logs:

```bash
docker compose logs signalarc-postgres
```

Then retry:

```bash
curl http://localhost:4000/readyz
```

### WalletConnect Project ID Missing

WalletConnect support requires `NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID` in `apps/web/.env.local` for local testing or deployment environment variables later. Do not commit a real project ID.

### Wrong Network

The frontend wallet flow is configured for Arc Testnet, chain ID `5042002`. Switch the connected wallet to Arc Testnet before testing the browser transaction flow.

### Local API Base URL

For local development, use http://localhost:4000.
