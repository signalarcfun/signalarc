import { ErrorState } from "@/components/shared"

export function AnalyticsErrorState({ message }: { message: string }) {
  return (
    <ErrorState title="Unable to load analytics snapshot" message={message} />
  )
}
