import type { Hash } from "viem"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import type { Market } from "@/lib/api"
import { TransactionLink } from "@/modules/wallet"

import { formatMarketDate } from "../format"

function ActivityRow({
  type,
  status,
  description,
  timestamp,
  txHash,
}: {
  type: string
  status: string
  description: string
  timestamp?: string
  txHash?: string | null
}) {
  return (
    <article className="grid gap-3 rounded-lg border border-border bg-muted/20 p-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
            {type}
          </p>
          <h3 className="mt-1 text-sm font-semibold text-foreground">{status}</h3>
        </div>
        {timestamp ? (
          <time className="text-xs text-muted-foreground" dateTime={timestamp}>
            {formatMarketDate(timestamp)}
          </time>
        ) : null}
      </div>
      <p className="text-sm leading-6 text-muted-foreground">{description}</p>
      {txHash ? (
        <div className="min-w-0 rounded-md border border-border bg-background/40 p-3">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
            Transaction
          </p>
          <div className="mt-1">
            <TransactionLink hash={txHash as Hash} />
          </div>
        </div>
      ) : null}
    </article>
  )
}

export function MarketActivityCard({ market }: { market: Market }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Activity</CardTitle>
        <CardDescription className="leading-6">
          Available market-level transaction references and indexed-history status.
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-3">
        {market.market_deployment_tx_hash ? (
          <ActivityRow
            type="Market Deployment"
            status={market.onchain_deployment_status}
            description="The market contract deployment transaction recorded by the existing market API."
            timestamp={market.updated_at}
            txHash={market.market_deployment_tx_hash}
          />
        ) : null}
        <ActivityRow
          type="Indexed Activity"
          status="Not available"
          description="Trade, lifecycle, claim, refund, and resolver activity are not indexed into a market activity feed yet. Live wallet actions still appear in their local transaction panels when submitted."
        />
      </CardContent>
    </Card>
  )
}
