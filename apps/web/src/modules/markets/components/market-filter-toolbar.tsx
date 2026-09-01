"use client"

import { useMemo } from "react"

import { Input } from "@/components/ui/input"
import type { MarketCategoryId } from "@/modules/categories"
import type { Market } from "@/lib/api"

import type { DiscoveryTabId } from "../types"
import { DiscoveryTabs } from "./discovery-tabs"
import { MarketCategoryFilter } from "./market-category-filter"

function SearchIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth={1.5}
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z"
      />
    </svg>
  )
}

export function MarketFilterToolbar({
  searchQuery,
  onSearchChange,
  discoveryTab,
  onDiscoveryTabChange,
  categoryFilter,
  onCategoryFilterChange,
  markets,
  filteredCount,
}: {
  searchQuery: string
  onSearchChange: (value: string) => void
  discoveryTab: DiscoveryTabId
  onDiscoveryTabChange: (value: DiscoveryTabId) => void
  categoryFilter: MarketCategoryId
  onCategoryFilterChange: (value: MarketCategoryId) => void
  markets: Market[]
  filteredCount: number
}) {
  const counts = useMemo(() => {
    const open = markets.filter(
      (market) => market.status.toLowerCase() === "open",
    ).length
    const resolved = markets.filter(
      (market) => market.status.toLowerCase() === "resolved",
    ).length

    return { total: markets.length, open, resolved }
  }, [markets])

  return (
    <div className="space-y-3">
      <div className="relative max-w-sm">
        <SearchIcon className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="text"
          placeholder="Search markets by title..."
          value={searchQuery}
          onChange={(event) => onSearchChange(event.target.value)}
          className="pl-8"
        />
      </div>

      <DiscoveryTabs value={discoveryTab} onChange={onDiscoveryTabChange} />

      <MarketCategoryFilter
        value={categoryFilter}
        onChange={onCategoryFilterChange}
      />

      <p className="text-xs font-medium text-muted-foreground">
        {filteredCount === counts.total ? (
          <>
            {counts.total} {counts.total === 1 ? "market" : "markets"}
            {" · "}
            {counts.open} live · {counts.resolved} resolved
          </>
        ) : (
          <>
            Showing {filteredCount} of {counts.total}{" "}
            {counts.total === 1 ? "market" : "markets"}
          </>
        )}
      </p>
    </div>
  )
}
