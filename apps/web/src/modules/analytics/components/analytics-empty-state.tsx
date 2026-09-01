import { EmptyState } from "@/components/shared"

export function AnalyticsEmptyState() {
  return (
    <EmptyState
      title="No analytics snapshot available"
      description="Analytics data will appear here when a snapshot is published."
      cardClassName="border-border/60 bg-card/60"
      contentClassName="py-12 text-center"
      titleClassName="text-sm font-medium"
    />
  )
}
