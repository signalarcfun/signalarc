import Link from "next/link"

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import type { Market, Position, Settlement } from "@/lib/api"

import {
  formatMarketStatus,
  formatPortfolioDate,
  truncatePortfolioId,
} from "../format"
import type { MarketsState } from "../types"

type ActivityTone = "active" | "resolved" | "claim" | "refund" | "empty"

function buildMarketLookup(marketsState: MarketsState) {
  if (marketsState.status !== "loaded") return new Map<string, Market>()
  return new Map(marketsState.markets.map((market) => [market.id, market]))
}

function getToneClasses(tone: ActivityTone) {
  switch (tone) {
    case "active":
      return "border-green-500/20 bg-green-500/5"
    case "resolved":
      return "border-indigo-500/20 bg-indigo-500/5"
    case "claim":
      return "border-blue-500/20 bg-blue-500/5"
    case "refund":
      return "border-yellow-500/20 bg-yellow-500/5"
    default:
      return "border-border bg-muted/10"
  }
}

function ActivityItem({
  tone,
  type,
  status,
  title,
  description,
  marketId,
  timestamp,
  txHash,
}: {
  tone: ActivityTone
  type: string
  status: string
  title: string
  description: string
  marketId?: string
  timestamp?: string | null
  txHash?: string | null
}) {
  return (
    <article className={`grid gap-3 rounded-lg border p-4 ${getToneClasses(tone)}`}>
      <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-start">
        <div className="min-w-0">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
            {type}
          </p>
          <h3 className="mt-1 break-words text-sm font-semibold text-foreground">
            {title}
          </h3>
        </div>
        <span className="rounded-md border border-border bg-background/50 px-2 py-1 text-xs font-medium text-muted-foreground">
          {status}
        </span>
      </div>
      <p className="text-sm leading-6 text-muted-foreground">{description}</p>
      <div className="grid gap-2 sm:grid-cols-3">
        <div className="min-w-0">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
            Timestamp
          </p>
          <p className="mt-1 text-sm text-foreground">
            {formatPortfolioDate(timestamp ?? null)}
          </p>
        </div>
        <div className="min-w-0">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
            Tx Hash
          </p>
          <p className="mt-1 break-words font-mono text-xs text-foreground">
            {txHash ? truncatePortfolioId(txHash) : "-"}
          </p>
        </div>
        {marketId ? (
          <div className="min-w-0">
            <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
              Market
            </p>
            <Link
              href={`/markets/${encodeURIComponent(marketId)}`}
              className="mt-1 inline-flex text-sm font-medium text-indigo-300 underline-offset-4 hover:text-indigo-200 hover:underline"
            >
              View Market
            </Link>
          </div>
        ) : null}
      </div>
    </article>
  )
}

export function PortfolioActivityCard({
  positions,
  settlements,
  marketsState,
}: {
  positions: Position[]
  settlements: Settlement[]
  marketsState: MarketsState
}) {
  const marketLookup = buildMarketLookup(marketsState)
  const hasActivity = positions.length > 0 || settlements.length > 0

  return (
    <Card>
      <CardHeader>
        <CardTitle>Activity History</CardTitle>
        <CardDescription className="leading-6">
          Existing API records shaped around future activity fields: type,
          status, market, side, amount, tx hash, and timestamp.
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-3">
        {!hasActivity ? (
          <ActivityItem
            tone="empty"
            type="Indexed Activity"
            status="Not available"
            title="No activity records returned"
            description="The current API lookup did not return position or settlement history. Wallet-indexed activity and backend indexing are planned for later phases."
          />
        ) : null}

        {positions.map((position) => {
          const market = marketLookup.get(position.market_id)
          const marketStatus = formatMarketStatus(market?.status)
          const isClosed = market ? market.status.toUpperCase() !== "OPEN" : false

          return (
            <ActivityItem
              key={`position-${position.id}`}
              tone={isClosed ? "resolved" : "active"}
              type={isClosed ? "Resolved Exposure" : "Active Position"}
              status={marketStatus}
              title={market?.title ?? `Market ${truncatePortfolioId(position.market_id)}`}
              description={`${position.outcome} exposure of ${position.quantity}. Average entry ${position.average_entry_price}.`}
              marketId={position.market_id}
              timestamp={position.updated_at}
            />
          )
        })}

        {settlements.map((settlement) => {
          const market = marketLookup.get(settlement.market_id)
          const isRefund =
            market?.status.toUpperCase() === "CANCELLED" ||
            settlement.status.toUpperCase().includes("REFUND")

          return (
            <ActivityItem
              key={`settlement-${settlement.id}`}
              tone={isRefund ? "refund" : "claim"}
              type={isRefund ? "Refundable Cancelled Market" : "Claimable Payout"}
              status={settlement.status}
              title={market?.title ?? `Market ${truncatePortfolioId(settlement.market_id)}`}
              description={`${settlement.outcome ?? "Settlement"} amount ${settlement.amount}.`}
              marketId={settlement.market_id}
              timestamp={settlement.settled_at ?? settlement.updated_at}
              txHash={settlement.tx_hash}
            />
          )
        })}
      </CardContent>
    </Card>
  )
}
