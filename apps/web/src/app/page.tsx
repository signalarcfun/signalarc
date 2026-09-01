"use client"

import { ConnectButton } from "@rainbow-me/rainbowkit"
import Link from "next/link"
import { useAccount } from "wagmi"

import { SiteFooter } from "@/components/layout/site-footer"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"

/* ─── Static signal terminal rows (interface preview only) ────────── */
const signalRows = [
  { label: "ETH > $5k by Dec 2025", status: "Open", probability: "—", category: "Crypto" },
  { label: "Fed cuts rate Q3 2025", status: "Open", probability: "—", category: "Macro" },
  { label: "Arc mainnet launch H2", status: "Closed", probability: "—", category: "Protocol" },
]

const infrastructureSurfaces = [
  {
    title: "Market Intelligence",
    description: "Structured probability signals accessible via the Agent API for automated consumption.",
    href: "/intelligence",
    badge: "API",
  },
  {
    title: "Resolver Workflows",
    description: "Configurable resolution sources and transparent settlement lifecycle per market.",
    href: "/markets",
    badge: "Lifecycle",
  },
  {
    title: "Agent Execution",
    description: "Programmatic market creation, trading intent, and execution through the Agent API.",
    href: "https://docs.signalarc.fun/AGENT_API",
    badge: "Agent",
    external: true,
  },
]

export default function Home() {
  const { isConnected } = useAccount()

  return (
    <>
      <div className="px-4 py-8 sm:px-6 lg:px-8">
        <div className="mx-auto flex w-full max-w-7xl flex-col gap-12">

          {/* ── Hero: two columns on desktop ─────────────────────── */}
          <section className="grid items-center gap-8 pt-8 lg:grid-cols-2 lg:gap-12 lg:pt-16">
            {/* Left: headline + CTAs */}
            <div className="flex flex-col gap-6">
              <Badge variant="outline" className="w-fit border-indigo-500/30 bg-indigo-500/10 text-indigo-300 text-xs">
                Arc-Native · Testnet Preview
              </Badge>
              <h1 className="text-3xl font-bold tracking-tight sm:text-4xl lg:text-5xl">
                Prediction Market
                <span className="block bg-gradient-to-r from-indigo-400 via-purple-400 to-indigo-300 bg-clip-text text-transparent">
                  Infrastructure on Arc
                </span>
              </h1>
              <p className="max-w-lg text-base leading-relaxed text-muted-foreground">
                SignalArc is an API-first infrastructure layer for USDC-settled event markets,
                real-time probability signals, resolver workflows, and agent execution on Arc.
              </p>
              <div className="flex flex-wrap items-center gap-3 pt-2">
                {!isConnected ? (
                  <ConnectButton label="Connect Wallet" />
                ) : (
                  <Button asChild size="lg">
                    <Link href="/markets">Explore Markets</Link>
                  </Button>
                )}
                <Button asChild variant="outline" size="lg">
                  <Link href="/intelligence">Intelligence</Link>
                </Button>
              </div>
            </div>

            {/* Right: Signal Terminal preview */}
            <div className="rounded-xl border border-border/60 bg-card/50 shadow-lg shadow-indigo-500/5">
              <div className="flex items-center justify-between border-b border-border/40 px-4 py-3">
                <div className="flex items-center gap-2">
                  <div className="h-2 w-2 rounded-full bg-indigo-400/60" />
                  <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    Market Signal Terminal
                  </span>
                </div>
                <Badge variant="outline" className="border-border bg-muted/40 text-muted-foreground text-[10px]">
                  Interface Preview
                </Badge>
              </div>
              <div className="divide-y divide-border/30">
                {signalRows.map((row) => (
                  <div key={row.label} className="flex items-center justify-between gap-4 px-4 py-3">
                    <div className="flex items-center gap-3 min-w-0">
                      <span className="text-sm font-medium text-foreground truncate">{row.label}</span>
                    </div>
                    <div className="flex shrink-0 items-center gap-2">
                      <Badge variant="outline" className="text-[10px] border-indigo-500/20 bg-indigo-500/5 text-indigo-300">
                        {row.category}
                      </Badge>
                      <span className={`text-[10px] font-medium px-2 py-0.5 rounded-full ${
                        row.status === "Open"
                          ? "bg-green-500/10 text-green-400 border border-green-500/20"
                          : "bg-yellow-500/10 text-yellow-400 border border-yellow-500/20"
                      }`}>
                        {row.status}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
              <div className="border-t border-border/40 px-4 py-2.5">
                <p className="text-[10px] text-muted-foreground/60">
                  Static preview — live signals available via the Intelligence API.
                </p>
              </div>
            </div>
          </section>

          {/* ── Metric cards ─────────────────────────────────────── */}
          <section className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {[
              { label: "Settlement", value: "USDC", sub: "Collateral-backed on Arc" },
              { label: "Market Type", value: "Binary", sub: "YES / NO outcomes" },
              { label: "Architecture", value: "API-First", sub: "Agents & developers" },
              { label: "Network", value: "Arc Testnet", sub: "Chain ID 5042002" },
            ].map((metric) => (
              <Card key={metric.label} className="border-border/50 bg-card/60">
                <CardContent className="py-4">
                  <p className="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground/70">{metric.label}</p>
                  <p className="mt-1 text-lg font-bold text-indigo-400">{metric.value}</p>
                  <p className="mt-0.5 text-xs text-muted-foreground">{metric.sub}</p>
                </CardContent>
              </Card>
            ))}
          </section>

          <Separator className="opacity-40" />

          {/* ── Infrastructure surfaces ───────────────────────────── */}
          <section className="space-y-6">
            <div>
              <h2 className="text-lg font-semibold tracking-tight sm:text-xl">Infrastructure Surfaces</h2>
              <p className="mt-1 text-sm text-muted-foreground">
                Core capabilities available through the SignalArc platform and API.
              </p>
            </div>
            <div className="grid gap-4 sm:grid-cols-3">
              {infrastructureSurfaces.map((surface) => (
                <Card key={surface.title} className="group border-border/50 transition-colors hover:border-indigo-500/30">
                  <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                      <CardTitle className="text-sm font-semibold">{surface.title}</CardTitle>
                      <Badge variant="outline" className="text-[10px] border-border bg-muted/30 text-muted-foreground">
                        {surface.badge}
                      </Badge>
                    </div>
                  </CardHeader>
                  <CardContent className="pt-0">
                    <CardDescription className="text-xs leading-relaxed">
                      {surface.description}
                    </CardDescription>
                    {surface.external ? (
                      <a
                        href={surface.href}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="mt-3 inline-flex items-center gap-1 text-xs font-medium text-indigo-400 transition-colors hover:text-indigo-300"
                      >
                        View docs →
                      </a>
                    ) : (
                      <Link
                        href={surface.href}
                        className="mt-3 inline-flex items-center gap-1 text-xs font-medium text-indigo-400 transition-colors hover:text-indigo-300"
                      >
                        Open →
                      </Link>
                    )}
                  </CardContent>
                </Card>
              ))}
            </div>
          </section>

          <Separator className="opacity-40" />

          {/* ── Market lifecycle ──────────────────────────────────── */}
          <section className="space-y-6">
            <div>
              <h2 className="text-lg font-semibold tracking-tight sm:text-xl">Market Lifecycle</h2>
              <p className="mt-1 text-sm text-muted-foreground">
                Three-step lifecycle from market creation to USDC settlement.
              </p>
            </div>
            <div className="grid gap-4 sm:grid-cols-3">
              {[
                {
                  step: "01",
                  title: "Create",
                  description: "Define a binary event market with YES/NO outcomes, resolution source, and close timestamp.",
                },
                {
                  step: "02",
                  title: "Trade",
                  description: "Submit position intents with USDC collateral. Positions are tracked per market per participant.",
                },
                {
                  step: "03",
                  title: "Resolve",
                  description: "Markets resolve against their defined source. USDC settlement distributes to winning positions.",
                },
              ].map((item) => (
                <div key={item.step} className="rounded-lg border border-border/50 bg-card/40 p-5">
                  <span className="text-[10px] font-bold uppercase tracking-widest text-indigo-400/50">{item.step}</span>
                  <h3 className="mt-2 text-base font-semibold text-foreground">{item.title}</h3>
                  <p className="mt-2 text-xs leading-relaxed text-muted-foreground">{item.description}</p>
                </div>
              ))}
            </div>
          </section>

          {/* ── Network info ─────────────────────────────────────── */}
          <section>
            <details className="group">
              <summary className="cursor-pointer text-sm font-medium text-muted-foreground hover:text-foreground transition-colors">
                Network &amp; Developer Info
              </summary>
              <div className="mt-4">
                <NetworkInfoPanel />
              </div>
            </details>
          </section>
        </div>
      </div>
      <SiteFooter />
    </>
  )
}

function NetworkInfoPanel() {
  return (
    <Card className="border-border/50">
      <CardHeader>
        <div className="flex items-center gap-2">
          <CardTitle className="text-sm">Arc Testnet</CardTitle>
          <Badge variant="outline" className="border-indigo-500/30 bg-indigo-500/10 text-indigo-300 text-xs">
            Testnet
          </Badge>
        </div>
        <CardDescription className="text-xs">
          SignalArc is deployed on Arc Testnet (Chain ID 5042002). Contract details are available via the backend API.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <dl className="grid gap-3 text-sm sm:grid-cols-3">
          <div>
            <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">Network</dt>
            <dd className="mt-0.5 text-foreground">Arc Testnet</dd>
          </div>
          <div>
            <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">Chain ID</dt>
            <dd className="mt-0.5 text-foreground">5042002</dd>
          </div>
          <div>
            <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">Explorer</dt>
            <dd className="mt-0.5">
              <a
                href="https://testnet.arcscan.app"
                target="_blank"
                rel="noopener noreferrer"
                className="text-indigo-400 hover:text-indigo-300 transition-colors"
              >
                testnet.arcscan.app
              </a>
            </dd>
          </div>
        </dl>
      </CardContent>
    </Card>
  )
}
