import type { ReactNode } from "react"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export function MetricCard({
  label,
  value,
  description,
  unit,
  featured = false,
}: {
  label: ReactNode
  value: ReactNode
  description?: ReactNode
  unit?: ReactNode
  featured?: boolean
}) {
  return (
    <Card
      size="sm"
      className={
        featured
          ? "border-indigo-500/30 bg-gradient-to-b from-indigo-500/10 to-card"
          : "border-border/60 bg-card/60"
      }
    >
      <CardHeader>
        <CardDescription className="text-xs font-medium">{label}</CardDescription>
        <CardTitle className="text-3xl font-semibold tracking-tight text-foreground">
          {value}
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col justify-between gap-3">
        {description ? <p className="text-xs leading-relaxed text-muted-foreground">{description}</p> : null}
        {unit ? (
          <p className="text-[10px] font-semibold uppercase tracking-widest text-indigo-300/70">
            {unit}
          </p>
        ) : null}
      </CardContent>
    </Card>
  )
}
