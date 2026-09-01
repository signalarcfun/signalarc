type RequestMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE"

export type ApiResponse<TData> = {
  data: TData
  requestId: string | null
}

export type ApiErrorBody = {
  code: string
  message: string
}

type BackendErrorResponse = {
  error?: Partial<ApiErrorBody>
}

export class ApiError extends Error {
  status: number
  code: string
  requestId: string | null
  body: unknown

  constructor({
    status,
    code,
    message,
    requestId,
    body,
  }: {
    status: number
    code: string
    message: string
    requestId: string | null
    body: unknown
  }) {
    super(message)
    this.name = "ApiError"
    this.status = status
    this.code = code
    this.requestId = requestId
    this.body = body
  }
}

export type HealthResponse = {
  status: string
}

export type SchemaValidationResponse = {
  status: string
  migration_version: number
  dirty: boolean
  missing_tables: string[]
  missing_columns: string[]
}

export type Market = {
  id: string
  creator_user_id: string
  title: string
  description: string | null
  category: string | null
  cover_image_url: string | null
  status: string
  outcome_yes_label: string
  outcome_no_label: string
  collateral_asset: string
  chain: string
  resolution_source: string | null
  opens_at: string | null
  closes_at: string
  resolved_at: string | null
  settled_at: string | null
  winning_outcome: string | null
  market_contract_address: string | null
  market_deployment_tx_hash: string | null
  market_factory_address: string | null
  resolver_address: string | null
  onchain_deployment_status: "NOT_DEPLOYED" | "DEPLOYED" | "FAILED"
  created_at: string
  updated_at: string
}

export type Trade = {
  id: string
  user_id: string
  market_id: string
  outcome: string
  side: string
  quantity: string
  price: string
  collateral_amount: string
  fee_amount: string
  status: string
  tx_hash: string | null
  created_at: string
  updated_at: string
}

export type Position = {
  id: string
  user_id: string
  market_id: string
  outcome: string
  quantity: string
  average_entry_price: string
  realized_pnl: string
  created_at: string
  updated_at: string
}

export type Settlement = {
  id: string
  market_id: string
  user_id: string | null
  resolution_id: string | null
  outcome: string | null
  amount: string
  status: string
  tx_hash: string | null
  settled_at: string | null
  created_at: string
  updated_at: string
}

export type Resolution = {
  id: string
  market_id: string
  winning_outcome: string | null
  status: string
  resolver_type: string | null
  evidence_reference: string | null
  resolved_at: string | null
  created_at: string
  updated_at: string
}

export type AgentMarket = {
  id: string
  title: string
  status: string
  category: string | null
  cover_image_url: string | null
  collateral_asset: string
  chain: string
  closes_at: string
  resolution_source: string | null
}

export type ArcContractStatus = {
  network: string
  chain_id: number
  signal_arc_market: string
  signal_arc_market_factory: string
  usdc_erc20_interface: string
  explorer: string
  prototype: boolean
  production_approved: boolean
  status: string
}

export type MarketsResponse = {
  markets: Market[]
}

export type MarketResponse = {
  market: Market
}

export type CreateMarketRequest = {
  id: string
  creator_user_id: string
  title: string
  description?: string
  category?: string
  cover_image_url?: string
  outcome_yes_label?: string
  outcome_no_label?: string
  collateral_asset?: string
  chain: string
  resolution_source?: string
  opens_at?: string
  closes_at: string
  market_contract_address: string
  market_deployment_tx_hash: string
  market_factory_address: string
  resolver_address: string
}

export type CreateTradeIntentRequest = {
  user_id: string
  market_id: string
  outcome: "YES" | "NO"
  side: "BUY" | "SELL"
  quantity: string
  price: string
}

export type AttachMarketContractRequest = {
  market_contract_address: string
  market_deployment_tx_hash: string
  market_factory_address: string
  resolver_address: string
}

export type TradeIntentResponse = {
  trade: Trade
  execution: {
    status: string
    reason: string
  }
}

export type UserPositionsResponse = {
  positions: Position[]
}

export type UserSettlementsResponse = {
  settlements: Settlement[]
}

export type MarketResolutionResponse = {
  resolution: Resolution | null
}

export type MarketSettlementsResponse = {
  settlements: Settlement[]
}

export type AgentMarketsResponse = {
  markets: AgentMarket[]
}

export type ArcContractResponse = ArcContractStatus

type ApiRequestOptions = Omit<RequestInit, "body" | "method"> & {
  body?: unknown
  method?: RequestMethod
}

export const localDemoUserId = "10000000-0000-4000-8000-000000000001"
export const localApiBaseUrl = "http://localhost:4000"

function getApiBaseUrl() {
  const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL?.trim()

  if (!baseUrl) {
    return localApiBaseUrl
  }

  const normalizedBaseUrl = baseUrl.replace(/\/+$/, "")

  if (normalizedBaseUrl === "http://127.0.0.1:4000") {
    return localApiBaseUrl
  }

  return normalizedBaseUrl
}

function buildUrl(path: string) {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`
  return `${getApiBaseUrl()}${normalizedPath}`
}

function getRequestId(response: Response) {
  return response.headers.get("X-Request-ID")
}

async function readJson(response: Response): Promise<unknown> {
  if (response.status === 204) {
    return null
  }

  const text = await response.text()
  if (!text) {
    return null
  }

  try {
    return JSON.parse(text)
  } catch {
    return text
  }
}

function normalizeApiError(response: Response, requestId: string | null, body: unknown) {
  const backendError = body as BackendErrorResponse
  const code = backendError?.error?.code ?? `http_${response.status}`
  const message = backendError?.error?.message ?? (response.statusText || "API request failed")

  return new ApiError({
    status: response.status,
    code,
    message,
    requestId,
    body,
  })
}

export async function apiRequest<TData>(
  path: string,
  options: ApiRequestOptions = {},
): Promise<ApiResponse<TData>> {
  const headers = new Headers(options.headers)
  const hasBody = options.body !== undefined

  if (hasBody && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json")
  }

  const response = await fetch(buildUrl(path), {
    ...options,
    method: options.method ?? "GET",
    headers,
    body: hasBody ? JSON.stringify(options.body) : undefined,
  })

  const requestId = getRequestId(response)
  const body = await readJson(response)

  if (!response.ok) {
    throw normalizeApiError(response, requestId, body)
  }

  return {
    data: body as TData,
    requestId,
  }
}

export function getHealth() {
  return apiRequest<HealthResponse>("/health")
}

export function getReadiness() {
  return apiRequest<HealthResponse>("/readyz")
}

export function validateSchema() {
  return apiRequest<SchemaValidationResponse>("/schema/validate")
}

export function getMarkets() {
  return apiRequest<MarketsResponse>("/markets")
}

export function getMarket(id: string) {
  return apiRequest<MarketResponse>(`/markets/${encodeURIComponent(id)}`)
}

export function createMarket(input: CreateMarketRequest) {
  return apiRequest<MarketResponse>("/markets", {
    method: "POST",
    body: input,
  })
}

export function attachMarketContract(marketId: string, input: AttachMarketContractRequest) {
  return apiRequest<MarketResponse>(`/markets/${encodeURIComponent(marketId)}/contract`, {
    method: "PATCH",
    body: input,
  })
}

export function createTradeIntent(input: CreateTradeIntentRequest) {
  return apiRequest<TradeIntentResponse>("/trade-intents", {
    method: "POST",
    body: input,
  })
}

export function getUserPositions(userId: string) {
  return apiRequest<UserPositionsResponse>(`/users/${encodeURIComponent(userId)}/positions`)
}

export function getUserSettlements(userId: string) {
  return apiRequest<UserSettlementsResponse>(`/users/${encodeURIComponent(userId)}/settlements`)
}

export function getMarketResolution(marketId: string) {
  return apiRequest<MarketResolutionResponse>(
    `/markets/${encodeURIComponent(marketId)}/resolution`,
  )
}

export function getMarketSettlements(marketId: string) {
  return apiRequest<MarketSettlementsResponse>(
    `/markets/${encodeURIComponent(marketId)}/settlements`,
  )
}

export function getAgentMarkets() {
  return apiRequest<AgentMarketsResponse>("/agent/markets")
}

export function getArcContract() {
  return apiRequest<ArcContractResponse>("/arc/contract")
}
