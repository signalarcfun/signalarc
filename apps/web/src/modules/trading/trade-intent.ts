import type { TradeOutcome } from "./types"

export const MARKET_STATUS_OPEN = 1
export const MARKET_STATUS_CLOSED = 2
export const MARKET_STATUS_RESOLVED = 3
export const MARKET_STATUS_CANCELLED = 4

export const marketStatusLabels: Record<number, string> = {
  1: "Open",
  2: "Closed",
  3: "Resolved",
  4: "Cancelled",
}

const outcomeSide: Record<TradeOutcome, 1 | 2> = {
  YES: 1,
  NO: 2,
}

export function getOutcomeSide(outcome: TradeOutcome) {
  return outcomeSide[outcome]
}
