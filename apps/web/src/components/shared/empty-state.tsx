import type { ReactNode } from "react"

import { Card, CardContent } from "@/components/ui/card"

export function EmptyState({
  title,
  description,
  children,
  cardClassName,
  contentClassName = "py-8 text-center",
  titleClassName = "text-sm text-muted-foreground",
  descriptionClassName = "mt-1 text-xs text-muted-foreground",
}: {
  title: ReactNode
  description?: ReactNode
  children?: ReactNode
  cardClassName?: string
  contentClassName?: string
  titleClassName?: string
  descriptionClassName?: string
}) {
  return (
    <Card className={cardClassName}>
      <CardContent className={contentClassName}>
        <p className={titleClassName}>{title}</p>
        {description ? <p className={descriptionClassName}>{description}</p> : null}
        {children}
      </CardContent>
    </Card>
  )
}
