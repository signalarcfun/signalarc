import type { Market } from "@/lib/api"

export type MarketDetailState =
  | { status: "loading" }
  | { status: "error"; message: string; requestId: string | null }
  | { status: "ready"; market: Market }
