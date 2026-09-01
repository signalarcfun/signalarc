"use client"

import Link from "next/link"

import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"

function ChartIcon({ className }: { className?: string }) {
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
        d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z"
      />
    </svg>
  )
}

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

export function MarketListEmptyState({
  kind,
  onClearFilters,
}: {
  kind: "no-markets" | "no-matches"
  onClearFilters?: () => void
}) {
  if (kind === "no-markets") {
    return (
      <Card>
        <CardContent className="flex flex-col items-center gap-5 py-16 text-center">
          <div className="flex h-14 w-14 items-center justify-center rounded-full bg-muted">
            <ChartIcon className="h-7 w-7 text-muted-foreground" />
          </div>

          <div className="max-w-sm space-y-1">
            <p className="font-medium">No markets available</p>
            <p className="text-sm text-muted-foreground">
              Markets will appear here once they are created through the API
              or creator dashboard. Use the platform API or web interface to
              launch a new event market.
            </p>
          </div>

          <Button asChild>
            <Link href="/markets/new">Create Market</Link>
          </Button>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardContent className="flex flex-col items-center gap-4 py-12 text-center">
        <div className="flex h-12 w-12 items-center justify-center rounded-full bg-muted">
          <SearchIcon className="h-6 w-6 text-muted-foreground" />
        </div>

        <div className="max-w-xs space-y-1">
          <p className="font-medium">No matching markets</p>
          <p className="text-sm text-muted-foreground">
            No markets match the current search or filter criteria. Adjust
            your filters to see results.
          </p>
        </div>

        <Button variant="outline" size="sm" onClick={onClearFilters}>
          Clear Filters
        </Button>
      </CardContent>
    </Card>
  )
}
