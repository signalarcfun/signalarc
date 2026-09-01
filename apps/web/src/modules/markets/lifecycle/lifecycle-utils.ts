import type { Address } from "viem"

import type { ClaimDisabledInput, ResolverActionDisabledInput } from "./types"

export const MARKET_STATUS_OPEN = 1
export const MARKET_STATUS_CLOSED = 2
export const MARKET_STATUS_RESOLVED = 3
export const MARKET_STATUS_CANCELLED = 4

export function getLifecycleErrorMessage(error: unknown) {
  if (error instanceof Error) {
    const message = error.message.toLowerCase()
    if (
      message.includes("user rejected") ||
      message.includes("user denied") ||
      message.includes("rejected the request") ||
      message.includes("request rejected")
    ) {
      return "Wallet transaction was rejected."
    }

    return error.message
  }

  return "Unable to execute the Arc Testnet contract action."
}

export function isSameAddress(
  left: Address | undefined,
  right: Address | undefined,
) {
  return Boolean(left && right && left.toLowerCase() === right.toLowerCase())
}

export function getClaimDisabledReason(input: ClaimDisabledInput) {
  if (!input.isConnected) return "Connect wallet to check claim eligibility."
  if (!input.isArcTestnet) return "Switch to Arc Testnet."
  if (!input.isResolved && !input.isCancelled) {
    return "Claims are available only after resolution or cancellation."
  }
  if (input.hasClaimed) return "This wallet has already claimed."
  if (input.claimableAmount === BigInt(0)) {
    return "No claimable USDC for this wallet."
  }

  return null
}

export function getResolverDisabledReason({
  isConnected,
  isArcTestnet,
  isResolver,
}: {
  isConnected: boolean
  isArcTestnet: boolean
  isResolver: boolean
}) {
  if (!isConnected) return "Connect the resolver wallet to manage lifecycle actions."
  if (!isArcTestnet) return "Switch to Arc Testnet."
  if (!isResolver) {
    return "Connected wallet is not the resolver for this Arc Testnet market."
  }

  return null
}

export function getResolverActionDisabledReason(
  input: ResolverActionDisabledInput,
) {
  if (input.isPending) return "Wait for the current transaction to confirm."
  if (!input.isConnected) {
    return "Connect the resolver wallet to manage lifecycle actions."
  }
  if (!input.isArcTestnet) return "Switch to Arc Testnet."
  if (!input.isResolver) {
    return "Connected wallet is not the resolver for this Arc Testnet market."
  }
  if (input.isResolved || input.isCancelled) {
    return "Resolver actions are disabled after resolution or cancellation."
  }
  if (input.statusValue === undefined) return "Loading onchain market status."

  if (input.action === "close") {
    if (!input.isOpen) return "Close is only available while the market is Open."
    if (input.closeTimestamp === undefined) return "Loading market close time."
    if (!input.hasReachedCloseTime) {
      return "Close is available after the market close time."
    }
  }

  if (input.action === "resolve" && !input.isClosed) {
    return "Resolve is only available after the market is Closed."
  }

  if (input.action === "cancel" && !input.isOpen && !input.isClosed) {
    return "Cancel is only available while the market is Open or Closed."
  }

  return null
}
