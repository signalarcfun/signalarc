"use client"

import { signMessage } from "wagmi/actions"
import type { Address } from "viem"

import {
  createWalletAuthChallenge,
  getWalletAuthSession,
  setWalletAuthSession,
  verifyWalletAuthChallenge,
} from "@/lib/api"
import { wagmiConfig } from "@/lib/wagmi"
import { ensureWalletAuthSessionWithDependencies } from "@/lib/wallet-auth-flow"

export async function ensureWalletAuthSession(address: Address) {
  return ensureWalletAuthSessionWithDependencies(address, {
    getSession: getWalletAuthSession,
    setSession: setWalletAuthSession,
    createChallenge: createWalletAuthChallenge,
    signChallenge: (account, message) =>
      signMessage(wagmiConfig, {
        account,
        message,
      }),
    verifyChallenge: verifyWalletAuthChallenge,
  })
}
