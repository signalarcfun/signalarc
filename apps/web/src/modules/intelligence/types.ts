import type { AgentMarket } from "@/lib/api"

export type IntelligenceSignal = AgentMarket

export type IntelligenceState =
  | { status: "loading" }
  | { status: "empty" }
  | { status: "error"; message: string; requestId: string | null }
  | { status: "loaded"; markets: IntelligenceSignal[] }
