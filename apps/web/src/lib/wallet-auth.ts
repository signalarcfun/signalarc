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

function normalizeAddress(address: string) {
  return address.trim().toLowerCase()
}

export async function ensureWalletAuthSession(address: Address) {
  const existing = getWalletAuthSession()
  if (
    existing &&
    normalizeAddress(existing.wallet_address) === normalizeAddress(address) &&
    new Date(existing.expires_at).getTime() > Date.now()
  ) {
    return existing
  }

  setWalletAuthSession(null)
  const challenge = await createWalletAuthChallenge(address)
  const signature = await signMessage(wagmiConfig, {
    account: address,
    message: challenge.data.challenge.message,
  })
  const verified = await verifyWalletAuthChallenge({
    challenge_id: challenge.data.challenge.id,
    address,
    signature,
  })

  setWalletAuthSession(verified.data.session)
  return verified.data.session
}
