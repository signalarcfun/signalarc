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

type MarketLookup = Map<string, Market>

function buildMarketLookup(marketsState: MarketsState): MarketLookup {
  if (marketsState.status !== "loaded") return new Map()
  return new Map(marketsState.markets.map((market) => [market.id, market]))
}

function getMarketTitle(market: Market | undefined, marketId: string) {
  return market?.title ?? `Market ${truncatePortfolioId(marketId)}`
}

function Field({
  label,
  value,
  mono = false,
}: {
  label: string
  value: string
  mono?: boolean
}) {
  return (
    <div className="min-w-0 rounded-lg border border-border bg-muted/20 p-3">
      <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
        {label}
      </p>
      <p
        className={`mt-1 break-words text-sm font-medium text-foreground ${
          mono ? "font-mono text-xs" : ""
        }`}
      >
        {value || "-"}
      </p>
    </div>
  )
}

function PositionCard({
  position,
  market,
}: {
  position: Position
  market: Market | undefined
}) {
  return (
    <article className="grid gap-4 rounded-lg border border-border bg-card/50 p-4">
      <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-start">
        <div className="min-w-0">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
            Active Position
          </p>
          <h3 className="mt-1 break-words text-base font-semibold leading-snug">
            {getMarketTitle(market, position.market_id)}
          </h3>
        </div>
        <Link
          href={`/markets/${encodeURIComponent(position.market_id)}`}
          className="inline-flex h-9 items-center justify-center rounded-md border border-border px-3 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
        >
          View Market
        </Link>
      </div>

      <dl className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <Field label="Side" value={position.outcome} />
        <Field label="Exposure" value={position.quantity} />
        <Field label="Average Entry" value={position.average_entry_price} />
        <Field
          label="Market Status"
          value={formatMarketStatus(market?.status)}
        />
        <Field
          label="Close Time"
          value={market ? formatPortfolioDate(market.closes_at) : "-"}
        />
        <Field label="Realized PnL" value={position.realized_pnl} />
        <Field
          label="Updated"
          value={formatPortfolioDate(position.updated_at)}
        />
        <Field
          label="Market ID"
          value={truncatePortfolioId(position.market_id)}
          mono
        />
      </dl>
    </article>
  )
}

function SettlementCard({
  settlement,
  market,
}: {
  settlement: Settlement
  market: Market | undefined
}) {
  const isRefund =
    market?.status?.toUpperCase() === "CANCELLED" ||
    settlement.status.toUpperCase().includes("REFUND")

  return (
    <article className="grid gap-4 rounded-lg border border-border bg-card/50 p-4">
      <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-start">
        <div className="min-w-0">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
            {isRefund ? "Refund Record" : "Settlement Record"}
          </p>
          <h3 className="mt-1 break-words text-base font-semibold leading-snug">
            {getMarketTitle(market, settlement.market_id)}
          </h3>
        </div>
        <Link
          href={`/markets/${encodeURIComponent(settlement.market_id)}`}
          className="inline-flex h-9 items-center justify-center rounded-md border border-border px-3 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
        >
          View Market
        </Link>
      </div>

      <dl className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <Field label="Outcome" value={settlement.outcome ?? "-"} />
        <Field label={isRefund ? "Refund Amount" : "Amount"} value={settlement.amount} />
        <Field label="Status" value={settlement.status} />
        <Field
          label="Market Status"
          value={formatMarketStatus(market?.status)}
        />
        <Field
          label="Close Time"
          value={market ? formatPortfolioDate(market.closes_at) : "-"}
        />
        <Field
          label="Settled"
          value={formatPortfolioDate(settlement.settled_at)}
        />
        <Field
          label="Tx Hash"
          value={settlement.tx_hash ? truncatePortfolioId(settlement.tx_hash) : "-"}
          mono
        />
        <Field
          label="Market ID"
          value={truncatePortfolioId(settlement.market_id)}
          mono
        />
      </dl>
    </article>
  )
}

function SectionEmpty({ children }: { children: string }) {
  return (
    <div className="rounded-lg border border-dashed border-border bg-muted/10 p-4 text-sm text-muted-foreground">
      {children}
    </div>
  )
}

export function PortfolioPositionCard({
  positions,
  settlements,
  marketsState,
}: {
  positions: Position[]
  settlements: Settlement[]
  marketsState: MarketsState
}) {
  const marketLookup = buildMarketLookup(marketsState)
  const activeExposure = positions.filter((position) => {
    const market = marketLookup.get(position.market_id)
    return !market || market.status.toUpperCase() === "OPEN"
  })
  const closedExposure = positions.filter((position) => {
    const market = marketLookup.get(position.market_id)
    return market ? market.status.toUpperCase() !== "OPEN" : false
  })

  return (
    <div className="grid gap-5 sm:gap-6">
      <Card>
        <CardHeader>
          <CardTitle>Active Positions</CardTitle>
          <CardDescription className="leading-6">
            Position records loaded from the existing portfolio API, enriched
            with market metadata when available.
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-3">
          {activeExposure.length === 0 ? (
            <SectionEmpty>No active position records were returned.</SectionEmpty>
          ) : (
            activeExposure.map((position) => (
              <PositionCard
                key={position.id}
                position={position}
                market={marketLookup.get(position.market_id)}
              />
            ))
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Claimable / Refundable Positions</CardTitle>
          <CardDescription className="leading-6">
            Settlement records indicate completed or pending payout/refund
            activity. Live claimability remains an onchain lifecycle concern.
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-3">
          {settlements.length === 0 ? (
            <SectionEmpty>No settlement or refund records were returned.</SectionEmpty>
          ) : (
            settlements.map((settlement) => (
              <SettlementCard
                key={settlement.id}
                settlement={settlement}
                market={marketLookup.get(settlement.market_id)}
              />
            ))
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Resolved or Closed Exposure</CardTitle>
          <CardDescription className="leading-6">
            Loaded positions whose market metadata is no longer open.
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-3">
          {closedExposure.length === 0 ? (
            <SectionEmpty>
              No closed, resolved, or cancelled position exposure is available
              in the loaded records.
            </SectionEmpty>
          ) : (
            closedExposure.map((position) => (
              <PositionCard
                key={position.id}
                position={position}
                market={marketLookup.get(position.market_id)}
              />
            ))
          )}
        </CardContent>
      </Card>
    </div>
  )
}
