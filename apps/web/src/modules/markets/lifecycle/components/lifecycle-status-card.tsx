import type { LifecycleStatusData } from "../types"
import {
  formatCloseTimestamp,
  formatLifecycleAddress,
  formatUsdc,
  marketStatusLabels,
  outcomeLabels,
} from "../format"

function OnchainField({
  label,
  value,
  mono = false,
}: {
  label: string
  value: string
  mono?: boolean
}) {
  return (
    <div className="min-w-0">
      <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
        {label}
      </dt>
      <dd
        className={`mt-1 break-all text-sm text-foreground ${
          mono ? "font-mono text-xs" : ""
        }`}
      >
        {value}
      </dd>
    </div>
  )
}

export function LifecycleStatusCard({ data }: { data: LifecycleStatusData }) {
  return (
    <dl className="grid gap-3 sm:grid-cols-2 lg:grid-cols-1 xl:grid-cols-2">
      <OnchainField
        label="Contract Address"
        value={data.deployedContractAddress}
        mono
      />
      <OnchainField
        label="Resolver Address"
        value={formatLifecycleAddress(data.resolverAddress)}
        mono
      />
      <OnchainField
        label="Connected Wallet"
        value={formatLifecycleAddress(data.connectedWallet)}
        mono
      />
      <OnchainField
        label="Status"
        value={
          data.statusValue === undefined
            ? "-"
            : marketStatusLabels[data.statusValue] ?? `Unknown (${data.statusValue})`
        }
      />
      <OnchainField
        label="Close Time"
        value={formatCloseTimestamp(data.closeTimestamp)}
      />
      <OnchainField
        label="Winning Outcome"
        value={
          data.isResolved
            ? outcomeLabels[data.winningOutcome ?? 0] ??
              `Unknown (${data.winningOutcome})`
            : "-"
        }
      />
      <OnchainField
        label="User YES"
        value={data.isConnected ? formatUsdc(data.userYes) : "-"}
      />
      <OnchainField
        label="User NO"
        value={data.isConnected ? formatUsdc(data.userNo) : "-"}
      />
      <OnchainField
        label="Claimable USDC"
        value={data.isConnected ? formatUsdc(data.claimableAmount) : "-"}
      />
      <OnchainField
        label="Claimed Status"
        value={data.isConnected ? (data.hasClaimed ? "Claimed" : "Not claimed") : "-"}
      />
      <OnchainField label="Total YES" value={formatUsdc(data.totalYes)} />
      <OnchainField label="Total NO" value={formatUsdc(data.totalNo)} />
      <OnchainField
        label="Total Collateral"
        value={formatUsdc(data.totalCollateral)}
      />
    </dl>
  )
}
