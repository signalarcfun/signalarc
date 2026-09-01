import { MetricCard } from "@/components/shared"

import { formatAnalyticsMetricValue } from "../format"
import type { AnalyticsMetric } from "../types"

export function AnalyticsMetricCard({ metric }: { metric: AnalyticsMetric }) {
  return (
    <MetricCard
      label={metric.label}
      value={formatAnalyticsMetricValue(metric.value)}
      description={metric.description}
      unit={metric.unit}
      featured={metric.featured}
    />
  )
}
