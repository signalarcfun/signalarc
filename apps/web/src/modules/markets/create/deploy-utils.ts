import { decodeEventLog, type Address, type TransactionReceipt } from "viem"

import { SIGNAL_ARC_MARKET_FACTORY_ABI } from "@/lib/contracts"

export function closeTimestampSeconds(value: string) {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    throw new Error("Close date must be valid.")
  }

  return BigInt(Math.floor(date.getTime() / 1000))
}

export function getDeployErrorMessage(error: unknown) {
  if (error instanceof Error) {
    const message = error.message.toLowerCase()
    if (
      message.includes("user rejected") ||
      message.includes("user denied") ||
      message.includes("rejected the request") ||
      message.includes("request rejected")
    ) {
      return "Wallet transaction was rejected. No backend market was created."
    }

    return error.message
  }

  return "Unable to deploy the Arc Testnet market contract."
}

export function getDeployedMarketAddress(
  receipt: TransactionReceipt,
): Address {
  for (const log of receipt.logs) {
    try {
      const decoded = decodeEventLog({
        abi: SIGNAL_ARC_MARKET_FACTORY_ABI,
        data: log.data,
        topics: log.topics,
      })

      if (decoded.eventName === "MarketDeployed") {
        return decoded.args.market
      }
    } catch {
      // Ignore logs from contracts other than the factory.
    }
  }

  throw new Error("MarketDeployed event was not found in the factory receipt.")
}
