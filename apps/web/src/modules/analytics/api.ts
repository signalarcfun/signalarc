import { apiRequest } from "@/lib/api"

import type { AnalyticsSummaryResponse } from "./types"

export function getAnalyticsSummary() {
  return apiRequest<AnalyticsSummaryResponse>("/analytics/summary")
}
