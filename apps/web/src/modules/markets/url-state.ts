import {
  normalizeMarketCategory,
  type MarketCategoryId,
} from "@/modules/categories"

import { DEFAULT_DISCOVERY_TAB, normalizeDiscoveryTab } from "./discovery"
import type { DiscoveryTabId } from "./types"

export type MarketDiscoveryUrlState = {
  category: MarketCategoryId
  tab: DiscoveryTabId
}

export function getCategoryFilterFromUrl(): MarketCategoryId {
  if (typeof window === "undefined") return "all"

  const category = new URLSearchParams(window.location.search).get("category")

  if (!category) return "all"

  return normalizeMarketCategory(category)
}

export function getDiscoveryTabFromUrl(): DiscoveryTabId {
  if (typeof window === "undefined") return DEFAULT_DISCOVERY_TAB

  const tab = new URLSearchParams(window.location.search).get("tab")

  return normalizeDiscoveryTab(tab)
}

export function getMarketDiscoveryUrlState(): MarketDiscoveryUrlState {
  return {
    category: getCategoryFilterFromUrl(),
    tab: getDiscoveryTabFromUrl(),
  }
}

export function setMarketDiscoveryUrlState(filters: MarketDiscoveryUrlState) {
  if (typeof window === "undefined") return

  const url = new URL(window.location.href)

  if (filters.category === "all") {
    url.searchParams.delete("category")
  } else {
    url.searchParams.set("category", filters.category)
  }

  url.searchParams.set("tab", filters.tab)

  window.history.pushState(null, "", `${url.pathname}${url.search}${url.hash}`)
}
