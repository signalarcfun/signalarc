import Link from "next/link"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { MarketList } from "@/features/markets/market-list"

export default function MarketsPage() {
  return (
    <div className="px-4 py-6 sm:px-6 lg:px-8">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-6">
        <section className="rounded-2xl border border-foreground/10 bg-card/60 p-5 shadow-sm sm:p-6 lg:p-8">
          <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
            <div className="max-w-3xl space-y-4">
              <div className="flex flex-wrap items-center gap-2">
                <Badge
                  variant="outline"
                  className="border-indigo-500/20 bg-indigo-500/5 text-[10px] uppercase tracking-[0.18em] text-indigo-300"
                >
                  Arc Testnet
                </Badge>
                <Badge
                  variant="outline"
                  className="border-muted-foreground/20 bg-muted/40 text-[10px] uppercase tracking-[0.18em] text-muted-foreground"
                >
                  YES / NO Markets
                </Badge>
              </div>

              <div className="space-y-2">
                <h1 className="text-2xl font-bold tracking-tight sm:text-3xl lg:text-4xl">
                  Discover prediction markets on Arc
                </h1>
                <p className="max-w-2xl text-sm leading-6 text-muted-foreground sm:text-base">
                  Browse USDC-settled YES/NO markets, filter by category, and
                  track live testnet activity before entering a position.
                </p>
              </div>

              <p className="text-xs text-muted-foreground">
                Testnet preview only. No real funds or production settlement.
              </p>
            </div>

            <div className="flex flex-col gap-2 sm:flex-row lg:flex-col">
              <Button asChild>
                <Link href="/markets/new">Create Market</Link>
              </Button>
              <Button asChild variant="outline">
                <Link href="https://docs.signalarc.fun" target="_blank" rel="noreferrer">Read Docs</Link>
              </Button>
            </div>
          </div>
        </section>

        <MarketList />
      </div>
    </div>
  )
}
