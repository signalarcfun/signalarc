# SignalArc Agent / MCP Roadmap

## Scope

Backend/API agent only. No frontend chat UI.

## Goal

Build a separate agent execution path for SignalArc.

Agent can later:
- read markets
- create market
- trade YES / NO
- claim refund
- claim payout
- resolve/cancel only if authorized

## Live AI Agent Transaction MVP

Status: IN PROGRESS - backend Circle Agent Wallet provider create_market, buy_yes, buy_no, close_market, resolve_market, claim_payout, cancel_market, and claim_refund are validated by real Arc Testnet runtime evidence; ChatGPT Custom GPT external client trigger is validated through a temporary ngrok tunnel for health, intent preview, confirm, and real create_market execution; Docker/Cloud Run Circle CLI/session strategy, production readiness, mainnet readiness, and WA/Telegram/Claude triggers remain pending.

Primary objective:
- A user owns or controls an agent.
- The agent has its own wallet address.
- The agent wallet is not the deployer wallet.
- The agent wallet is not the user personal wallet unless an explicitly documented user-controlled custody link is implemented.
- ChatGPT, Claude, Telegram, or other external clients are command interfaces only.
- SignalArc Backend Agent API remains the source of truth for intent validation, policy, permissions, balances, market status, execution state, and audit state.
- Every milestone transaction must include a real Arc Testnet tx hash, receipt success, onchain readback, and proof that the signer/wallet is the agent wallet.

### Agent Wallet Boundary

Do not use `contracts/.env` `PRIVATE_KEY` as an agent wallet.

Do not map `PRIVATE_KEY` to `AGENT_EXECUTOR_PRIVATE_KEY`.

Do not use the deployer, resolver, or user wallet as the agent wallet.

Do not treat `AGENT_EXECUTOR_PRIVATE_KEY` as the final Agent Wallet design. The existing backend EOA executor is legacy execution plumbing only and is not valid proof of the Live AI Agent Transaction MVP.

Deployer/resolver wallet explicitly forbidden as an agent wallet:
- `0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC`

Preferred wallet provider:
- `circle_agent_wallet`, if the live Circle Agent Wallet CLI path is authenticated and proven on `ARC-TESTNET`.

Documented fallback provider name only:
- `temporary_testnet_agent_eoa`

The fallback must never use the deployer private key, resolver wallet, user wallet, or any key copied from `contracts/.env`.

### Official Documentation Findings

Circle Agent Wallets:
- Designed for agents to hold, spend, trade, and earn USDC/tokens within spending controls.
- Built on Circle user-controlled wallets with 2-of-2 MPC.
- Agent key shares are not exposed to the agent.
- Users retain custody.
- Operated through Circle CLI.

Circle Agent Wallet ARC-TESTNET support:
- Circle documents `ARC-TESTNET` as the Arc Testnet chain identifier for Agent Wallet CLI commands.
- Circle documents `circle wallet login you@example.com --testnet`.
- Circle documents `circle wallet list --type agent --chain ARC-TESTNET`.
- Circle documents `circle wallet fund --address 0xYourWalletAddress --chain ARC-TESTNET` for testnet funding, with testnet wallets auto-funded from the Circle faucet.
- Circle documents `circle wallet balance --address 0xYourWalletAddress --chain ARC-TESTNET`.
- Circle documents `circle wallet execute <abiFunctionSignature> ... --address <agent_wallet> --contract <contract> --chain ARC-TESTNET`.
- Circle documents `circle contract query <abiFunctionSignature> ... --contract <contract> --chain ARC-TESTNET`.

These are official-doc and CLI help command shapes only. They are not authenticated local validation results unless this document records the exact command output, wallet address, tx hash or Circle transaction id, receipt status, and onchain readback.

Arc:
- Arc MCP is documentation lookup tooling only.
- Arc MCP is not the runtime trading agent.
- Arc Agentic Economy is relevant to the long-term SignalArc design.

Unknown / not documented:
- Whether SignalArc should store Circle policy IDs, detailed spending caps, or Circle-side wallet policy state in the backend before a real Circle Agent Wallet is authenticated and inspected.
- Whether Circle Agent Wallet contract execution for this exact SignalArc contract call shape succeeds on this local account before CLI login and OTP authentication.

### Checkpoint 2 - Circle Agent Wallet POC Discovery

Local discovery result:
- `circle` CLI was installed globally with `npm install -g @circle-fin/cli` because it is an external/global tool and not a SignalArc project dependency.
- `circle --version` returned `0.0.3`.
- Safe help discovery was run for `circle wallet`, `wallet login`, `wallet list`, `wallet create`, `wallet fund`, `wallet balance`, `wallet execute`, and `contract query`.
- Additional safe help discovery was run for `circle --help` and `circle terms --help`.
- Safe auth/chain discovery commands were blocked before login by Circle CLI Terms acceptance:
  - `circle wallet status --type agent --output json`
  - `circle blockchain list -q`
  - `circle blockchain list --output json`
- Circle CLI testnet authentication could not be tested because accepting Terms of Use and Privacy Policy is a user-controlled manual step.
- No Circle Agent Wallet address was created or listed.
- No Circle Agent Wallet was funded.
- No Circle Agent Wallet contract execution was performed.

Manual setup commands from official docs:

```bash
npm install -g @circle-fin/cli
circle wallet login you@example.com --testnet
circle wallet list --type agent --chain ARC-TESTNET
circle wallet fund --address 0xYourWalletAddress --chain ARC-TESTNET
circle wallet balance --address 0xYourWalletAddress --chain ARC-TESTNET
circle wallet execute "createMarket(string,string,uint256,address,address)" <marketId> <question> <closeTimestamp> <resolver> <collateralToken> --address 0xAgentWallet --contract 0x69aE770e8b2F96297101FeC4dc123B3801dA7d80 --chain ARC-TESTNET
circle contract query "marketCount()" --contract 0x69aE770e8b2F96297101FeC4dc123B3801dA7d80 --chain ARC-TESTNET
```

If OTP/manual login is required, stop and let the user complete it. Do not bypass user control.

Current manual blocker:

```bash
# User-controlled step; do not run automatically.
circle terms accept
circle wallet login you@example.com --testnet
```

Circle CLI help says login emails an OTP. The user must accept Circle Terms/Privacy Policy and complete OTP manually before SignalArc can list, fund, or execute from an agent wallet.

Note: official Circle docs show `circle wallet login you@example.com --testnet`; the installed `circle wallet login --help` output describes separate mainnet/testnet sessions but did not list a `--testnet` option. Treat the exact testnet login flag as blocked until the user completes the CLI Terms step and the command can be checked interactively.

### Backend Agent Wallet Model

Persistent backend model added for:
- `agent_id`
- `user_wallet`
- `user_email`
- `agent_wallet_address`
- `wallet_provider`
- `chain`
- `allowed_actions`
- `status`
- `policy_metadata`
- `source_client`
- `created_at`
- `updated_at`

Current implementation uses the DB-backed `agent_wallets` table for production routing. Test code uses an in-memory registry with the same interface for handler coverage only.

Multi-tenant onboarding/session foundation:
- `agent_onboarding_sessions` stores pending per-user, per-agent onboarding state before final wallet registration.
- `agent_sessions` stores activated per-agent runtime ownership/session boundaries without Circle session secrets.
- State is keyed by explicit user email, user wallet, agent id, source client, and channel so external clients do not imply one shared global Circle Agent Wallet session.
- Circle Agent Wallet OTP start skeleton is added behind `CIRCLE_AGENT_ONBOARDING_OTP_START_ENABLED=false` by default.
- When enabled for controlled host-shell/dev runtime, the OTP start skeleton uses a Circle onboarding runner abstraction, stores only `circle_request_id_hash` and expiry, and keeps any raw request ID in local process memory only until restart.
- Circle Agent Wallet OTP verify skeleton is added behind the same disabled-by-default onboarding OTP flag.
- Verify uses the in-memory request ID from start, consumes it on success, and updates onboarding status to `verified`.
- Verify uses the documented Circle completion command shape: `circle wallet login --request <request-id> --otp <code>`.
- Verify failures now emit sanitized server-side diagnostics for Circle CLI output while redacting the raw request ID and OTP; API responses remain generic.
- Backend restart before verify requires onboarding restart because raw request IDs are not stored in the database.
- Agent wallet resolution, wallet provisioning readback, `agent_sessions` creation, and Circle session persistence are not implemented.

No private keys, Circle OTPs, secret tokens, or undocumented Circle session material are stored.

### Agent Wallet Onboarding API

`POST /agent/onboarding/start`
- Creates a pending onboarding session only.
- Requires `agent_id` and `user_email`.
- Does not require `user_wallet`; user wallets belong to frontend wallet-connect and user transaction flows, not initial agent onboarding.
- Defaults to `chain == ARC-TESTNET`, `wallet_provider == circle_agent_wallet`, and `status == pending_otp`.
- With OTP start disabled, returns `circle_otp_verification_not_implemented` and does not call Circle CLI.
- With OTP start explicitly enabled, calls the Circle onboarding runner and returns `circle_otp_required` with expiry and non-secret request reference.
- Does not expose or store raw OTP or raw request ID.

`POST /agent/onboarding/verify`
- Disabled by default unless onboarding OTP start is enabled.
- Completes the OTP skeleton using the in-memory request ID from `POST /agent/onboarding/start`.
- Returns `agent_wallet_resolution_not_implemented` after marking onboarding `verified`.
- Does not expose or store raw OTP or raw request ID.
- Does not create an agent session or resolve an agent wallet address yet.

Custom GPT schema note:
- `startAgentOnboarding` should not require `user_wallet`.

`GET /agent/onboarding/{onboarding_id}`
- Returns pending onboarding status.
- Does not return Circle secrets.

`GET /agent/sessions/{agent_id}`
- Returns the known activated agent-session boundary if one exists.
- Does not return Circle secrets.

`POST /agent/onboarding/register`
- Registry-only convenience endpoint for final agent wallet mapping with default Arc Testnet Circle provider policy.
- Does not create a pending OTP session and does not authenticate Circle.

`POST /agent/wallets`
- Registers an already-created user-owned Circle Agent Wallet with SignalArc.
- Does not create the Circle Agent Wallet.
- Does not authenticate Circle.
- Does not store secrets.
- Rejects deployer/resolver wallet reuse.
- Rejects agent wallet reuse of `user_wallet` until a documented custody-link model exists.
- Requires `wallet_provider == circle_agent_wallet`.
- Requires `chain == ARC-TESTNET`.
- Requires non-empty `allowed_actions`.

`GET /agent/wallets/{agent_id}`
- Returns registered wallet metadata only.
- Does not return secrets.

`POST /agent/wallets/{agent_id}/disable`
- Marks a registered agent wallet disabled.
- Disabled wallets cannot execute.

Onboarding flow:
1. User controls email and Circle Agent Wallet setup.
2. User completes Circle CLI login/OTP outside SignalArc backend.
3. User obtains Circle Agent Wallet address on `ARC-TESTNET`.
4. Client registers wallet metadata with SignalArc through `POST /agent/wallets`.
5. Any external client can submit intents by `agent_id`.
6. Backend validates wallet status, chain, provider, action allowlist, and wallet boundary before execution readiness.

### Backend Execution Guardrails

`POST /agent/intents` now accepts:
- `agent_id`
- `agent_wallet_address`
- `source_client`
- `client_request_id`
- existing action, market, and amount fields

`POST /agent/intents/{id}/execute` rejects execution when:
- `agent_id` has no registered agent wallet.
- agent wallet is inactive.
- agent wallet chain is not `ARC-TESTNET`.
- action is not in `allowed_actions`.
- agent wallet equals the deployer/resolver wallet.
- agent wallet equals the registered user wallet.
- agent wallet matches `SIGNALARC_FORBIDDEN_AGENT_WALLETS`.
- request `agent_wallet_address` does not match the registered agent wallet.
- provider is `circle_agent_wallet` but Circle CLI authentication and live wallet proof are not complete.
- provider is `temporary_testnet_agent_eoa`; this is only a documented fallback name and is not enabled.

### Channel-Agnostic Client Contract

External clients can include:
- WhatsApp agent
- Telegram bot
- ChatGPT custom action
- Claude tool
- Web client

Client responsibilities:
- Call SignalArc Backend Agent API.
- Reference a registered `agent_id`.
- Provide `source_client` and optional `client_request_id` for auditability.
- Never hold private keys.
- Never receive OTPs or Circle session tokens from SignalArc.
- Never execute contracts directly.

Backend responsibilities:
- Maintain the user/agent wallet registry.
- Maintain pending onboarding sessions and activated agent-session boundaries without storing Circle secrets.
- Validate intent shape and registered wallet state.
- Enforce provider, chain, status, and action allowlist checks.
- Return an execution plan or execution readiness failure.
- Keep Circle Agent Wallet execution fail-closed until a safe server-side Circle auth/session strategy is designed and proven.

Chosen MVP client remains: ChatGPT custom action / API-call style client.

ChatGPT Custom GPT live external agent validation is complete for `getHealth`, create intent preview, confirm intent, and `executeAgentIntent` for `create_market` through the backend-approved path. WA, Telegram, and Claude live client triggers remain untested.

Temporary Custom GPT test path:
- ChatGPT Custom GPT Action -> ngrok HTTPS -> `localhost:4001` SignalArc backend -> Circle Agent Wallet CLI provider -> Arc Testnet.
- Authentication was set to None only for temporary local tunnel testing. This is not production API authentication.
- The ngrok tunnel is temporary and is not a production API endpoint.

### Backend Circle Provider Decision

Do not automate Circle CLI execution in backend by relying on a user's local interactive CLI session in production.

Current backend behavior:
- `circle_agent_wallet` execution uses a guarded Circle CLI adapter only when explicitly enabled.
- Registered wallet metadata and intent validation are DB-backed.
- Onchain Circle Agent Wallet transactions are proven through the backend Circle CLI provider in a host-shell runtime with an authenticated Circle Agent Wallet session.
- By default, `CIRCLE_AGENT_WALLET_EXECUTION_ENABLED=false`, so execution fails closed.
- The backend never calls Circle login.
- The backend never accepts OTP.
- The backend never stores or logs Circle session files, OTPs, Circle tokens, private keys, or deployer keys.

Backend Circle CLI Provider:
- Provider: `circle_agent_wallet`
- Execution mode: `circle_agent_wallet_cli`
- Chain: `ARC-TESTNET`
- CLI path config: `CIRCLE_CLI_PATH`, default `circle`
- Enable config: `CIRCLE_AGENT_WALLET_EXECUTION_ENABLED`, default `false`
- Chain config: `CIRCLE_AGENT_WALLET_CHAIN`, default `ARC-TESTNET`
- Timeout config: `CIRCLE_AGENT_WALLET_TIMEOUT_SECONDS`, default `120`

Supported backend Circle CLI actions in this MVP:
- `create_market`
- `buy_yes`
- `buy_no`
- `close_market`
- `resolve_market`
- `claim_payout`
- `cancel_market`
- `claim_refund`

Circle CLI command mapping:
- `create_market`: `circle wallet execute "createMarket(string,string,uint256,address,address)" ... --address <agent_wallet> --contract <factory> --chain ARC-TESTNET --output json`
- `buy_yes`: `approve(address,uint256)` against Arc Testnet USDC, then `buyYes(uint256)` against the market contract.
- `buy_no`: `approve(address,uint256)` against Arc Testnet USDC, then `buyNo(uint256)` against the market contract.
- `close_market`: `closeMarket()` against the market contract.
- `resolve_market`: `resolve(uint8)` against the market contract, where the repo contract enum maps `Yes` to `1` and `No` to `2`.
- `claim_payout`: `claimPayout()` against the market contract.
- `cancel_market`: `cancelMarket()` against the market contract.
- `claim_refund`: `claimRefund()` against the market contract.

Readback mapping:
- After `create_market`: `marketCount()`, `allMarkets(lastIndex)`, `isMarket(created_market)`.
- After `buy_yes`: `yesPositions(agent_wallet)`, `totalYes()`, `totalCollateral()`, `USDC.balanceOf(market)`.
- After `buy_no`: `noPositions(agent_wallet)`, `totalNo()`, `totalCollateral()`, `USDC.balanceOf(market)`.
- After `close_market`: `status()`, `isOpen()`.
- After `resolve_market`: `status()`, `winningOutcome()`, `claimablePayout(agent_wallet)`, `hasClaimed(agent_wallet)`, `USDC.balanceOf(market)`.
- After `claim_payout`: `status()`, `winningOutcome()`, `claimablePayout(agent_wallet)`, `hasClaimed(agent_wallet)`, `USDC.balanceOf(market)`.
- After `cancel_market`: `status()`, `claimableRefund(agent_wallet)`, `hasClaimed(agent_wallet)`, `USDC.balanceOf(market)`.
- After `claim_refund`: `status()`, `claimableRefund(agent_wallet)`, `hasClaimed(agent_wallet)`, `USDC.balanceOf(market)`.

Runtime requirement:
- The backend runtime must have Circle CLI installed and an operator-authenticated Circle Agent Wallet session available.
- The validated backend provider run used the host shell backend on port `4001`, not the Docker backend container, because the Docker backend container does not have Circle CLI installed.
- If official docs do not document a non-CLI server API for this exact execution path, production auth/session strategy remains a separate checkpoint.
- This remains a testnet/dev execution path until production auth, policy, audit, and lifecycle behavior are designed and proven.

### Current Non-Claims

Not complete:
- WhatsApp, Telegram, and Claude live client triggers have not been tested yet.
- Docker/Cloud Run runtime does not yet include a Circle CLI/session strategy.
- SignalArc is not mainnet ready.
- SignalArc makes no production readiness claim.
- SignalArc makes no Circle policy limit claim on `ARC-TESTNET`.
- SignalArc is not claiming autonomous trading.
- The temporary ngrok tunnel is not a production API endpoint.

### User-Proven Circle Agent Wallet Evidence

Status: BACKEND PROVIDER FULL MARKET LIFECYCLE VALIDATED ON ARC TESTNET.

This evidence supersedes the earlier deployer-signed smoke tests and the manual Circle CLI-only checkpoint for create-market, buy-position, payout, cancel, and refund lifecycle portions of the Live AI Agent Transaction MVP. Later ChatGPT Custom GPT evidence completes the ChatGPT external-client trigger portion for health, intent preview, confirm, and real create_market execution only. It does not complete WA/Telegram/Claude external-client validation, Docker/Cloud Run runtime strategy, production readiness, mainnet readiness, or Circle policy-limit validation.

Runtime setup:
- Backend was run from the host shell on port `4001`, not Docker.
- Reason: Docker backend container does not have Circle CLI installed.
- Host Circle Agent Wallet session was valid.
- DB migration was advanced to version `15`, `dirty=false`.
- Agent wallet was registered through DB-backed `POST /agent/wallets`.
- Latest full lifecycle validation also used host shell runtime with `APP_PORT=4001`.
- Execution mode: `circle_agent_wallet_cli`.

Agent wallet:
- Agent ID: `agent_desi_001`
- User email: `desi33905@gmail.com`
- User wallet: `0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC`
- Chain: `ARC-TESTNET`
- Wallet provider: `circle_agent_wallet`
- Circle CLI version observed: `0.0.3`
- Agent wallet address: `0x96d5051a005547eba149f71604ccf58ae1a7c950`
- Deployer/resolver wallet: `0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC`
- Boundary proof: agent wallet is different from deployer/resolver wallet.
- Allowed actions: `create_market`, `buy_yes`, `buy_no`
- Source client: `manual_backend_runtime`

Factory:
- `SignalArcAgentMarketFactory`: `0x69aE770e8b2F96297101FeC4dc123B3801dA7d80`

Backend provider `create_market`:
- Intent ID: `agent_intent_f76bd9653e9ce3f6269023a25e7c6b8c`
- Execution mode: `circle_agent_wallet_cli`
- Transaction: `0x7aa51a0d19b163a3a88ae16ac4a88a1cdbb3090cad5d0ccc54f828b166f74e5d`
- Created market: `0x38aE7E0133e9594F597F913884cbDa619A950523`
- Post-create factory readback:
  - `market_count == 9`
  - `is_market == true`

Backend provider `buy_yes`:
- Intent ID: `agent_intent_d6f9cd78896886f954a05a164c17c067`
- Execution mode: `circle_agent_wallet_cli`
- Market: `0x38aE7E0133e9594F597F913884cbDa619A950523`
- USDC approve transaction: `0x09d0a418c34b0e54a31bc3a2a7bfba85218eba27e84becdcbd89e5b63b8bb387`
- `buyYes` transaction: `0xa1fadb400aa8b4babca0c936698e686eeaac3ae408b22d1e37960901a5c48ade`
- Readback:
  - `yes_positions == 1000000`
  - `total_yes == 1000000`
  - `total_collateral == 1000000`
  - `USDC.balanceOf(market) == 1000000`

Backend provider `buy_no`:
- Intent ID: `agent_intent_ed7050305f62798d8472b7f48e538ff8`
- Execution mode: `circle_agent_wallet_cli`
- Market: `0x38aE7E0133e9594F597F913884cbDa619A950523`
- USDC approve transaction: `0x40ba807e24e2fcfeba22f21575920d3dfe5389f7f00320ab5a74fb46f06c6dc8`
- `buyNo` transaction: `0xb36288fc40a69765d62679982fdd3319d09b27c010e2eb9caafa9c6508d03e9c`
- Readback:
  - `no_positions == 1000000`
  - `total_no == 1000000`
  - `total_collateral == 2000000`
  - `USDC.balanceOf(market) == 2000000`

Backend provider resolved payout lifecycle:
- `create_market`
  - Intent ID: `agent_intent_ecc88160f7e2b908fc498c3dff66fbe7`
  - Execution mode: `circle_agent_wallet_cli`
  - Transaction: `0x7913fd51b38b147cfc6936da7eb7166a97a156351c9ccce4264d48d31fc91ae9`
  - Created market: `0x38D4317fcB0C82e5EC2407a89c311b3Be8059CD0`
  - Readback: `market_count == 10`, `is_market == true`
- `buy_yes`
  - Intent ID: `agent_intent_e1d3889f3252ac318b7b435b2a789d9d`
  - USDC approve transaction: `0x9c66332ad2d798126118b961ad9005ab7f2055649b46d808cc979ccf40eee3f7`
  - `buyYes` transaction: `0xdd4448fdf237f13bc9e90737f27b7a1e912ee8d34195bae311cc2bac15aaa95d`
  - Readback: `yes_positions == 1000000`, `total_yes == 1000000`, `total_collateral == 1000000`, `USDC.balanceOf(market) == 1000000`
- `close_market`
  - Intent ID: `agent_intent_6cab73b950296ac2dcc3a5414f4e1613`
  - Transaction: `0x5accb4dafee2a27be032709e427e25eede4e7eb67f54a06bdf0ba82e5f4a013e`
  - Readback: `market_status == 1`, `is_open == false`
- `resolve_market`
  - Intent ID: `agent_intent_4073e0f29f842a0110a1e099d9fa50b0`
  - Transaction: `0xc45d8612768591736425a743520b5e432b9a424ef9552cfbbc1bb04d785c874b`
  - Readback: `market_status == 2`, `winning_outcome == 1`, `claimable_payout == 1000000`, `has_claimed == false`, `USDC.balanceOf(market) == 1000000`
- `claim_payout`
  - Intent ID: `agent_intent_65284f22e52fce32d6a3efc2fa6163cd`
  - Transaction: `0xc80bb9dd7f6924c93c1722d7c5f1136c076403d052ed792f98ed2d7abd59568f`
  - Readback: `market_status == 2`, `winning_outcome == 1`, `claimable_payout == 1000000`, `has_claimed == true`, `USDC.balanceOf(market) == 0`

Backend provider refund lifecycle:
- `create_market`
  - Intent ID: `agent_intent_b167e466b198c7f909c388c76b683f5d`
  - Transaction: `0x1340817f922aaa7ae181789bbaf2b7bff13a426b0397524150d01ae869d2a033`
  - Created market: `0xbfd93169DAFf0610EA10E1221B9a2a6552379648`
  - Readback: `market_count == 11`, `is_market == true`
- `buy_yes`
  - Intent ID: `agent_intent_b06dd357e62778ed1a0527800690473a`
  - USDC approve transaction: `0x83bdc164512979296c385f51b4b6b1df51c741f2157cc93944b7f15c1328f487`
  - `buyYes` transaction: `0xa85c60d903a1b1e30f39da95a37ca356d61d11cc20c942aaa22f0740a06945ab`
  - Readback: `yes_positions == 1000000`, `total_yes == 1000000`, `total_collateral == 1000000`, `USDC.balanceOf(market) == 1000000`
- First backend provider `cancel_market` attempt failed.
  - Intent ID: `agent_intent_634dad5fc81b7a155e3854d8e14bb135`
  - Backend response: `502 agent_execution_failed`
  - Onchain check after failure showed the market still Open.
  - This is recorded as a failed backend attempt, not validated backend cancel evidence.
- Manual Circle CLI `cancelMarket` succeeded on the refund market.
  - Transaction: `0xbf8b98862ed691c0023643ab72ee71d8422e434868b2806f3390b4ffc88fe21b`
  - This was manual Circle CLI, not backend provider evidence.
- Backend provider `claim_refund` succeeded after manual cancel.
  - Intent ID: `agent_intent_c519c98114f8a5c113ab4bc6dfefbcae`
  - Transaction: `0x40f3f4e1737340dbbbff92e3020d0cfa6dbd7d7bbca8a1ecb580b1c0cdc43dfd`
  - Readback: `market_status == 3`, `claimable_refund == 1000000`, `has_claimed == true`, `USDC.balanceOf(market) == 0`

Backend provider cancel-only validation:
- `create_market`
  - Intent ID: `agent_intent_666c781a7a3fe2f939bd5342e331fd67`
  - Transaction: `0x62f5f43b834a09f4f4c78e3bb365403a2b76374d0030a91ccd3e3571fd3c9d12`
  - Created market: `0x928F3F9Cb43811837C0e8D4FA40c24A4f083B3Ed`
  - Readback: `market_count == 12`, `is_market == true`
- `cancel_market`
  - Intent ID: `agent_intent_34f54290e5f59ffaff4a167986868c56`
  - Transaction: `0x3eae0d0508397e5ea515d417bdc5be5c38f40f3b0296b1cf424d99060cb92de4`
  - Readback: `market_status == 3`, `claimable_refund == 0`, `has_claimed == false`, `USDC.balanceOf(market) == 0`

Current proven status:
- Backend Circle provider `create_market` validated with a real Arc Testnet transaction.
- Backend Circle provider `buy_yes` validated with a real Arc Testnet transaction.
- Backend Circle provider `buy_no` validated with a real Arc Testnet transaction.
- Backend Circle provider `close_market` validated with a real Arc Testnet transaction.
- Backend Circle provider `resolve_market` validated with a real Arc Testnet transaction.
- Backend Circle provider `claim_payout` validated with a real Arc Testnet transaction.
- Backend Circle provider `cancel_market` validated after one failed backend provider attempt on the refund market and a later successful backend-only cancel-only test.
- Backend Circle provider `claim_refund` validated after manual Circle CLI cancellation of the refund market.
- Backend-to-Circle-to-Arc path is proven for create, trade, close, resolve, payout claim, cancel, and refund claim flows on Arc Testnet.

### ChatGPT Custom GPT External Client Trigger Evidence

Status: VALIDATED FOR HEALTH, INTENT PREVIEW, CONFIRM, AND CREATE_MARKET EXECUTION ON ARC TESTNET.

This was ChatGPT Custom GPT via a temporary ngrok HTTPS tunnel to the local backend. It is not a production API endpoint and does not prove production authentication.

Connectivity:
- External client: ChatGPT Custom GPT Action.
- Backend runtime: host shell on `APP_PORT=4001`.
- Execution mode: `circle_agent_wallet_cli`.
- Public test path: ChatGPT Custom GPT Action -> ngrok HTTPS -> `localhost:4001` SignalArc backend -> Circle Agent Wallet CLI provider -> Arc Testnet.
- Local backend health returned 200: `http://127.0.0.1:4001/health` -> `{"status":"ok"}`.
- Ngrok public health returned 200: `https://undamaged-commerce-juggling.ngrok-free.dev/health` -> `{"status":"ok"}`.
- Custom GPT `getHealth` action returned `{"status":"ok"}`.

Custom GPT Action schema:
- Operations exposed:
  - `getHealth`
  - `createAgentIntent`
  - `confirmAgentIntent`
  - `executeAgentIntent`
- Authentication was set to None only for temporary local tunnel testing.
- This is not production API authentication.

Intent preview and failure evidence:
- Initial preview and confirm were tested.
- One confirm failed due to a typo in a manually copied intent ID. This was a user/operator copy error, not a backend failure.
- A later execution attempt failed because `close_timestamp` was already in the past:
  - `close_timestamp` used: `1779439999`
  - current timestamp checked later: `1779452713`
  - probable root cause: `createMarket` reverted due to `closeTimestamp <= block.timestamp`
  - Circle Agent Wallet session was checked and testnet `tokenStatus` was `VALID`
  - This is stale input evidence, not a proven provider failure.
- New valid preview used future `close_timestamp`: `1779456313`.

Successful Custom GPT external execution:
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

Current external-client trigger status:
- ChatGPT Custom GPT trigger validated for health, create intent preview, confirm intent, and execute `create_market` real Arc Testnet transaction.
- WA, Telegram, and Claude triggers are still not tested.

## Core Rule

Backend remains source of truth.

Agent must not bypass backend.

## Contract Separation

Current user/frontend contracts stay untouched:

SignalArcMarketFactory -> SignalArcMarket

Future agent contracts:

SignalArcAgentMarketFactory -> SignalArcAgentMarket

Planned files:
- contracts/src/agent/SignalArcAgentMarket.sol
- contracts/src/agent/SignalArcAgentMarketFactory.sol

Rules:
- Do not modify current live contracts.
- Do not reuse current factory address.
- Deploy new agent factory later.
- Agent contracts stay testnet-only until validated.
- Add Foundry tests before deployment.

## Planned Backend Agent API

- POST /agent/intents
- POST /agent/intents/{id}/confirm
- GET /agent/intents/{id}
- GET /agent/activity
- GET /agent/markets
- GET /agent/positions
- GET /agent/claimable

## Planned Actions

- create_market
- buy_yes
- buy_no
- cancel_market
- close_market
- resolve_market
- claim_refund
- claim_payout

## MCP Boundary

Arc MCP is only for developer tooling and official docs lookup.

Arc MCP is not the runtime trading agent.

## Circle Agent Wallet Boundary

Circle Agent Wallet is integrated behind a guarded, disabled-by-default backend Circle CLI provider for Arc Testnet. Docker/Cloud Run session strategy, production policy controls, and mainnet readiness are not complete.

Before implementation:
- verify supported blockchains
- verify Arc Testnet support
- verify SDK/CLI path
- verify faucet/funding path

If not documented, mark as unknown / not documented.

## Implementation Order

1. Separate agent contract prototype — DONE
2. Foundry tests — DONE
3. Agent factory deploy on Arc Testnet — DONE
4. Backend Agent API intent model — DONE
5. Backend execution path to agent contract - DONE
6. Circle Agent Wallet proof of concept - DONE for full backend provider lifecycle on Arc Testnet
7. Backend Circle Agent Wallet lifecycle executor - DONE for host-shell Arc Testnet runtime validation
8. Policy-limited agent wallet execution - PENDING



## Agent Factory Deployment

Status: DONE.

- Network: Arc Testnet
- Agent factory: `0x69aE770e8b2F96297101FeC4dc123B3801dA7d80`
- Deploy transaction: `0x8e4dfa481a2863a08a749fe4add30c4f030f178a4d6ba4658419df6730ebf10b`
- Read validation: `marketCount() == 0`


## Real Arc Testnet Agent Market Smoke

Status: DONE.

This is a real Arc Testnet transaction, not a mock or local simulation.

- Action: `SignalArcAgentMarketFactory.createMarket`
- Factory: `0x69aE770e8b2F96297101FeC4dc123B3801dA7d80`
- Transaction: `0xff4d6eb644792a1c064992704ba767b6712b7cc02c1b44635859e199efdfc69d`
- Created agent market: `0x4e26143A63457cf06A34112b8B9044F3760d3007`
- Read validation:
  - `marketCount() == 1`
  - `allMarkets(0) == 0x4e26143A63457cf06A34112b8B9044F3760d3007`
  - `isMarket(0x4e26143A63457cf06A34112b8B9044F3760d3007) == true`

Note: this validates real factory lifecycle only. Trading validation still requires a valid Arc Testnet collateral token.

## Real Arc Testnet Agent USDC Collateral / Trading Smoke

Status: DONE.

This is a real Arc Testnet transaction sequence, not a mock, fake smoke, dry run, or local simulation.

Official docs finding:
- Arc docs document Arc Testnet USDC ERC-20 interface: `0x3600000000000000000000000000000000000000`.
- Circle docs document Arc Testnet USDC token address: `0x3600000000000000000000000000000000000000`.
- Token readback on Arc Testnet:
  - `name() == "USDC"`
  - `symbol() == "USDC"`
  - `decimals() == 6`

Previous agent market collateral read:
- Market: `0x4e26143A63457cf06A34112b8B9044F3760d3007`
- `collateralToken() == 0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC`
- Result: not usable as ERC-20 collateral; this is the deployer/resolver EOA, not the documented Arc Testnet USDC contract.

USDC-backed agent market:
- Action: `SignalArcAgentMarketFactory.createMarket`
- Factory: `0x69aE770e8b2F96297101FeC4dc123B3801dA7d80`
- Transaction: `0x9e1dc4b2b65ea6220605f8960dd13ad1be3907b69b72560083e39d3f2c77f579`
- Created agent market: `0xd76c5633c3D8C1761F7edae46506B44cDeEe43a7`
- Collateral token: `0x3600000000000000000000000000000000000000`
- Read validation:
  - `marketCount() == 2`
  - `collateralToken() == 0x3600000000000000000000000000000000000000`
  - `admin() == 0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC`
  - `resolver() == 0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC`
  - `isOpen() == true`

Real buyYes / buyNo validation:
- USDC approve transaction: `0xc5963b4ad20b8c66dfb7787c260fcf04deb25fcf53667349958109a10f7584c9`
- `buyYes(1000000)` transaction: `0x1a23107407b17a4dced81d8fd79c2ead8acb8148190a03aaae2f34662621e79f`
- `buyNo(1000000)` transaction: `0x41e0a39271940265894c45893733f794936a6e49cf3f475ae8881d5f9a7bb073`
- Read validation:
  - `yesPositions(0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC) == 1000000`
  - `noPositions(0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC) == 1000000`
  - `totalYes() == 1000000`
  - `totalNo() == 1000000`
  - `totalCollateral() == 2000000`
  - `USDC.balanceOf(0xd76c5633c3D8C1761F7edae46506B44cDeEe43a7) == 2000000`
  - `USDC.allowance(deployer, market) == 0`

Real cancel / refund lifecycle validation:
- Pre-cancel read validation:
  - `status() == 0`
  - `claimableRefund(0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC) == 0`
  - `USDC.balanceOf(0xd76c5633c3D8C1761F7edae46506B44cDeEe43a7) == 2000000`
  - `USDC.balanceOf(0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC) == 32622253`
- `cancelMarket()` transaction: `0x93df5450fcfc50054b7cbc4f260fafc876a29d339d3923807cdb3d4f6323274d`
- Cancel receipt status: `1 (success)`
- Post-cancel read validation:
  - `status() == 3`
  - `claimableRefund(0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC) == 2000000`
  - `USDC.balanceOf(0xd76c5633c3D8C1761F7edae46506B44cDeEe43a7) == 2000000`
- `claimRefund()` transaction: `0x54bb3939995f613212531e45345d42c31d57e0ce61eccb26eb5d526921ce4453`
- Claim refund receipt status: `1 (success)`
- Post-refund read validation:
  - `hasClaimed(0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC) == true`
  - `USDC.balanceOf(0xd76c5633c3D8C1761F7edae46506B44cDeEe43a7) == 0`
  - `USDC.balanceOf(0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC) == 34619899`
  - `claimableRefund(0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC) == 2000000`
- Note: `claimableRefund` reports the cancelled position amount and does not check `hasClaimed`; refund completion is proven by `hasClaimed == true`, market USDC balance `0`, and the successful USDC transfer in the claim transaction.

## Real Arc Testnet Agent YES Payout Lifecycle Smoke

Status: DONE.

This is a real Arc Testnet transaction sequence, not a mock, fake smoke, dry run, or local simulation.

Fresh USDC-backed payout market:
- Action: `SignalArcAgentMarketFactory.createMarket`
- Factory: `0x69aE770e8b2F96297101FeC4dc123B3801dA7d80`
- Transaction: `0x4ac895622bc802ac6639095707675269d5b3de8a08e60991fe8a002c794aa75d`
- Created agent market: `0xcCE012A74Cdf7d17138cd6A514394c79b092B6E7`
- Close timestamp: `1779357303`
- Collateral token: `0x3600000000000000000000000000000000000000`
- Read validation:
  - `marketCount() == 4`
  - `allMarkets(3) == 0xcCE012A74Cdf7d17138cd6A514394c79b092B6E7`
  - `collateralToken() == 0x3600000000000000000000000000000000000000`
  - `status() == 0`
  - `isOpen() == true`

Real YES position:
- USDC approve transaction: `0x73d9ac7f00cda86c02acc9d8082dafd8432874818696fc7588bb705305ce31ad`
- `buyYes(1000000)` transaction: `0x584d9d762b4e5279e6db475ffe2a0ab43b0c06220c86ed799ba7e0dbc8d15311`
- Read validation:
  - `yesPositions(0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC) == 1000000`
  - `totalYes() == 1000000`
  - `totalCollateral() == 1000000`
  - `USDC.balanceOf(0xcCE012A74Cdf7d17138cd6A514394c79b092B6E7) == 1000000`

Real close / resolve / payout lifecycle validation:
- Wait validation:
  - latest block timestamp before close: `1779357317`
  - close timestamp: `1779357303`
- `closeMarket()` transaction: `0x0fb4194bbb9b75097180bcae078ecaf4ce20e3c9425a75de7358ac2d2f8f34e4`
- Close receipt status: `1 (success)`
- `resolve(YES)` transaction: `0x961153bc68239ffcdfccfe7472e7498305d98d1765fe90fb3b4a7013f8ba7afb`
- Resolve receipt status: `1 (success)`
- Pre-claim read validation:
  - `status() == 2`
  - `winningOutcome() == 1`
  - `claimablePayout(0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC) == 1000000`
  - `USDC.balanceOf(0xcCE012A74Cdf7d17138cd6A514394c79b092B6E7) == 1000000`
- `claimPayout()` transaction: `0xca6e837d455a43e99136a6f2c50bdeb2ba76aef41c5235d5d9994baadc83b631`
- Claim payout receipt status: `1 (success)`
- Post-payout read validation:
  - `status() == 2`
  - `winningOutcome() == 1`
  - `hasClaimed(0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC) == true`
  - `USDC.balanceOf(0xcCE012A74Cdf7d17138cd6A514394c79b092B6E7) == 0`
  - `USDC.balanceOf(0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC) == 34570580`
  - `claimablePayout(0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC) == 1000000`
  - `USDC.allowance(deployer, market) == 0`
- Note: `claimablePayout` reports the winning position amount and does not check `hasClaimed`; payout completion is proven by `hasClaimed == true`, market USDC balance `0`, and the successful USDC transfer in the claim transaction.

## Backend Agent create_market Execution Endpoint

Status: DONE.

This validates the backend `POST /agent/intents/{id}/execute` path for `action=create_market` with a real Arc Testnet transaction, not a mock, fake smoke, dry run, or local simulation.

Implementation boundary:
- Implemented action: `create_market`
- Other actions: return `501 not_implemented`
- Runtime env names:
  - `ARC_TESTNET_RPC_URL`
  - `AGENT_EXECUTOR_PRIVATE_KEY`
  - `AGENT_FACTORY_ADDRESS`
- Private key is read from env only and is not logged or returned.

Backend local validation:
- Backend route: `POST /agent/intents/{id}/execute`
- Local validation port: `4107`
- Intent id: `agent_intent_4e6c6445c3d542cd6b6cf2dff24611d9`
- Market id: `backend-create-market-1779358755`
- Question: `Will SignalArc backend execute create_market on Arc Testnet?`
- Close timestamp: `1779445152`
- Resolver: `0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC`
- Collateral token: `0x3600000000000000000000000000000000000000`

Execution result:
- Transaction: `0x0f89ff6e50e31769c0708d3f725c86c4612ba553f7e2bdab7721f8e6dd2674c4`
- Created agent market: `0x0d4aF15Bee1Caf6FB61F55668c2Cd8CB7a051e81`
- Backend response readback:
  - `status == executed`
  - `broadcast_performed == true`
  - `transaction_hash == 0x0f89ff6e50e31769c0708d3f725c86c4612ba553f7e2bdab7721f8e6dd2674c4`
  - `network == arc_testnet`
  - `readback.market_count == 5`
  - `readback.created_market == 0x0d4aF15Bee1Caf6FB61F55668c2Cd8CB7a051e81`
  - `readback.is_market == true`

Independent onchain readback:
- Receipt status: `1 (success)`
- Factory `marketCount() == 5`
- Factory `isMarket(0x0d4aF15Bee1Caf6FB61F55668c2Cd8CB7a051e81) == true`
- Created market `collateralToken() == 0x3600000000000000000000000000000000000000`

## Backend Agent buy_yes Execution Endpoint

Status: DONE.

This validates the backend `POST /agent/intents/{id}/execute` path for `action=buy_yes` with real Arc Testnet transactions, not a mock, fake smoke, dry run, or local simulation.

Implementation boundary:
- Implemented additional action: `buy_yes`
- Historical note: at this earlier phase, `buy_no`, `cancel_market`, `close_market`, `resolve_market`, `claim_refund`, and `claim_payout` were not implemented. Current status is recorded above.
- Runtime env names:
  - `ARC_TESTNET_RPC_URL`
  - `AGENT_EXECUTOR_PRIVATE_KEY`
  - `AGENT_FACTORY_ADDRESS`
- Private key is read from env only and is not logged or returned.
- Execution uses Arc Testnet USDC: `0x3600000000000000000000000000000000000000`

Fresh market for backend buy_yes validation:
- Created through backend `create_market` execution endpoint.
- Intent id: `agent_intent_3c7ccaed030100b61eb3818aec277883`
- Market id: `backend-buy-yes-market-1779359535`
- Create transaction: `0xb7acd12fe390c1f3cc59543938710cd549855d4636682f9a841fd16fa500f8f3`
- Created agent market: `0xCB8D34fFdA32a3b58f355b92c3e720deCCF1C437`
- Backend create readback:
  - `readback.market_count == 6`
  - `readback.created_market == 0xCB8D34fFdA32a3b58f355b92c3e720deCCF1C437`
  - `readback.is_market == true`

Backend buy_yes validation:
- Intent id: `agent_intent_296671404fb958240b433ea31c79f0b5`
- Market contract address: `0xCB8D34fFdA32a3b58f355b92c3e720deCCF1C437`
- Amount: `1000000`
- USDC approve transaction: `0x5b3668792b391a7ec5fa892168569c26cf690e939ee8afa333ce54396871c1aa`
- `buyYes(1000000)` transaction: `0xc5d455a6f31bee34c443045575c2dedc5e62dfa53576d52b05916206a5fa8825`
- Backend response readback:
  - `status == executed`
  - `action == buy_yes`
  - `network == arc_testnet`
  - `broadcast_performed == true`
  - `approve_transaction_hash == 0x5b3668792b391a7ec5fa892168569c26cf690e939ee8afa333ce54396871c1aa`
  - `transaction_hash == 0xc5d455a6f31bee34c443045575c2dedc5e62dfa53576d52b05916206a5fa8825`
  - `readback.yes_positions == 1000000`
  - `readback.total_yes == 1000000`
  - `readback.total_collateral == 1000000`
  - `readback.usdc_balance == 1000000`

Independent onchain readback:
- Approve receipt status: `1 (success)`
- Buy YES receipt status: `1 (success)`
- `yesPositions(0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC) == 1000000`
- `totalYes() == 1000000`
- `totalCollateral() == 1000000`
- `USDC.balanceOf(0xCB8D34fFdA32a3b58f355b92c3e720deCCF1C437) == 1000000`
- `USDC.allowance(deployer, market) == 0`

## Non-Claims

Not implemented yet:
- frontend chat UI
- Circle Agent Wallet
- autonomous trading
- mainnet settlement
- real funds
- smart contract audit
