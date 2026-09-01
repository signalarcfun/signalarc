export function IntelligenceLoadingSkeleton() {
  return (
    <div className="grid gap-3">
      <div className="h-24 animate-pulse rounded-lg border border-border/50 bg-muted/20" />
      {[1, 2, 3].map((i) => (
        <div key={i} className="animate-pulse rounded-lg border border-border/30 p-4">
          <div className="h-4 w-2/3 rounded bg-muted" />
          <div className="mt-2 flex gap-2">
            <div className="h-4 w-12 rounded bg-muted" />
            <div className="h-4 w-16 rounded bg-muted" />
          </div>
          <div className="mt-3 grid grid-cols-4 gap-4">
            <div className="h-3 rounded bg-muted/60" />
            <div className="h-3 rounded bg-muted/60" />
            <div className="h-3 rounded bg-muted/60" />
            <div className="h-3 rounded bg-muted/60" />
          </div>
        </div>
      ))}
    </div>
  )
}
