import { EmptyState } from "@/components/shared"

export function PortfolioEmptyState({
  title = "No positions or settlements found",
  description = "The loaded API records did not include active positions, settlement history, or refundable activity.",
}: {
  title?: string
  description?: string
}) {
  return (
    <EmptyState
      title={title}
      description={description}
      contentClassName="p-6 text-left"
      titleClassName="text-sm font-medium text-foreground"
      descriptionClassName="mt-2 text-sm leading-6 text-muted-foreground"
    />
  )
}
