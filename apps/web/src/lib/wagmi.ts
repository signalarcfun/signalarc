"use client"

import { getDefaultConfig } from "@rainbow-me/rainbowkit"
import { defineChain } from "viem"

import {
  ARC_TESTNET_CHAIN_ID,
  ARC_TESTNET_EXPLORER_NAME,
  ARC_TESTNET_EXPLORER_URL,
  ARC_TESTNET_NAME,
  ARC_TESTNET_RPC_URL,
} from "@/lib/contracts"

export const arcTestnet = defineChain({
  id: ARC_TESTNET_CHAIN_ID,
  name: ARC_TESTNET_NAME,
  nativeCurrency: {
    name: "USDC",
    symbol: "USDC",
    decimals: 18,
  },
  rpcUrls: {
    default: {
      http: [ARC_TESTNET_RPC_URL],
    },
  },
  blockExplorers: {
    default: {
      name: ARC_TESTNET_EXPLORER_NAME,
      url: ARC_TESTNET_EXPLORER_URL,
    },
  },
  testnet: true,
})

export const wagmiConfig = getDefaultConfig({
  appName: "SignalArc",
  projectId: process.env.NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID ?? "",
  chains: [arcTestnet],
  ssr: true,
})
