import type { Resolution, Settlement } from "@/lib/api"

export type ResolutionState =
  | { status: "loading" }
  | { status: "empty"; settlements: Settlement[] }
  | { status: "loaded"; resolution: Resolution; settlements: Settlement[] }
  | { status: "error"; message: string; requestId: string | null }
