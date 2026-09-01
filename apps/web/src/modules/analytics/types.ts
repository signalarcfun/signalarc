export type AnalyticsMetric = {
  label: string
  value: string
  description: string
  unit: string
  featured: boolean
}

export type AnalyticsFactorySnapshot = {
  name: string
  address: string
  explorerUrl: string
  deploymentTx: string
  deploymentTxUrl: string
  deployer: string
  deploymentBlock: string
  deploymentTimestamp: string
  totalTransactions: number
  latestMarket: string
}

export type AnalyticsActivitySnapshot = {
  timestamp: string
  tx: string
  txUrl: string
  market: string
}

export type AnalyticsTopMarket = {
  address: string
  question: string
  collateral: string
  positionEvents: number
  explorerUrl: string
}

export type AnalyticsPublicLink = {
  label: string
  href: string
}

export type AnalyticsSummaryMetrics = {
  markets_created: number
  market_contracts_found: number
  total_trades: number
  position_events: number
  yes_position_events: number
  no_position_events: number
  unique_wallets: number
  testnet_usdc_volume: string
  resolved_markets: number
  cancelled_markets: number
  claim_events: number
  payouts_claimed: number
  refunds_claimed: number
}

export type AnalyticsSummaryResponse = {
  status: string
  source_status: string
  factory_address: string
  generated_at: string | null
  latest_event_at: string | null
  latest_block: number | null
  metrics: AnalyticsSummaryMetrics
}
