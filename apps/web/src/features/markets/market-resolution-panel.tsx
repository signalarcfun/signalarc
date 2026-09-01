"use client"

import { useEffect, useState } from "react"

import {
  getMarketResolution,
  getMarketSettlements,
  type Resolution,
} from "@/lib/api"
import {
  getResolutionErrorState,
  isResolutionNotFoundError,
  ResolutionPanel,
  type ResolutionState,
} from "@/modules/markets/resolution"

async function loadResolution(marketId: string): Promise<Resolution | null> {
  try {
    const response = await getMarketResolution(marketId)
    return response.data.resolution
  } catch (error) {
    if (isResolutionNotFoundError(error)) {
      return null
    }

    throw error
  }
}

export function MarketResolutionPanel({ marketId }: { marketId: string }) {
  const [state, setState] = useState<ResolutionState>({ status: "loading" })

  useEffect(() => {
    let isActive = true

    async function loadState() {
      setState({ status: "loading" })

      try {
        const [resolution, settlementsResponse] = await Promise.all([
          loadResolution(marketId),
          getMarketSettlements(marketId),
        ])
        const settlements = settlementsResponse.data.settlements

        if (!isActive) {
          return
        }

        if (!resolution) {
          setState({ status: "empty", settlements })
          return
        }

        setState({ status: "loaded", resolution, settlements })
      } catch (error) {
        if (isActive) {
          setState(getResolutionErrorState(error))
        }
      }
    }

    void loadState()

    return () => {
      isActive = false
    }
  }, [marketId])

  return <ResolutionPanel state={state} />
}
