import { Card, CardContent, CardHeader } from "@/components/ui/card"
import type { Market } from "@/lib/api"

export function MarketOutcomeCard({ market }: { market: Market }) {
  return (
    <Card>
      <CardHeader className="pb-3">
        <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          Outcomes &amp; Probability
        </h3>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-3 sm:grid-cols-2">
          <div className="min-w-0 rounded-lg border border-green-500/20 bg-green-500/5 p-4">
            <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              YES
            </p>
            <p className="mt-1 break-words text-xl font-bold leading-snug text-green-400">
              {market.outcome_yes_label}
            </p>
          </div>
          <div className="min-w-0 rounded-lg border border-red-500/20 bg-red-500/5 p-4">
            <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              NO
            </p>
            <p className="mt-1 break-words text-xl font-bold leading-snug text-red-400">
              {market.outcome_no_label}
            </p>
          </div>
        </div>
        <p className="text-xs text-muted-foreground">
          Probability signal will be derived from market position data when available.
        </p>
      </CardContent>
    </Card>
  )
}
