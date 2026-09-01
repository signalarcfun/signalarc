import { ApiError } from "@/lib/api"

import type { ResolutionState } from "./types"

export function isResolutionNotFoundError(error: unknown) {
  return (
    error instanceof ApiError &&
    (error.code === "resolution_not_found" || error.status === 404)
  )
}

export function getResolutionErrorState(
  error: unknown,
): Extract<ResolutionState, { status: "error" }> {
  if (error instanceof ApiError) {
    return {
      status: "error",
      message: error.message,
      requestId: error.requestId,
    }
  }

  if (error instanceof Error) {
    return {
      status: "error",
      message: error.message,
      requestId: null,
    }
  }

  return {
    status: "error",
    message: "Unable to load resolution state.",
    requestId: null,
  }
}
