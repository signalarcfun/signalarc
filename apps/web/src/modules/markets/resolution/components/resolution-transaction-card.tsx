import type { ReactNode } from "react"

export function ResolutionTransactionCard({
  children,
  status = "available",
  label = "Resolution transaction",
}: {
  children: ReactNode
  status?: "pending" | "success" | "error" | "available" | "not_available"
  label?: string
}) {
  const statusLabel = status.replace("_", " ")

  return (
    <div className="grid gap-2 rounded-lg border border-border bg-muted/20 p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
          {label}
        </p>
        <span className="rounded-md border border-border bg-background/50 px-2 py-1 text-xs capitalize text-muted-foreground">
          {statusLabel}
        </span>
      </div>
      <div className="min-w-0 text-sm text-muted-foreground">{children}</div>
    </div>
  )
}
