import { formatUnits, type Address, type Hash } from "viem"

import {
  USDC_ERC20_DECIMALS,
  formatShortHash as formatContractShortHash,
  getArcscanTxUrl,
} from "@/lib/contracts"

export const marketStatusLabels: Record<number, string> = {
  0: "Draft",
  1: "Open",
  2: "Closed",
  3: "Resolved",
  4: "Cancelled",
}

export const outcomeLabels: Record<number, string> = {
  0: "None",
  1: "Yes",
  2: "No",
}

export function formatLifecycleAddress(address: Address | undefined) {
  return address ?? "-"
}

export function formatUsdc(value: bigint | undefined) {
  if (value === undefined) return "-"
  return `${formatUnits(value, USDC_ERC20_DECIMALS)} USDC`
}

export function formatCloseTimestamp(value: bigint | undefined) {
  if (value === undefined) return "-"

  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(Number(value) * 1000))
}

export function formatLifecycleTxUrl(hash: Hash) {
  return getArcscanTxUrl(hash)
}

export function formatShortHash(hash: Hash) {
  return formatContractShortHash(hash)
}
