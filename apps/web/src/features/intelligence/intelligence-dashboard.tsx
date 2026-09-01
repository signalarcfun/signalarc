"use client"

import { useEffect, useState } from "react"

import { getAgentMarkets } from "@/lib/api"
import {
  getIntelligenceErrorState,
  IntelligenceEmptyState,
  IntelligenceErrorState,
  IntelligenceLoadingSkeleton,
  IntelligenceSignalCard,
  IntelligenceSummaryCard,
  type IntelligenceState,
} from "@/modules/intelligence"

export function IntelligenceDashboard() {
  const [state, setState] = useState<IntelligenceState>({ status: "loading" })

  useEffect(() => {
    let isActive = true

    async function loadMarkets() {
      setState({ status: "loading" })

      try {
        const response = await getAgentMarkets()

        if (!isActive) {
          return
        }

        if (response.data.markets.length === 0) {
          setState({ status: "empty" })
          return
        }

        setState({ status: "loaded", markets: response.data.markets })
      } catch (error) {
        if (isActive) {
          setState(getIntelligenceErrorState(error))
        }
      }
    }

    void loadMarkets()

    return () => {
      isActive = false
    }
  }, [])

  if (state.status === "loading") {
    return <IntelligenceLoadingSkeleton />
  }

  if (state.status === "empty") {
    return <IntelligenceEmptyState />
  }

  if (state.status === "error") {
    return <IntelligenceErrorState message={state.message} requestId={state.requestId} />
  }

  return (
    <div className="grid gap-4">
      <IntelligenceSummaryCard marketCount={state.markets.length} />
      {state.markets.map((market) => (
        <IntelligenceSignalCard key={market.id} market={market} />
      ))}
    </div>
  )
}
