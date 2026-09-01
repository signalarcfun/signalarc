import { ARC_TESTNET_CHAIN_ID, ARC_TESTNET_EXPLORER_URL } from "./chains"

export function getArcscanTxUrl(hash: string) {
  return `${ARC_TESTNET_EXPLORER_URL}/tx/${hash}`
}

export function getArcscanAddressUrl(address: string) {
  return `${ARC_TESTNET_EXPLORER_URL}/address/${address}`
}

export function formatShortHash(hash: string) {
  return `${hash.slice(0, 10)}...${hash.slice(-8)}`
}

export function formatShortAddress(address: string) {
  return `${address.slice(0, 6)}...${address.slice(-4)}`
}

export function isArcTestnetChain(chainId: number | undefined) {
  return chainId === ARC_TESTNET_CHAIN_ID
}
