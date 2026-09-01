import { isAddress, type Address } from "viem"

export const ARC_TESTNET_USDC_ADDRESS =
  "0x3600000000000000000000000000000000000000" as Address
export const USDC_ERC20_DECIMALS = 6

const configuredFactoryAddress = process.env.NEXT_PUBLIC_SIGNAL_ARC_MARKET_FACTORY_ADDRESS?.trim()

export const SIGNAL_ARC_MARKET_FACTORY_ADDRESS =
  configuredFactoryAddress && isAddress(configuredFactoryAddress)
    ? (configuredFactoryAddress as Address)
    : null
