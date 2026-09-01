import type { Address, Hex } from "viem"

import type { WalletAuthSession } from "@/lib/api"

type ChallengeResponse = {
  data: {
    challenge: unknown
  }
}

type VerifiedSessionResponse = {
  data: {
    session: WalletAuthSession
  }
}

export type WalletAuthFlowDependencies = {
  getSession: () => WalletAuthSession | null
  setSession: (session: WalletAuthSession | null) => void
  createChallenge: (address: string) => Promise<ChallengeResponse>
  signChallenge: (address: Address, message: string) => Promise<Hex>
  verifyChallenge: (input: {
    challenge_id: string
    address: string
    signature: string
  }) => Promise<VerifiedSessionResponse>
  now?: () => number
}

function normalizeAddress(address: string) {
  return address.trim().toLowerCase()
}

function readChallenge(value: unknown): { id: string; message: string } {
  if (
    typeof value !== "object" ||
    value === null ||
    !("id" in value) ||
    typeof value.id !== "string" ||
    value.id.trim() === "" ||
    !("message" in value) ||
    typeof value.message !== "string" ||
    value.message.trim() === ""
  ) {
    throw new Error("Wallet authentication challenge response is invalid.")
  }

  return { id: value.id, message: value.message }
}

export async function ensureWalletAuthSessionWithDependencies(
  address: Address,
  dependencies: WalletAuthFlowDependencies,
) {
  const existing = dependencies.getSession()
  const now = dependencies.now?.() ?? Date.now()
  if (
    existing &&
    normalizeAddress(existing.wallet_address) === normalizeAddress(address) &&
    new Date(existing.expires_at).getTime() > now
  ) {
    return existing
  }

  dependencies.setSession(null)
  const response = await dependencies.createChallenge(address)
  const challenge = readChallenge(response.data.challenge)
  const signature = await dependencies.signChallenge(address, challenge.message)
  const verified = await dependencies.verifyChallenge({
    challenge_id: challenge.id,
    address,
    signature,
  })

  dependencies.setSession(verified.data.session)
  return verified.data.session
}
