# SignalArc Agent API

SignalArc Agent API is a framework-neutral HTTP API for Arc Testnet YES/NO prediction market workflows. External agents can integrate through ordinary HTTP requests from any runtime: custom services, scripts, agent frameworks, chatbots, schedulers, dashboards, or internal tools.

SignalArc handles the API surface, wallet and session records, durable intents, execution coordination, and read models. External agents own their prompt loop, user interaction, scheduling, messaging channel, and higher-level autonomous policy logic.

## Base URLs

| Environment | Base URL |
| --- | --- |
| Production | `https://api.signalarc.fun` |
| Local development | `http://localhost:4000` |

Health and schema checks:

```bash
BASE_URL="https://api.signalarc.fun"

curl "$BASE_URL/health"
curl "$BASE_URL/readyz"
curl "$BASE_URL/schema/validate"
```

For local development, set:

```bash
BASE_URL="http://localhost:4000"
```

## Security And Custody Boundaries

- External agents never send private keys to SignalArc.
- Public API responses must not expose Circle API keys, entity secrets, `entitySecretCiphertext`, credential paths, database URLs, deploy tokens, recovery files, or other backend secrets.
- Keep your own API keys, wallet credentials, prompt secrets, and service tokens out of prompts, logs, repositories, and client-side code.
- SignalArc Agent API does not support arbitrary transfers, withdrawals, swaps, deposits, or mainnet funding.
- Current execution scope is `ARC-TESTNET`.
- Existing smart contracts and ABIs remain unchanged.
- Agent execution requires an active registered wallet and a configured backend wallet provider.
- Production execution direction is backend-managed Circle Developer-Controlled Wallet API or direct Circle API integration. Circle CLI is only an old/manual/dev fallback and must not be treated as a production runtime dependency.

## Agent Identity

`agent_id` is the stable caller-supplied identity for an external agent/user pair. Reuse the same value across onboarding, wallet registration, session reads, intent creation, confirmation, execution, portfolio, and activity.

Validation rules:

- Required where an agent identity is needed.
- Must start with `agent_`.
- Must be at least 10 characters long.
- May contain ASCII letters, digits, underscores, and hyphens.
- Must not be a generic placeholder.

Safe example:

```text
agent_demo_custom_001
```

## Wallet And Session Model

Agent wallet records are SignalArc backend records keyed by `agent_id`. The current executable wallet provider is `circle_agent_wallet`.

| Field | Meaning |
| --- | --- |
| `wallet_provider` | Wallet provider for the agent record. Current executable provider: `circle_agent_wallet`. |
| `chain` | Execution and faucet flows currently require `ARC-TESTNET`. |
| `allowed_actions` | Per-wallet allowlist for executable actions. |
| `policy_metadata.max_trade_amount` | Optional max amount policy for `buy_yes` and `buy_no`. |
| `status` | Must be `active` for execution and faucet flows. |

`ARC-TESTNET` is currently required for execution and faucet flows.

## Supported Actions

| Action | Required fields |
| --- | --- |
| `create_market` | `action`, `agent_id`, `user_wallet`, `market_id`, `question`, `close_timestamp`, `resolver`, `collateral_token` |
| `buy_yes` | `action`, `agent_id`, `user_wallet`, `market_id`, `market_contract_address`, `amount` |
| `buy_no` | `action`, `agent_id`, `user_wallet`, `market_id`, `market_contract_address`, `amount` |
| `close_market` | `action`, `agent_id`, `user_wallet`, `market_id`, `market_contract_address` |
| `resolve_market` | `action`, `agent_id`, `user_wallet`, `market_id`, `market_contract_address`, `outcome` |
| `claim_payout` | `action`, `agent_id`, `user_wallet`, `market_id`, `market_contract_address` |
| `cancel_market` | `action`, `agent_id`, `user_wallet`, `market_id`, `market_contract_address` |
| `claim_refund` | `action`, `agent_id`, `user_wallet`, `market_id`, `market_contract_address` |

`amount` is a positive decimal string. `close_timestamp` accepts UTC RFC3339 or unix-seconds strings.

## Outcome Handling

`resolve_market` accepts `yes`, `no`, `1`, or `2`.

Executor normalization:

| Input | Contract value |
| --- | --- |
| `yes` | `uint8 1` |
| `1` | `uint8 1` |
| `no` | `uint8 2` |
| `2` | `uint8 2` |

## Intent Lifecycle

Transaction actions follow a three-step lifecycle:

1. `POST /agent/intents` creates a preview only and does not broadcast.
2. `POST /agent/intents/{intent_id}/confirm` confirms the preview and returns an execution plan only.
3. `POST /agent/intents/{intent_id}/execute` executes only a confirmed intent.

Calling execute before confirm returns `409 agent_intent_not_confirmed`.

Read endpoints:

- `GET /agent/intents/{intent_id}` reads the current durable intent state.
- `GET /agent/intents/{intent_id}/executions` lists execution attempts.

## Idempotency

Intent creation supports idempotency with `agent_id`, `source_client`, and `client_request_id`. When durable storage is configured, repeating the same `source_client + client_request_id` for the same agent can return the existing intent instead of creating a duplicate.

```json
{
  "agent_id": "agent_demo_custom_001",
  "source_client": "custom-agent",
  "client_request_id": "demo-001"
}
```

## Endpoint Summary

| Method | Endpoint |
| --- | --- |
| GET | `/health` |
| GET | `/readyz` |
| GET | `/schema/validate` |
| GET | `/agent/markets` |
| GET | `/markets` |
| GET | `/markets/{market_id}` |
| POST | `/agent/onboarding/start` |
| POST | `/agent/onboarding/verify` |
| GET | `/agent/onboarding/{onboarding_id}` |
| GET | `/agent/sessions/{agent_id}` |
| POST | `/agent/wallets` |
| GET | `/agent/wallets/{agent_id}` |
| GET | `/agent/wallets/{agent_id}/balance` |
| POST | `/agent/wallets/{agent_id}/faucet` |
| POST | `/agent/wallets/{agent_id}/disable` |
| POST | `/agent/intents` |
| GET | `/agent/intents/{intent_id}` |
| POST | `/agent/intents/{intent_id}/confirm` |
| POST | `/agent/intents/{intent_id}/execute` |
| GET | `/agent/intents/{intent_id}/executions` |
| GET | `/agent/portfolio/{agent_id}` |
| GET | `/agent/activity/{agent_id}` |

## Curl Examples

Set shared variables:

```bash
BASE_URL="https://api.signalarc.fun"
AGENT_ID="agent_demo_custom_001"
USER_WALLET="0x1111111111111111111111111111111111111111"
RESOLVER="0x1111111111111111111111111111111111111111"
COLLATERAL_TOKEN="0x3600000000000000000000000000000000000000"
MARKET_ID="example-market-001"
MARKET_CONTRACT_ADDRESS="0x3333333333333333333333333333333333333333"
```

For local development:

```bash
BASE_URL="http://localhost:4000"
```

Health check:

```bash
curl "$BASE_URL/health"
```

Create market intent:

```bash
curl -X POST "$BASE_URL/agent/intents" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent_demo_custom_001",
    "source_client": "custom-agent",
    "client_request_id": "create-market-001",
    "action": "create_market",
    "user_wallet": "0x1111111111111111111111111111111111111111",
    "market_id": "example-market-001",
    "question": "Will this example market resolve yes?",
    "close_timestamp": "2030-01-01T00:00:00Z",
    "resolver": "0x1111111111111111111111111111111111111111",
    "collateral_token": "0x3600000000000000000000000000000000000000"
  }'
```

Confirm intent:

```bash
INTENT_ID="agent_intent_example"
curl -X POST "$BASE_URL/agent/intents/$INTENT_ID/confirm"
```

Execute intent:

```bash
curl -X POST "$BASE_URL/agent/intents/$INTENT_ID/execute"
```

Create `buy_yes` intent:

```bash
curl -X POST "$BASE_URL/agent/intents" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent_demo_custom_001",
    "source_client": "custom-agent",
    "client_request_id": "buy-yes-001",
    "action": "buy_yes",
    "user_wallet": "0x1111111111111111111111111111111111111111",
    "market_id": "example-market-001",
    "market_contract_address": "0x3333333333333333333333333333333333333333",
    "amount": "1"
  }'
```

Create `buy_no` intent:

```bash
curl -X POST "$BASE_URL/agent/intents" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent_demo_custom_001",
    "source_client": "custom-agent",
    "client_request_id": "buy-no-001",
    "action": "buy_no",
    "user_wallet": "0x1111111111111111111111111111111111111111",
    "market_id": "example-market-001",
    "market_contract_address": "0x3333333333333333333333333333333333333333",
    "amount": "1"
  }'
```

Create `close_market` intent:

```bash
curl -X POST "$BASE_URL/agent/intents" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent_demo_custom_001",
    "source_client": "custom-agent",
    "client_request_id": "close-market-001",
    "action": "close_market",
    "user_wallet": "0x1111111111111111111111111111111111111111",
    "market_id": "example-market-001",
    "market_contract_address": "0x3333333333333333333333333333333333333333"
  }'
```

Create `resolve_market` yes intent:

```bash
curl -X POST "$BASE_URL/agent/intents" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent_demo_custom_001",
    "source_client": "custom-agent",
    "client_request_id": "resolve-market-yes-001",
    "action": "resolve_market",
    "user_wallet": "0x1111111111111111111111111111111111111111",
    "market_id": "example-market-001",
    "market_contract_address": "0x3333333333333333333333333333333333333333",
    "outcome": "yes"
  }'
```

Create `claim_payout` intent:

```bash
curl -X POST "$BASE_URL/agent/intents" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent_demo_custom_001",
    "source_client": "custom-agent",
    "client_request_id": "claim-payout-001",
    "action": "claim_payout",
    "user_wallet": "0x1111111111111111111111111111111111111111",
    "market_id": "example-market-001",
    "market_contract_address": "0x3333333333333333333333333333333333333333"
  }'
```

Create `cancel_market` intent:

```bash
curl -X POST "$BASE_URL/agent/intents" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent_demo_custom_001",
    "source_client": "custom-agent",
    "client_request_id": "cancel-market-001",
    "action": "cancel_market",
    "user_wallet": "0x1111111111111111111111111111111111111111",
    "market_id": "example-market-001",
    "market_contract_address": "0x3333333333333333333333333333333333333333"
  }'
```

Create `claim_refund` intent:

```bash
curl -X POST "$BASE_URL/agent/intents" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent_demo_custom_001",
    "source_client": "custom-agent",
    "client_request_id": "claim-refund-001",
    "action": "claim_refund",
    "user_wallet": "0x1111111111111111111111111111111111111111",
    "market_id": "example-market-001",
    "market_contract_address": "0x3333333333333333333333333333333333333333"
  }'
```

Read intent and executions:

```bash
curl "$BASE_URL/agent/intents/$INTENT_ID"
curl "$BASE_URL/agent/intents/$INTENT_ID/executions"
```

Read portfolio and activity:

```bash
curl "$BASE_URL/agent/portfolio/$AGENT_ID"
curl "$BASE_URL/agent/activity/$AGENT_ID"
```

## Response Examples

Intent preview response:

```json
{
  "intent": {
    "intent_id": "agent_intent_example",
    "action": "buy_yes",
    "status": "preview",
    "requires_confirmation": true,
    "validation_result": {
      "valid": true,
      "errors": []
    }
  }
}
```

Confirm response:

```json
{
  "execution_plan": {
    "intent_id": "agent_intent_example",
    "action": "buy_yes",
    "status": "confirmed",
    "execution_mode": "circle_developer_wallet_api",
    "network": "ARC-TESTNET",
    "broadcast_performed": false,
    "transaction_hash": null,
    "transaction_request": {
      "function": "buyYes",
      "chain": "ARC-TESTNET",
      "broadcast_performed": false
    }
  }
}
```

Execute response:

```json
{
  "execution": {
    "intent_id": "agent_intent_example",
    "action": "buy_yes",
    "status": "executed",
    "execution_mode": "circle_developer_wallet_api",
    "network": "ARC-TESTNET",
    "broadcast_performed": true,
    "approve_transaction_hash": "0xapprove...",
    "transaction_hash": "0xtransaction...",
    "readback": {
      "yes_positions": "1000000",
      "total_collateral": "1000000"
    }
  }
}
```

## Read Models

`GET /agent/portfolio/{agent_id}` returns a compact portfolio view for a registered agent wallet. Current portfolio data is derived from durable Agent API records and current read models. Do not assume full live onchain wallet-indexed indexing unless a later release explicitly documents it.

`GET /agent/activity/{agent_id}` returns recent intent and execution activity derived from durable records.

`GET /agent/intents/{intent_id}/executions` returns execution attempts for one intent, including status, transaction hashes, readback JSON, and sanitized error fields when available.

## Safety Model

- `agent_id` validation rejects malformed or generic IDs.
- `allowed_actions` enforcement gates preview, confirmation, and execution.
- Execution requires a confirmed intent.
- `policy_metadata.max_trade_amount` can reject buy amounts above wallet policy.
- Provider errors are sanitized before they appear in logs, API responses, or durable execution records.
- Execution is pinned to `ARC-TESTNET`.
- Private keys and Circle secrets are not returned in API responses.

## Error Catalog

| Status | Code | Meaning |
| --- | --- | --- |
| 400 | `invalid_json` | Request body is not valid JSON. |
| 400 | `agent_id_invalid` | `agent_id` is missing, malformed, too short, or generic. |
| 400 | `agent_intent_invalid` | Intent validation failed. |
| 400 | `agent_wallet_missing` | Execution needs a registered agent wallet. |
| 400 | `agent_onboarding_invalid` | Onboarding request validation failed. |
| 400 | `agent_wallet_invalid` | Wallet registration validation failed. |
| 403 | `agent_action_forbidden` | Action is outside `allowed_actions`. |
| 403 | `agent_policy_violation` | Wallet policy blocks the request. |
| 404 | `agent_intent_not_found` | No intent exists for the supplied id. |
| 404 | `agent_wallet_not_found` | No wallet exists for the supplied agent. |
| 404 | `agent_session_not_found` | No active session exists for the supplied agent. |
| 409 | `agent_intent_not_confirmed` | Execute was called before confirm. |
| 409 | `agent_wallet_status_invalid` | Wallet is not active. |
| 409 | `agent_wallet_chain_invalid` | Wallet is not on `ARC-TESTNET`. |
| 500 | `agent_activity_get_failed` | Activity read failed. |
| 500 | `agent_portfolio_get_failed` | Portfolio read failed. |
| 501 | `circle_agent_wallet_execution_not_enabled` | Wallet execution is disabled in this runtime. |
| 501 | `circle_agent_wallet_balance_not_configured` | Balance lookup is not configured. |
| 501 | `circle_agent_wallet_faucet_not_configured` | Faucet helper is not configured. |
| 502 | `agent_execution_failed` | Wallet provider execution failed; response is sanitized. |
| 502 | `circle_agent_wallet_balance_failed` | Balance lookup failed; response is sanitized. |
| 502 | `circle_agent_wallet_faucet_failed` | Faucet request failed; response is sanitized. |
| 503 | `agent_execution_config_invalid` | Execution environment is not configured. |

Common validation examples:

- `400 agent_intent_invalid`: `create_market` omitted `resolver` or `collateral_token`.
- `403 agent_policy_violation`: buy amount exceeds `policy_metadata.max_trade_amount`.
- `409 agent_intent_not_confirmed`: execute was called before confirm.
- `502 agent_execution_failed`: sanitized wallet provider failure.

## Local Docker Smoke Test Results

These are local Docker validation results. They validate backend integration behavior in the local environment and are not by themselves a production deployment certification.

Validated local context:

| Item | Value |
| --- | --- |
| Backend URL | `http://127.0.0.1:4000` |
| Backend container | `signalarc-backend` |
| Postgres container | `signalarc-postgres` |
| Wallet provider | `circle_agent_wallet` |
| Execution mode | `circle_developer_wallet_api` |
| Chain | `ARC-TESTNET` |
| USDC collateral token | `0x3600000000000000000000000000000000000000` |

Action matrix passed locally:

| Action | Local result |
| --- | --- |
| `create_market` | Passed |
| `buy_yes` | Passed |
| `buy_no` | Passed |
| `buy_yes` with 10 USDC | Passed |
| `close_market` | Passed |
| `resolve_market` with yes outcome | Passed |
| `claim_payout` | Passed |
| `cancel_market` | Passed |
| `claim_refund` | Passed |

Final lifecycle actions returned HTTP 200 in backend logs. Earlier `400`, `403`, and `502` responses during local testing were expected validation, policy, or timing examples and were not final blockers.

Selected local evidence:

- 10 USDC `buy_yes` on market `0x2E68a7F1B09e67574b7d25277e78325b9798Fd2e` produced `readback.yes_positions: 10000000` and `readback.total_collateral: 10000000`.
- `close_market` produced `readback.is_open: false`.
- `resolve_market` with yes produced `readback.claimable_payout: 10000000` and `readback.has_claimed: false`.
- `claim_payout` produced `readback.has_claimed: true`.
- `cancel_market` followed by `claim_refund` produced `readback.claimable_refund: 1` and then `readback.has_claimed: true`.

## Production Readiness Caveat

Before claiming production readiness, the team must:

- audit the diff;
- run backend tests;
- run local smoke checks;
- ensure production secrets are configured in a secret manager;
- ensure production database migrations are current;
- ensure local docker-compose overrides are not committed;
- ensure Circle CLI is not a production runtime dependency;
- deploy only after explicit approval.

## OpenAPI

The Custom GPT action schema lives at:

```text
project-roadmap/signalarc-custom-gpt-openapi.json
```

The schema is useful for any HTTP client, not only Custom GPTs. Custom GPT-specific behavior belongs in integration instructions, not in the core API contract.

## Known Limitations

- API key enforcement, paid access, rate limits, and production SLA are not implemented.
- Portfolio positions are derived from durable intent/execution records and current read models, not full live onchain wallet-indexed indexing.
- Claimable/refundable eligibility remains limited until wallet-indexed claim/refund state is indexed.
- Mainnet funding and Arc mainnet contract execution are not supported.
- Behavior beyond the documented endpoints is unknown or not documented.
