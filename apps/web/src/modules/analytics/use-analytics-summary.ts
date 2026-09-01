"use client"

import { useEffect, useState } from "react"

import { ApiError } from "@/lib/api"

import { getAnalyticsSummary } from "./api"
import type { AnalyticsSummaryResponse } from "./types"

export type AnalyticsSummaryState =
  | { status: "loading" }
  | { status: "loaded"; summary: AnalyticsSummaryResponse }
  | { status: "error"; message: string; requestId: string | null }

function getErrorState(error: unknown): Extract<AnalyticsSummaryState, { status: "error" }> {
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
    message: "Unable to load analytics summary.",
    requestId: null,
  }
}

export function useAnalyticsSummary() {
  const [state, setState] = useState<AnalyticsSummaryState>({ status: "loading" })

  useEffect(() => {
    let isActive = true

    async function loadSummary() {
      try {
        const response = await getAnalyticsSummary()

        if (isActive) {
          setState({ status: "loaded", summary: response.data })
        }
      } catch (error) {
        if (isActive) {
          setState(getErrorState(error))
        }
      }
    }

    void loadSummary()

    return () => {
      isActive = false
    }
  }, [])

  return state
}
