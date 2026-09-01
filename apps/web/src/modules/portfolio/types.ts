import type { Market, Position, Settlement } from "@/lib/api"

export type PortfolioData = {
  positions: Position[]
  settlements: Settlement[]
}

export type PortfolioState =
  | { status: "idle" }
  | { status: "loading"; userId: string }
  | { status: "empty"; userId: string }
  | { status: "loaded"; userId: string; data: PortfolioData }
  | {
      status: "error"
      userId: string
      message: string
      requestId: string | null
    }

export type MarketsState =
  | { status: "loading" }
  | { status: "loaded"; markets: Market[] }
  | { status: "error"; message: string; requestId: string | null }
