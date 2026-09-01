import { ApiError } from "@/lib/api"

import type { MarketsState, PortfolioState } from "./types"

export function getPortfolioErrorState(
  userId: string,
  error: unknown,
): Extract<PortfolioState, { status: "error" }> {
  if (error instanceof ApiError) {
    return {
      status: "error",
      userId,
      message: error.message,
      requestId: error.requestId,
    }
  }

  if (error instanceof Error) {
    return {
      status: "error",
      userId,
      message: error.message,
      requestId: null,
    }
  }

  return {
    status: "error",
    userId,
    message: "Unable to load portfolio data.",
    requestId: null,
  }
}

export function getMarketsErrorState(error: unknown): Extract<MarketsState, { status: "error" }> {
  if (error instanceof ApiError) {
    return {
      status: "error",
      message: error.message,
      requestId: error.requestId,
    }
  }

  return {
    status: "error",
    message: error instanceof Error ? error.message : "Unable to load markets.",
    requestId: null,
  }
}
