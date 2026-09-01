import type { ReactNode } from "react"

export function LoadingSkeletonBlock({
  className = "animate-pulse rounded bg-muted",
}: {
  className?: string
}) {
  return <div className={className} />
}

export function LoadingSkeletonList({
  count,
  renderItem,
}: {
  count: number
  renderItem: (index: number) => ReactNode
}) {
  return <>{Array.from({ length: count }, (_, index) => renderItem(index))}</>
}
