import { Card, CardContent } from "@/components/ui/card"

function SkeletonBlock({ className }: { className: string }) {
  return <div className={`rounded bg-muted ${className}`} />
}

export function PortfolioLoadingSkeleton() {
  return (
    <div className="grid gap-4">
      {Array.from({ length: 2 }).map((_, index) => (
        <Card key={index} className="animate-pulse">
          <CardContent className="grid gap-4 p-5 sm:p-6">
            <SkeletonBlock className="h-5 w-40" />
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
              {Array.from({ length: 4 }).map((__, tileIndex) => (
                <SkeletonBlock key={tileIndex} className="h-20" />
              ))}
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
