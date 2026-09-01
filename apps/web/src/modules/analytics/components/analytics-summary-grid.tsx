import type { AnalyticsMetric } from "../types"

import { AnalyticsMetricCard } from "./analytics-metric-card"

export function AnalyticsSummaryGrid({ metrics }: { metrics: AnalyticsMetric[] }) {
  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-5">
      {metrics.map((metric) => (
        <AnalyticsMetricCard key={metric.label} metric={metric} />
      ))}
    </div>
  )
}
