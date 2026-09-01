export function shortenAnalyticsAddress(value: string) {
  return `${value.slice(0, 8)}\u2026${value.slice(-6)}`
}

export function formatAnalyticsMetricValue(value: string | number) {
  return String(value)
}
