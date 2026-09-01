import type { Metadata } from "next"
import Link from "next/link"

import { Button } from "@/components/ui/button"
import { MarketDetail } from "@/features/markets/market-detail"
import { getMarket } from "@/lib/api"

type MarketDetailPageProps = {
  params: Promise<{
    id: string
  }>
}

const SITE_URL = "https://signalarc.fun"
const DEFAULT_SHARE_IMAGE = `${SITE_URL}/og-image.png`

function truncate(value: string, maxLength: number) {
  return value.length > maxLength ? `${value.slice(0, maxLength - 1)}…` : value
}

export async function generateMetadata({
  params,
}: MarketDetailPageProps): Promise<Metadata> {
  const { id } = await params
  const marketUrl = `${SITE_URL}/markets/${encodeURIComponent(id)}`

  try {
    const response = await getMarket(id)
    const market = response.data.market
    const title = truncate(`${market.title} | SignalArc`, 70)
    const description = truncate(
      market.description ||
        `YES/NO prediction market on SignalArc. Trade ${market.outcome_yes_label} or ${market.outcome_no_label} on Arc Testnet.`,
      180,
    )
    const image = market.cover_image_url || DEFAULT_SHARE_IMAGE

    return {
      title,
      description,
      alternates: {
        canonical: marketUrl,
      },
      openGraph: {
        title,
        description,
        url: marketUrl,
        siteName: "SignalArc",
        type: "website",
        images: [
          {
            url: image,
            width: 1200,
            height: 630,
            alt: market.title,
          },
        ],
      },
      twitter: {
        card: "summary_large_image",
        title,
        description,
        images: [image],
      },
    }
  } catch {
    return {
      title: "SignalArc Market",
      description: "YES/NO prediction market on Arc Testnet.",
      alternates: {
        canonical: marketUrl,
      },
      openGraph: {
        title: "SignalArc Market",
        description: "YES/NO prediction market on Arc Testnet.",
        url: marketUrl,
        siteName: "SignalArc",
        type: "website",
        images: [
          {
            url: DEFAULT_SHARE_IMAGE,
            width: 1200,
            height: 630,
            alt: "SignalArc",
          },
        ],
      },
      twitter: {
        card: "summary_large_image",
        title: "SignalArc Market",
        description: "YES/NO prediction market on Arc Testnet.",
        images: [DEFAULT_SHARE_IMAGE],
      },
    }
  }
}

export default async function MarketDetailPage({ params }: MarketDetailPageProps) {
  const { id } = await params

  return (
    <div className="px-4 py-8 sm:px-6 lg:px-8">
      <div className="mx-auto flex w-full max-w-5xl flex-col gap-6">
        <div>
          <Button asChild size="sm" variant="ghost">
            <Link href="/markets">&larr; Back to markets</Link>
          </Button>
        </div>
        <MarketDetail marketId={id} />
      </div>
    </div>
  )
}
