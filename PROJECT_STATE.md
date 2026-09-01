# SignalArc Project State

This file is the handoff source of truth for continuing SignalArc work in a new chat or coding session.

## Project

- Name: SignalArc
- Repository: https://github.com/wahyu241205/SignalArc
- Positioning: Arc-native prediction market infrastructure, not a Polymarket clone
- Core idea: Convert market opinions into structured probability signals, support USDC-settled event markets, expose market intelligence through APIs, and make the system usable by creators, institutions, developers, and AI agents.

## Working Style

- Use English for repository-facing guidance and project notes.
- Use controlled debugging.
- One step at a time.
- Do not jump phases without justification.
- Do not provide multiple unrelated solutions at once.
- For Arc and Circle behavior, use official documentation only.
- If behavior is not documented, state: unknown / not documented.
- Do not invent integrations or system behavior.
- Keep commits small and reviewable.

## Current Overall Status

Core MVP, CI, backend CD, Cloud Run deployment, and runtime health/readiness are operational.

Current production schema validation is not clean yet: `/schema/validate` currently returns 503 with migration version 18, `dirty=false`, no missing tables, and no missing columns.

Next work should focus on institutional product polish, backend hardening, contract mechanism maturity, and production-grade operational readiness.

## Phase 0 — Baseline Stabilization

Goal: make the current production baseline clean before new feature expansion.

Scope:

- Frontend: verify all obvious user-facing routes load, render usable states, and avoid broken navigation.
- Backend: verify health, readiness, schema validation, structured errors, and diagnosable runtime logs.
- Contracts: verify current compiled/tested contract state and document any prototype-only assumptions.
- Infra: verify deployed service configuration, Cloud Run health/readiness, environment variables, CI/CD state, and rollback path.

Blockers:

- `/schema/validate` production blocker resolved. Endpoint now returns 200 with `status=ok`, `migration_version=18`, `dirty=false`, `missing_tables=[]`, and `missing_columns=[]`.

Exit criteria:

- `/health`, `/readyz`, and `/schema/validate` all return 200.
- No obvious broken frontend route in the primary product flow.
- Logs are sufficient to diagnose backend, deployment, and schema failures without guessing.

## Phase 1 — Professional Frontend Shell

Goal: make SignalArc feel like an institutional product-grade prediction market infrastructure interface.

Scope:

- Homepage with clear Arc-native infrastructure positioning.
- Market list UI with credible scanning, filtering hooks, and useful loading states.
- Market detail shell with probability, rules, liquidity, trading, resolution, and activity areas.
- Account state for disconnected, connected, wrong-network, and unavailable-backend cases.
- Loading states, empty states, error states, and disabled states across primary surfaces.
- Responsive layout for desktop and mobile without visual clutter.
- Restrained visual hierarchy suitable for reviewers, partners, institutions, and operators.

Exclusions:

- No backend business logic changes unless a UI route cannot consume an existing or clearly needed read endpoint.
- No frontend settlement, resolution, private key, Circle secret, or database logic.

## Phase 2 — Market Data and Trading UX

Goal: make market browsing and trading intent flows clear and trustworthy.

Scope:

- Probability display that distinguishes current probability, volume/liquidity context, and market status.
- YES/NO trading panel with clear intent entry, disabled states, and validation feedback.
- Trade preview with side, amount, expected outcome exposure, fees or unknown fee status, and risk language.
- Confirmation state that separates intent submission from settlement or execution finality.
- Transaction status timeline for pending, accepted, submitted, confirmed, failed, cancelled, and claimable states where supported.
- Positions panel showing user exposure, market status, and settlement or claim availability.
- Settlement and claim status UI for resolved, cancelled, claimable, claimed, and unavailable states.

Backend allowance:

- Add richer read endpoints if needed to support trustworthy market, position, transaction, or settlement display.

Contract allowance:

- Change contracts only if an existing frontend/backend/contract mismatch requires it and the mismatch is confirmed from repo evidence.

## Phase 3 — Institutional Backend Hardening

Goal: improve backend reliability, auditability, and consistency before expanding the product surface.

Scope:

- Standardize API error envelopes across backend routes.
- Add or verify request IDs and correlation IDs in logs and responses.
- Expand audit logs for critical writes and lifecycle actions.
- Add idempotency for critical writes such as market creation, trading intent submission, settlement actions, and claim paths.
- Strengthen request validation and domain-level validation boundaries.
- Align OpenAPI documentation with implemented backend behavior.
- Harden migration validation, including dirty migration handling and schema drift diagnostics.
- Review database indexes for market browsing, positions, audit logs, lifecycle status, and operational queries.
- Add monitoring and alerting for health, readiness, schema validation, critical route failures, and external execution failures.

Exit criteria:

- Critical API routes return consistent errors, are traceable by request ID, and have auditable state transitions.
- Migration validation distinguishes unhealthy runtime state from missing schema objects and dirty migrations.

## Phase 4 — Contract and Market Mechanism Upgrade

Goal: move from prototype mechanics toward a production-grade market lifecycle.

Scope:

- Review payout model, accounting model, liquidity assumptions, fee handling, and claim mechanics.
- Define market lifecycle states and allowed transitions clearly.
- Emit and document events for market creation, trading, close, resolve, cancel, refund, payout, and claim actions.
- Review access control for creators, resolvers, admins, agents, and emergency paths.
- Define cancel/refund behavior and ensure it is consistent across contracts, backend, and UI.
- Document settlement invariants and failure modes.
- Add focused unit tests for lifecycle transitions and payout/claim behavior.
- Add invariant tests for accounting, settlement, refund, and double-claim prevention.
- Add edge-case tests for zero liquidity, late trades, cancelled markets, invalid outcomes, repeated actions, and unauthorized callers.

Production-readiness note:

- The current production-grade economic model must be validated before mainnet or institutional launch.

## Phase 5 — Agent, Circle, and Arc Production Path

Goal: make the agent wallet and Arc execution path reliable enough for controlled production workflows.

Scope:

- Agent wallet onboarding flow and operator runbook.
- OTP, session, liveness, and recovery model for agent execution.
- Execution status lifecycle from requested to accepted, submitted, confirmed, failed, cancelled, or unknown.
- Circle CLI/API error mapping into stable backend error categories.
- Arc transaction coordination and confirmation tracking based only on official Arc documentation.
- Circle behavior documented only from official Circle developer documentation.
- Wallet dashboard for agent wallet identity, chain, balance, status, recent actions, and failure history.
- Execution history with request IDs, action IDs, transaction hashes where available, readbacks, and sanitized errors.

Documentation rule:

- Undocumented Arc or Circle behavior must be marked unknown / not documented.

## Phase 6 — Trust, Compliance, and Institutional Presentation

Goal: make the platform understandable and credible to reviewers, partners, and institutions.

Scope:

- Public documentation explaining SignalArc as Arc-native prediction market infrastructure.
- Risk disclosure for markets, trading, resolution, settlement, smart contracts, and external execution providers.
- Market rules and resolution criteria templates.
- Architecture overview covering frontend, backend, database, contracts, Arc, Circle, agents, and operational boundaries.
- Threat model for custody, execution, settlement, admin actions, webhooks, API keys, and abuse cases.
- Contract lifecycle documentation from market creation through resolution, cancel, refund, payout, and claim.
- API reference aligned with implemented backend behavior.
- Deployment runbook for CI, backend CD, Cloud Run, environment variables, migrations, and rollback.
- Incident runbook for degraded backend, schema validation failure, external execution failure, stuck transaction, and contract pause/cancel scenarios.
- Status page or status endpoint plan for public and internal operational visibility.

## Phase 7 — Scale and Controlled Beta Readiness

Goal: prepare SignalArc for a controlled beta with bounded operational and product risk.

Scope:

- Beta user flow from onboarding to market discovery, trading intent, position tracking, resolution, and claim.
- Market creation governance, including who can create markets, who approves them, and how market rules are reviewed.
- Admin review workflow for market approval, resolution, cancellation, and incident response.
- Featured markets and curated market surfaces for early beta quality control.
- Search, filter, and sort for market discovery.
- Monitoring for backend health, schema health, route failures, latency, external execution failures, and contract lifecycle anomalies.
- Backup and restore plan for PostgreSQL and operational data.
- Rollback plan for backend deployments, frontend deployments, migrations, and configuration changes.
- Staging environment aligned with production configuration and test data practices.
- Dependency audit for frontend, backend, contracts, infrastructure, and CI/CD.
- Secrets audit for environment variables, API keys, service accounts, deploy credentials, and local developer machines.
- Contract audit preparation, including threat model, tests, invariants, known limitations, and deployment notes.
- Abuse and rate-limit strategy for public APIs, trading intents, market creation, agent execution, and admin routes.

## Frontend Wallet and Production-Facing UI Polish

Status: COMPLETE.

Done:

- Added external wallet connection UI using RainbowKit, Wagmi, and Viem.
- Added Arc Testnet frontend chain configuration.
- Added app-level Web3 providers for wallet connection.
- Added responsive dark Web3 layout and navigation.
- Added visible wallet connection control in the header.
- Added connected wallet address display through RainbowKit.
- Added wrong-network warning for non-Arc Testnet wallet connections.
- Reworked the landing page into a production-facing SignalArc product homepage.
- Reduced internal developer wording from user-facing pages.
- Moved contract metadata away from the primary landing hero into lower-priority developer/network context.
- Reworked markets, market detail, create market, portfolio, and intelligence pages for a cleaner user-facing experience.
- Kept trade flow intent-only.
- No Circle SDK added.
- No Circle API key added.
- No onchain write execution added.
- No USDC approve, open position, claim, refund, or resolver transaction UI added.
- No backend, Docker, contract, DNS, deployment, or mainnet changes added in this step.

Validation:

- `pnpm --dir apps/web lint` passed.
- `pnpm --dir apps/web exec tsc --noEmit` passed.
- `pnpm --dir apps/web build` passed.
- Browser wallet connection UI rendered.
- Wallet address displayed after connection.
- Existing frontend pages rendered with the polished dark Web3 UI.

## Local Docker Backend Step

Status: COMPLETE.

Done:

- Added `backend/Dockerfile` for local backend container builds.
- Added `backend/.dockerignore` to exclude local env and unnecessary files.
- Added `signalarc-backend` service to `docker-compose.yml`.
- Local Docker Compose now runs PostgreSQL and backend API.
- Backend container connects to PostgreSQL through Docker network using `signalarc-postgres:5432`.
- Backend is exposed only on host loopback at `127.0.0.1:4000`.
- Frontend remains a manual Next.js dev server.
- Smart contract is not Dockerized.
- No production deployment, DNS change, contract redeploy, or mainnet config added.

Validation:

- `docker compose build signalarc-backend` passed.
- `docker compose up -d signalarc-postgres signalarc-backend` passed.
- `GET /health` passed.
- `GET /readyz` passed.
- `GET /arc/contract` passed.
- `GET /markets` passed.

## Local MVP Integration Fix

Status: COMPLETE.

Done:

- Local frontend/backend MVP usability smoke test completed.
- Backend local CORS added for the local Next.js frontend.
- Backend read-only `GET /arc/contract` endpoint added.
- Frontend root page now shows SignalArc local MVP instead of the default Next.js template.
- Frontend can load markets from backend without "Failed to fetch".
- Local browser flow passed:
  - `/` — landing page renders
  - `/markets` — market list loads from backend
  - `/markets/new` — create market form renders
  - Created one test market successfully
  - Opened market detail page
  - Submitted one trade intent with status `not_executed`
  - `/portfolio` — portfolio page renders
  - `/intelligence` — intelligence dashboard renders
- Arc Testnet contract reference visible as prototype/testnet metadata only.
- No wallet/onchain write execution added.
- No Circle integration added.
- No production deployment approved.
- No contract redeploy performed.
- Temporary demo domain purchased: `signalarc.fun`.
- DNS/live deployment remains pending and not approved.

## CI / Quality Gate

Status: COMPLETE.

Done:

- Added first CI-only GitHub Actions workflow at `.github/workflows/ci.yml`.
- Workflow runs on `pull_request` and push to `main`.
- Added least-privilege workflow permission: `contents: read`.
- Frontend CI validates pnpm frozen install, lint, TypeScript check, and production build.
- Backend CI runs `go test ./...` from `backend`.
- Contracts CI installs Foundry and runs `forge test` from `contracts`.
- No deployment, production secrets, Docker publish, Vercel, or GCP automation added in this step.

Validation:

- `pnpm install --frozen-lockfile` passed.
- `pnpm --dir apps/web lint` passed.
- `pnpm --dir apps/web exec tsc --noEmit` passed.
- `pnpm --dir apps/web build` passed under Node 22.
- `cd backend && go test ./...` passed.
- `cd contracts && forge test` passed.

- GitHub branch ruleset `Protect main with CI` created and set to Active.
- Ruleset targets default branch `main`.
- Required status checks configured: `Frontend`, `Backend`, `Contracts`.
- Force pushes and branch deletions are blocked for `main`.

## Current Last Completed Step

- Multi-tenant agent onboarding session foundation added for isolated pending onboarding sessions and durable per-agent session boundaries. Circle OTP/provisioning remains not implemented.

## Live AI Agent Transaction MVP

Status: IN PROGRESS - backend Circle Agent Wallet provider trading, payout, cancel, and refund lifecycle actions are validated on Arc Testnet from the host-shell backend runtime; ChatGPT Custom GPT external client trigger is validated for health, intent preview, confirm, and real `create_market` execution through a temporary ngrok tunnel; Docker/Cloud Run runtime strategy, production readiness, and mainnet readiness remain pending.

Current checkpoint state:

- Agent Wallet Boundary documented in `project-roadmap/agent-mcp.md`.
- Previous deployer-signed Arc Testnet transactions are explicitly not accepted as agent-wallet validation.
- `contracts/.env` `PRIVATE_KEY` must not be used as an agent wallet.
- `AGENT_EXECUTOR_PRIVATE_KEY` is not the final Agent Wallet design.
- Circle Agent Wallet is the preferred provider if the official CLI path works on `ARC-TESTNET`.
- Circle CLI was installed globally as an external tool with `npm install -g @circle-fin/cli`; no SignalArc package manager files were intentionally changed for Circle CLI.
- `circle --version` returned `0.0.3`.
- Circle CLI safe help discovery was completed.
- Circle CLI auth/chain discovery is blocked by user-controlled Terms acceptance before login or OTP; `circle terms accept` must be run manually by the user before authenticated discovery continues.
- Official Circle docs show `circle wallet login you@example.com --testnet`, but the installed `circle wallet login --help` output did not list a `--testnet` option before Terms/login validation.
- User-provided Circle CLI evidence now proves a real `ARC-TESTNET` Circle Agent Wallet at `0x96d5051a005547eba149f71604ccf58ae1a7c950`, distinct from deployer/resolver `0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC`.
- User-provided funding proof showed `20 USDC`.
- User-provided create-market evidence recorded tx `0x7142dbd7eebe7cbfb19199d9984efa5cef814d0e6038c17b98f2e98cc731cacf`, source address `0x96d5051a005547eba149f71604ccf58ae1a7c950`, state `COMPLETE`, `marketCount() == 8`, and `allMarkets(7) == 0xAbCf081E456C1a11106deF590666A07B76D456f8`.
- User-provided market readback proved market `0xAbCf081E456C1a11106deF590666A07B76D456f8`, collateral token `0x3600000000000000000000000000000000000000`, admin/resolver both `0x96d5051a005547eba149f71604ccf58ae1a7c950`, and `isOpen == true`.
- User-provided buy YES evidence recorded approve tx `0xeb7304b0a1be9f5dc575f62fb705dfaf384bc720da13f7e4ffe9563442c036ca`, buy tx `0xe311d999e15e6f34fa6f623a8f27bc724c665d7c3296632460339326b6094b16`, and readback `yesPositions(agent wallet) == 1000000`, `totalYes == 1000000`, `totalCollateral == 1000000`.
- User-provided buy NO evidence recorded approve tx `0x6ea6a10293a4df5d7ed50e077821115571787d8e9d6b9507a984ebf33fc52a9b`, buy tx `0xaefe8bcdcec794c811d615517f0dfa800b9e263631200a74c85d000374aa8f24`, and readback `noPositions(agent wallet) == 1000000`, `totalNo == 1000000`, `totalCollateral == 2000000`, `USDC.balanceOf(market) == 2000000`.
- Continuation from this Codex shell is blocked because Circle CLI returns `AUTH_REQUIRED` or no active agent session for status/list/balance, even when `CIRCLE_ACCEPT_TERMS=1` is set for the process.
- Added persistent `agent_wallets` migration for user-owned Circle Agent Wallet onboarding.
- Added migration `000016_create_agent_onboarding_sessions` for `agent_onboarding_sessions` and `agent_sessions`.
- `POST /agent/onboarding/register` remains a registry-only convenience endpoint for creating the final agent wallet mapping; it does not model OTP or Circle session isolation by itself.
- Added `POST /agent/onboarding/start` as the pending-session foundation for per-user, per-agent onboarding. It creates `pending_otp` state only and returns `circle_otp_verification_not_implemented`.
- Corrected `POST /agent/onboarding/start` to be agent-first and email-based: `user_wallet` is no longer required for initial agent onboarding and belongs to frontend wallet-connect or user transaction flows.
- Added disabled-by-default Circle Agent Wallet OTP start skeleton behind `CIRCLE_AGENT_ONBOARDING_OTP_START_ENABLED=false` by default. When enabled for a controlled dev runtime, it can call the Circle CLI login init runner, store only a hashed request reference plus expiry, and return `circle_otp_required` without exposing the raw request ID.
- Added disabled-by-default Circle Agent Wallet OTP verify skeleton at `POST /agent/onboarding/verify`. It uses the in-memory request ID from OTP start, consumes it on successful fake-runner verification, updates onboarding status to `verified`, and returns `agent_wallet_resolution_not_implemented`.
- Corrected OTP verify CLI completion command shape to the documented Circle form: `circle wallet login --request <request-id> --otp <code>`.
- Added sanitized server-side diagnostics for Circle OTP verify CLI failures; diagnostics redact the raw request ID and OTP and the API response remains the generic `circle_otp_verify_failed` error.
- After commit `b1280f0`, Custom GPT onboarding start returned `circle_otp_start_failed` even though the OTP email was delivered.
- Root cause identified: Circle CLI OTP init can print the request ID as text, while backend OTP start handling only accepted JSON request ID fields.
- Updated OTP start handling to accept JSON `request_id` / `requestId` and documented text-style printed request IDs.
- Raw Circle request IDs remain hidden from API responses and are stored only in memory for verify; the database stores only the hash.
- Added sanitized Circle CLI OTP start diagnostics that redact email and request ID.
- Added tests for JSON/text request ID extraction, sanitized start diagnostics, and successful `/agent/onboarding/start` with text request ID output.
- Validation: `go test ./...` passed.
- After commit `a14f17d`, Custom GPT onboarding start still returned `circle_otp_start_failed` although the OTP email was delivered.
- Observed Circle CLI output placed the request ID inside a nested message containing the documented completion command shape: `circle wallet login --request <id> --otp <code>`.
- Updated OTP start handling to extract request IDs from completion-command output and treat that as success even if the CLI runner returned an error.
- Strengthened sanitized diagnostics to redact request IDs embedded in command text.
- Validation: `go test ./...` passed.
- Runtime validation after commit `65cebdb` passed through Custom GPT/ngrok:
  - `/agent/onboarding/start` returned `pending_otp` and `next_step=circle_otp_required`.
  - Onboarding ID: `agent_onboarding_ad068b4c0538605db092819aa55df08d`.
  - Agent ID: `agent_adenhusen65_chatgpt_008`.
  - `/agent/onboarding/verify` accepted the emailed OTP and returned status `verified`.
  - Backend returned expected boundary `agent_wallet_resolution_not_implemented`.
  - No market, intent, confirmation, execution, or user wallet request was performed.
- Post-OTP wallet/session activation added.
- Verified OTP onboarding can now resolve a Circle Agent Wallet address from read-only Circle CLI list output.
- Backend registers the resolved wallet and creates an active agent session.
- Backend exposes read-only wallet/session status and balance for Custom GPT.
- No funding, transfer, market creation, or transaction execution is performed by the activation step.
- Validation: `go test ./...` passed.
- Added read-only onboarding/session status APIs: `GET /agent/onboarding/{onboarding_id}` and `GET /agent/sessions/{agent_id}`.
- Multi-tenant/session isolation state now separates user email, user wallet, source client, channel, pending onboarding, and activated agent-session boundaries without storing Circle session secrets.
- Backend now registers agent wallets through DB-backed `POST /agent/wallets` in production routing.
- Backend now returns registered wallet metadata through `GET /agent/wallets/{agent_id}`.
- Backend now disables registered wallets through `POST /agent/wallets/{agent_id}/disable`.
- Test handlers use an in-memory registry only as a test double; production registration is no longer the temporary in-memory development path.
- Backend agent intents now carry `agent_id` and optional `agent_wallet_address`.
- Backend agent intents now carry `source_client` and optional `client_request_id` so WhatsApp, Telegram, ChatGPT, Claude, and web clients can share one channel-agnostic API contract.
- Backend execution now rejects missing agent wallets, deployer/resolver wallet reuse, user-wallet reuse, disallowed actions, wrong chain, inactive wallet, mismatched wallet address, and unsupported provider execution.
- Added a guarded backend Circle CLI executor provider for `circle_agent_wallet`.
- Provider execution mode is `circle_agent_wallet_cli`.
- Provider is disabled by default with `CIRCLE_AGENT_WALLET_EXECUTION_ENABLED=false`.
- Provider config added: `CIRCLE_CLI_PATH`, `CIRCLE_AGENT_WALLET_CHAIN`, and `CIRCLE_AGENT_WALLET_TIMEOUT_SECONDS`.
- Provider supports `create_market`, `buy_yes`, `buy_no`, `close_market`, `resolve_market`, `claim_payout`, `cancel_market`, and `claim_refund` when the registered agent wallet allowlist permits the action.
- Lifecycle support uses repo-confirmed `SignalArcAgentMarket` functions only:
  - `close_market`: `closeMarket()`
  - `resolve_market`: `resolve(uint8)` from contract `resolve(Outcome winningOutcome_)`
  - `claim_payout`: `claimPayout()`
  - `cancel_market`: `cancelMarket()`
  - `claim_refund`: `claimRefund()`
- Lifecycle readbacks use repo-confirmed getters only: `status()`, `winningOutcome()`, `claimablePayout(address)`, `claimableRefund(address)`, `hasClaimed(address)`, `isOpen()`, and USDC `balanceOf(address)`.
- Provider uses Circle CLI `wallet execute` for writes and `contract query` for readbacks, with JSON-only parsing and sanitized errors.
- Provider never calls Circle login, never accepts OTP, and never stores Circle tokens, session files, private keys, or deployer keys.
- Runtime validation was performed from the host shell backend on port `4001`, not the Docker backend container, because the Docker backend container does not have Circle CLI installed.
- Host Circle Agent Wallet session was valid for the recorded backend provider run.
- DB migration was advanced to version `15`, `dirty=false`.
- Agent wallet was registered through DB-backed `POST /agent/wallets`.
- Registered agent:
  - `agent_id`: `agent_desi_001`
  - `user_email`: `desi33905@gmail.com`
  - `user_wallet`: `0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC`
  - `agent_wallet_address`: `0x96d5051a005547eba149f71604ccf58ae1a7c950`
  - `wallet_provider`: `circle_agent_wallet`
  - `chain`: `ARC-TESTNET`
  - `allowed_actions`: `create_market`, `buy_yes`, `buy_no`
  - `source_client`: `manual_backend_runtime`
- Backend provider `create_market` evidence:
  - `intent_id`: `agent_intent_f76bd9653e9ce3f6269023a25e7c6b8c`
  - `execution_mode`: `circle_agent_wallet_cli`
  - Transaction: `0x7aa51a0d19b163a3a88ae16ac4a88a1cdbb3090cad5d0ccc54f828b166f74e5d`
  - Created market: `0x38aE7E0133e9594F597F913884cbDa619A950523`
  - Readback: `market_count == 9`, `is_market == true`
- Backend provider `buy_yes` evidence:
  - `intent_id`: `agent_intent_d6f9cd78896886f954a05a164c17c067`
  - `execution_mode`: `circle_agent_wallet_cli`
  - Market: `0x38aE7E0133e9594F597F913884cbDa619A950523`
  - Approve transaction: `0x09d0a418c34b0e54a31bc3a2a7bfba85218eba27e84becdcbd89e5b63b8bb387`
  - `buyYes` transaction: `0xa1fadb400aa8b4babca0c936698e686eeaac3ae408b22d1e37960901a5c48ade`
  - Readback: `yes_positions == 1000000`, `total_yes == 1000000`, `total_collateral == 1000000`, `USDC.balanceOf(market) == 1000000`
- Backend provider `buy_no` evidence:
  - `intent_id`: `agent_intent_ed7050305f62798d8472b7f48e538ff8`
  - `execution_mode`: `circle_agent_wallet_cli`
  - Market: `0x38aE7E0133e9594F597F913884cbDa619A950523`
  - Approve transaction: `0x40ba807e24e2fcfeba22f21575920d3dfe5389f7f00320ab5a74fb46f06c6dc8`
  - `buyNo` transaction: `0xb36288fc40a69765d62679982fdd3319d09b27c010e2eb9caafa9c6508d03e9c`
  - Readback: `no_positions == 1000000`, `total_no == 1000000`, `total_collateral == 2000000`, `USDC.balanceOf(market) == 2000000`
- Latest full lifecycle runtime evidence was performed from the host shell backend on `APP_PORT=4001` with execution mode `circle_agent_wallet_cli`, agent `agent_desi_001`, agent wallet `0x96d5051a005547eba149f71604ccf58ae1a7c950`, wallet provider `circle_agent_wallet`, and chain `ARC-TESTNET`.
- Docker backend was not used for the latest execution because the Docker backend container still does not include the Circle CLI/session strategy.
- Backend provider payout lifecycle evidence:
  - `create_market` intent `agent_intent_ecc88160f7e2b908fc498c3dff66fbe7`, transaction `0x7913fd51b38b147cfc6936da7eb7166a97a156351c9ccce4264d48d31fc91ae9`, market `0x38D4317fcB0C82e5EC2407a89c311b3Be8059CD0`, readback `market_count == 10`, `is_market == true`
  - `buy_yes` intent `agent_intent_e1d3889f3252ac318b7b435b2a789d9d`, approve transaction `0x9c66332ad2d798126118b961ad9005ab7f2055649b46d808cc979ccf40eee3f7`, buy transaction `0xdd4448fdf237f13bc9e90737f27b7a1e912ee8d34195bae311cc2bac15aaa95d`, readback `yes_positions == 1000000`, `total_yes == 1000000`, `total_collateral == 1000000`, `USDC.balanceOf(market) == 1000000`
  - `close_market` intent `agent_intent_6cab73b950296ac2dcc3a5414f4e1613`, transaction `0x5accb4dafee2a27be032709e427e25eede4e7eb67f54a06bdf0ba82e5f4a013e`, readback `market_status == 1`, `is_open == false`
  - `resolve_market` intent `agent_intent_4073e0f29f842a0110a1e099d9fa50b0`, transaction `0xc45d8612768591736425a743520b5e432b9a424ef9552cfbbc1bb04d785c874b`, readback `market_status == 2`, `winning_outcome == 1`, `claimable_payout == 1000000`, `has_claimed == false`, `USDC.balanceOf(market) == 1000000`
  - `claim_payout` intent `agent_intent_65284f22e52fce32d6a3efc2fa6163cd`, transaction `0xc80bb9dd7f6924c93c1722d7c5f1136c076403d052ed792f98ed2d7abd59568f`, readback `market_status == 2`, `winning_outcome == 1`, `claimable_payout == 1000000`, `has_claimed == true`, `USDC.balanceOf(market) == 0`
- Backend provider refund lifecycle evidence:
  - `create_market` intent `agent_intent_b167e466b198c7f909c388c76b683f5d`, transaction `0x1340817f922aaa7ae181789bbaf2b7bff13a426b0397524150d01ae869d2a033`, market `0xbfd93169DAFf0610EA10E1221B9a2a6552379648`, readback `market_count == 11`, `is_market == true`
  - `buy_yes` intent `agent_intent_b06dd357e62778ed1a0527800690473a`, approve transaction `0x83bdc164512979296c385f51b4b6b1df51c741f2157cc93944b7f15c1328f487`, buy transaction `0xa85c60d903a1b1e30f39da95a37ca356d61d11cc20c942aaa22f0740a06945ab`, readback `yes_positions == 1000000`, `total_yes == 1000000`, `total_collateral == 1000000`, `USDC.balanceOf(market) == 1000000`
  - First backend provider `cancel_market` attempt failed: intent `agent_intent_634dad5fc81b7a155e3854d8e14bb135`, backend response `502 agent_execution_failed`; onchain check after failure showed the market was still Open. This failed attempt is not validated backend cancel evidence.
  - Manual Circle CLI `cancelMarket` then succeeded on the refund market with transaction `0xbf8b98862ed691c0023643ab72ee71d8422e434868b2806f3390b4ffc88fe21b`; this was manual Circle CLI, not backend provider evidence.
  - Backend provider `claim_refund` succeeded after manual cancel: intent `agent_intent_c519c98114f8a5c113ab4bc6dfefbcae`, transaction `0x40f3f4e1737340dbbbff92e3020d0cfa6dbd7d7bbca8a1ecb580b1c0cdc43dfd`, readback `market_status == 3`, `claimable_refund == 1000000`, `has_claimed == true`, `USDC.balanceOf(market) == 0`
- Backend provider cancel-only validation evidence:
  - `create_market` intent `agent_intent_666c781a7a3fe2f939bd5342e331fd67`, transaction `0x62f5f43b834a09f4f4c78e3bb365403a2b76374d0030a91ccd3e3571fd3c9d12`, market `0x928F3F9Cb43811837C0e8D4FA40c24A4f083B3Ed`, readback `market_count == 12`, `is_market == true`
  - `cancel_market` intent `agent_intent_34f54290e5f59ffaff4a167986868c56`, transaction `0x3eae0d0508397e5ea515d417bdc5be5c38f40f3b0296b1cf424d99060cb92de4`, readback `market_status == 3`, `claimable_refund == 0`, `has_claimed == false`, `USDC.balanceOf(market) == 0`
- Current proven status:
  - Backend Circle provider `create_market` validated with a real Arc Testnet transaction.
  - Backend Circle provider `buy_yes` validated with a real Arc Testnet transaction.
  - Backend Circle provider `buy_no` validated with a real Arc Testnet transaction.
  - Backend Circle provider `close_market` validated with a real Arc Testnet transaction.
  - Backend Circle provider `resolve_market` validated with a real Arc Testnet transaction.
  - Backend Circle provider `claim_payout` validated with a real Arc Testnet transaction.
  - Backend Circle provider `cancel_market` validated after one failed backend provider attempt on the refund market and a later successful backend-only cancel-only test.
  - Backend Circle provider `claim_refund` validated after manual Circle CLI cancellation of the refund market.
- Backend-to-Circle-to-Arc path is proven for create, trade, close, resolve, payout claim, cancel, and refund claim flows on Arc Testnet.
- ChatGPT Custom GPT Action external client trigger evidence:
  - External client: ChatGPT Custom GPT Action.
  - Temporary tunnel: ngrok HTTPS URL to local backend.
  - Backend runtime: host shell on `APP_PORT=4001`.
  - Execution mode: `circle_agent_wallet_cli`.
  - Backend public test path: ChatGPT Custom GPT Action -> ngrok HTTPS -> `localhost:4001` SignalArc backend -> Circle Agent Wallet CLI provider -> Arc Testnet.
  - Local backend health returned 200: `http://127.0.0.1:4001/health` -> `{"status":"ok"}`.
  - Ngrok public health returned 200: `https://undamaged-commerce-juggling.ngrok-free.dev/health` -> `{"status":"ok"}`.
  - Custom GPT `getHealth` action returned `{"status":"ok"}`.
  - Custom GPT Action schema exposed `getHealth`, `createAgentIntent`, `confirmAgentIntent`, and `executeAgentIntent`.
  - Authentication was set to None only for temporary local tunnel testing. This is not production API authentication.
  - Initial intent preview and confirm were tested.
  - One confirm failed due to a typo in a manually copied intent ID. This was a user/operator copy error, not a backend failure.
  - One later execution attempt failed because `close_timestamp` was already in the past:
    - `close_timestamp` used: `1779439999`
    - current timestamp checked later: `1779452713`
    - probable root cause: `createMarket` reverted due to `closeTimestamp <= block.timestamp`
    - Circle Agent Wallet session was checked and testnet `tokenStatus` was `VALID`
    - This is recorded as stale input evidence, not a proven provider failure.
  - New valid preview used future `close_timestamp`: `1779456313`.
- Successful ChatGPT Custom GPT external execution:
  - Intent ID: `agent_intent_7d0bfd385329ba97cb7c1b88ada6f049`
  - Action: `create_market`
  - Status: `executed`
  - Execution mode: `circle_agent_wallet_cli`
  - Network: `arc_testnet`
  - Agent factory: `0x69aE770e8b2F96297101FeC4dc123B3801dA7d80`
  - Agent ID: `agent_desi_001`
  - Agent wallet: `0x96d5051a005547eba149f71604ccf58ae1a7c950`
  - Broadcast performed: `true`
  - Transaction hash: `0x1062e254f8640ffdc2d75d368754e5d698d42cb935bc7bcb113547c7a501aec2`
  - Created market: `0x6cef2f33F0F2a5e01E885176bAa17709d6A6a299`
  - Readback:
    - `market_count == 13`
    - `is_market == true`
- Current external client trigger status:
  - ChatGPT Custom GPT trigger validated for health, create intent preview, confirm intent, and execute `create_market` real Arc Testnet transaction.
  - WA, Telegram, and Claude triggers are still not tested.
- No secrets, `.env` files, frontend code, production deployment config, commits, pushes, or deploys were changed.

Current non-claims:

- WhatsApp, Telegram, and Claude live client triggers have not been tested yet.
- Docker/Cloud Run runtime does not yet include a Circle CLI/session strategy.
- No production readiness claim.
- No mainnet claim.
- No Circle policy limit claim on `ARC-TESTNET`.
- Circle session storage is not implemented.
- Circle wallet provisioning readback beyond read-only CLI wallet list output is unknown / not documented.
- Custom GPT / Claude / Telegram / WA onboarding schemas should not require `user_wallet` for start-agent onboarding.
- Backend restart before OTP verify requires onboarding restart because raw request IDs are not stored in the database.
- No real Circle CLI run was performed for the OTP start skeleton tests.
- The temporary ngrok tunnel is not a production API endpoint.
- Circle CLI command shapes in `project-roadmap/agent-mcp.md` are official-doc/help-discovery shapes only unless accompanied by exact authenticated output and onchain evidence.
- The phase is not complete.

## Circle CLI Warning-Prefix Wallet Resolver Fix

Status: COMPLETE.

Done:

- Runtime verification after commit `5697295` reached wallet resolution but failed because Circle CLI prepended Node deprecation warning text before JSON output.
- Root cause: `parseCircleAgentWallets` and `parseCircleAgentWalletBalances` in `circle_wallet_resolver.go` called `json.Unmarshal` directly on raw CLI output, which fails when Node.js prints `(node:...) [DEP0040] DeprecationWarning...` lines before the JSON payload.
- Added `extractJSONFromCLIOutput` helper that scans for the first `{` or `[` that starts valid JSON, skipping warning text and bracket characters in deprecation labels like `[DEP0040]`.
- Updated `parseCircleAgentWallets` and `parseCircleAgentWalletBalances` to use the helper before unmarshalling.
- Clean JSON output still parses identically (fast path).
- Output without any valid JSON object/array still returns a parse error.
- Sanitized diagnostics behavior unchanged.
- Validation: `go test ./...` passed.

## Cloud Run Backend Image Includes Circle CLI

Status: COMPLETE (image-level only; durable Circle CLI session persistence remains a production runtime concern).

Done:

- Cloud Run production reached the live domain `api.signalarc.fun` and the production database was migrated to version `18`.
- Production agent endpoints then reached the Circle runtime boundary because the Cloud Run runtime image did not include Node.js, npm, or the Circle CLI binary, so commands like `circle wallet login`, `circle wallet list`, and `circle wallet balance` could not be executed inside the deployed container.
- `backend/Dockerfile` runtime stage was updated to install `nodejs` and `npm` from Alpine packages and then install the Circle CLI globally with `npm install -g @circle-fin/cli`, then clean the npm cache.
- The runtime stage continues to copy the compiled `signalarc-api` binary from the Go builder stage into `/usr/local/bin/signalarc-api`.
- The image final layout exposes both `signalarc-api` and `circle` on `PATH` at `/usr/local/bin`.
- The existing `ENTRYPOINT ["signalarc-api"]` was preserved.
- No Circle credentials, OTP request IDs, session files, private keys, or API keys are baked into the image.
- This change does not solve durable Circle CLI session persistence; session strategy remains a production runtime concern outside the image.

Validation:

- `cd backend && go test ./...` passed; no Go test files were modified.
- `docker build -t signalarc-backend:circle-cli-test backend` succeeded.
- Inside the built image:
  - `command -v circle` returned `/usr/local/bin/circle`.
  - `circle --version` returned `0.0.3`.
  - `command -v signalarc-api` returned `/usr/local/bin/signalarc-api`.
  - `node --version` returned `v20.15.1` and `npm --version` returned `10.9.1`.

Not done:

- Production deploy of the updated image to Cloud Run was not performed in this step.
- Circle CLI session/login persistence inside Cloud Run is still unresolved.

## ARC-TESTNET Faucet Helper Endpoint

Status: COMPLETE.

Done:

- Added ARC-TESTNET faucet helper endpoint for active registered agent wallets at `POST /agent/wallets/{agent_id}/faucet`.
- Endpoint uses the registered `agent_wallet_address` from the SignalArc database only and does not accept any arbitrary recipient address from the request body.
- Endpoint requires the registered wallet to exist (`404 agent_wallet_not_found`), to be `active` (`409 agent_wallet_status_invalid`), and to be on `ARC-TESTNET` (`409 agent_wallet_chain_invalid`).
- Endpoint calls only the documented Circle CLI testnet faucet command shape `circle wallet fund --address <registered_agent_wallet_address> --chain ARC-TESTNET --token usdc --output json`.
- Endpoint never passes `--amount`, `--method`, `--open`, `--export`, `transfer`, `swap`, `execute`, or any mainnet funding option, and does not transfer, swap, execute contracts, create markets, or use mainnet funding.
- Faucet runner is disabled by default behind `CIRCLE_AGENT_WALLET_FAUCET_ENABLED=false`. When disabled or not wired, the endpoint returns `501 circle_agent_wallet_faucet_not_configured`.
- Provider failures return `502 circle_agent_wallet_faucet_failed` with sanitized API responses; CLI/runner output is never echoed back to clients and credential paths, tokens, OTP-like values, request IDs, session material, and emails are redacted from server-side diagnostics.
- Output parser reuses the existing warning-prefixed JSON extraction helper used by wallet list/balance, so Node deprecation warnings before JSON still parse cleanly.
- Successful JSON CLI output is returned under `result`. Successful text-only CLI output is returned under `result.message` after sanitization.
- Updated Custom GPT OpenAPI schema `project-roadmap/signalarc-custom-gpt-openapi.json` with the `requestAgentWalletFaucet` action, including `404`, `409`, `501`, and `502` error responses; server URL remains `https://api.signalarc.fun` and `/agent/onboarding/register` was not re-added.
- Validation: `go test ./...` passed.

Not done:

- Production deploy of the updated image to Cloud Run was not performed in this step.
- No funding, transfer, swap, contract execution, market creation, mainnet funding, login/logout, or session persistence was added.

## Public Documentation Refresh For Live Architecture And Faucet

Status: COMPLETE.

Done:

- Updated `docs/index.mdx`, `docs/API.md`, `docs/AGENT_API.md`, `docs/DEPLOYMENT_PLAN.md`, `docs/ARCHITECTURE.md`, and `docs/GRANT_READINESS.md` to describe the live architecture: `https://signalarc.fun` on Vercel for the frontend, `https://api.signalarc.fun` on GCP Cloud Run service `signalarc-backend-api` for the backend, GCP Cloud SQL PostgreSQL migrated to version 18, and the Cloud Run image bundling Node/npm and the Circle CLI (`@circle-fin/cli`) on `PATH`.
- Documented the agent flow `onboarding start -> OTP verify -> active session -> wallet -> balance -> faucet -> create intent -> confirm intent -> execute intent` in the agent API reference.
- Documented the new `POST /agent/wallets/{agent_id}/faucet` endpoint, response shape, failure codes (`404 agent_wallet_not_found`, `409 agent_wallet_status_invalid`, `409 agent_wallet_chain_invalid`, `501 circle_agent_wallet_faucet_not_configured`, `502 circle_agent_wallet_faucet_failed`), token-fixed-to-USDC behavior, ARC-TESTNET-only behavior, and that the endpoint never accepts arbitrary recipient addresses, transfers, swaps, contract executions, market creations, or mainnet funding.
- Added a judge/user testing guide that walks through opening the SignalArc GPT Agent, connecting an account, providing email, entering OTP, checking wallet, checking balance, requesting the faucet, creating a draft market intent, confirming, and executing only after explicit approval.
- Documented the maintainer-only OpenAPI import URL `https://raw.githubusercontent.com/wahyu241205/SignalArc/main/project-roadmap/signalarc-custom-gpt-openapi.json` and clarified that judges/users do not need to import it because the published GPT Agent is already wired to https://api.signalarc.fun.
- Refreshed the docs status table to show available capabilities (health, onboarding, OTP verify, session, wallet, balance, faucet, market intent lifecycle) and explicit out-of-scope capabilities (arbitrary transfer, withdraw/deposit, logout/session management, mainnet funding).
- ngrok references in docs are now explicitly marked as local development only.
- `docs/docs.json` navigation was reviewed; no structural change was required because API and AGENT_API are already linked.
- Validation: `grep -R "ngrok-free\\|undamaged-commerce\\|signalarc-backend-973633221696\\|localhost:4001" docs project-roadmap/signalarc-custom-gpt-openapi.json` returned no matches.

Not done:

- No backend, frontend, or contract logic was changed.
- No commits were created and nothing was pushed.

## Next Recommended Step

- Design and validate the Docker/Cloud Run Circle CLI/session strategy before treating the backend provider as deployable. Do not capture OTP or store Circle session material in SignalArc.
- Validate WA, Telegram, or Claude external client triggers through the existing backend-approved path when explicitly approved.
- Do not add Circle API keys, deployer/user private keys, DNS, live deployment, contract redeploy, frontend execution UI, or mainnet configuration yet.

Do not start unrelated coding before checking:

```bash
git status
```

## Important Rules

1. Frontend is UI only.
2. Backend owns business logic.
3. Database schema must be migration-based.
4. Arc and Circle behavior must use official documentation only.
5. Unknown or undocumented behavior must be marked as unknown / not documented.
6. Do not invent integrations.
7. Do not commit secrets.
8. Do not commit `node_modules`, `.next`, `dist`, `out`, or `.env`.
9. Keep commits small and reviewable.
10. Build MVP first, then polish.


## SignalArc Custom GPT / Agent Onboarding Stability Pass

Status: COMPLETE.

Done:

- Added backend `agent_id` validation helper in `backend/internal/api/agent_id_validation.go` that rejects empty values, generic placeholder values such as `signalarc-gpt-agent`, `agent_desi_001`, `default`, `defaultagent`, `default_agent`, `test`, `testagent`, `test_agent`, `demo`, `demoagent`, `demo_agent`, `user`, `useragent`, `user_agent`, `agent`, `chatgpt`, `chatgpt_agent`, and any case-insensitive equivalent, requires the `agent_<slug>` SignalArc shape, requires at least 10 characters, and limits the alphabet to ASCII letters, digits, underscore, and hyphen.
- Wired the helper into `validateAgentOnboardingSessionInput` and `validateAgentWalletRegistrationInput` in `backend/internal/api/agent_handlers.go`. `POST /agent/onboarding/start` now returns HTTP 400 `agent_onboarding_invalid` for generic/short/wrong-shape `agent_id` values and `POST /agent/wallets` returns HTTP 400 `agent_wallet_invalid` for the same.
- Added `normalizeCloseTimestamp` in `backend/internal/agent/intents.go` that accepts either a base-10 unix-seconds string (preserved unchanged for backward compatibility with the existing executor) or a UTC RFC3339 timestamp such as `2026-05-31T23:59:00Z`, normalizing the latter to unix seconds before the validator runs. The on-chain `SignalArcAgentMarketFactory.createMarket(uint256)` signature is unchanged because the executor still passes a base-10 unix-seconds string to `big.Int.SetString(value, 10)`.
- Tightened `validateIntent` so a `create_market` intent with a non-numeric `close_timestamp` returns a stable validation error such as `close_timestamp must be a unix-seconds integer or an RFC3339 timestamp such as 2026-05-31T23:59:00Z`, instead of a confusing `invalid_json`.
- Replaced the generic `invalid_json` message on `POST /agent/intents` with `request body must be valid JSON; for create_market send close_timestamp as RFC3339 (for example 2026-05-31T23:59:00Z) or unix-seconds integer string`, so non-developer Custom GPT users get a clear hint about how to convert natural-language dates.
- Hardened `project-roadmap/signalarc-custom-gpt-openapi.json` (now `0.2.0`):
  - Added top-level Custom GPT instruction language about generating a unique `agent_id` per user, never using generic placeholders, and converting natural-language dates to RFC3339 before calling create market intent.
  - Added explicit `pattern`, `minLength`, `description`, and `examples` for `agent_id` on `startAgentOnboarding` request body and on `agent_id` path parameters in `getAgentSession`, `getAgentWallet`, `getAgentWalletBalance`, and `requestAgentWalletFaucet`.
  - Added explicit `pattern` and `description` for `agent_id` on `createAgentIntent`, `confirmAgentIntent`, and `executeAgentIntent`.
  - Changed `createAgentIntent.close_timestamp` to a `string` with description and examples for both RFC3339 and unix-seconds shapes, and added inline `examples` for `create_market` and `buy_yes` payloads.
  - Documented the read-only market discovery surface as `listMarkets` (`GET /markets`), `getMarket` (`GET /markets/{market_id}`), and `listAgentMarkets` (`GET /agent/markets`) so the Custom GPT can show open/trending markets by calling an Action instead of refusing.
  - Added 201 Created responses for `startAgentOnboarding` and `createAgentIntent` matching the actual backend status codes.
  - Added `Market` and `AgentMarket` component schemas.
- Updated `docs/AGENT_API.md`:
  - Added a Custom GPT instruction block, an explicit `agent_id` naming rule list with the documented blocklist, and live recommended shapes `agent_sanatarau21_chatgpt_001` / `agent_adenhusen65_live_002`.
  - Documented the RFC3339 / unix-seconds rule for `create_market.close_timestamp` and the new stable error message.
  - Documented `GET /markets` and `GET /markets/{market_id}` as exposed read-only Actions, alongside `GET /agent/markets`.
  - Replaced the demo `agent_demo_001` example with the recommended `agent_sanatarau21_chatgpt_001` shape.
- Added Go tests in `backend/internal/api/agent_id_validation_test.go`:
  - `TestStartAgentOnboardingRejectsGenericAgentIDValues` covers each documented placeholder including case variants.
  - `TestStartAgentOnboardingRejectsShortAgentID` covers the minimum length rule.
  - `TestStartAgentOnboardingAcceptsRecommendedAgentIDShape` covers the recommended shape.
  - `TestRegisterAgentWalletRejectsGenericAgentID` covers the explicit `/agent/wallets` registration path.
  - `TestCreateAgentIntentRejectsInvalidJSON` covers the natural-language date case where the body itself is not valid JSON.
  - `TestCreateAgentIntentRejectsNonRFC3339CloseTimestamp` covers a body that decodes but contains a natural-language `close_timestamp`.
  - `TestCreateAgentIntentAcceptsRFC3339CloseTimestamp` covers the recommended RFC3339 input and verifies the response normalizes to unix seconds.
  - `TestCreateAgentIntentAcceptsUnixSecondsCloseTimestamp` covers the existing unix-seconds clients.
  - `TestValidateAgentIDDirect` table-tests the helper in isolation.

Validation results:

- `cd backend && go vet ./...` passed.
- `cd backend && go test ./...` passed (agent + api packages, including the new tests).
- `python3 -m json.tool project-roadmap/signalarc-custom-gpt-openapi.json > /tmp/signalarc-openapi.validated.json` passed.
- `grep -R 'ngrok-free\|undamaged-commerce' project-roadmap/signalarc-custom-gpt-openapi.json docs` returned no matches.

Out of scope for this pass:

- No frontend code change.
- No Solidity / contract change.
- No new arbitrary transfer, withdraw, deposit, swap, or mainnet feature.
- No commit and no push.

## Backend stopgap — Circle CLI session liveness

Problem:
- Production Cloud Run logs showed Circle CLI `AUTH_REQUIRED`.
- SignalArc DB could report an agent session as `active`, while Circle CLI runtime session was unavailable on the Cloud Run instance handling balance/execute.
- This caused misleading session status and opaque `agent_execution_failed` / `circle_agent_wallet_balance_failed` errors.

Completed:
- Added backend-only Circle CLI session liveness detection for active Circle Agent Wallet sessions.
- `GET /agent/sessions/{agent_id}` no longer blindly returns `active` when local Circle CLI session is unavailable.
- Balance and execute public error codes are preserved.
- Backend logs now include sanitized structured Circle provider failure details.

Important limitation:
- This does not fix Cloud Run Circle CLI session persistence.
- It only makes runtime session loss visible and non-misleading.
- Circle Agent Wallet HTTP API replacement remains unknown / not documented.

Validation:
- `go build ./...` passed.
- `go test ./...` passed.
- `go vet ./...` passed.
- `gofmt` clean on changed Go files.

## Backend runtime hardening — Circle CLI Node engine

Problem:
- Cloud Build for image tag `948ca29` succeeded, but npm emitted `EBADENGINE` warnings.
- `@circle-fin/cli@0.0.3` requires Node `>=20.18.2`.
- The previous runtime image installed Node `v20.15.1` from `alpine:3.20`.
- Because Circle CLI is a backend runtime dependency for Agent Wallet balance/execute paths, this mismatch is not acceptable as a long-term baseline.

Completed:
- Updated backend runtime Docker stage to use pinned `node:20.18.2-alpine3.20`.
- Removed Alpine `nodejs`/`npm` package install from the runtime stage.
- Kept `ca-certificates` and global `@circle-fin/cli` install.
- No backend Go logic, frontend, OpenAPI, docs, contracts, migrations, or database behavior changed.

Validation required before deploy:
- Build backend image again.
- Confirm the previous `EBADENGINE` warning is gone.
- Deploy only the new image if build succeeds.

## Custom GPT Actions schema compatibility fix

Problem:
- Backend and Cloud Run were healthy.
- `curl https://api.signalarc.fun/health` returned HTTP 200 with `{"status":"ok"}`.
- A minimal Custom GPT Action schema containing only `GET /health` worked.
- The full SignalArc OpenAPI schema failed in GPT Actions with `ClientResponseError` and `Failed to Parse JSON`.
- Cloud Run logs showed that failing full-schema Action calls did not reach the backend.
- Root cause was therefore schema compatibility with GPT Actions parser/runtime, not backend, Cloud Run, custom domain, or deploy.

Completed:
- Updated `project-roadmap/signalarc-custom-gpt-openapi.json` for GPT Actions compatibility.
- Changed OpenAPI version from `3.1.0` to `3.0.3`.
- Removed JSON Schema / OpenAPI 3.1 type arrays such as `["string", "null"]`.
- Replaced nullable-style fields with plain string schemas where nullable behavior is not required for Action routing.
- Replaced empty array item schema `{}` with explicit object item schema.
- Replaced schema-level `examples` with OpenAPI 3.0-compatible `example`.
- Preserved all 14 operationIds and kept server URL as `https://api.signalarc.fun`.

Validation:
- JSON parses successfully with Python.
- All 14 operationIds are present and unique.
- No `type` arrays remain.
- No empty `items` schemas remain.
- No `oneOf`, `anyOf`, or `allOf` remain.

Important limitation:
- Exact GPT Actions parser feature support is unknown / not documented.
- This fix intentionally uses conservative OpenAPI 3.0.3-compatible schema constructs.

## Backend diagnostics — Circle CLI execute failures

Problem:
- Agent onboarding, OTP verification, session, wallet, balance, market listing, intent preview, and intent confirm work.
- `executeAgentIntent` still fails with `agent_execution_failed`.
- Production logs showed `operation=agent_execution`, `error_class=unknown`, and generic summary `Circle CLI command failed`.
- Existing logs were not enough to distinguish contract revert, ABI/function mismatch, gas/funds issue, Circle CLI syntax, process crash, or unsupported Circle CLI execute behavior.

Completed:
- Added diagnostics-only context for Circle CLI execute failures.
- Public API behavior is unchanged: execute failures still return HTTP 502 with public error code `agent_execution_failed`.
- Raw Circle CLI stdout/stderr is not returned to API callers.
- Structured logs now include sanitized execute context where available:
  - action
  - function signature
  - redacted contract address
  - redacted wallet address
  - chain
  - command category
  - exit status
  - raw output length
  - sanitized process error
  - sanitized stdout/stderr summary
- Added sanitizer behavior to preserve `0x`-prefixed 64-hex transaction hashes while still redacting bare long hex secrets.

Validation:
- `go build ./...` passed.
- `go test ./...` passed.
- `go vet ./...` passed.
- `gofmt` clean on changed Go files.

Important limitation:
- This does not fix execute failure.
- It only makes the next execute failure diagnosable from production logs.
- Exact Circle CLI stderr format for execute failure modes remains unknown / not documented.

## Backend guard — create market stale close timestamp

Problem:
- After execute diagnostics were deployed, `create_market` execution failed inside Circle CLI with `Transaction failed: ESTIMATION_ERROR`.
- Diagnostics showed the call reached `wallet_execute` for `createMarket(string,string,uint256,address,address)`.
- ABI/signature and argument order matched the deployed factory.
- Most likely revert cause was stale `close_timestamp`, because the contract requires close timestamp to be greater than block timestamp.

Completed:
- Added backend-only pre-execution guard for `create_market`.
- `ExecuteCreateMarket` now validates `close_timestamp` is still sufficiently in the future before calling Circle CLI.
- Uses a 60-second safety margin.
- If stale, backend does not call Circle CLI.
- Public response is now HTTP 400 with error code `create_market_close_timestamp_stale`.
- Existing provider/Circle CLI failures still use HTTP 502 with `agent_execution_failed`.
- Non-`create_market` actions are not affected.
- Tests use dynamic future/past timestamps to avoid time-bomb failures.
- Timestamp comparison uses `big.Int.Cmp` instead of `Int64()` to avoid overflow behavior for very large timestamp inputs.

Validation:
- `go test ./...` passed.
- `go vet ./...` passed.
- `gofmt` clean on changed Go files.

Important limitation:
- This does not fix every possible `ESTIMATION_ERROR`.
- Duplicate `marketId`, unsupported collateral, authorization, or other contract reverts remain possible and should be handled separately if observed.

## Backend CD — Cloud Build to Cloud Run

Status:
- ACTIVE.

Completed:
- Added `cloudbuild.backend.yaml` to `main`.
- Created Cloud Build 2nd gen GitHub connection `signalarc-github` in `asia-southeast1`.
- Linked repository `wahyu241205/SignalArc` as Cloud Build repository `SignalArc`.
- Created trigger `signalarc-backend-cloud-run-main`.
- Trigger event is push to `main`.
- Trigger build config is `cloudbuild.backend.yaml`.
- Trigger service account is `973633221696-compute@developer.gserviceaccount.com`.
- Manual trigger test succeeded for commit `3ed8112`.
- Cloud Run deployed revision `signalarc-backend-api-00015-7h9`.
- Active image is `asia-southeast1-docker.pkg.dev/signalarc-prod-241205/cloud-run-source-deploy/signalarc-backend-api:3ed8112`.

Validation:
- Cloud Build status: SUCCESS.
- Cloud Run `/health`: HTTP 200 with `{"status":"ok"}`.
- Cloud Run `/readyz`: HTTP 200 with `{"status":"ok"}`.

Important notes:
- The old Cloud Run / Cloud Build 1st gen integration is not used for the new backend CD path.
- The active backend CD path is GitHub main push → Cloud Build 2nd gen trigger → `cloudbuild.backend.yaml` → Artifact Registry → Cloud Run.
- `/status` is not a backend route; the valid health endpoints are `/health` and `/readyz`.
