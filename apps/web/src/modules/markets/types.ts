import type { MarketCategoryId } from "@/modules/categories"

export type DiscoveryTabId = "live" | "trending" | "new" | "ending-soon" | "resolved"

export type MarketDiscoveryTab = {
  id: DiscoveryTabId
  label: string
}

export type MarketDiscoveryFilters = {
  tab: DiscoveryTabId
  category: MarketCategoryId
  searchQuery: string
}

export type DiscoverableMarket = {
  title: string
  category: string | null
  status: string
  closes_at: string
  created_at?: string | null
  updated_at?: string | null
}
