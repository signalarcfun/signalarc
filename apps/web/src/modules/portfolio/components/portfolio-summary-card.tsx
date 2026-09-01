import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { InlineErrorState } from "@/components/shared"

import {
  formatPortfolioAmount,
  formatWalletAddress,
} from "../format"
import type { MarketsState, PortfolioState } from "../types"

function sumPositionExposure(
  portfolioState: PortfolioState,
): { value: string; description: string } {
  if (portfolioState.status === "empty") {
    return {
      value: "0",
      description: "No loaded position exposure.",
    }
  }

  if (portfolioState.status !== "loaded") {
    return {
      value: "-",
      description: "Load API position records to calculate exposure.",
    }
  }

  const total = portfolioState.data.positions.reduce(
    (sum, position) => sum + Number(position.quantity || 0),
    0,
  )

  return {
    value: formatPortfolioAmount(total),
    description: "Sum of loaded position quantities.",
  }
}

function MetricTile({
  label,
  value,
  description,
}: {
  label: string
  value: string
  description: string
}) {
  return (
    <div className="min-w-0 rounded-lg border border-border bg-muted/20 p-3">
      <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
        {label}
      </p>
      <p className="mt-2 break-words text-xl font-semibold text-foreground">
        {value}
      </p>
      <p className="mt-1 text-xs leading-5 text-muted-foreground">
        {description}
      </p>
    </div>
  )
}

export function PortfolioSummaryCard({
  address,
  isConnected,
  marketsState,
  portfolioState,
}: {
  address: string | undefined
  isConnected: boolean
  marketsState: MarketsState
  portfolioState: PortfolioState
}) {
  const activePositions =
    portfolioState.status === "loaded"
      ? portfolioState.data.positions.length
      : portfolioState.status === "empty"
        ? 0
        : null
  const settlementCount =
    portfolioState.status === "loaded"
      ? portfolioState.data.settlements.length
      : portfolioState.status === "empty"
        ? 0
        : null
  const exposure = sumPositionExposure(portfolioState)

  return (
    <Card>
      <CardHeader>
        <CardTitle>Wallet Overview</CardTitle>
        <CardDescription className="leading-6">
          Portfolio data currently combines wallet identity with API position and
          settlement records when a backend user ID is loaded.
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-4">
        {marketsState.status === "error" ? (
          <InlineErrorState
            title="Unable to load market context"
            message={marketsState.message}
            requestId={marketsState.requestId}
          />
        ) : null}

        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <MetricTile
            label="Connected Wallet"
            value={isConnected && address ? formatWalletAddress(address) : "Not connected"}
            description={
              isConnected
                ? "Wallet identity is available for onchain context."
                : "Connect a wallet to anchor this portfolio view."
            }
          />
          <MetricTile
            label="Active Positions"
            value={activePositions === null ? "-" : String(activePositions)}
            description={
              activePositions === null
                ? "Load API records to show positions."
                : "Loaded position rows from the backend."
            }
          />
          <MetricTile
            label="Total Exposure"
            value={exposure.value}
            description={exposure.description}
          />
          <MetricTile
            label="Claim / Refund Records"
            value={settlementCount === null ? "-" : String(settlementCount)}
            description={
              settlementCount === null
                ? "Load settlement records for claim history."
                : "Loaded settlement rows; live claimability remains onchain."
            }
          />
        </div>

        <div className="rounded-lg border border-border bg-background/40 p-3 text-sm text-muted-foreground">
          {marketsState.status === "loading" ? (
            <span>Loading market metadata for portfolio cards...</span>
          ) : null}
          {marketsState.status === "loaded" ? (
            <span>
              {marketsState.markets.length} market
              {marketsState.markets.length === 1 ? "" : "s"} available for title,
              status, close time, and detail links.
            </span>
          ) : null}
          {marketsState.status === "error" ? (
            <span>Position cards will use raw market IDs until market metadata loads.</span>
          ) : null}
        </div>
      </CardContent>
    </Card>
  )
}
