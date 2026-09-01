"use client"

import { type FormEvent, useEffect, useMemo, useState } from "react"
import { isAddress, zeroAddress, type Address, type Hash } from "viem"
import { waitForTransactionReceipt, writeContract } from "wagmi/actions"
import { useAccount, useChainId, useConfig, useReadContract, useSwitchChain } from "wagmi"

import {
  ARC_TESTNET_CHAIN_ID,
  ARC_TESTNET_USDC_ADDRESS,
  ERC20_APPROVE_ABI,
  SIGNAL_ARC_MARKET_ABI,
  isArcTestnetChain,
} from "@/lib/contracts"
import { arcTestnet } from "@/lib/wagmi"
import {
  getOutcomeSide,
  getTradeDisabledReason,
  getTradeErrorMessage,
  MARKET_STATUS_OPEN,
  parseUsdcAmount,
  TradePanel,
  type TradeOutcome,
  type TradeSubmitState,
} from "@/modules/trading"

export function TradeIntentPanel({
  marketId,
  marketTitle,
  marketStatus,
  marketContractAddress,
}: {
  marketId: string
  marketTitle?: string
  marketStatus: string
  marketContractAddress: string | null
}) {
  const config = useConfig()
  const { address, isConnected } = useAccount()
  const chainId = useChainId()
  const { switchChain, isPending: isSwitchingChain } = useSwitchChain()
  const [outcome, setOutcome] = useState<TradeOutcome>("YES")
  const [amount, setAmount] = useState("1")
  const [state, setState] = useState<TradeSubmitState>({ status: "idle" })
  const [nowSeconds, setNowSeconds] = useState(() => Math.floor(Date.now() / 1000))
  const isTradingOpen = marketStatus.toUpperCase() === "OPEN"
  const isArcTestnet = isArcTestnetChain(chainId)
  const isPending = state.status === "approving" || state.status === "opening"
  const contractAddress =
    marketContractAddress && isAddress(marketContractAddress)
      ? (marketContractAddress as Address)
      : null
  const userAddress = address ?? zeroAddress
  const readsEnabled = Boolean(contractAddress)
  const walletReadsEnabled = readsEnabled && isConnected
  const parsedAmount = useMemo(() => {
    try {
      return parseUsdcAmount(amount)
    } catch {
      return null
    }
  }, [amount])
  const { data: isContractOpen } = useReadContract({
    address: contractAddress ?? undefined,
    abi: SIGNAL_ARC_MARKET_ABI,
    functionName: "isOpen",
    chainId: ARC_TESTNET_CHAIN_ID,
    query: {
      enabled: readsEnabled,
    },
  })
  const { data: contractStatus } = useReadContract({
    address: contractAddress ?? undefined,
    abi: SIGNAL_ARC_MARKET_ABI,
    functionName: "status",
    chainId: ARC_TESTNET_CHAIN_ID,
    query: {
      enabled: readsEnabled,
    },
  })
  const { data: closeTimestamp } = useReadContract({
    address: contractAddress ?? undefined,
    abi: SIGNAL_ARC_MARKET_ABI,
    functionName: "closeTimestamp",
    chainId: ARC_TESTNET_CHAIN_ID,
    query: {
      enabled: readsEnabled,
    },
  })
  const { data: yesPosition } = useReadContract({
    address: contractAddress ?? undefined,
    abi: SIGNAL_ARC_MARKET_ABI,
    functionName: "yesPositions",
    args: [userAddress],
    chainId: ARC_TESTNET_CHAIN_ID,
    query: {
      enabled: walletReadsEnabled,
    },
  })
  const { data: noPosition } = useReadContract({
    address: contractAddress ?? undefined,
    abi: SIGNAL_ARC_MARKET_ABI,
    functionName: "noPositions",
    args: [userAddress],
    chainId: ARC_TESTNET_CHAIN_ID,
    query: {
      enabled: walletReadsEnabled,
    },
  })
  const hasReachedCloseTime =
    closeTimestamp !== undefined && BigInt(nowSeconds) >= closeTimestamp
  const isContractTradingClosed =
    contractStatus !== undefined && contractStatus !== MARKET_STATUS_OPEN

  useEffect(() => {
    const intervalId = window.setInterval(() => {
      setNowSeconds(Math.floor(Date.now() / 1000))
    }, 30_000)

    return () => window.clearInterval(intervalId)
  }, [])

  const disabledReason = getTradeDisabledReason({
    contractAddress,
    contractStatus,
    isContractTradingClosed,
    hasReachedCloseTime,
    isConnected,
    isArcTestnet,
    isTradingOpen,
    isContractOpen,
    parsedAmount,
  })
  const canSubmit = !disabledReason && !isPending

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!canSubmit || !address || !contractAddress) return

    let approveHash: Hash | undefined
    let openHash: Hash | undefined
    try {
      const usdcAmount = parseUsdcAmount(amount)
      setState({ status: "approving" })

      approveHash = await writeContract(config, {
        address: ARC_TESTNET_USDC_ADDRESS,
        abi: ERC20_APPROVE_ABI,
        functionName: "approve",
        args: [contractAddress, usdcAmount],
        chainId: ARC_TESTNET_CHAIN_ID,
        account: address,
      })
      setState({ status: "approving", approveHash })
      await waitForTransactionReceipt(config, {
        hash: approveHash,
        chainId: ARC_TESTNET_CHAIN_ID,
      })

      setState({ status: "opening", approveHash })
      openHash = await writeContract(config, {
        address: contractAddress,
        abi: SIGNAL_ARC_MARKET_ABI,
        functionName: "openPosition",
        args: [getOutcomeSide(outcome), usdcAmount],
        chainId: ARC_TESTNET_CHAIN_ID,
        account: address,
      })
      setState({ status: "opening", approveHash, openHash })
      await waitForTransactionReceipt(config, {
        hash: openHash,
        chainId: ARC_TESTNET_CHAIN_ID,
      })

      setState({ status: "success", approveHash, openHash })
    } catch (error) {
      setState({
        status: "error",
        message: getTradeErrorMessage(error),
        approveHash,
        openHash,
      })
    }
  }

  return (
    <TradePanel
      marketId={marketId}
      marketTitle={marketTitle}
      contractAddress={contractAddress}
      walletAddress={address}
      outcome={outcome}
      onOutcomeChange={setOutcome}
      amount={amount}
      onAmountChange={setAmount}
      parsedAmount={parsedAmount}
      yesPosition={yesPosition}
      noPosition={noPosition}
      state={state}
      disabledReason={disabledReason}
      canSubmit={canSubmit}
      isPending={isPending}
      isConnected={isConnected}
      isArcTestnet={isArcTestnet}
      isSwitchingChain={isSwitchingChain}
      onSwitchNetwork={() => switchChain({ chainId: arcTestnet.id })}
      onSubmit={handleSubmit}
    />
  )
}
