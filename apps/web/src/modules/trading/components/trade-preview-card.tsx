import type { Address } from "viem"

import { formatShortAddress } from "@/lib/contracts"

import type { TradeOutcome } from "../types"

function PreviewRow({
  label,
  value,
  title,
  mono = false,
}: {
  label: string
  value: string
  title?: string
  mono?: boolean
}) {
  return (
    <div className="min-w-0 rounded-lg border border-border bg-background/40 p-3">
      <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
        {label}
      </dt>
      <dd
        className={`mt-1 break-words text-sm font-medium text-foreground ${
          mono ? "font-mono text-xs" : ""
        }`}
        title={title}
      >
        {value}
      </dd>
    </div>
  )
}

export function TradePreviewCard({
  contractAddress,
  marketId,
  walletAddress,
  outcome,
  amount,
  parsedAmount,
}: {
  contractAddress: Address | null
  marketId: string
  walletAddress: Address | undefined
  outcome: TradeOutcome
  amount: string
  parsedAmount: bigint | null
}) {
  return (
    <section
      aria-labelledby="trade-preview-heading"
      className="grid gap-3 rounded-lg border border-border bg-muted/20 p-4 text-sm"
    >
      <div>
        <h3 id="trade-preview-heading" className="text-sm font-medium">
          Trade Preview
        </h3>
        <p className="mt-1 text-xs text-muted-foreground">
          This opens a {outcome} position on Arc Testnet after USDC approval.
        </p>
      </div>
      <dl className="grid gap-2 sm:grid-cols-2">
        <PreviewRow label="Selected side" value={outcome} />
        <PreviewRow label="Order amount" value={`${amount || "0"} USDC`} />
        <PreviewRow
          label="Base units"
          value={parsedAmount ? parsedAmount.toString() : "-"}
          mono
        />
        <PreviewRow
          label="Contract"
          value={contractAddress ? formatShortAddress(contractAddress) : "Not deployed"}
          title={contractAddress ?? undefined}
          mono
        />
        <PreviewRow
          label="Market ID"
          value={`${marketId.slice(0, 8)}...${marketId.slice(-4)}`}
          title={marketId}
          mono
        />
        {walletAddress ? (
          <PreviewRow
            label="Connected wallet"
            value={formatShortAddress(walletAddress)}
            title={walletAddress}
            mono
          />
        ) : null}
      </dl>
    </section>
  )
}
