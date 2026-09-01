import Link from "next/link"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { getMarketCategoryLabel } from "@/modules/categories"

import { formatIntelligenceDate } from "../format"
import { getIntelligenceStatusClassName } from "../intelligence-utils"
import type { IntelligenceSignal } from "../types"

export function IntelligenceSignalCard({ market }: { market: IntelligenceSignal }) {
  return (
    <div className="group rounded-lg border border-border/50 bg-card/40 transition-colors hover:border-indigo-500/20">
      <div className="flex items-start justify-between gap-3 p-4">
        <div className="min-w-0 space-y-1.5">
          <Link
            className="text-sm font-semibold text-foreground hover:text-indigo-300 transition-colors"
            href={`/markets/${market.id}`}
          >
            {market.title}
          </Link>
          <div className="flex flex-wrap items-center gap-1.5">
            <Badge variant="outline" className={`text-[10px] ${getIntelligenceStatusClassName(market.status)}`}>
              {market.status}
            </Badge>
            <Badge variant="outline" className="text-[10px] border-border bg-muted/30 text-muted-foreground">
              {getMarketCategoryLabel(market.category)}
            </Badge>
          </div>
        </div>
        <Button
          asChild
          size="sm"
          variant="ghost"
          className="shrink-0 text-xs opacity-0 transition-opacity group-hover:opacity-100"
        >
          <Link href={`/markets/${market.id}`}>View {"\u2192"}</Link>
        </Button>
      </div>
      <Separator className="opacity-30" />
      <dl className="grid gap-x-4 gap-y-1 px-4 py-3 text-xs sm:grid-cols-4">
        <div>
          <dt className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground/60">Collateral</dt>
          <dd className="font-medium text-foreground">{market.collateral_asset}</dd>
        </div>
        <div>
          <dt className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground/60">Chain</dt>
          <dd className="font-medium text-foreground">{market.chain}</dd>
        </div>
        <div>
          <dt className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground/60">Closes</dt>
          <dd className="font-medium text-foreground">{formatIntelligenceDate(market.closes_at)}</dd>
        </div>
        {market.resolution_source ? (
          <div>
            <dt className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground/60">Resolution</dt>
            <dd className="font-medium text-foreground truncate" title={market.resolution_source}>
              {market.resolution_source}
            </dd>
          </div>
        ) : null}
      </dl>
    </div>
  )
}
