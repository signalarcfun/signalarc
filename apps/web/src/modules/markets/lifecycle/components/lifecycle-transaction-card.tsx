import type { Hash } from "viem"

import { TransactionLink } from "@/modules/wallet"

export function LifecycleTransactionLink({ hash }: { hash: Hash }) {
  return <TransactionLink hash={hash} />
}

export function LifecycleTransactionCard({
  hash,
  status = "available",
  label = "Transaction",
}: {
  hash?: Hash
  status?: "pending" | "success" | "error" | "available" | "not_available"
  label?: string
}) {
  const statusLabel = status.replace("_", " ")

  return (
    <div className="grid gap-2 rounded-lg border border-border bg-muted/20 p-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
          {label}
        </p>
        <span className="rounded-md border border-border bg-background/50 px-2 py-1 text-xs capitalize text-muted-foreground">
          {statusLabel}
        </span>
      </div>
      {hash ? (
        <LifecycleTransactionLink hash={hash} />
      ) : (
        <p className="text-sm text-muted-foreground">Transaction hash not available.</p>
      )}
    </div>
  )
}
