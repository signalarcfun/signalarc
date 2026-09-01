# Arc Testnet Deployment

## Purpose

This document records the manual owner-run Arc Testnet deployment and onchain smoke test results for the SignalArc Phase 6 contract prototype.

No private keys, seed phrases, `.env` contents, Circle credentials, or RPC secrets are recorded here.

## Network

- Network: Arc Testnet
- Chain ID: 5042002
- RPC URL used locally: `https://rpc.testnet.arc.network`
- Explorer: `https://testnet.arcscan.app`
- Collateral token used: `0x3600000000000000000000000000000000000000`

## Deployment Boundary

- Deployment was run manually by the repo owner from a local terminal.
- Private key remained local-only.
- `.env` remained local-only and was not committed.
- No Circle API key or Circle wallet credential was used.
- This deployment does not approve production usage.
- This deployment does not represent a production custody model.

## Primary Market Deployment

- Contract: `SignalArcMarket`
- Address: `0xf4ccc11A9e24fb996679F946C23C04AFd2797F26`
- Deploy transaction: `0xb102abd10d865b5215112774fe748bae656aead6f6b5151eeb64bf99650ec658`
- Deployer / resolver: `0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC`
- Question: `Will SignalArc deploy a local prototype to Arc Testnet?`
- Source verification: passed on Arcscan.
- Constructor/read verification:
  - code size was non-empty
  - resolver matched deployer/resolver address
  - collateral token matched Arc Testnet USDC address
  - status was `1`
  - `isOpen()` returned `true`

## Primary Market Open Position Test

- Approve transaction: `0x6640823970142c88146e91c780c999886dd742e7b1da4d6134b5cbf15db26a88`
- Open YES position transaction: `0xeb2cfc774ef604a0d808ec4f02aac69c25db9a02fdfd816fa56d70e37074d869`
- Amount: `1000000` USDC base units
- Result:
  - YES position: `1000000`
  - total YES: `1000000`
  - total collateral: `1000000`
  - market USDC balance: `1000000`

## Short Lifecycle Market

- Address: `0x5d936e93809474C04b9F78F857D6AcFd10b0d11b`
- Deploy transaction: `0x27f6fdbc5e6b50fd12d9dc8303705de465ca322a55b50bc180b1a68c10b1d595`
- Close transaction: `0x3483bbf14a55ffc46054095f4944de509d39bd13575d968686e21f05f0ef70c3`
- Resolve transaction: `0x4caa8c54a33f8b1c4e36b4c09c4fc45cdbd4864a0cd5838ea32f844d63932694`
- Result:
  - status after close: `2`
  - status after resolve: `3`
  - winning outcome: `1`

## Claim Path Market

- Address: `0x144f21A9E7D83f94b36377285263634192d38801`
- Deploy transaction: `0x40dc8d8bb58f5714789a303d8e080e5acd8dc34b258ae34e7cbbc1848f50f148`
- Approve transaction: `0x15eb98150d1213c4f829e8ae4dd96faad2de526b503316bc5ab712b8e7dd5083`
- Open YES position transaction: `0x0c952f05c6f9f5a915ea65a2470ef7f8ff446ce68e2f3bca43596c5e423096be`
- Close transaction: `0x8a0a59830a5ef11e66bf057df289e4c8360c576d1b3131f920592ad03053b1ec`
- Resolve transaction: `0x52f68c8776feedbb553dd7cc786386c9cc7ec76b6c412fdba9a0215894e61691`
- Claim transaction: `0xe235dd783cfb69ac5701bca59c9a41c5fde402f8be56f0dc1a1245a19c592d6f`
- Result:
  - status after resolve: `3`
  - winning outcome: `1`
  - claimable amount before claim: `1000000`
  - has claimed: `true`
  - market USDC balance after claim: `0`

## Cancel / Refund Market

- Address: `0x2A59100e22a41F804d4eB25E8E5464803CbFfb45`
- Deploy transaction: `0x2e30aafc8a53edea6db99c58e480acd22025dbf0d510f5e9924dc72c4921253b`
- Approve transaction: `0x1f712378f6b88bd070a6b57e6aa4494b82ad8e39803a6c8c744a93a0c2d56bd4`
- Open YES position transaction: `0xc0bab83b9fcc925ac7f9d8815c5b838f93ae1e69a5a8f1fd99ce43eb438849ff`
- Cancel transaction: `0xd25806f75b0eea9dba40de24bcd0c90dc32d5cd3aa961f22041f4b29f157344e`
- Refund claim transaction: `0xcb2689f32c1cb5773acb1e44c222f486daa0b7b2297096ad346d3030a9e812f8`
- Result:
  - status after cancel: `4`
  - claimable refund: `1000000`
  - has claimed: `true`
  - market USDC balance after refund: `0`

## Onchain Smoke Test Summary

Passed:

- Deploy
- Source verification
- Constructor/read verification
- USDC approve
- Open YES position
- Position accounting
- Close market
- Resolve market
- Claim payout
- Cancel market
- Refund claim

## Explicit Limitations

- This is still a prototype.
- This is not audited.
- This is not production custody.
- This is not production settlement.
- Payout behavior is fixed 1:1.
- Resolver is a single address.
- No oracle integration exists.
- No dispute flow exists.
- No frontend/backend integration is completed in this step.
