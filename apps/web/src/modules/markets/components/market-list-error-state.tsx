"use client"

import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"

function AlertIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth={1.5}
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z"
      />
    </svg>
  )
}

export function MarketListErrorState({
  message,
  requestId,
  onRetry,
}: {
  message: string
  requestId: string | null
  onRetry: () => void
}) {
  return (
    <Card className="border-destructive/30 bg-destructive/5">
      <CardContent className="flex flex-col items-center gap-4 py-10 text-center">
        <div className="flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10">
          <AlertIcon className="h-6 w-6 text-destructive" />
        </div>

        <div className="space-y-1">
          <h2 className="text-base font-medium text-destructive">
            Unable to load markets
          </h2>
          <p className="text-sm text-muted-foreground">{message}</p>
        </div>

        {requestId ? (
          <p className="font-mono text-xs text-muted-foreground">
            Request ID: {requestId}
          </p>
        ) : null}

        <Button variant="outline" size="sm" onClick={onRetry}>
          Retry
        </Button>
      </CardContent>
    </Card>
  )
}
