import Link from "next/link"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { getMarketCategoryLabel } from "@/modules/categories"
import type { Market } from "@/lib/api"

function formatDate(value: string) {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return value
  }

  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date)
}

function statusColor(status: string) {
  switch (status.toLowerCase()) {
    case "open":
      return "border-green-500/30 bg-green-500/10 text-green-300"
    case "closed":
      return "border-yellow-500/30 bg-yellow-500/10 text-yellow-300"
    case "resolved":
      return "border-indigo-500/30 bg-indigo-500/10 text-indigo-300"
    case "cancelled":
      return "border-red-500/30 bg-red-500/10 text-red-300"
    default:
      return ""
  }
}

function deploymentColor(status: string) {
  switch (status) {
    case "DEPLOYED":
      return "border-green-500/20 bg-green-500/5 text-green-400"
    case "FAILED":
      return "border-red-500/20 bg-red-500/5 text-red-400"
    default:
      return "border-border bg-muted/30 text-muted-foreground"
  }
}

function deploymentLabel(status: string) {
  switch (status) {
    case "DEPLOYED":
      return "Onchain"
    case "FAILED":
      return "Deploy Failed"
    default:
      return "Not Deployed"
  }
}

export function MarketCard({ market }: { market: Market }) {
  const marketHref = `/markets/${market.id}`

  return (
    <Card className="group gap-3 transition-colors hover:ring-indigo-500/30">
      <div className="px-4 pt-4">
        {market.cover_image_url ? (
          // Plain img is intentional for v1 user-provided remote URLs.
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={market.cover_image_url}
            alt={`Cover image for ${market.title}`}
            className="h-32 w-full rounded-lg object-cover sm:h-36"
            loading="lazy"
          />
        ) : (
          <div className="flex h-32 w-full flex-col justify-between rounded-lg border border-dashed border-border bg-muted/40 p-4 sm:h-36">
            <div className="flex items-center justify-between gap-3">
              <span className="rounded-full border bg-background/80 px-2.5 py-1 text-[0.68rem] font-medium uppercase tracking-wider text-muted-foreground">
                SignalArc
              </span>
              <span className="text-xs font-medium text-muted-foreground">
                {market.collateral_asset}
              </span>
            </div>
            <div>
              <p className="text-sm font-semibold text-foreground">
                YES / NO prediction market
              </p>
              <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">
                {market.title}
              </p>
            </div>
          </div>
        )}
      </div>

      <CardHeader className="space-y-3 pb-0">
        <CardDescription className="flex flex-wrap items-center gap-1.5">
          <Badge variant="outline" className={statusColor(market.status)}>
            {market.status}
          </Badge>

          <Badge
            variant="outline"
            className="border-muted-foreground/20 bg-muted/40 text-muted-foreground"
          >
            {getMarketCategoryLabel(market.category)}
          </Badge>

          <Badge
            variant="outline"
            className={`text-xs ${deploymentColor(market.onchain_deployment_status)}`}
          >
            {deploymentLabel(market.onchain_deployment_status)}
          </Badge>
        </CardDescription>

        <div className="space-y-1.5">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
            YES / NO prediction market
          </p>
          <CardTitle className="line-clamp-2 text-base font-semibold leading-snug">
            <Link
              className="transition-colors hover:text-indigo-300"
              href={marketHref}
            >
              {market.title}
            </Link>
          </CardTitle>
        </div>
      </CardHeader>

      <CardContent className="pt-1">
        <dl className="grid gap-3 text-sm text-muted-foreground sm:grid-cols-3">
          <div className="space-y-0.5">
            <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
              Collateral
            </dt>
            <dd className="font-medium text-foreground">
              {market.collateral_asset}
            </dd>
          </div>
          <div className="space-y-0.5">
            <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
              Chain
            </dt>
            <dd className="font-medium text-foreground">{market.chain}</dd>
          </div>
          <div className="space-y-0.5">
            <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
              Closes
            </dt>
            <dd className="font-medium text-foreground">
              {formatDate(market.closes_at)}
            </dd>
          </div>
        </dl>
      </CardContent>

      <CardFooter className="mt-auto bg-muted/30">
        <Button asChild size="sm" variant="outline" className="w-full">
          <Link href={marketHref} aria-label={`View market: ${market.title}`}>
            View market
          </Link>
        </Button>
      </CardFooter>
    </Card>
  )
}
