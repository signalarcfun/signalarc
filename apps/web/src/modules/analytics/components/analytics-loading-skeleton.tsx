export function AnalyticsLoadingSkeleton() {
  return (
    <div className="grid gap-6">
      <div className="h-56 animate-pulse rounded-lg border border-border/50 bg-muted/20" />
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-5">
        {[1, 2, 3, 4, 5].map((item) => (
          <div key={item} className="h-36 animate-pulse rounded-lg border border-border/50 bg-muted/20" />
        ))}
      </div>
    </div>
  )
}
