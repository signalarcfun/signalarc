/**
 * Official market category types for SignalArc.
 *
 * MarketCategoryId uses lowercase identifiers for internal
 * consistency. Display labels are defined in constants.ts.
 */

export type MarketCategoryId =
  | "all"
  | "crypto"
  | "sports"
  | "politics"
  | "macro"
  | "ai"
  | "tech"
  | "arc"
  | "other"

export interface MarketCategory {
  id: MarketCategoryId
  label: string
  description: string
}
