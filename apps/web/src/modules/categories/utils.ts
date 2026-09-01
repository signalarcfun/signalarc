import type { MarketCategoryId, MarketCategory } from "./types"
import { MARKET_CATEGORIES } from "./constants"

/**
 * Mapping of legacy/free-text category values to official MarketCategoryId.
 * Keys are normalized with normalizeCategoryKey() for case-insensitive and
 * separator-insensitive matching.
 */
const LEGACY_CATEGORY_MAP: Record<string, MarketCategoryId> = {
  // Crypto
  crypto: "crypto",
  cryptocurrency: "crypto",
  blockchain: "crypto",
  web3: "crypto",
  defi: "crypto",
  nft: "crypto",
  token: "crypto",
  btc: "crypto",
  eth: "crypto",
  bitcoin: "crypto",
  ethereum: "crypto",

  // Sports
  sports: "sports",
  sport: "sports",
  football: "sports",
  soccer: "sports",
  basketball: "sports",
  baseball: "sports",
  tennis: "sports",
  mma: "sports",
  ufc: "sports",
  nfl: "sports",
  nba: "sports",
  mlb: "sports",
  fifa: "sports",

  // Politics
  politics: "politics",
  political: "politics",
  election: "politics",
  elections: "politics",
  government: "politics",
  policy: "politics",
  vote: "politics",
  voting: "politics",
  democrat: "politics",
  republican: "politics",
  congress: "politics",
  senate: "politics",
  presidential: "politics",

  // Macro
  macro: "macro",
  economy: "macro",
  economics: "macro",
  finance: "macro",
  fed: "macro",
  "federal reserve": "macro",
  "interest rate": "macro",
  inflation: "macro",
  gdp: "macro",
  employment: "macro",
  markets: "macro",
  stock: "macro",
  tradfi: "macro",

  // AI
  ai: "ai",
  "artificial intelligence": "ai",
  "machine learning": "ai",
  llm: "ai",
  gpt: "ai",
  openai: "ai",
  "deep learning": "ai",

  // Tech
  tech: "tech",
  technology: "tech",
  software: "tech",
  startup: "tech",
  product: "tech",
  saas: "tech",
  app: "tech",
  internet: "tech",
  cyber: "tech",
  cybersecurity: "tech",
  space: "tech",

  // Arc
  arc: "arc",
  protocol: "arc",
  staking: "arc",
  "proof of work": "arc",
  "proof of stake": "arc",

  // Other (explicit fallbacks)
  other: "other",
  miscellaneous: "other",
  misc: "other",
  general: "other",
}

/**
 * Normalizes any free-text category value into an official MarketCategoryId.
 *
 * Handles:
 * - Official ids directly ("crypto", "sports", etc.)
 * - Official labels with case variance ("Crypto", "CRYPTO", "aI")
 * - Legacy/free-text values ("cryptocurrency", "football", "election", etc.)
 * - null/undefined (returns "other")
 */
function normalizeCategoryKey(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/&/g, "and")
    .replace(/[^a-z0-9]+/g, " ")
    .trim()
}

export function normalizeMarketCategory(
  value: string | null | undefined,
): MarketCategoryId {
  if (!value) return "other"

  const normalized = normalizeCategoryKey(value)
  if (!normalized) return "other"

  // Direct match on official id.
  if (isMarketCategoryId(normalized)) return normalized

  // Direct match on official label.
  for (const category of MARKET_CATEGORIES) {
    if (category.id !== "all" && normalizeCategoryKey(category.label) === normalized) {
      return category.id
    }
  }

  // Legacy/free-text lookup.
  const mapped = LEGACY_CATEGORY_MAP[normalized]
  if (mapped) return mapped

  return "other"
}

/**
 * Returns the display label for a category value.
 * Handles old/free-text data by normalizing first.
 */
export function getMarketCategoryLabel(
  value: string | null | undefined,
): string {
  const id = normalizeMarketCategory(value)
  const category = MARKET_CATEGORIES.find((c) => c.id === id)
  return category?.label ?? "Other"
}

/**
 * Returns the full MarketCategory object for a category value.
 * Handles old/free-text data by normalizing first.
 */
export function getMarketCategory(
  value: string | null | undefined,
): MarketCategory {
  const id = normalizeMarketCategory(value)
  return MARKET_CATEGORIES.find((c) => c.id === id) ?? {
    id: "other",
    label: "Other",
    description: "Markets that do not fit another category",
  }
}

/**
 * Type guard: checks if a string is a valid MarketCategoryId.
 */
export function isMarketCategoryId(
  value: string,
): value is MarketCategoryId {
  const validIds: readonly string[] = [
    "all",
    "crypto",
    "sports",
    "politics",
    "macro",
    "ai",
    "tech",
    "arc",
    "other",
  ]
  return validIds.includes(value)
}
