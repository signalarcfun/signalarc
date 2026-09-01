"use client"

import { useEffect, useState } from "react"
import { isAddress, zeroAddress, type Address, type Hash } from "viem"
import { waitForTransactionReceipt, writeContract } from "wagmi/actions"
import { useAccount, useChainId, useConfig, useReadContract, useSwitchChain } from "wagmi"

import {
  ARC_TESTNET_CHAIN_ID,
  SIGNAL_ARC_MARKET_ABI,
  isArcTestnetChain,
} from "@/lib/contracts"
import { arcTestnet } from "@/lib/wagmi"
import {
  getClaimDisabledReason,
  getLifecycleErrorMessage,
  getResolverActionDisabledReason,
  getResolverDisabledReason,
  isSameAddress,
  LifecycleNotDeployedCard,
  LifecyclePanel,
  MARKET_STATUS_CANCELLED,
  MARKET_STATUS_CLOSED,
  MARKET_STATUS_OPEN,
  MARKET_STATUS_RESOLVED,
  type LifecycleActionState,
} from "@/modules/markets/lifecycle"

export function OnchainMarketLifecyclePanel({
  marketId,
  marketTitle,
  marketContractAddress,
}: {
  marketId: string
  marketTitle?: string
  marketContractAddress: string | null
}) {
  const config = useConfig()
  const chainId = useChainId()
  const { address, isConnected } = useAccount()
  const { switchChain, isPending: isSwitchingChain } = useSwitchChain()
  const [actionState, setActionState] = useState<LifecycleActionState>({
    status: "idle",
  })
  const [nowSeconds, setNowSeconds] = useState(() => Math.floor(Date.now() / 1000))
  const isArcTestnet = isArcTestnetChain(chainId)
  const contractAddress =
    marketContractAddress && isAddress(marketContractAddress)
      ? (marketContractAddress as Address)
      : null
  const userAddress = address ?? zeroAddress
  const readsEnabled = Boolean(contractAddress)
  const walletReadsEnabled = readsEnabled && isConnected

  const resolverRead = useReadContract({
    address: contractAddress ?? undefined,
    abi: SIGNAL_ARC_MARKET_ABI,
    functionName: "resolver",
    chainId: ARC_TESTNET_CHAIN_ID,
    query: { enabled: readsEnabled },
  })
  const statusRead = useReadContract({
    address: contractAddress ?? undefined,
    abi: SIGNAL_ARC_MARKET_ABI,
    functionName: "status",
    chainId: ARC_TESTNET_CHAIN_ID,
    query: { enabled: readsEnabled },
  })
  const closeTimestampRead = useReadContract({
    address: contractAddress ?? undefined,
    abi: SIGNAL_ARC_MARKET_ABI,
    functionName: "closeTimestamp",
    chainId: ARC_TESTNET_CHAIN_ID,
    query: { enabled: readsEnabled },
  })
  const winningOutcomeRead = useReadContract({
    address: contractAddress ?? undefined,
    abi: SIGNAL_ARC_MARKET_ABI,
    functionName: "winningOutcome",
    chainId: ARC_TESTNET_CHAIN_ID,
    query: { enabled: readsEnabled },
  })
  const totalYesRead = useReadContract({
    address: contractAddress ?? undefined,
    abi: SIGNAL_ARC_MARKET_ABI,
    functionName: "totalYes",
    chainId: ARC_TESTNET_CHAIN_ID,
    query: { enabled: readsEnabled },
  })
  const totalNoRead = useReadContract({
    address: contractAddress ?? undefined,
    abi: SIGNAL_ARC_MARKET_ABI,
    functionName: "totalNo",
    chainId: ARC_TESTNET_CHAIN_ID,
    query: { enabled: readsEnabled },
  })
  const totalCollateralRead = useReadContract({
    address: contractAddress ?? undefined,
    abi: SIGNAL_ARC_MARKET_ABI,
    functionName: "totalCollateral",
    chainId: ARC_TESTNET_CHAIN_ID,
    query: { enabled: readsEnabled },
  })
  const yesPositionRead = useReadContract({
    address: contractAddress ?? undefined,
    abi: SIGNAL_ARC_MARKET_ABI,
    functionName: "yesPositions",
    args: [userAddress],
    chainId: ARC_TESTNET_CHAIN_ID,
    query: { enabled: walletReadsEnabled },
  })
  const noPositionRead = useReadContract({
    address: contractAddress ?? undefined,
    abi: SIGNAL_ARC_MARKET_ABI,
    functionName: "noPositions",
    args: [userAddress],
    chainId: ARC_TESTNET_CHAIN_ID,
    query: { enabled: walletReadsEnabled },
  })
  const claimableRead = useReadContract({
    address: contractAddress ?? undefined,
    abi: SIGNAL_ARC_MARKET_ABI,
    functionName: "claimableAmount",
    args: [userAddress],
    chainId: ARC_TESTNET_CHAIN_ID,
    query: { enabled: walletReadsEnabled },
  })
  const hasClaimedRead = useReadContract({
    address: contractAddress ?? undefined,
    abi: SIGNAL_ARC_MARKET_ABI,
    functionName: "hasClaimed",
    args: [userAddress],
    chainId: ARC_TESTNET_CHAIN_ID,
    query: { enabled: walletReadsEnabled },
  })

  useEffect(() => {
    const intervalId = window.setInterval(() => {
      setNowSeconds(Math.floor(Date.now() / 1000))
    }, 30_000)

    return () => window.clearInterval(intervalId)
  }, [])

  if (!contractAddress) {
    return <LifecycleNotDeployedCard />
  }

  const deployedContractAddress = contractAddress
  const statusValue = statusRead.data
  const closeTimestamp = closeTimestampRead.data
  const claimableAmount = claimableRead.data ?? BigInt(0)
  const hasClaimed = hasClaimedRead.data ?? false
  const isOpen = statusValue === MARKET_STATUS_OPEN
  const isClosed = statusValue === MARKET_STATUS_CLOSED
  const isResolved = statusValue === MARKET_STATUS_RESOLVED
  const isCancelled = statusValue === MARKET_STATUS_CANCELLED
  const hasReachedCloseTime =
    closeTimestamp !== undefined && BigInt(nowSeconds) >= closeTimestamp
  const isResolver = isSameAddress(address, resolverRead.data)
  const isPending = actionState.status === "pending"
  const canClaim =
    isConnected &&
    isArcTestnet &&
    (isResolved || isCancelled) &&
    claimableAmount > BigInt(0) &&
    !hasClaimed

  const claimDisabledReason = getClaimDisabledReason({
    isConnected,
    isArcTestnet,
    isResolved,
    isCancelled,
    hasClaimed,
    claimableAmount,
  })
  const resolverDisabledReason = getResolverDisabledReason({
    isConnected,
    isArcTestnet,
    isResolver,
  })

  async function refetchOnchainReads() {
    await Promise.all([
      resolverRead.refetch(),
      statusRead.refetch(),
      closeTimestampRead.refetch(),
      winningOutcomeRead.refetch(),
      totalYesRead.refetch(),
      totalNoRead.refetch(),
      totalCollateralRead.refetch(),
      yesPositionRead.refetch(),
      noPositionRead.refetch(),
      claimableRead.refetch(),
      hasClaimedRead.refetch(),
    ])
  }

  function getActionDisabledReason(action: "close" | "resolve" | "cancel") {
    return getResolverActionDisabledReason({
      action,
      isPending,
      isConnected,
      isArcTestnet,
      isResolver,
      isResolved,
      isCancelled,
      statusValue,
      isOpen,
      isClosed,
      closeTimestamp,
      hasReachedCloseTime,
    })
  }

  async function runMarketAction(
    label: string,
    functionName: "closeMarket" | "cancelMarket" | "claim",
  ) {
    if (!address) return

    let hash: Hash | undefined
    try {
      setActionState({ status: "pending", label })
      hash = await writeContract(config, {
        address: deployedContractAddress,
        abi: SIGNAL_ARC_MARKET_ABI,
        functionName,
        chainId: ARC_TESTNET_CHAIN_ID,
        account: address,
      })
      setActionState({ status: "pending", label, hash })
      await waitForTransactionReceipt(config, {
        hash,
        chainId: ARC_TESTNET_CHAIN_ID,
      })
      setActionState({ status: "success", label, hash })
      await refetchOnchainReads()
    } catch (error) {
      setActionState({
        status: "error",
        label,
        message: getLifecycleErrorMessage(error),
        hash,
      })
    }
  }

  async function runResolveAction(winningOutcome: 1 | 2) {
    if (!address) return

    const label = winningOutcome === 1 ? "Resolve YES" : "Resolve NO"
    let hash: Hash | undefined
    try {
      setActionState({ status: "pending", label })
      hash = await writeContract(config, {
        address: deployedContractAddress,
        abi: SIGNAL_ARC_MARKET_ABI,
        functionName: "resolve",
        args: [winningOutcome],
        chainId: ARC_TESTNET_CHAIN_ID,
        account: address,
      })
      setActionState({ status: "pending", label, hash })
      await waitForTransactionReceipt(config, {
        hash,
        chainId: ARC_TESTNET_CHAIN_ID,
      })
      setActionState({ status: "success", label, hash })
      await refetchOnchainReads()
    } catch (error) {
      setActionState({
        status: "error",
        label,
        message: getLifecycleErrorMessage(error),
        hash,
      })
    }
  }

  const claimButtonLabel = isCancelled ? "Claim Refund" : "Claim Payout"
  const closeDisabledReason = getActionDisabledReason("close")
  const resolveDisabledReason = getActionDisabledReason("resolve")
  const cancelDisabledReason = getActionDisabledReason("cancel")
  const resolverActionReason =
    closeDisabledReason ?? resolveDisabledReason ?? cancelDisabledReason

  return (
    <LifecyclePanel
      marketId={marketId}
      marketTitle={marketTitle}
      data={{
        deployedContractAddress,
        resolverAddress: resolverRead.data,
        connectedWallet: address,
        statusValue,
        closeTimestamp,
        winningOutcome: winningOutcomeRead.data,
        userYes: yesPositionRead.data,
        userNo: noPositionRead.data,
        claimableAmount,
        hasClaimed,
        isConnected,
        totalYes: totalYesRead.data,
        totalNo: totalNoRead.data,
        totalCollateral: totalCollateralRead.data,
        isResolved,
      }}
      isArcTestnet={isArcTestnet}
      isSwitchingChain={isSwitchingChain}
      onSwitchNetwork={() => switchChain({ chainId: arcTestnet.id })}
      claimDisabledReason={claimDisabledReason}
      canClaim={canClaim}
      claimButtonLabel={claimButtonLabel}
      resolverDisabledReason={resolverDisabledReason}
      resolverActionReason={resolverActionReason}
      closeDisabledReason={closeDisabledReason}
      resolveDisabledReason={resolveDisabledReason}
      cancelDisabledReason={cancelDisabledReason}
      isPending={isPending}
      actionState={actionState}
      onClaim={() => runMarketAction(claimButtonLabel, "claim")}
      onCloseMarket={() => runMarketAction("Close Market", "closeMarket")}
      onResolveYes={() => runResolveAction(1)}
      onResolveNo={() => runResolveAction(2)}
      onCancelMarket={() => runMarketAction("Cancel Market", "cancelMarket")}
    />
  )
}
