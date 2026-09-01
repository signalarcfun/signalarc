import { formatShortAddress, formatShortHash, getArcscanTxUrl } from "@/lib/contracts"

import type { TransactionHash, WalletAddress } from "./types"

export function getTransactionExplorerUrl(hash: TransactionHash) {
  return getArcscanTxUrl(hash)
}

export function formatTransactionHash(hash: TransactionHash) {
  return formatShortHash(hash)
}

export function formatWalletAddress(address: WalletAddress | null | undefined) {
  return address ? formatShortAddress(address) : "-"
}

export function getWalletConnectionMessage(isConnected: boolean) {
  return isConnected ? "Wallet connected." : "Use the wallet control in the header to connect."
}

export function getArcTestnetSwitchLabel(isSwitchingChain: boolean) {
  return isSwitchingChain ? "Switching..." : "Switch network"
}
