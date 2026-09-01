# Contract Security Boundaries

## Purpose

This is a Phase 6.6 public-safe security boundary note for the local contract prototype before any Arc Testnet deployment review.

## Current Prototype Scope

- Local Solidity / Foundry prototype.
- Binary YES / NO market.
- Test-only MockUSDC.
- Manual resolver.
- Fixed 1:1 local claim/refund behavior.
- No production custody.
- No production settlement.
- No mainnet deployment.

## Explicit Non-Goals

- No Arc Testnet deployment in Phase 6.6.
- No Arc mainnet deployment.
- No private keys in repo.
- No .env committed.
- No Circle API keys.
- No Circle Wallets integration.
- No Circle Agent Wallets integration.
- No backend integration.
- No frontend integration.
- No production oracle.
- No AMM.
- No order book.
- No fees.
- No admin withdrawal.
- No production collateral custody claim.

## Known Prototype Limitations

- Payout is fixed 1:1 and not a production market payout model.
- MockUSDC is test-only.
- Resolver is a single address.
- No dispute process.
- No oracle integration.
- No liquidity model.
- No fee model.
- No secondary trading.
- No production access control system.
- No audit has been performed.

## Required Before Arc Testnet Deployment

- All local tests must pass.
- Deployment must be manual by the repo owner.
- Private key must remain local-only.
- RPC URL must remain local-only.
- .env must not be committed.
- Deployment command must be reviewed without secrets.
- Only non-secret outputs may be shared:
  - Contract address.
  - Transaction hash.
  - Explorer link.
  - Success/failure status.

## Required Before Production

- Professional smart contract audit or equivalent review.
- Complete threat model.
- Production custody design.
- Oracle/resolution policy.
- Dispute/cancellation policy.
- Access-control review.
- Decimal handling review for Arc USDC.
- Integration tests.
- Deployment runbook.
- Monitoring and incident response plan.

## Current Decision

- Phase 6.6 does not approve production deployment.
- Phase 6.6 only prepares the prototype for later manual Arc Testnet deployment review.
