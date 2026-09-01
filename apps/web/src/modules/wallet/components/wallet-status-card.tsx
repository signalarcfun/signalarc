import type { ReactNode } from "react"

const toneClassNames = {
  warning: "border-yellow-500/30 bg-yellow-500/5 text-yellow-300",
  success: "border-green-500/20 bg-green-500/5 text-green-300",
  error: "border-destructive/30 bg-destructive/5 text-destructive",
  info: "border-blue-500/20 bg-blue-500/5 text-blue-200",
} as const

export function WalletStatusCard({
  title,
  children,
  tone = "warning",
}: {
  title: string
  children?: ReactNode
  tone?: keyof typeof toneClassNames
}) {
  return (
    <div className={`rounded-lg border p-4 ${toneClassNames[tone]}`}>
      <p className="text-sm font-medium">{title}</p>
      {children ? <div className="mt-1 text-sm text-muted-foreground">{children}</div> : null}
    </div>
  )
}
