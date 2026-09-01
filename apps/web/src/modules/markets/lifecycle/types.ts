import type { Address, Hash } from "viem"

export type LifecycleActionState =
  | { status: "idle" }
  | { status: "pending"; label: string; hash?: Hash }
  | { status: "success"; label: string; hash: Hash }
  | { status: "error"; label: string; message: string; hash?: Hash }

export type ResolverAction = "close" | "resolve" | "cancel"

export type LifecycleStatusData = {
  deployedContractAddress: Address
  resolverAddress: Address | undefined
  connectedWallet: Address | undefined
  statusValue: number | undefined
  closeTimestamp: bigint | undefined
  winningOutcome: number | undefined
  userYes: bigint | undefined
  userNo: bigint | undefined
  claimableAmount: bigint | undefined
  hasClaimed: boolean
  isConnected: boolean
  totalYes: bigint | undefined
  totalNo: bigint | undefined
  totalCollateral: bigint | undefined
  isResolved: boolean
}

export type ResolverActionDisabledInput = {
  action: ResolverAction
  isPending: boolean
  isConnected: boolean
  isArcTestnet: boolean
  isResolver: boolean
  isResolved: boolean
  isCancelled: boolean
  statusValue: number | undefined
  isOpen: boolean
  isClosed: boolean
  closeTimestamp: bigint | undefined
  hasReachedCloseTime: boolean
}

export type ClaimDisabledInput = {
  isConnected: boolean
  isArcTestnet: boolean
  isResolved: boolean
  isCancelled: boolean
  hasClaimed: boolean
  claimableAmount: bigint
}
