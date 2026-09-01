import { ErrorState } from "@/components/shared"

export function IntelligenceErrorState({
  message,
  requestId,
}: {
  message: string
  requestId: string | null
}) {
  return (
    <ErrorState
      title="Unable to load market signals"
      message={message}
      requestId={requestId}
    />
  )
}
