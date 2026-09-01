import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { formatShortAddress, formatShortHash } from "@/lib/contracts"
import type { Market } from "@/lib/api"

import {
  arcscanContractUrl,
  arcscanTransactionUrl,
  formatDeploymentStatus,
  formatMarketDate,
  onchainDeploymentBadgeClass,
} from "../format"

function DetailItem({ label, value }: { label: string; value: string | null }) {
  return (
    <div className="min-w-0 rounded-lg border border-border bg-muted/20 p-3">
      <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
        {label}
      </dt>
      <dd className="mt-1 break-words text-sm font-medium text-foreground">
        {value || "-"}
      </dd>
    </div>
  )
}

function ExplorerItem({
  label,
  value,
  href,
  displayValue,
}: {
  label: string
  value: string | null
  href: string | null
  displayValue: string
}) {
  return (
    <div className="min-w-0 rounded-lg border border-border bg-muted/20 p-3">
      <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
        {label}
      </dt>
      <dd className="mt-1 min-w-0 text-sm font-medium">
        {value && href ? (
          <a
            href={href}
            target="_blank"
            rel="noopener noreferrer"
            title={value}
            className="inline-flex max-w-full items-center font-mono text-xs text-indigo-300 underline-offset-4 transition-colors hover:text-indigo-200 hover:underline"
          >
            <span className="truncate">{displayValue}</span>
          </a>
        ) : (
          <span className="text-muted-foreground">-</span>
        )}
      </dd>
    </div>
  )
}

export function MarketMetadataCard({ market }: { market: Market }) {
  return (
    <>
      <Card>
        <CardHeader className="space-y-3">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Market Parameters
            </h3>
            <Badge
              variant="outline"
              className={onchainDeploymentBadgeClass(market.onchain_deployment_status)}
            >
              {formatDeploymentStatus(market.onchain_deployment_status)}
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <dl className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <DetailItem label="Closes" value={formatMarketDate(market.closes_at)} />
            <DetailItem
              label="Opens"
              value={market.opens_at ? formatMarketDate(market.opens_at) : null}
            />
            <DetailItem
              label="Resolved"
              value={market.resolved_at ? formatMarketDate(market.resolved_at) : null}
            />
            <DetailItem label="Winning Outcome" value={market.winning_outcome} />
            <DetailItem label="Created" value={formatMarketDate(market.created_at)} />
            <DetailItem label="Collateral" value={market.collateral_asset} />
            <DetailItem label="Chain" value={market.chain} />
            <ExplorerItem
              label="Market Contract"
              value={market.market_contract_address}
              href={
                market.market_contract_address
                  ? arcscanContractUrl(market.market_contract_address)
                  : null
              }
              displayValue={
                market.market_contract_address
                  ? formatShortAddress(market.market_contract_address as `0x${string}`)
                  : "-"
              }
            />
            <ExplorerItem
              label="Deployment Tx"
              value={market.market_deployment_tx_hash}
              href={
                market.market_deployment_tx_hash
                  ? arcscanTransactionUrl(market.market_deployment_tx_hash)
                  : null
              }
              displayValue={
                market.market_deployment_tx_hash
                  ? formatShortHash(market.market_deployment_tx_hash as `0x${string}`)
                  : "-"
              }
            />
          </dl>
        </CardContent>
      </Card>

      {market.resolution_source ? (
        <Card>
          <CardHeader>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Resolution
            </h3>
          </CardHeader>
          <CardContent>
            <dl>
              <DetailItem
                label="Resolution Source"
                value={market.resolution_source}
              />
            </dl>
          </CardContent>
        </Card>
      ) : null}
    </>
  )
}
