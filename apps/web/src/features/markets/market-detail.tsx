"use client"

import { useEffect, useRef, useState } from "react"

import { MarketResolutionPanel } from "@/features/markets/market-resolution-panel"
import { OnchainMarketLifecyclePanel } from "@/features/markets/onchain-market-lifecycle-panel"
import { TradeIntentPanel } from "@/features/markets/trade-intent-panel"
import { ApiError, getMarket } from "@/lib/api"
import {
  MarketDetailErrorState,
  MarketDetailLoadingSkeleton,
  MarketActivityCard,
  MarketMetadataCard,
  MarketOutcomeCard,
  MarketSummaryCard,
  type MarketDetailState,
} from "@/modules/markets/detail"

function getErrorState(
  error: unknown,
): Extract<MarketDetailState, { status: "error" }> {
  if (error instanceof ApiError) {
    return {
      status: "error",
      message: error.message,
      requestId: error.requestId,
    }
  }

  if (error instanceof Error) {
    return {
      status: "error",
      message: error.message,
      requestId: null,
    }
  }

  return {
    status: "error",
    message: "Unable to load market.",
    requestId: null,
  }
}

export function MarketDetail({ marketId }: { marketId: string }) {
  const [state, setState] = useState<MarketDetailState>({ status: "loading" })
  const [isTradeTicketVisible, setIsTradeTicketVisible] = useState(false)
  const tradeTicketRef = useRef<HTMLElement | null>(null)

  useEffect(() => {
    let isActive = true

    async function loadMarket() {
      setState({ status: "loading" })

      try {
        const response = await getMarket(marketId)

        if (isActive) {
          setState({ status: "ready", market: response.data.market })
        }
      } catch (error) {
        if (isActive) {
          setState(getErrorState(error))
        }
      }
    }

    void loadMarket()

    return () => {
      isActive = false
    }
  }, [marketId])

  useEffect(() => {
    if (state.status !== "ready" || !tradeTicketRef.current) {
      setIsTradeTicketVisible(false)
      return
    }

    const target = tradeTicketRef.current

    if (!("IntersectionObserver" in window)) {
      return
    }

    const observer = new IntersectionObserver(
      ([entry]) => setIsTradeTicketVisible(entry.isIntersecting),
      { rootMargin: "0px 0px -35% 0px", threshold: 0.1 },
    )

    observer.observe(target)

    return () => observer.disconnect()
  }, [state.status])

  if (state.status === "loading") {
    return <MarketDetailLoadingSkeleton />
  }

  if (state.status === "error") {
    return (
      <MarketDetailErrorState
        message={state.message}
        requestId={state.requestId}
      />
    )
  }

  const { market } = state
  const isTradingOpen = market.status.toUpperCase() === "OPEN"

  return (
    <div className="relative pb-40 md:pb-20 lg:pb-0">
      <div className="grid gap-4 lg:grid-cols-[minmax(0,2fr)_minmax(320px,1fr)] lg:gap-6">
        <div className="space-y-4 lg:space-y-6">
          <MarketSummaryCard market={market} />
          <MarketOutcomeCard market={market} />
          <MarketMetadataCard market={market} />
          <MarketActivityCard market={market} />
        </div>

        <aside
          id="trade-ticket"
          ref={tradeTicketRef}
          className="scroll-mt-24 space-y-4 lg:space-y-6"
        >
          <TradeIntentPanel
            marketId={market.id}
            marketTitle={market.title}
            marketStatus={market.status}
            marketContractAddress={market.market_contract_address}
          />
          <OnchainMarketLifecyclePanel
            marketId={market.id}
            marketTitle={market.title}
            marketContractAddress={market.market_contract_address}
          />
          <MarketResolutionPanel marketId={market.id} />
        </aside>
      </div>

      {isTradingOpen && !isTradeTicketVisible ? (
        <div className="fixed inset-x-0 bottom-[calc(env(safe-area-inset-bottom)+4.75rem)] z-[60] border-t border-border bg-background/95 p-3 shadow-lg backdrop-blur md:bottom-0 lg:hidden">
          <a
            href="#trade-ticket"
            className="flex h-11 items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          >
            Trade this market
          </a>
        </div>
      ) : null}
    </div>
  )
}
