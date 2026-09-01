import type { ReactNode } from "react"

import { Card, CardContent } from "@/components/ui/card"

export function ErrorState({
  title,
  message,
  requestId,
  children,
  cardClassName = "border-destructive/30 bg-destructive/5",
  contentClassName = "pt-6",
  titleClassName = "text-sm font-semibold text-destructive",
  messageClassName = "mt-2 text-xs text-muted-foreground",
  requestIdClassName = "mt-3 font-mono text-[10px] text-muted-foreground",
}: {
  title: ReactNode
  message?: ReactNode
  requestId?: string | null
  children?: ReactNode
  cardClassName?: string
  contentClassName?: string
  titleClassName?: string
  messageClassName?: string
  requestIdClassName?: string
}) {
  return (
    <Card className={cardClassName}>
      <CardContent className={contentClassName}>
        <h2 className={titleClassName}>{title}</h2>
        {message ? <p className={messageClassName}>{message}</p> : null}
        {requestId ? (
          <p className={requestIdClassName}>Request ID: {requestId}</p>
        ) : null}
        {children}
      </CardContent>
    </Card>
  )
}

export function InlineErrorState({
  title,
  message,
  requestId,
  className = "rounded-lg border border-destructive/30 bg-destructive/5 p-4",
  titleClassName = "text-sm font-medium text-destructive",
  messageClassName = "mt-1 text-sm text-muted-foreground",
  requestIdClassName = "mt-2 font-mono text-xs text-muted-foreground",
}: {
  title: ReactNode
  message?: ReactNode
  requestId?: string | null
  className?: string
  titleClassName?: string
  messageClassName?: string
  requestIdClassName?: string
}) {
  return (
    <div className={className}>
      <p className={titleClassName}>{title}</p>
      {message ? <p className={messageClassName}>{message}</p> : null}
      {requestId ? (
        <p className={requestIdClassName}>Request ID: {requestId}</p>
      ) : null}
    </div>
  )
}
