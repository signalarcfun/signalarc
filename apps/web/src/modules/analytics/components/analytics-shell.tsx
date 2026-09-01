"use client"

import { useMemo, type ComponentProps, type ReactNode } from "react"
import {
  ArrowUpRight,
  Bot,
  Check,
  CircleDollarSign,
  Database,
  ExternalLink,
  Factory,
  FileCheck2,
  Network,
  ShieldCheck,
  WalletCards,
} from "lucide-react"

import { SiteFooter } from "@/components/layout/site-footer"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"

import {
  analyticsAgentIntegrationChecklist,
  analyticsBackendMetrics,
  buildAnalyticsLifecycleMetrics,
  buildAnalyticsMetrics,
  formatAnalyticsTimestamp,
  getAnalyticsFactoryAddress,
  getAnalyticsFactoryExplorerUrl,
  analyticsFactory,
  analyticsFactoryProofPoints,
  analyticsLatestActivity,
  analyticsLifecycleMetrics,
  analyticsLimitations,
  analyticsMetrics,
  analyticsPublicLinks,
  analyticsStatusBadges,
  isIndexedAnalyticsSummary,
} from "../analytics-utils"
import type { AnalyticsMetric, AnalyticsSummaryResponse } from "../types"
import { useAnalyticsSummary } from "../use-analytics-summary"

import { AnalyticsDisclaimerCard } from "./analytics-disclaimer-card"
import { AnalyticsEmptyState } from "./analytics-empty-state"
import { AnalyticsErrorState } from "./analytics-error-state"
import { AnalyticsLoadingSkeleton } from "./analytics-loading-skeleton"
import { AnalyticsSummaryGrid } from "./analytics-summary-grid"

type ButtonVariant = ComponentProps<typeof Button>["variant"]

function SectionHeading({ title, description }: { title: string; description: string }) {
  return (
    <div className="flex max-w-3xl flex-col gap-2">
      <h2 className="text-xl font-semibold tracking-tight sm:text-2xl">{title}</h2>
      <p className="text-sm leading-relaxed text-muted-foreground">{description}</p>
    </div>
  )
}

function ExternalButton({
  href,
  children,
  variant = "outline",
}: {
  href: string
  children: ReactNode
  variant?: ButtonVariant
}) {
  return (
    <Button asChild variant={variant}>
      <a href={href} target="_blank" rel="noreferrer">
        {children}
        <ExternalLink data-icon="inline-end" />
      </a>
    </Button>
  )
}

function DataRow({
  label,
  value,
  mono = false,
}: {
  label: string
  value: ReactNode
  mono?: boolean
}) {
  return (
    <div className="grid gap-1 border-b border-border/40 py-3 last:border-b-0 sm:grid-cols-[180px_1fr] sm:gap-5">
      <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground">{label}</dt>
      <dd className={mono ? "min-w-0 break-all font-mono text-xs text-foreground" : "text-sm text-foreground"}>
        {value}
      </dd>
    </div>
  )
}

function getMetricValue(metrics: AnalyticsMetric[], label: string, fallback: string) {
  return metrics.find((metric) => metric.label === label)?.value ?? fallback
}

function AnalyticsHero({
  summary,
  metrics,
  isLive,
}: {
  summary: AnalyticsSummaryResponse | null
  metrics: AnalyticsMetric[]
  isLive: boolean
}) {
  const createdMarkets = getMetricValue(metrics, "Markets Created", "126")
  const testnetUsdc = getMetricValue(metrics, "Testnet USDC Collateral Volume", "149.77")

  return (
    <section className="grid gap-10 lg:grid-cols-[minmax(0,1.3fr)_minmax(320px,0.7fr)] lg:items-end">
      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-2">
          <p className="text-sm font-medium text-indigo-300">Arc Testnet Proof-of-Activity Dashboard</p>
          <h1 className="max-w-4xl text-4xl font-bold tracking-tight sm:text-5xl lg:text-6xl">
            SignalArc{" "}
            <span className="bg-gradient-to-r from-indigo-300 via-purple-300 to-indigo-400 bg-clip-text text-transparent">
              Analytics
            </span>
          </h1>
        </div>

        <p className="max-w-3xl text-base leading-relaxed text-muted-foreground sm:text-lg">
          SignalArc publishes this dashboard as a transparent proof-of-activity layer for its Arc Testnet deployment.
          The dashboard summarizes verified factory deployment, created YES/NO market contracts, market-level position
          activity, testnet USDC collateral movement, lifecycle events, and Circle Agent Wallet integration readiness.
        </p>

        <div className="flex flex-wrap gap-2">
          {analyticsStatusBadges.map((badge, index) => (
            <Badge
              key={badge}
              variant="outline"
              className={
                index < 3
                  ? "border-indigo-500/30 bg-indigo-500/10 text-indigo-200"
                  : "border-border/70 bg-card/50 text-muted-foreground"
              }
            >
              {badge}
            </Badge>
          ))}
        </div>
      </div>

      <Card className="border-indigo-500/20 bg-card/70 shadow-2xl shadow-indigo-950/20 backdrop-blur">
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="flex size-9 items-center justify-center rounded-lg bg-indigo-500/10 text-indigo-300 ring-1 ring-indigo-500/20">
              <Database className="size-4" aria-hidden="true" />
            </div>
            <div>
              <CardTitle>Data provenance</CardTitle>
              <CardDescription>
                {isLive ? "Backend indexed analytics cache" : "Static fallback analytics snapshot"}
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <p className="text-sm leading-relaxed text-muted-foreground">
            {isLive
              ? "This page reads SignalArc's public backend analytics summary, which is rebuilt from Arcscan/Blockscout event ingestion owned by the backend."
              : "The live backend analytics cache is unavailable or not indexed yet, so this page keeps the historical static snapshot visible as a fallback."}
          </p>
          <div className="grid grid-cols-2 gap-3 border-t border-border/50 pt-4">
            <div>
              <p className="text-2xl font-semibold text-foreground">{createdMarkets}</p>
              <p className="text-xs text-muted-foreground">Created markets</p>
            </div>
            <div>
              <p className="text-2xl font-semibold text-foreground">{testnetUsdc}</p>
              <p className="text-xs text-muted-foreground">Testnet USDC</p>
            </div>
          </div>
          <div className="border-t border-border/50 pt-4">
            <p className="break-all font-mono text-xs text-muted-foreground">
              Factory: {getAnalyticsFactoryAddress(summary)}
            </p>
          </div>
        </CardContent>
      </Card>
    </section>
  )
}

function AnalyticsFreshnessSection({
  summary,
  isLive,
  status,
}: {
  summary: AnalyticsSummaryResponse | null
  isLive: boolean
  status: "loading" | "loaded" | "error"
}) {
  const sourceStatus = summary?.source_status ?? (status === "loading" ? "loading" : "fallback")
  const generatedAt = summary ? formatAnalyticsTimestamp(summary.generated_at) : "Not available"
  const latestEventAt = summary ? formatAnalyticsTimestamp(summary.latest_event_at) : "Not available"
  const latestBlock = summary?.latest_block ? summary.latest_block.toLocaleString("en-US") : "Not indexed yet"

  return (
    <Card className="border-border/60 bg-card/60">
      <CardContent className="grid gap-4 py-5 sm:grid-cols-2 lg:grid-cols-5">
        <div>
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">Source status</p>
          <div className="mt-2 flex flex-wrap items-center gap-2">
            <Badge
              variant="outline"
              className={
                isLive
                  ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
                  : "border-amber-500/30 bg-amber-500/10 text-amber-300"
              }
            >
              {sourceStatus}
            </Badge>
            {!isLive ? <span className="text-xs text-muted-foreground">Static fallback visible</span> : null}
          </div>
        </div>
        <div>
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">Generated at</p>
          <p className="mt-2 text-sm text-foreground">{generatedAt}</p>
        </div>
        <div>
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">Latest event</p>
          <p className="mt-2 text-sm text-foreground">{latestEventAt}</p>
        </div>
        <div>
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">Latest block</p>
          <p className="mt-2 text-sm text-foreground">{latestBlock}</p>
        </div>
        <div>
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">Factory</p>
          <a
            href={getAnalyticsFactoryExplorerUrl(summary)}
            target="_blank"
            rel="noreferrer"
            className="mt-2 block break-all font-mono text-xs text-indigo-200 hover:text-indigo-100"
          >
            {getAnalyticsFactoryAddress(summary)}
          </a>
        </div>
      </CardContent>
    </Card>
  )
}

function ExecutiveMetricsSection({
  state,
  metrics,
  isLive,
}: {
  state: ReturnType<typeof useAnalyticsSummary>
  metrics: AnalyticsMetric[]
  isLive: boolean
}) {
  return (
    <section className="flex flex-col gap-7">
      <SectionHeading
        title="Executive Metrics"
        description="A contract-derived summary of verified deployment, created YES/NO markets, participating wallets, testnet collateral movement, and market lifecycle activity."
      />
      {state.status === "loading" ? <AnalyticsLoadingSkeleton /> : null}
      {state.status === "error" ? (
        <AnalyticsErrorState
          message={`${state.message} The historical static analytics snapshot is shown below as a fallback.`}
        />
      ) : null}
      {state.status === "loaded" && !isLive ? <AnalyticsEmptyState /> : null}
      {state.status !== "loading" ? <AnalyticsSummaryGrid metrics={metrics} /> : null}
    </section>
  )
}

function FactorySection({ summary, isLive }: { summary: AnalyticsSummaryResponse | null; isLive: boolean }) {
  const factoryAddress = getAnalyticsFactoryAddress(summary)
  const factoryExplorerUrl = getAnalyticsFactoryExplorerUrl(summary)

  return (
    <section className="flex flex-col gap-7">
      <SectionHeading
        title={isLive ? "Active Analytics Factory" : "Legacy Verified Factory Snapshot"}
        description={
          isLive
            ? "Live analytics are keyed to the active factory used by the backend indexer. Child market activity is aggregated from contracts discovered through that factory."
            : "When the backend cache is unavailable or not indexed, SignalArc keeps the historical Arc Testnet proof-of-activity snapshot visible."
        }
      />

      <div className="grid gap-5 lg:grid-cols-[minmax(0,1.25fr)_minmax(280px,0.75fr)]">
        <Card className="border-indigo-500/20 bg-card/70">
          <CardHeader>
            <div className="flex items-start justify-between gap-4">
              <div className="flex items-center gap-3">
                <div className="flex size-10 items-center justify-center rounded-lg bg-indigo-500/10 text-indigo-300 ring-1 ring-indigo-500/20">
                  <Factory className="size-5" aria-hidden="true" />
                </div>
                <div>
                  <CardTitle>{isLive ? "Active SignalArcMarketFactory" : analyticsFactory.name}</CardTitle>
                  <CardDescription>Canonical market deployment registry</CardDescription>
                </div>
              </div>
              <Badge variant="outline" className="border-emerald-500/30 bg-emerald-500/10 text-emerald-300">
                <ShieldCheck data-icon="inline-start" />
                Verified
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <dl>
              <DataRow label="Contract name" value={isLive ? "SignalArcMarketFactory" : analyticsFactory.name} />
              <DataRow label="Address" value={factoryAddress} mono />
              <DataRow label="Verified" value="Yes" />
              <DataRow label="Network" value="Arc Testnet" />
              {isLive ? (
                <>
                  <DataRow label="Source status" value={summary?.source_status ?? "not_indexed"} />
                  <DataRow label="Latest indexed block" value={summary?.latest_block?.toLocaleString("en-US") ?? "Not indexed yet"} />
                  <DataRow label="Latest indexed event" value={formatAnalyticsTimestamp(summary?.latest_event_at ?? null)} mono />
                  <DataRow label="Cache generated" value={formatAnalyticsTimestamp(summary?.generated_at ?? null)} mono />
                  <DataRow
                    label="Market contracts"
                    value={summary?.metrics.market_contracts_found.toLocaleString("en-US") ?? "0"}
                  />
                </>
              ) : (
                <>
                  <DataRow label="Deployment tx" value={analyticsFactory.deploymentTx} mono />
                  <DataRow label="Deployer" value={analyticsFactory.deployer} mono />
                  <DataRow label="Deployment block" value={analyticsFactory.deploymentBlock} />
                  <DataRow label="Deployment timestamp" value={analyticsFactory.deploymentTimestamp} mono />
                  <DataRow label="Factory transactions" value={analyticsFactory.totalTransactions} />
                  <DataRow label="Latest market created" value={analyticsFactory.latestMarket} mono />
                </>
              )}
            </dl>
          </CardContent>
        </Card>

        <div className="flex flex-col gap-5">
          <Card className="border-border/60 bg-card/60">
            <CardHeader>
              <CardTitle>Factory proof points</CardTitle>
              <CardDescription>Explorer-verifiable deployment references</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
              {analyticsFactoryProofPoints.map((item) => (
                <div key={item.label} className="flex items-center gap-3">
                  <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-muted text-indigo-300">
                    <Factory className="size-4" aria-hidden="true" />
                  </div>
                  <div>
                    <p className="text-sm font-medium">{item.value}</p>
                    <p className="text-xs text-muted-foreground">{item.label}</p>
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>

          <div className="flex flex-col gap-2">
            <ExternalButton href={factoryExplorerUrl}>View Factory on Arcscan</ExternalButton>
            {!isLive ? (
              <>
                <ExternalButton href={analyticsFactory.deploymentTxUrl}>View Legacy Deployment Transaction</ExternalButton>
                <ExternalButton href={analyticsLatestActivity.txUrl}>View Legacy Latest Activity</ExternalButton>
              </>
            ) : null}
          </div>
        </div>
      </div>
    </section>
  )
}

function MarketsSection() {
  return null
}

function LifecycleSection({
  lifecycleMetrics,
  summary,
  isLive,
}: {
  lifecycleMetrics: ReadonlyArray<{ label: string; value: string }>
  summary: AnalyticsSummaryResponse | null
  isLive: boolean
}) {
  return (
    <section className="flex flex-col gap-7">
      <SectionHeading
        title="Lifecycle Activity"
        description="SignalArc's Arc Testnet deployment includes market creation, position activity, market resolution, cancellation, and claim/refund lifecycle events. These metrics demonstrate end-to-end contract lifecycle coverage across the testnet deployment."
      />

      <div className="grid gap-5 lg:grid-cols-[0.8fr_1.2fr]">
        <div className="grid grid-cols-3 gap-3 self-start">
          {lifecycleMetrics.map((item) => (
            <Card key={item.label} size="sm" className="border-border/60 bg-card/60">
              <CardHeader>
                <CardTitle className="text-3xl font-semibold">{item.value}</CardTitle>
                <CardDescription className="text-xs">{item.label}</CardDescription>
              </CardHeader>
            </Card>
          ))}
        </div>

        <Card className="border-indigo-500/20 bg-gradient-to-r from-card to-indigo-500/5">
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="flex size-9 items-center justify-center rounded-lg bg-indigo-500/10 text-indigo-300">
                <Network className="size-4" aria-hidden="true" />
              </div>
              <div>
                <CardTitle>Latest recorded activity</CardTitle>
                <CardDescription>
                  {isLive ? "Most recent indexed analytics event" : "Most recent market creation in the fallback snapshot"}
                </CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <dl>
              <DataRow
                label="Timestamp"
                value={isLive ? formatAnalyticsTimestamp(summary?.latest_event_at ?? null) : analyticsLatestActivity.timestamp}
                mono
              />
              <DataRow
                label={isLive ? "Latest block" : "Transaction"}
                value={isLive ? summary?.latest_block?.toLocaleString("en-US") ?? "Not indexed yet" : analyticsLatestActivity.tx}
                mono
              />
              <DataRow
                label={isLive ? "Factory" : "Market"}
                value={isLive ? getAnalyticsFactoryAddress(summary) : analyticsLatestActivity.market}
                mono
              />
            </dl>
            <div className="pt-4">
              <ExternalButton href={isLive ? getAnalyticsFactoryExplorerUrl(summary) : analyticsLatestActivity.txUrl} variant="default">
                Inspect {isLive ? "Factory" : "Latest Activity"}
              </ExternalButton>
            </div>
          </CardContent>
        </Card>
      </div>
    </section>
  )
}

function AgentWalletSection() {
  return (
    <section className="flex flex-col gap-7">
      <SectionHeading
        title="Circle Agent Wallet Integration"
        description="SignalArc integrates Circle Agent Wallet / Programmable Wallet flows in the backend for agent wallet onboarding, OTP verification, wallet resolution, session activation, wallet balance lookup, faucet requests, and backend-driven smart contract execution on Arc Testnet."
      />

      <div className="grid gap-5 lg:grid-cols-2">
        <Card className="border-indigo-500/20 bg-card/70">
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="flex size-10 items-center justify-center rounded-lg bg-indigo-500/10 text-indigo-300">
                <Bot className="size-5" aria-hidden="true" />
              </div>
              <div>
                <CardTitle>Integration readiness</CardTitle>
                <CardDescription>Backend and agent execution surfaces</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="grid gap-3 sm:grid-cols-2">
            {analyticsAgentIntegrationChecklist.map((item) => (
              <div key={item} className="flex items-start gap-2 rounded-lg border border-border/50 bg-background/30 p-3">
                <div className="mt-0.5 flex size-5 shrink-0 items-center justify-center rounded-full bg-indigo-500/10 text-indigo-300">
                  <Check className="size-3" aria-hidden="true" />
                </div>
                <span className="text-sm leading-snug">{item}</span>
              </div>
            ))}
          </CardContent>
        </Card>

        <div className="flex flex-col gap-5">
          <Card className="border-amber-500/20 bg-amber-500/5">
            <CardHeader>
              <div className="flex items-center gap-3">
                <WalletCards className="size-5 text-amber-300" aria-hidden="true" />
                <CardTitle>Attribution boundary</CardTitle>
              </div>
            </CardHeader>
            <CardContent>
              <p className="text-sm leading-relaxed text-muted-foreground">
                Direct Circle Agent Wallet attribution is not visible from chain data alone. On-chain analytics show
                wallet and contract interactions, while Circle Agent Wallet session, onboarding, and intent metrics are
                sourced from SignalArc backend data.
              </p>
            </CardContent>
          </Card>

          <Card className="border-border/60 bg-card/60">
            <CardHeader>
              <CardTitle>Backend-sourced metrics to add</CardTitle>
              <CardDescription>
                These operational metrics require backend references and are not represented as onchain totals.
              </CardDescription>
            </CardHeader>
            <CardContent className="grid gap-x-5 gap-y-2 sm:grid-cols-2">
              {analyticsBackendMetrics.map((metric) => (
                <div key={metric} className="flex items-center gap-2 text-xs text-muted-foreground">
                  <span className="size-1.5 rounded-full bg-indigo-400/70" />
                  {metric}
                </div>
              ))}
            </CardContent>
          </Card>
        </div>
      </div>
    </section>
  )
}

function PublicLinksSection({ metrics, isLive }: { metrics: AnalyticsMetric[]; isLive: boolean }) {
  const createdMarkets = getMetricValue(metrics, "Markets Created", "126")
  const totalTrades = getMetricValue(metrics, "Total Trades", "806")
  const uniqueWallets = getMetricValue(metrics, "Unique Participating Wallets", "218")
  const testnetUsdc = getMetricValue(metrics, "Testnet USDC Collateral Volume", "149.77")
  const resolvedMarkets = getMetricValue(metrics, "Resolved Markets", "105")
  const claimEvents = getMetricValue(metrics, "Claim Events", "393")

  return (
    <section className="grid gap-5 lg:grid-cols-[1.1fr_0.9fr]">
      <Card className="border-border/60 bg-card/60">
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="flex size-9 items-center justify-center rounded-lg bg-muted text-indigo-300">
              <ArrowUpRight className="size-4" aria-hidden="true" />
            </div>
            <div>
              <CardTitle>Public Links</CardTitle>
              <CardDescription>Product, documentation, source, and explorer references</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="grid gap-2 sm:grid-cols-2">
          {analyticsPublicLinks.map((link) => (
            <a
              key={link.label}
              href={link.href}
              target="_blank"
              rel="noreferrer"
              className="group flex items-center justify-between gap-3 rounded-lg border border-border/50 bg-background/30 px-3 py-3 text-sm transition-colors hover:border-indigo-500/30 hover:bg-indigo-500/5"
            >
              <span>{link.label}</span>
              <ArrowUpRight
                className="size-4 text-muted-foreground transition-colors group-hover:text-indigo-300"
                aria-hidden="true"
              />
            </a>
          ))}
        </CardContent>
      </Card>

      <AnalyticsDisclaimerCard
        title="Proof, not production claims"
        description="Arc Testnet context is preserved throughout"
        className="border-indigo-500/20 bg-gradient-to-br from-indigo-500/10 via-card to-purple-500/5"
        icon={
          <div className="flex size-9 items-center justify-center rounded-lg bg-indigo-500/10 text-indigo-300">
            <CircleDollarSign className="size-4" aria-hidden="true" />
          </div>
        }
      >
        <p className="text-sm leading-relaxed text-muted-foreground">
          The strongest signals in this {isLive ? "backend cache" : "fallback snapshot"} are a verified factory,{" "}
          {createdMarkets} created markets, {totalTrades} total trades, {uniqueWallets} unique participating wallets,{" "}
          {testnetUsdc} testnet USDC in aggregate collateral movement, {resolvedMarkets} resolved markets, and{" "}
          {claimEvents} claim events.
        </p>
      </AnalyticsDisclaimerCard>
    </section>
  )
}

function NotesSection() {
  return (
    <section className="grid gap-5 lg:grid-cols-2">
      <AnalyticsDisclaimerCard
        title="Dune Availability Note"
        description="Current Arc Testnet indexing constraint"
        icon={
          <div className="flex size-9 items-center justify-center rounded-lg bg-muted text-indigo-300">
            <Database className="size-4" aria-hidden="true" />
          </div>
        }
      >
        <div className="flex flex-col gap-4 text-sm leading-relaxed text-muted-foreground">
          <p>
            SignalArc currently operates on Arc Testnet. At this stage, Arc Testnet data is not yet reliably queryable
            through Dune for the contract-level analytics required by this dashboard. For that reason, SignalArc
            publishes this public proof-of-activity page using Arcscan-derived contract data, verified factory
            references, market-level activity, backend integration references, and explorer links.
          </p>
          <p>
            A Dune dashboard or another third-party analytics dashboard will be added once Arc Testnet indexing becomes
            reliably available.
          </p>
        </div>
      </AnalyticsDisclaimerCard>

      <AnalyticsDisclaimerCard
        title="Methodology and Limitations"
        description="Transparent interpretation boundaries"
        icon={
          <div className="flex size-9 items-center justify-center rounded-lg bg-muted text-indigo-300">
            <FileCheck2 className="size-4" aria-hidden="true" />
          </div>
        }
      >
        <ul className="flex flex-col gap-3">
          {analyticsLimitations.map((limitation) => (
            <li key={limitation} className="flex items-start gap-3 text-sm leading-relaxed text-muted-foreground">
              <span className="mt-2 size-1.5 shrink-0 rounded-full bg-indigo-400/70" />
              {limitation}
            </li>
          ))}
        </ul>
      </AnalyticsDisclaimerCard>
    </section>
  )
}

export function AnalyticsShell() {
  const state = useAnalyticsSummary()
  const liveSummary =
    state.status === "loaded" && isIndexedAnalyticsSummary(state.summary) ? state.summary : null
  const isLive = Boolean(liveSummary)
  const metrics = useMemo(
    () => (liveSummary ? buildAnalyticsMetrics(liveSummary) : analyticsMetrics),
    [liveSummary],
  )
  const lifecycleMetrics = useMemo(
    () => (liveSummary ? buildAnalyticsLifecycleMetrics(liveSummary) : analyticsLifecycleMetrics),
    [liveSummary],
  )
  const freshnessStatus = state.status === "loaded" ? "loaded" : state.status

  return (
    <>
      <div className="relative overflow-hidden">
        <div className="pointer-events-none absolute inset-x-0 top-0 h-[560px] bg-[radial-gradient(circle_at_top_left,rgba(99,102,241,0.12),transparent_38%),radial-gradient(circle_at_82%_10%,rgba(168,85,247,0.1),transparent_34%)]" />

        <div className="relative px-4 py-10 sm:px-6 sm:py-14 lg:px-8">
          <div className="mx-auto flex w-full max-w-7xl flex-col gap-16 sm:gap-20">
            <AnalyticsHero summary={liveSummary} metrics={metrics} isLive={isLive} />
            <AnalyticsFreshnessSection summary={state.status === "loaded" ? state.summary : null} isLive={isLive} status={freshnessStatus} />
            <ExecutiveMetricsSection state={state} metrics={metrics} isLive={isLive} />
            <Separator className="opacity-40" />
            <FactorySection summary={liveSummary} isLive={isLive} />
            <MarketsSection />
            <LifecycleSection lifecycleMetrics={lifecycleMetrics} summary={liveSummary} isLive={isLive} />
            <Separator className="opacity-40" />
            <AgentWalletSection />
            <PublicLinksSection metrics={metrics} isLive={isLive} />
            <NotesSection />
          </div>
        </div>
      </div>
      <SiteFooter />
    </>
  )
}
