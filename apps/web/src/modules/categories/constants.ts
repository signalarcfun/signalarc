import type { MarketCategory } from "./types"

/**
 * Full list of official market categories, including the virtual "All" filter.
 */
export const MARKET_CATEGORIES: MarketCategory[] = [
  {
    id: "all",
    label: "All",
    description: "All categories",
  },
  {
    id: "crypto",
    label: "Crypto",
    description: "Cryptocurrency markets",
  },
  {
    id: "sports",
    label: "Sports",
    description: "Sports event markets",
  },
  {
    id: "politics",
    label: "Politics",
    description: "Political event markets",
  },
  {
    id: "macro",
    label: "Macro",
    description: "Macro-economics and finance markets",
  },
  {
    id: "ai",
    label: "AI",
    description: "Artificial intelligence markets",
  },
  {
    id: "tech",
    label: "Tech",
    description: "Technology markets",
  },
  {
    id: "arc",
    label: "Arc",
    description: "Arc protocol and ecosystem markets",
  },
  {
    id: "other",
    label: "Other",
    description: "Markets that do not fit another category",
  },
]

/**
 * Categories available for market creation.
 * Excludes the virtual "All" filter.
 */
export const CREATE_MARKET_CATEGORIES = MARKET_CATEGORIES.filter(
  (c) => c.id !== "all",
)

/**
 * Default category for new markets.
 */
export const DEFAULT_CATEGORY_ID = "crypto" as const
