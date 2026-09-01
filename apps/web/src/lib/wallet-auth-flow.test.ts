import { describe, expect, it, vi } from "vitest"
import type { Address, Hex } from "viem"

import type { WalletAuthSession } from "./api"
import {
  ensureWalletAuthSessionWithDependencies,
  type WalletAuthFlowDependencies,
} from "./wallet-auth-flow"

const address = "0x1111111111111111111111111111111111111111" as Address
const signature = `0x${"ab".repeat(65)}` as Hex
const session: WalletAuthSession = {
  token: "session-token",
  user_id: "user-id",
  wallet_address: address,
  expires_at: "2030-01-01T00:00:00Z",
}

function dependencies(
  challenge: unknown,
): WalletAuthFlowDependencies & {
  signChallenge: ReturnType<typeof vi.fn>
  verifyChallenge: ReturnType<typeof vi.fn>
  setSession: ReturnType<typeof vi.fn>
} {
  return {
    getSession: () => null,
    setSession: vi.fn(),
    createChallenge: vi.fn().mockResolvedValue({ data: { challenge } }),
    signChallenge: vi.fn().mockResolvedValue(signature),
    verifyChallenge: vi.fn().mockResolvedValue({ data: { session } }),
  }
}

describe("wallet authentication flow", () => {
  it("rejects the former PascalCase backend response before calling the wallet signer", async () => {
    const deps = dependencies({
      ID: "challenge-id",
      Message: "Sign this challenge",
    })

    await expect(
      ensureWalletAuthSessionWithDependencies(address, deps),
    ).rejects.toThrow("Wallet authentication challenge response is invalid.")
    expect(deps.signChallenge).not.toHaveBeenCalled()
    expect(deps.verifyChallenge).not.toHaveBeenCalled()
  })

  it("signs the snake_case challenge message and verifies the returned hex signature", async () => {
    const deps = dependencies({
      id: "challenge-id",
      message: "Sign this challenge",
    })

    await expect(
      ensureWalletAuthSessionWithDependencies(address, deps),
    ).resolves.toEqual(session)
    expect(deps.signChallenge).toHaveBeenCalledWith(address, "Sign this challenge")
    expect(deps.verifyChallenge).toHaveBeenCalledWith({
      challenge_id: "challenge-id",
      address,
      signature,
    })
    expect(deps.setSession).toHaveBeenLastCalledWith(session)
  })
})
