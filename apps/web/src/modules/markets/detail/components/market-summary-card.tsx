import { Card, CardHeader } from "@/components/ui/card"
import type { Market } from "@/lib/api"

import { MarketDetailHeader } from "./market-detail-header"
import { MarketDetailHero } from "./market-detail-hero"

export function MarketSummaryCard({ market }: { market: Market }) {
  return (
    <Card className="overflow-hidden">
      <CardHeader className="space-y-4 p-4 sm:p-6">
        <MarketDetailHero market={market} />
        <MarketDetailHeader market={market} />
      </CardHeader>
    </Card>
  )
}
