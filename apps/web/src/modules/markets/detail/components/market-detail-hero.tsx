import type { Market } from "@/lib/api"

export function MarketDetailHero({ market }: { market: Market }) {
  if (market.cover_image_url) {
    return (
      // Plain img is intentional for v1 user-provided remote URLs.
      // eslint-disable-next-line @next/next/no-img-element
      <img
        src={market.cover_image_url}
        alt={market.title}
        className="aspect-[16/10] w-full rounded-lg object-cover sm:aspect-[16/7]"
      />
    )
  }

  return (
    <div
      className="aspect-[16/10] w-full rounded-lg bg-muted sm:aspect-[16/7]"
      aria-hidden="true"
    />
  )
}
