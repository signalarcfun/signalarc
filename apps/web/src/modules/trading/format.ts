import type { Hash } from "viem"

import { formatShortHash as formatContractShortHash, getArcscanTxUrl } from "@/lib/contracts"

export function getTradeErrorMessage(error: unknown) {
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

  return "Unable to execute the Arc Testnet trade."
}

export function getTxUrl(hash: Hash) {
  return getArcscanTxUrl(hash)
}

export function formatShortHash(hash: Hash) {
  return formatContractShortHash(hash)
}
