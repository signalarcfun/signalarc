import { parseUnits } from "viem"

import { USDC_ERC20_DECIMALS } from "@/lib/contracts"

import type { TradeDisabledReasonInput } from "./types"
import {
  MARKET_STATUS_CANCELLED,
  MARKET_STATUS_CLOSED,
  MARKET_STATUS_RESOLVED,
} from "./trade-intent"

export function parseUsdcAmount(value: string) {
  const normalized = value.trim()

  if (!normalized) {
    throw new Error("Enter a USDC amount.")
  }

  if (!/^\d+(\.\d{1,6})?$/.test(normalized)) {
    throw new Error("Enter a valid USDC amount with up to 6 decimals.")
  }

  const amount = parseUnits(normalized, USDC_ERC20_DECIMALS)
  if (amount <= BigInt(0)) {
    throw new Error("USDC amount must be greater than 0.")
  }

  return amount
}

export function getTradeDisabledReason(input: TradeDisabledReasonInput) {
  if (!input.contractAddress) {
    return "Onchain contract not deployed for this market."
  }
  if (input.contractStatus === MARKET_STATUS_RESOLVED) {
    return "Market resolved. Trading is closed; check payout eligibility below."
  }
  if (input.contractStatus === MARKET_STATUS_CANCELLED) {
    return "Market cancelled. Trading is closed; check refund eligibility below."
  }
  if (
    input.contractStatus === MARKET_STATUS_CLOSED ||
    input.isContractOpen === false ||
    input.hasReachedCloseTime
  ) {
    return "Market closed. Trading is unavailable while resolution is pending."
  }
  if (input.isContractTradingClosed) {
    return "Onchain market status is not open."
  }
  if (!input.isConnected) return "Connect wallet to trade."
  if (!input.isArcTestnet) return "Switch to Arc Testnet."
  if (!input.isTradingOpen) {
    return "Backend market status is not open for trading."
  }
  if (!input.parsedAmount) return "Enter a valid USDC amount."

  return null
}
