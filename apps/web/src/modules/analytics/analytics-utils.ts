import type {
  AnalyticsActivitySnapshot,
  AnalyticsFactorySnapshot,
  AnalyticsMetric,
  AnalyticsPublicLink,
  AnalyticsSummaryResponse,
  AnalyticsTopMarket,
} from "./types"

const ACTIVE_FACTORY_ADDRESS = "0x02555FC5EE3c53938f2F0356e963865503442A56"
const ARCSCAN_BASE_URL = "https://testnet.arcscan.app"

function formatInteger(value: number) {
  return new Intl.NumberFormat("en-US").format(value)
}

function formatWholeDigits(value: string) {
  return value.replace(/\B(?=(\d{3})+(?!\d))/g, ",")
}

export function formatTestnetUsdcBaseUnits(value: string, decimals = 6) {
  const rawValue = value.trim()

  if (!/^\d+$/.test(rawValue)) {
    return "0"
  }

  const paddedValue = rawValue.padStart(decimals + 1, "0")
  const wholeDigits = paddedValue.slice(0, -decimals) || "0"
  const fractionalDigits = paddedValue.slice(-decimals).replace(/0+$/, "")
  const whole = formatWholeDigits(wholeDigits.replace(/^0+(?=\d)/, ""))

  if (!fractionalDigits) {
    return whole
  }

  return `${whole}.${fractionalDigits}`
}

export function isIndexedAnalyticsSummary(summary: AnalyticsSummaryResponse | null) {
  return Boolean(
    summary &&
      summary.status === "ok" &&
      (summary.source_status === "indexed" || summary.source_status === "cached"),
  )
}

export function formatAnalyticsTimestamp(value: string | null) {
  if (!value) {
    return "Not indexed yet"
  }

  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return value
  }

  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
    timeZoneName: "short",
  }).format(date)
}

export function buildAnalyticsMetrics(summary: AnalyticsSummaryResponse): AnalyticsMetric[] {
  const { metrics } = summary

  return [
    {
      label: "Markets Created",
      value: formatInteger(metrics.markets_created),
      description: "MarketDeployed events indexed from the active SignalArc factory",
      unit: "testnet markets",
      featured: false,
    },
    {
      label: "Market Contracts Found",
      value: formatInteger(metrics.market_contracts_found),
      description: "Market contract addresses discovered from indexed factory events",
      unit: "verified records",
      featured: false,
    },
    {
      label: "Total Trades",
      value: formatInteger(metrics.total_trades),
      description: "Aggregate YES/NO position events indexed from child market contracts",
      unit: "onchain events",
      featured: true,
    },
    {
      label: "Position Events",
      value: formatInteger(metrics.position_events),
      description: "YES/NO position events across discovered market contracts",
      unit: "onchain events",
      featured: false,
    },
    {
      label: "Unique Participating Wallets",
      value: formatInteger(metrics.unique_wallets),
      description: "Unique wallet addresses participating in indexed market activity",
      unit: "wallets",
      featured: false,
    },
    {
      label: "Testnet USDC Collateral Volume",
      value: formatTestnetUsdcBaseUnits(metrics.testnet_usdc_volume),
      description: "Aggregate Arc Testnet USDC collateral movement from base-unit event amounts",
      unit: "testnet USDC",
      featured: true,
    },
    {
      label: "YES Position Events",
      value: formatInteger(metrics.yes_position_events),
      description: "YES-side position events",
      unit: "events",
      featured: false,
    },
    {
      label: "NO Position Events",
      value: formatInteger(metrics.no_position_events),
      description: "NO-side position events",
      unit: "events",
      featured: false,
    },
    {
      label: "Resolved Markets",
      value: formatInteger(metrics.resolved_markets),
      description: "Markets with indexed resolution events",
      unit: "testnet markets",
      featured: false,
    },
    {
      label: "Cancelled Markets",
      value: formatInteger(metrics.cancelled_markets),
      description: "Markets with indexed cancellation events",
      unit: "testnet markets",
      featured: false,
    },
    {
      label: "Claim Events",
      value: formatInteger(metrics.claim_events),
      description: "Payout/refund claim events detected",
      unit: "onchain events",
      featured: false,
    },
    {
      label: "Payouts Claimed",
      value: formatInteger(metrics.payouts_claimed),
      description: "Payout claim events from resolved markets",
      unit: "onchain events",
      featured: false,
    },
    {
      label: "Refunds Claimed",
      value: formatInteger(metrics.refunds_claimed),
      description: "Refund claim events from cancelled markets",
      unit: "onchain events",
      featured: false,
    },
  ]
}

export function buildAnalyticsLifecycleMetrics(summary: AnalyticsSummaryResponse) {
  const { metrics } = summary

  return [
    { label: "Resolved markets", value: formatInteger(metrics.resolved_markets) },
    { label: "Cancelled markets", value: formatInteger(metrics.cancelled_markets) },
    { label: "Claim events", value: formatInteger(metrics.claim_events) },
  ]
}

export function getAnalyticsFactoryAddress(summary: AnalyticsSummaryResponse | null) {
  return summary?.factory_address || ACTIVE_FACTORY_ADDRESS
}

export function getAnalyticsFactoryExplorerUrl(summary: AnalyticsSummaryResponse | null) {
  return `${ARCSCAN_BASE_URL}/address/${getAnalyticsFactoryAddress(summary)}`
}

export const analyticsFactory: AnalyticsFactorySnapshot = {
  name: "Legacy SignalArcMarketFactory",
  address: "0x837e09E8D7806E0e7b740b798173756315E51206",
  explorerUrl:
    "https://testnet.arcscan.app/address/0x837e09E8D7806E0e7b740b798173756315E51206",
  deploymentTx:
    "0x85870afd1e8c3d7d8574a10a21aef5ca919fffc44d7e20b6bce2a792a572e38e",
  deploymentTxUrl:
    "https://testnet.arcscan.app/tx/0x85870afd1e8c3d7d8574a10a21aef5ca919fffc44d7e20b6bce2a792a572e38e",
  deployer: "0x153D2Fc8334a84a37B7A7cF9deFA5Cb401a36FdC",
  deploymentBlock: "43,221,323",
  deploymentTimestamp: "2026-05-20T18:49:53Z",
  totalTransactions: 128,
  latestMarket: "0x6127d26e322b50e0e1ced9e22EAA55EC8AE087ea",
}

export const analyticsLatestActivity: AnalyticsActivitySnapshot = {
  timestamp: "2026-06-21T07:51:30Z",
  tx: "0xd8e9f40d6b95d1ed3eee69e65b41af877361b614c6e96bf4b2ff4bf0e3fb248f",
  txUrl:
    "https://testnet.arcscan.app/tx/0xd8e9f40d6b95d1ed3eee69e65b41af877361b614c6e96bf4b2ff4bf0e3fb248f",
  market: "0x6127d26e322b50e0e1ced9e22EAA55EC8AE087ea",
}

export const analyticsStatusBadges = [
  "Arc Testnet",
  "Backend Cache",
  "Active Factory",
  "Testnet USDC",
  "Public Proof-of-Activity",
  "Explorer-Indexed Logs",
] as const

export const analyticsMetrics: AnalyticsMetric[] = [
  {
    label: "Markets Created",
    value: "126",
    description: "Historical YES/NO market contracts created from the legacy SignalArc factory",
    unit: "testnet markets",
    featured: false,
  },
  {
    label: "Market Contracts Found",
    value: "126",
    description: "Historical market contracts detected from legacy factory deployment events",
    unit: "verified records",
    featured: false,
  },
  {
    label: "Total Trades",
    value: "806",
    description: "Aggregate YES/NO position events across created markets",
    unit: "onchain events",
    featured: true,
  },
  {
    label: "Position Events",
    value: "806",
    description: "YES/NO position events across created markets",
    unit: "onchain events",
    featured: false,
  },
  {
    label: "Unique Participating Wallets",
    value: "218",
    description: "Wallet addresses participating in market activity",
    unit: "wallets",
    featured: false,
  },
  {
    label: "Testnet USDC Collateral Volume",
    value: "149.77",
    description: "Aggregate Arc Testnet USDC collateral movement",
    unit: "testnet USDC",
    featured: true,
  },
  {
    label: "YES Position Events",
    value: "374",
    description: "YES-side position events",
    unit: "events",
    featured: false,
  },
  {
    label: "NO Position Events",
    value: "432",
    description: "NO-side position events",
    unit: "events",
    featured: false,
  },
  {
    label: "Resolved Markets",
    value: "105",
    description: "Markets that completed resolution lifecycle",
    unit: "testnet markets",
    featured: false,
  },
  {
    label: "Cancelled Markets",
    value: "12",
    description: "Markets cancelled during testnet lifecycle",
    unit: "testnet markets",
    featured: false,
  },
  {
    label: "Claim Events",
    value: "393",
    description: "Payout/refund claim events detected",
    unit: "onchain events",
    featured: false,
  },
  {
    label: "Payouts Claimed",
    value: "382",
    description: "Payout claim events from resolved markets",
    unit: "onchain events",
    featured: false,
  },
  {
    label: "Refunds Claimed",
    value: "11",
    description: "Refund claim events from cancelled markets",
    unit: "onchain events",
    featured: false,
  },
]

export const analyticsTopMarkets: AnalyticsTopMarket[] = [
  {
    address: "0xB2d3D059Cb1d9ebAaeC9751e654f1DBA99eb5c27",
    question: "when SignalArc launch?",
    collateral: "5",
    positionEvents: 1,
    explorerUrl:
      "https://testnet.arcscan.app/address/0xB2d3D059Cb1d9ebAaeC9751e654f1DBA99eb5c27",
  },
  {
    address: "0xcDba50C6E74B2798607375844728a71084D65aEE",
    question: "Will Doge june end?",
    collateral: "5",
    positionEvents: 2,
    explorerUrl:
      "https://testnet.arcscan.app/address/0xcDba50C6E74B2798607375844728a71084D65aEE",
  },
  {
    address: "0x2302D7AD177a574bd8f688b0debC81355F80E998",
    question: "Yes",
    collateral: "3",
    positionEvents: 3,
    explorerUrl:
      "https://testnet.arcscan.app/address/0x2302D7AD177a574bd8f688b0debC81355F80E998",
  },
  {
    address: "0xB748Cd9810429d0756c235686E620E7C783727d9",
    question:
      "Phase 5 random QA market phase5-qa-wallet-015-1781105544-530980: will the automated lifecycle complete?",
    collateral: "2.4",
    positionEvents: 3,
    explorerUrl:
      "https://testnet.arcscan.app/address/0xB748Cd9810429d0756c235686E620E7C783727d9",
  },
  {
    address: "0x1243A49e746702aa2226f7Abd828f47e3119EF88",
    question:
      "Phase 5 random QA market phase5-qa-wallet-042-1781139890-562129: will the automated lifecycle complete?",
    collateral: "2.32",
    positionEvents: 3,
    explorerUrl:
      "https://testnet.arcscan.app/address/0x1243A49e746702aa2226f7Abd828f47e3119EF88",
  },
]

export const analyticsAgentIntegrationChecklist = [
  "Agent wallet onboarding",
  "OTP verification",
  "Agent wallet session activation",
  "Wallet balance lookup",
  "Faucet request flow",
  "Backend-driven smart contract execution",
  "Testnet USDC collateral usage",
  "Agent intent \u2192 confirmation \u2192 execution flow",
] as const

export const analyticsBackendMetrics = [
  "Agent intents created",
  "Agent intents confirmed",
  "Agent executions attempted",
  "Agent executions successful",
  "Circle agent wallets registered",
  "Circle agent wallet sessions",
  "Onboarding attempts",
  "OTP verifications",
  "Wallet balance checks",
  "Faucet requests",
] as const

export const analyticsPublicLinks: AnalyticsPublicLink[] = [
  { label: "Website", href: "https://www.signalarc.fun" },
  { label: "Docs", href: "https://docs.signalarc.fun" },
  { label: "GitHub", href: "https://github.com/wahyu241205/SignalArc" },
  { label: "Active factory on Arcscan", href: `${ARCSCAN_BASE_URL}/address/${ACTIVE_FACTORY_ADDRESS}` },
  { label: "Legacy analytics factory on Arcscan", href: analyticsFactory.explorerUrl },
  { label: "Latest activity", href: analyticsLatestActivity.txUrl },
]

export const analyticsLimitations = [
  "Live metrics are sourced from SignalArc's backend analytics cache when indexed data is available.",
  "The historical static snapshot remains a fallback when the backend cache is unavailable or not indexed.",
  "Volume represents testnet USDC collateral activity, not production or mainnet trading volume.",
  "Unique participants are counted as wallet addresses, not real-world users.",
  "Circle Agent Wallet session and onboarding metrics are sourced from backend data, not chain data alone.",
  "Direct Circle Agent Wallet attribution is not visible from chain data alone.",
  "Arc Testnet data is not yet reliably queryable through Dune for the required contract-level analytics.",
  "This dashboard does not imply audited production custody or mainnet financial activity.",
] as const

export const analyticsLifecycleMetrics = [
  { label: "Resolved markets", value: "105" },
  { label: "Cancelled markets", value: "12" },
  { label: "Claim events", value: "393" },
] as const

export const analyticsFactoryProofPoints = [
  {
    label: "Legacy verified source",
    value: "Published on Arcscan",
  },
  {
    label: "Deployment network",
    value: "Arc Testnet",
  },
  {
    label: "Historical created contracts",
    value: "126 YES/NO markets",
  },
] as const
