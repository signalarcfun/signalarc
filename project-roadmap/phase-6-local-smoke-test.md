# Phase 6 Local Smoke Test

This checklist validates the Phase 6 Agent API developer surface without changing smart contracts, ABIs, deployment config, or migrations.

## Static Validation

Run from the repository root:

```bash
cd backend && go test ./...
pnpm --dir apps/web lint
pnpm --dir apps/web build
python3 -m json.tool project-roadmap/signalarc-custom-gpt-openapi.json
git diff --check
```

Run `gofmt` only when Go files changed:

```bash
gofmt -w backend/internal/api/agent_handlers.go backend/internal/api/agent_handlers_test.go
```

## Backend Boot

Prepare local environment:

```bash
cp backend/.env.example backend/.env
```

Set a local `DATABASE_URL` and keep `APP_PORT=4000` unless testing on a different port.

Apply migrations using the repo's existing migration workflow. Do not create Phase 6E migrations; this subphase should use the existing Phase 6B/6C/6D tables and policies.

Start the backend:

```bash
cd backend
go run ./cmd/api
```

Health checks:

```bash
BASE_URL="http://localhost:4000"

curl "$BASE_URL/health"
curl "$BASE_URL/readyz"
curl "$BASE_URL/schema/validate"
```

Expected healthy responses return `{"status":"ok"}` for `/health` and `/readyz`; `/schema/validate` should report schema status `ok`.

## Basic Agent API Curl Sequence

Use a valid placeholder-safe agent id:

```bash
BASE_URL="http://localhost:4000"
AGENT_ID="agent_demo_custom_001"
SOURCE_CLIENT="custom-agent"
CLIENT_REQUEST_ID="demo-001"
```

Read-only market discovery:

```bash
curl "$BASE_URL/agent/markets"
```

Start onboarding:

```bash
curl -X POST "$BASE_URL/agent/onboarding/start" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent_demo_custom_001",
    "user_email": "user@example.com",
    "source_client": "custom-agent",
    "channel": "custom-agent"
  }'
```

If an agent wallet/session already exists:

```bash
curl "$BASE_URL/agent/sessions/$AGENT_ID"
curl "$BASE_URL/agent/wallets/$AGENT_ID"
curl "$BASE_URL/agent/wallets/$AGENT_ID/balance"
```

Create a `buy_yes` preview:

```bash
curl -X POST "$BASE_URL/agent/intents" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent_demo_custom_001",
    "source_client": "custom-agent",
    "client_request_id": "demo-001",
    "action": "buy_yes",
    "user_wallet": "0x1111111111111111111111111111111111111111",
    "market_id": "example-market-id",
    "market_contract_address": "0x3333333333333333333333333333333333333333",
    "amount": "1"
  }'
```

Save the returned `intent.intent_id`:

```bash
INTENT_ID="agent_intent_example"
```

Confirm, execute, and read back:

```bash
curl -X POST "$BASE_URL/agent/intents/$INTENT_ID/confirm"
curl -X POST "$BASE_URL/agent/intents/$INTENT_ID/execute"
curl "$BASE_URL/agent/intents/$INTENT_ID"
curl "$BASE_URL/agent/intents/$INTENT_ID/executions"
curl "$BASE_URL/agent/portfolio/$AGENT_ID"
curl "$BASE_URL/agent/activity/$AGENT_ID"
```

Execution may return a safe configuration/provider failure in local environments without Circle CLI authentication. That is acceptable for smoke testing as long as the error code is stable and sanitized.

## Expected Safe Failure Cases

Invalid or generic `agent_id`:

```bash
curl "$BASE_URL/agent/portfolio/agent_desi_001"
```

Expected: `400 agent_id_invalid`.

Unconfirmed execution:

```bash
# Use a fresh preview intent that has not been confirmed yet.
curl -X POST "$BASE_URL/agent/intents/$INTENT_ID/execute"
```

Expected before confirm: `409 agent_intent_not_confirmed`.

Forbidden action:

1. Register or seed a wallet whose `allowed_actions` does not include the requested action.
2. Create or execute an intent for that forbidden action.

Expected: `403 agent_action_forbidden`.

Duplicate `client_request_id`:

```bash
curl -X POST "$BASE_URL/agent/intents" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent_demo_custom_001",
    "source_client": "custom-agent",
    "client_request_id": "demo-001",
    "action": "buy_yes",
    "user_wallet": "0x1111111111111111111111111111111111111111",
    "market_id": "example-market-id",
    "market_contract_address": "0x3333333333333333333333333333333333333333",
    "amount": "1"
  }'
```

Expected with durable intent storage: the same idempotent intent is returned for the same `agent_id + source_client + client_request_id` instead of creating a duplicate.

Policy max amount violation:

1. Register or seed a wallet with `policy_metadata.max_trade_amount`.
2. Create a `buy_yes` or `buy_no` intent with `amount` above that cap.

Expected: `403 agent_policy_violation`.

## Pass Criteria

- Backend tests pass.
- Frontend lint and build pass.
- OpenAPI JSON parses.
- Health/readiness/schema endpoints return expected local status.
- Agent API curl sequence reaches stable success or stable sanitized failure responses.
- No smart contracts, ABIs, deployment config, or migrations changed for Phase 6E.
