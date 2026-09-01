import {
  Card,
  CardContent,
  CardHeader,
} from "@/components/ui/card"

export function MarketListLoadingSkeleton() {
  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <div className="h-8 w-full max-w-sm animate-pulse rounded-lg bg-muted" />
        <div className="flex gap-2">
          {[1, 2, 3, 4].map((i) => (
            <div
              key={i}
              className="h-8 w-16 animate-pulse rounded-md bg-muted"
            />
          ))}
        </div>
        <div className="h-4 w-64 animate-pulse rounded bg-muted" />
      </div>

      <div className="grid gap-4">
        {[1, 2, 3, 4].map((i) => (
          <Card key={i} className="animate-pulse">
            <div className="px-6 pt-6">
              <div className="h-48 w-full rounded-xl bg-muted" />
            </div>
            <CardHeader className="pb-3">
              <div className="space-y-2.5">
                <div className="h-5 w-3/4 rounded bg-muted" />
                <div className="flex gap-1.5">
                  <div className="h-5 w-14 rounded-full bg-muted" />
                  <div className="h-5 w-20 rounded-full bg-muted" />
                  <div className="h-5 w-20 rounded-full bg-muted" />
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 sm:grid-cols-3">
                {[1, 2, 3].map((j) => (
                  <div key={j} className="space-y-1.5">
                    <div className="h-3 w-16 rounded bg-muted/70" />
                    <div className="h-4 w-24 rounded bg-muted" />
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}
