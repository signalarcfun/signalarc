import { InlineErrorState } from "@/components/shared"

import type { ResolutionState } from "../types"

export function ResolutionActionStatus({
  state,
}: {
  state: Extract<ResolutionState, { status: "loading" | "error" }>
}) {
  if (state.status === "loading") {
    return (
      <div className="animate-pulse space-y-3">
        <div className="h-4 w-1/3 rounded bg-muted" />
        <div className="h-4 w-1/2 rounded bg-muted" />
      </div>
    )
  }

  return (
    <InlineErrorState
      title="Unable to load resolution data"
      message={state.message}
      requestId={state.requestId}
    />
  )
}
