import type { Hash } from "viem"

export type TradeSubmitState =
  | { status: "idle" }
  | { status: "approving"; approveHash?: Hash }
  | { status: "opening"; approveHash: Hash; openHash?: Hash }
  | { status: "success"; approveHash: Hash; openHash: Hash }
  | { status: "error"; message: string; approveHash?: Hash; openHash?: Hash }

export type TradeOutcome = "YES" | "NO"

export type TradeDisabledReasonInput = {
  contractAddress: string | null
  contractStatus: number | undefined
  isContractTradingClosed: boolean
  hasReachedCloseTime: boolean
  isConnected: boolean
  isArcTestnet: boolean
  isTradingOpen: boolean
  isContractOpen: boolean | undefined
  parsedAmount: bigint | null
}
