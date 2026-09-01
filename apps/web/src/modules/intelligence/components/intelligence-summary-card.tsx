import { Card, CardContent } from "@/components/ui/card"

export function IntelligenceSummaryCard({ marketCount }: { marketCount: number }) {
  return (
    <Card className="border-indigo-500/10 bg-gradient-to-r from-indigo-500/5 via-transparent to-purple-500/5">
      <CardContent className="py-5">
        <div className="grid gap-6 sm:grid-cols-[1fr_auto]">
          <div className="space-y-2">
            <h2 className="text-sm font-semibold text-foreground">Market Intelligence API</h2>
            <p className="text-xs leading-relaxed text-muted-foreground">
              Structured, API-accessible probability signals derived from prediction market activity.
              Available for programmatic consumption through the{" "}
              <a
                href="https://docs.signalarc.fun/AGENT_API"
                target="_blank"
                rel="noopener noreferrer"
                className="text-indigo-400 hover:text-indigo-300 transition-colors"
              >
                Agent API
              </a>
              . Signals include market status, category, collateral, resolution source, and close timestamps.
            </p>
          </div>
          <div className="flex items-center gap-4">
            <div className="rounded-lg border border-border/50 bg-card/60 px-4 py-2 text-center">
              <p className="text-lg font-bold text-indigo-400">{marketCount}</p>
              <p className="text-[10px] font-medium uppercase tracking-widest text-muted-foreground">
                {marketCount === 1 ? "Signal" : "Signals"}
              </p>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
