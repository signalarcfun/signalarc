/**
 * Category module — centralized market category definitions, normalization,
 * and display utilities.
 *
 * Usage:
 *   import { getMarketCategoryLabel, CREATE_MARKET_CATEGORIES } from "@/modules/categories"
 */

export type { MarketCategoryId, MarketCategory } from "./types"
export {
  MARKET_CATEGORIES,
  CREATE_MARKET_CATEGORIES,
  DEFAULT_CATEGORY_ID,
} from "./constants"
export {
  normalizeMarketCategory,
  getMarketCategoryLabel,
  getMarketCategory,
  isMarketCategoryId,
} from "./utils"
