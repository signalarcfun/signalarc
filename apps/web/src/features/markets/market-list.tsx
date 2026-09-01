"use client"

import { useEffect, useMemo, useState } from "react"

import {
  DEFAULT_DISCOVERY_TAB,
  discoverMarkets,
  getMarketDiscoveryUrlState,
  MarketCard,
  MarketFilterToolbar,
  MarketListEmptyState,
  MarketListErrorState,
  MarketListLoadingSkeleton,
  setMarketDiscoveryUrlState,
  type DiscoveryTabId,
} from "@/modules/markets"
import type { MarketCategoryId } from "@/modules/categories"
import { ApiError, getMarkets, type Market } from "@/lib/api"

type MarketListState =
  | { status: "loading" }
  | { status: "error"; message: string; requestId: string | null }
  | { status: "ready"; markets: Market[] }

function getErrorState(
  error: unknown,
): Extract<MarketListState, { status: "error" }> {
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
    message: "Unable to load markets.",
    requestId: null,
  }
}

export function MarketList() {
  const [state, setState] = useState<MarketListState>({ status: "loading" })
  const [searchQuery, setSearchQuery] = useState("")
  const [discoveryTab, setDiscoveryTab] =
    useState<DiscoveryTabId>(DEFAULT_DISCOVERY_TAB)
  const [categoryFilter, setCategoryFilter] =
    useState<MarketCategoryId>("all")

  useEffect(() => {
    function syncDiscoveryFiltersFromUrl() {
      const urlState = getMarketDiscoveryUrlState()

      setCategoryFilter(urlState.category)
      setDiscoveryTab(urlState.tab)
    }

    syncDiscoveryFiltersFromUrl()
    window.addEventListener("popstate", syncDiscoveryFiltersFromUrl)

    return () => {
      window.removeEventListener("popstate", syncDiscoveryFiltersFromUrl)
    }
  }, [])

  useEffect(() => {
    let isActive = true

    async function loadMarkets() {
      setState({ status: "loading" })

      try {
        const response = await getMarkets()

        if (isActive) {
          setState({ status: "ready", markets: response.data.markets })
        }
      } catch (error) {
        if (isActive) {
          setState(getErrorState(error))
        }
      }
    }

    void loadMarkets()

    return () => {
      isActive = false
    }
  }, [])

  const filteredMarkets = useMemo(() => {
    if (state.status !== "ready") return []

    return discoverMarkets(state.markets, {
      tab: discoveryTab,
      category: categoryFilter,
      searchQuery,
    })
  }, [state, searchQuery, discoveryTab, categoryFilter])

  function updateDiscoveryTab(value: DiscoveryTabId) {
    setDiscoveryTab(value)
    setMarketDiscoveryUrlState({ tab: value, category: categoryFilter })
  }

  function updateCategoryFilter(value: MarketCategoryId) {
    setCategoryFilter(value)
    setMarketDiscoveryUrlState({ tab: discoveryTab, category: value })
  }

  function clearFilters() {
    setSearchQuery("")
    setDiscoveryTab(DEFAULT_DISCOVERY_TAB)
    setCategoryFilter("all")
    setMarketDiscoveryUrlState({
      tab: DEFAULT_DISCOVERY_TAB,
      category: "all",
    })
  }

  if (state.status === "loading") {
    return <MarketListLoadingSkeleton />
  }

  if (state.status === "error") {
    return (
      <MarketListErrorState
        message={state.message}
        requestId={state.requestId}
        onRetry={() => window.location.reload()}
      />
    )
  }

  if (state.markets.length === 0) {
    return <MarketListEmptyState kind="no-markets" />
  }

  return (
    <div className="space-y-6">
      <MarketFilterToolbar
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        discoveryTab={discoveryTab}
        onDiscoveryTabChange={updateDiscoveryTab}
        categoryFilter={categoryFilter}
        onCategoryFilterChange={updateCategoryFilter}
        markets={state.markets}
        filteredCount={filteredMarkets.length}
      />

      {filteredMarkets.length === 0 ? (
        <MarketListEmptyState kind="no-matches" onClearFilters={clearFilters} />
      ) : (
        <div className="grid gap-4">
          {filteredMarkets.map((market) => (
            <MarketCard key={market.id} market={market} />
          ))}
        </div>
      )}
    </div>
  )
}
