import type { ReactNode } from "react"

import { Badge } from "@/components/ui/badge"

export function StatusBadge({
  children,
  className,
}: {
  children: ReactNode
  className?: string
}) {
  return (
    <Badge variant="outline" className={className}>
      {children}
    </Badge>
  )
}
