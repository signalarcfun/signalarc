import { ErrorState } from "@/components/shared"

export function PortfolioErrorState({
  message,
  requestId,
}: {
  message: string
  requestId: string | null
}) {
  return (
    <ErrorState
      title="Unable to load portfolio"
      message={message}
      requestId={requestId}
      contentClassName="p-6"
      titleClassName="text-base font-medium text-destructive"
      messageClassName="mt-2 text-sm text-muted-foreground"
      requestIdClassName="mt-3 font-mono text-xs text-muted-foreground"
    />
  )
}
