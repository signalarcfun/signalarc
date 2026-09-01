import { normalizeMarketCategory } from "@/modules/categories"

import type {
  DiscoverableMarket,
  DiscoveryTabId,
  MarketDiscoveryFilters,
  MarketDiscoveryTab,
} from "./types"

export const DEFAULT_DISCOVERY_TAB: DiscoveryTabId = "live"

export const MARKET_DISCOVERY_TABS: MarketDiscoveryTab[] = [
  { id: "live", label: "Live" },
  { id: "trending", label: "Trending" },
  { id: "new", label: "New" },
  { id: "ending-soon", label: "Ending Soon" },
  { id: "resolved", label: "Resolved" },
]

const DISCOVERY_TAB_IDS = new Set<DiscoveryTabId>(
  MARKET_DISCOVERY_TABS.map((tab) => tab.id),
)

function parseTime(value: string | null | undefined) {
  if (!value) return null

  const time = new Date(value).getTime()

  return Number.isNaN(time) ? null : time
}

function byDateDesc<T>(
  entries: { market: T; index: number }[],
  getDate: (market: T) => string | null | undefined,
) {
  return entries.sort((a, b) => {
    const aTime = parseTime(getDate(a.market))
    const bTime = parseTime(getDate(b.market))

    if (aTime === null || bTime === null) {
      return a.index - b.index
    }

    return bTime - aTime || a.index - b.index
  })
}

function byDateAsc<T>(
  entries: { market: T; index: number }[],
  getDate: (market: T) => string | null | undefined,
) {
  return entries.sort((a, b) => {
    const aTime = parseTime(getDate(a.market))
    const bTime = parseTime(getDate(b.market))

    if (aTime === null || bTime === null) {
      return a.index - b.index
    }

    return aTime - bTime || a.index - b.index
  })
}

export function normalizeDiscoveryTab(
  value: string | null | undefined,
): DiscoveryTabId {
  if (!value) return DEFAULT_DISCOVERY_TAB

  return DISCOVERY_TAB_IDS.has(value as DiscoveryTabId)
    ? (value as DiscoveryTabId)
    : DEFAULT_DISCOVERY_TAB
}

export function discoverMarkets<TMarket extends DiscoverableMarket>(
  markets: TMarket[],
  filters: MarketDiscoveryFilters,
) {
  const query = filters.searchQuery.trim().toLowerCase()

  let entries = markets.map((market, index) => ({ market, index }))

  if (filters.tab === "live") {
    entries = entries.filter(
      ({ market }) => market.status.toLowerCase() === "open",
    )
  }

  if (filters.tab === "trending") {
    entries = entries.filter(
      ({ market }) => market.status.toLowerCase() === "open",
    )
    entries = byDateDesc(
      entries,
      (market) => market.updated_at ?? market.created_at,
    )
  }

  if (filters.tab === "ending-soon") {
    entries = entries.filter(
      ({ market }) => market.status.toLowerCase() === "open",
    )
    entries = byDateAsc(entries, (market) => market.closes_at)
  }

  if (filters.tab === "resolved") {
    entries = entries.filter(
      ({ market }) => market.status.toLowerCase() === "resolved",
    )
  }

  if (filters.tab === "new") {
    entries = byDateDesc(entries, (market) => market.created_at)
  }

  if (filters.category !== "all") {
    entries = entries.filter(
      ({ market }) => normalizeMarketCategory(market.category) === filters.category,
    )
  }

  if (query) {
    entries = entries.filter(({ market }) =>
      market.title.toLowerCase().includes(query),
    )
  }

  return entries.map(({ market }) => market)
}
