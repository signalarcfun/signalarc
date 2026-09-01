# Frontend Wallet Integration

The current frontend working tree includes external wallet integration for Arc Testnet using RainbowKit, Wagmi, and Viem.

Planned public frontend:

```text
https://signalarc.fun
```

The planned frontend domain is not live yet. DNS is not configured.

## Wallet Stack

| Layer | Implementation |
| --- | --- |
| Wallet UI | RainbowKit |
| Wallet state/actions | Wagmi |
| EVM utilities | Viem |
| Config file | `apps/web/src/lib/wagmi.ts` |
| Contract constants | `apps/web/src/lib/contracts.ts` |

## Chain

| Field | Value |
| --- | --- |
| Network | Arc Testnet |
| Chain ID | `5042002` |
| RPC | `https://rpc.testnet.arc.network` |
| Explorer | `https://testnet.arcscan.app` |
| Native currency label in frontend config | USDC |

## WalletConnect Project ID

WalletConnect support requires a WalletConnect Project ID.

- Local value: `apps/web/.env.local`
- Variable: `NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID`
- Deployment value: configure in Vercel environment variables later.
- Do not commit a real project ID.

`NEXT_PUBLIC_*` values are exposed to the browser. They must not contain private keys, Circle API keys, database URLs, RPC secrets, or server credentials.

## User Flow

1. Open the app.
2. Connect an external wallet through the header wallet control.
3. Verify the connected address appears.
4. Verify the network state.
5. Switch to Arc Testnet if the wallet is on another network.
6. Open a market detail page.
7. If the market is open and the contract reports `isOpen()`, enter a USDC amount.
8. Approve USDC for the `SignalArcMarket` contract.
9. Submit `openPosition`.
10. Wait for transaction receipts.
11. View transaction hashes through Arcscan links.

## Implemented Arc Testnet Trade Flow

The current `trade-intent-panel.tsx` uses browser wallet actions for:

- `approve(address,uint256)` on the Arc Testnet USDC ERC20 interface.
- `openPosition(uint8,uint256)` on the deployed `SignalArcMarket` prototype.
- `isOpen()` read check before enabling execution.
- Arcscan transaction links after submission.

Outcome mapping in the frontend:

| UI outcome | Contract enum value |
| --- | --- |
| YES | `1` |
| NO | `2` |

USDC input is parsed with 6 decimals. For example, `1` USDC becomes `1000000` base units.

## Safety Boundaries

- Browser wallet signs user transactions.
- Backend does not sign user transactions.
- Frontend and backend must not contain private keys.
- Circle embedded wallet is not implemented.
- Circle SDK/API integration is not implemented.
- This is Arc Testnet prototype behavior only.
- No production settlement approval is claimed.
- No audit is claimed.
