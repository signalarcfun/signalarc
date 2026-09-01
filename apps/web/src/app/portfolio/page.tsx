import { PortfolioView } from "@/features/portfolio/portfolio-view"

export default function PortfolioPage() {
  return (
    <div className="px-4 py-5 sm:px-6 sm:py-6 lg:px-8">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-5 sm:gap-6">
        <header className="space-y-2">
          <p className="text-xs font-semibold uppercase tracking-wider text-indigo-300/80">
            Wallet Activity
          </p>
          <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
            Portfolio
          </h1>
          <p className="max-w-2xl text-sm leading-6 text-muted-foreground">
            Review connected-wallet context, active position records, and
            settlement activity from the existing SignalArc API data.
          </p>
        </header>
        <PortfolioView />
      </div>
    </div>
  )
}
