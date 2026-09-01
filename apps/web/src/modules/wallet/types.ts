import type { Address, Hash } from "viem"

export type WalletAddress = Address
export type TransactionHash = Hash

export type WalletStatusTone = "warning" | "success" | "error" | "info"
