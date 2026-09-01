import { Card, CardContent } from "@/components/ui/card"

import { IntelligenceSummaryCard } from "./intelligence-summary-card"

function SignalIcon() {
  return (
    <svg className="h-5 w-5 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09z"
      />
    </svg>
  )
}

export function IntelligenceEmptyState() {
  return (
    <div className="grid gap-4">
      <IntelligenceSummaryCard marketCount={0} />
      <Card className="border-border/30">
        <CardContent className="flex flex-col items-center gap-4 py-12 text-center">
          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-muted">
            <SignalIcon />
          </div>
          <div>
            <p className="text-sm font-medium">No market signals available</p>
            <p className="mt-1 text-xs text-muted-foreground">
              Signals will appear here as markets become active on the platform.
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
