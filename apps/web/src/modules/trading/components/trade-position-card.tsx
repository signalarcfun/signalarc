import { formatUnits, type Address } from "viem"

import { USDC_ERC20_DECIMALS, formatShortAddress } from "@/lib/contracts"

function formatUsdcPosition(value: bigint | undefined) {
  if (value === undefined) return "-"
  return `${formatUnits(value, USDC_ERC20_DECIMALS)} USDC`
}

function PositionAmount({
  label,
  value,
  tone,
}: {
  label: string
  value: bigint | undefined
  tone: "yes" | "no"
}) {
  const toneClass = tone === "yes" ? "text-green-300" : "text-red-300"

  return (
    <div className="min-w-0 rounded-lg border border-border bg-background/40 p-3">
      <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
        {label}
      </dt>
      <dd className={`mt-1 break-words text-sm font-semibold ${toneClass}`}>
        {formatUsdcPosition(value)}
      </dd>
    </div>
  )
}

export function TradePositionCard({
  walletAddress,
  yesPosition,
  noPosition,
  isConnected,
}: {
  walletAddress: Address | undefined
  yesPosition: bigint | undefined
  noPosition: bigint | undefined
  isConnected: boolean
}) {
  const hasReadablePosition = yesPosition !== undefined || noPosition !== undefined
  const hasPosition =
    (yesPosition !== undefined && yesPosition > BigInt(0)) ||
    (noPosition !== undefined && noPosition > BigInt(0))

  return (
    <section
      aria-labelledby="trade-position-heading"
      className="grid gap-3 rounded-lg border border-border bg-muted/20 p-4"
    >
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div>
          <h3 id="trade-position-heading" className="text-sm font-medium">
            Your Position
          </h3>
          <p className="mt-1 text-xs text-muted-foreground">
            Connected-wallet exposure in this market.
          </p>
        </div>
        {walletAddress ? (
          <span
            className="max-w-full truncate rounded-md border border-border bg-background/60 px-2 py-1 font-mono text-xs text-muted-foreground"
            title={walletAddress}
          >
            {formatShortAddress(walletAddress)}
          </span>
        ) : null}
      </div>

      {!isConnected ? (
        <p className="rounded-md border border-border bg-background/40 p-3 text-sm text-muted-foreground">
          Connect a wallet to view your YES/NO position.
        </p>
      ) : (
        <>
          <dl className="grid gap-2 sm:grid-cols-2">
            <PositionAmount label="YES Position" value={yesPosition} tone="yes" />
            <PositionAmount label="NO Position" value={noPosition} tone="no" />
          </dl>
          <p className="text-xs text-muted-foreground">
            {hasReadablePosition
              ? hasPosition
                ? "Position values come from the market contract on Arc Testnet."
                : "No connected-wallet position is currently recorded for this market."
              : "Position data is loading from the market contract."}
          </p>
        </>
      )}
    </section>
  )
}
