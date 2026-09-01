"use client"

import { type FormEvent, useEffect, useState } from "react"
import { useAccount } from "wagmi"

import { getMarkets, getUserPositions, getUserSettlements } from "@/lib/api"
import {
  getMarketsErrorState,
  getPortfolioErrorState,
  PortfolioActivityCard,
  PortfolioAdvancedLookup,
  PortfolioEmptyState,
  PortfolioErrorState,
  PortfolioLoadingSkeleton,
  PortfolioPositionCard,
  PortfolioShell,
  PortfolioSummaryCard,
  WalletIdentityCard,
  WalletNotConnectedState,
  type MarketsState,
  type PortfolioState,
} from "@/modules/portfolio"

export function PortfolioView() {
  const { address, isConnected } = useAccount()
  const [state, setState] = useState<PortfolioState>({ status: "idle" })
  const [marketsState, setMarketsState] = useState<MarketsState>({ status: "loading" })
  const [showAdvanced, setShowAdvanced] = useState(false)

  useEffect(() => {
    let isActive = true

    async function loadMarkets() {
      try {
        const response = await getMarkets()
        if (isActive) {
          setMarketsState({ status: "loaded", markets: response.data.markets })
        }
      } catch (error) {
        if (isActive) {
          setMarketsState(getMarketsErrorState(error))
        }
      }
    }

    void loadMarkets()

    return () => {
      isActive = false
    }
  }, [])

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const formData = new FormData(event.currentTarget)
    const userId = String(formData.get("user_id") ?? "").trim()

    if (!userId) {
      setState({
        status: "error",
        userId: "",
        message: "User ID is required.",
        requestId: null,
      })
      return
    }

    setState({ status: "loading", userId })

    try {
      const [positionsResponse, settlementsResponse] = await Promise.all([
        getUserPositions(userId),
        getUserSettlements(userId),
      ])

      const data = {
        positions: positionsResponse.data.positions,
        settlements: settlementsResponse.data.settlements,
      }

      if (data.positions.length === 0 && data.settlements.length === 0) {
        setState({ status: "empty", userId })
        return
      }

      setState({ status: "loaded", userId, data })
    } catch (error) {
      setState(getPortfolioErrorState(userId, error))
    }
  }

  return (
    <PortfolioShell>
      {isConnected && address ? (
        <WalletIdentityCard address={address} />
      ) : (
        <WalletNotConnectedState />
      )}
      <PortfolioSummaryCard
        address={address}
        isConnected={isConnected}
        marketsState={marketsState}
        portfolioState={state}
      />
      <PortfolioAdvancedLookup
        showAdvanced={showAdvanced}
        isLoading={state.status === "loading"}
        onToggleAdvanced={() => setShowAdvanced(!showAdvanced)}
        onSubmit={handleSubmit}
      />

      {state.status === "idle" ? (
        <PortfolioEmptyState
          title="No portfolio lookup loaded"
          description="Connect a wallet for identity context, then use the demo lookup to load existing API position and settlement records."
        />
      ) : null}
      {state.status === "loading" ? <PortfolioLoadingSkeleton /> : null}
      {state.status === "empty" ? (
        <>
          <PortfolioEmptyState />
          <PortfolioPositionCard
            positions={[]}
            settlements={[]}
            marketsState={marketsState}
          />
          <PortfolioActivityCard
            positions={[]}
            settlements={[]}
            marketsState={marketsState}
          />
        </>
      ) : null}
      {state.status === "error" ? <PortfolioErrorState message={state.message} requestId={state.requestId} /> : null}
      {state.status === "loaded" ? (
        <>
          <PortfolioPositionCard
            positions={state.data.positions}
            settlements={state.data.settlements}
            marketsState={marketsState}
          />
          <PortfolioActivityCard
            positions={state.data.positions}
            settlements={state.data.settlements}
            marketsState={marketsState}
          />
        </>
      ) : null}
    </PortfolioShell>
  )
}
