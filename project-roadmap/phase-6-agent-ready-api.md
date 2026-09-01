# Phase 6 Agent-Ready API

## Phase 6 Objective

SignalArc should expose a framework-neutral Agent API and execution layer that any external AI agent or tool framework can call through stable HTTP contracts.

External clients may include Hermes, OpenClaw, Claude Code, Custom GPT, LangChain, ElizaOS, Telegram or WhatsApp bots, institutional dashboards, or custom agent runtimes. SignalArc does not own the agent runtime, messaging channel, prompt loop, or chat UI.

Target flow:

```text
External AI agent/framework
  -> SignalArc Agent API
  -> intent preview / confirm / execute
  -> Circle Agent Wallet / executor
  -> Arc Testnet contracts
  -> portfolio / activity / result read models
```

## Non-Goals

- No smart contract changes.
- No Smart Contract V2.
- No channel-specific bot design.
- No frontend chat UI.
- No autonomous unattended trading by default.
- No mainnet funding.
- No deploy, push, or production runtime change in Phase 6A.

## Current Architecture Summary

The current backend has three separate layers:

| Layer | Current backing | Status |
| --- | --- | --- |
| Agent wallet registry | Postgres `agent_wallets` through `backend/internal/repository/agent_wallets.go` | DB-backed. |
| Agent onboarding/session boundary | Postgres `agent_onboarding_sessions` and `agent_sessions` through `backend/internal/repository/agent_sessions.go` | DB-backed metadata; Circle secrets are not stored. |
| Agent intent lifecycle | Postgres `agent_intents` through `backend/internal/repository/agent_intents.go`; `agent.Store` remains fallback/test plumbing | DB-backed in configured backend runtime. |
| Circle Agent Wallet execution | `CircleCLIExecutor` in `backend/internal/agent/circle_cli_executor.go` | Guarded by config and Circle CLI/session availability. |
| Market discovery | `GET /markets`, `GET /markets/{id}`, `GET /agent/markets` | DB-backed read model. |
| User positions/settlements | `/users/{user_id}/positions`, `/users/{user_id}/settlements` | DB-backed but user-id keyed, not agent-wallet keyed. |
| Trades | `backend/internal/repository/trades.go` | DB-backed trade intent records for existing backend trade routes, not Agent API intent records. |

Execution is intentionally fail-closed unless the backend runtime has Circle CLI support and `CIRCLE_AGENT_WALLET_EXECUTION_ENABLED=true`. The backend never stores Circle OTPs, Circle session files, Circle tokens, private keys, or deployer keys.

## Existing Endpoints

| Endpoint | Current purpose | Backing |
| --- | --- | --- |
| `GET /health` | Service health. | Runtime/database health. |
| `GET /markets` | Full market list. | DB-backed. |
| `GET /markets/{id}` | Full market detail. | DB-backed. |
| `GET /agent/markets` | Compact agent-readable market list. | DB-backed. |
| `POST /agent/onboarding/start` | Create pending onboarding session; optionally start Circle OTP when enabled. | DB-backed onboarding metadata; raw Circle request id is in-memory only. |
| `POST /agent/onboarding/verify` | Verify OTP when enabled; may resolve wallet and create session when resolver is configured. | DB-backed status/session metadata; no Circle secrets stored. |
| `GET /agent/onboarding/{onboarding_id}` | Read onboarding status. | DB-backed. |
| `GET /agent/sessions/{agent_id}` | Read activated agent-session boundary and optional Circle CLI liveness downgrade. | DB-backed plus runtime liveness check. |
| `POST /agent/onboarding/register` | Registry-only wallet registration shortcut. | DB-backed wallet upsert. |
| `POST /agent/wallets` | Register an already-created Circle Agent Wallet. | DB-backed wallet upsert. |
| `GET /agent/wallets/{agent_id}` | Read registered wallet metadata. | DB-backed. |
| `GET /agent/wallets/{agent_id}/balance` | Read Circle Agent Wallet balance. | DB-backed wallet lookup plus Circle CLI balance runner. |
| `POST /agent/wallets/{agent_id}/faucet` | Request ARC-TESTNET USDC faucet for the registered agent wallet only. | DB-backed wallet lookup plus Circle CLI faucet runner. |
| `POST /agent/wallets/{agent_id}/disable` | Disable a registered agent wallet. | DB-backed. |
| `POST /agent/intents` | Create intent preview. | In-memory only. |
| `GET /agent/intents/{intent_id}` | Read intent preview/confirmed state. | In-memory only. |
| `POST /agent/intents/{intent_id}/confirm` | Confirm intent and return execution plan. | In-memory only. |
| `POST /agent/intents/{intent_id}/execute` | Execute confirmed intent through registered wallet provider. | In-memory intent plus DB-backed wallet lookup; execution result is not durably stored. |
| `GET /users/{user_id}/positions` | User positions. | DB-backed, user-id keyed. |
| `GET /users/{user_id}/settlements` | User settlements. | DB-backed, user-id keyed. |

## Capability Matrix

| Capability | Current status | Notes |
| --- | --- | --- |
| Agent identity | Implemented for onboarding and wallet registration. | `agent_id` validation rejects generic placeholders for onboarding and wallet registration. Intent creation does not validate the same shape yet. |
| Agent wallet/session | Partly implemented. | Wallet/session records are DB-backed. Circle session material remains outside SignalArc and is not durably managed by the backend. |
| Allowed actions | Implemented. | Registered wallets carry `allowed_actions`; execution enforces the allowlist. |
| Market discovery | Implemented. | `GET /agent/markets`, `GET /markets`, and `GET /markets/{id}` are DB-backed. |
| Market detail | Implemented through generic market route. | No separate compact `/agent/markets/{id}` route yet. |
| Create intent | Implemented as preview. | In-memory only; response includes validation result and warnings. |
| Confirm intent | Implemented. | Produces an execution plan only; no transaction is broadcast. |
| Execute intent | Implemented for configured Circle CLI runtime. | Execution supports `create_market`, `buy_yes`, `buy_no`, `close_market`, `resolve_market`, `claim_payout`, `cancel_market`, and `claim_refund` through `circle_agent_wallet_cli` when enabled. |
| Execution result | Implemented and persisted. | `agent_executions` records pending, executed, and failed attempts with tx hashes, readback JSON, and sanitized errors. |
| Portfolio/positions read model | Implemented as a first durable read model. | `GET /agent/portfolio/{agent_id}` uses registered wallet metadata and executed agent buy intents/executions; live wallet-indexed contract balances remain unavailable. |
| Activity/history read model | Implemented for Agent API records. | `GET /agent/activity/{agent_id}` and `GET /agent/intents/{intent_id}/executions` expose durable intents/executions with tx hashes, readbacks, and sanitized failures. |
| Stable error codes | Partly implemented. | Many handler errors use stable codes; Phase 6B/6D should formalize and document the complete error catalog. |
| OpenAPI surface | Partly aligned. | Custom GPT schema covers most live agent routes. Phase 6A added `getAgentIntent` and aligned `ExecutionResult` fields with the handler response. |

## Current Execution Status

Preview-only:

- `POST /agent/intents`
- `POST /agent/intents/{intent_id}/confirm`

Executes through Circle CLI when enabled and authenticated:

- `create_market`
- `buy_yes`
- `buy_no`
- `close_market`
- `resolve_market`
- `claim_payout`
- `cancel_market`
- `claim_refund`

DB-backed:

- Agent wallet registry.
- Agent onboarding session metadata.
- Agent session metadata.
- Market list/detail.
- Existing user-id-keyed positions and settlements.
- Existing trade intent records outside the Agent API intent store.

In-memory only:

- Raw Circle OTP request id between onboarding start and verify.

Documented but incomplete:

- Unified agent portfolio and activity read models.
- Production Circle CLI/session strategy.
- API authentication, rate limiting, and policy enforcement.
- Framework-neutral developer docs beyond the current Custom GPT-oriented schema.

Unsafe or not production-ready:

- Circle CLI execution depends on runtime CLI/session availability.
- Existing unconfigured/test fallback intent storage remains in-process.
- Agent API routes do not yet enforce API keys or rate limits.
- Phase 6D now enforces the existing `agent_id` shape on intent, wallet, portfolio, activity, balance, faucet, session, confirm, and execute paths where an `agent_id` is present or read from a durable intent.

## Framework-Neutral Agent API Contract

Phase 6 should keep these concepts stable and channel-agnostic:

| Concept | Contract direction |
| --- | --- |
| Agent identity | `agent_id` is caller supplied, unique, non-generic, and stable across a session. |
| Agent wallet/session | `agent_id` maps to a registered agent wallet and optional active session boundary; SignalArc never returns Circle secrets. |
| Allowed actions | Wallet-scoped `allowed_actions` gate preview, confirmation when a wallet is available, and execution. |
| Market discovery | Agents can list compact markets and fetch full market detail without wallet setup. |
| Create intent | `POST /agent/intents` validates intent shape and returns preview plus warnings. |
| Confirm intent | `POST /agent/intents/{intent_id}/confirm` returns an execution plan and still performs no broadcast. |
| Execute intent | `POST /agent/intents/{intent_id}/execute` performs the configured provider execution only after explicit confirmation. |
| Portfolio/positions | Phase 6C should expose agent-wallet keyed positions/claim/refund state, not only internal user-id keyed routes. |
| Activity/history | Phase 6C should expose agent activity across intents, executions, tx hashes, readbacks, and sanitized failures. |
| Execution result | Phase 6B should persist execution attempts and results with stable status transitions. |
| Error codes | Phase 6D should publish a stable error-code catalog and keep public errors sanitized. |

The API contract must not contain WhatsApp, Telegram, ChatGPT, or Claude-specific required fields. `source_client` may describe the caller for auditability, but it must not change the domain model.

## OpenAPI / Documentation Alignment

Phase 6A audit found:

- `project-roadmap/signalarc-custom-gpt-openapi.json` already covered health, markets, onboarding, session, wallet, balance, faucet, create intent, confirm, and execute.
- The backend also exposes `GET /agent/intents/{intent_id}`; Phase 6A added this as `getAgentIntent`.
- The `ExecutionResult` schema had stale fields (`tx_hash`, `block_number`, `chain`, `executed_at`, `error`) and omitted current handler fields (`agent_id`, `action`, `network`, `agent_factory_address`, `market_contract_address`, `broadcast_performed`, `approve_transaction_hash`). Phase 6A aligned the schema with `newAgentExecutionResponse`.
- `docs/AGENT_API.md` remains useful but is oriented toward the published Custom GPT flow. Phase 6E should split a framework-neutral Agent API guide from Custom GPT-specific operator instructions.

## Security And Safety Requirements

- Keep Circle OTPs, request IDs, session files, tokens, API keys, private keys, and deployer keys out of database records, API responses, docs examples, and logs.
- Keep execution fail-closed when Circle CLI execution is disabled or the runtime session is unavailable.
- Reject deployer/resolver wallet reuse for agent wallets.
- Reject agent wallet reuse of `user_wallet` until a documented custody-link model exists.
- Enforce chain `ARC-TESTNET` for all current execution and faucet flows.
- Keep faucet endpoint pinned to the registered agent wallet and fixed testnet USDC behavior.
- Require explicit confirm before execute.
- Reject unconfirmed execute calls with `agent_intent_not_confirmed`.
- Enforce wallet/session `allowed_actions` with `agent_action_forbidden`.
- Enforce optional wallet `policy_metadata.max_trade_amount` for `buy_yes` / `buy_no` with `agent_policy_violation`.
- Persist execution attempts before production use so failures and successful tx hashes are auditable.
- Publish stable error codes and keep provider output sanitized.

## Proposed Subphases

### 6A Agent API Audit & Contract

Status: this document.

Deliverables:

- Audit current Agent API implementation, docs, OpenAPI, and repository backing.
- Define framework-neutral concepts and non-goals.
- Patch small schema drift without changing contracts, migrations, or deployment config.

### 6B Durable Intent & Execution Records

Status: IMPLEMENTED LOCALLY in this phase.

Goal:

- Move Agent API intents and execution attempts out of process memory.

Implemented:

- Added migration `000020_create_agent_intents_executions` for `agent_intents` and `agent_executions`.
- Added repository methods for create/get/confirm/execute/fail and list-by-agent support.
- Wired configured backend routing to use durable intent/execution storage while preserving `agent.Store` fallback behavior for tests/unconfigured paths.
- Added minimal idempotency using `agent_id + source_client + client_request_id`.
- Persisted execution pending, success tx/readback, and sanitized failure code/message.

### 6C Agent Market / Portfolio / Activity API

Status: IMPLEMENTED LOCALLY in this phase.

Goal:

- Give external agents stable read models for markets, positions, claims, refunds, and history.

Implemented:

- Added `GET /agent/portfolio/{agent_id}` with registered wallet metadata, intent/execution-derived positions, total exposure where amounts parse, empty settlements, and explicit unavailable field notes.
- Added `GET /agent/activity/{agent_id}` with framework-neutral activity items derived from durable `agent_intents` and `agent_executions`.
- Added `GET /agent/intents/{intent_id}/executions` for intent-scoped execution history.
- Extended compact `GET /agent/markets` responses with `market_contract_address` while preserving existing fields and the deployed-market filter.
- Reused existing Phase 6B tables; no new migration was required.

Known limitations:

- Portfolio positions are derived from executed buy intents/executions, not live onchain balance readbacks.
- Existing `positions`, `settlements`, and `trades` tables are internal-user keyed and are not yet safely joinable to `agent_id` or agent wallet address.
- Claimable/refundable eligibility remains unavailable until wallet-indexed claim/refund state is indexed.

### 6D Agent Safety & Policy Layer

Status: IMPLEMENTED LOCALLY in this phase.

Goal:

- Formalize safety policy before broader agent use.

Implemented:

- Applied stable `agent_id` validation to wallet/session, portfolio/activity, intent create/read/confirm/execute, balance, faucet, and disable paths where practical.
- Rejected generic placeholder agent IDs with `400 agent_id_invalid`.
- Enforced wallet `allowed_actions` for executable intent preview, confirm when a wallet exists, and execute with `403 agent_action_forbidden`.
- Preserved preview -> confirm -> execute; unconfirmed executions fail with `409 agent_intent_not_confirmed`.
- Kept existing amount validation for buy intents and added optional `policy_metadata.max_trade_amount` enforcement with `403 agent_policy_violation`.
- Kept Circle/provider errors sanitized in public responses and durable execution failure records.
- Added backend tests for invalid/generic agent IDs, disallowed actions, valid execution flow preservation, unconfirmed execution rejection, invalid amounts, max amount policy, and sanitized execution failures.

Deferred:

- API keys, rate limits, and paid access are still out of scope for this phase.
- No autonomous unattended trading mode is enabled by default.

### 6E Agent Developer Surface

Status: IMPLEMENTED LOCALLY in this phase.

Goal:

- Make SignalArc usable by any external agent framework without channel lock-in.

Implemented:

- Reworked `docs/AGENT_API.md` as the canonical framework-neutral HTTP integration guide.
- Added concrete cURL examples for market discovery, onboarding, wallet/session reads, buy intent creation, confirm, execute, intent lookup, intent executions, portfolio, and activity.
- Documented base URLs, local development URL, identity/wallet/session model, supported actions, durable intent/execution records, idempotency, safety policy, stable errors, limitations, and non-goals.
- Aligned OpenAPI schema naming around `AgentIntent`, `AgentExecution`, `AgentPortfolio`, `AgentActivityItem`, and `AgentError`.
- Added `project-roadmap/phase-6-local-smoke-test.md` with local validation commands, backend boot checks, Agent API curl sequence, and expected safe failures.

Deferred:

- No new SDK package was added.
- Custom GPT remains supported by the existing schema, but channel-specific bot behavior is not part of the core Agent API contract.
