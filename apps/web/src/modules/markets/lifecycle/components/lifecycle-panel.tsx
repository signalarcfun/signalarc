"use client"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  TransactionResultDialog,
  type TransactionResultDialogState,
} from "@/components/shared"
import { ChainStatusCard } from "@/modules/wallet"

import type { LifecycleActionState, LifecycleStatusData } from "../types"
import { formatUsdc } from "../format"
import { LifecycleActionStatus } from "./lifecycle-action-status"
import { LifecycleStatusCard } from "./lifecycle-status-card"

export function LifecycleNotDeployedCard() {
  return (
    <Card className="border-yellow-500/20">
      <CardHeader>
        <CardTitle>Onchain Status</CardTitle>
        <CardDescription>Onchain contract not deployed for this market.</CardDescription>
      </CardHeader>
    </Card>
  )
}

export function LifecyclePanel({
  marketId,
  marketTitle,
  data,
  isArcTestnet,
  isSwitchingChain,
  onSwitchNetwork,
  claimDisabledReason,
  canClaim,
  claimButtonLabel,
  resolverDisabledReason,
  resolverActionReason,
  closeDisabledReason,
  resolveDisabledReason,
  cancelDisabledReason,
  isPending,
  actionState,
  onClaim,
  onCloseMarket,
  onResolveYes,
  onResolveNo,
  onCancelMarket,
}: {
  marketId: string
  marketTitle?: string
  data: LifecycleStatusData
  isArcTestnet: boolean
  isSwitchingChain: boolean
  onSwitchNetwork: () => void
  claimDisabledReason: string | null
  canClaim: boolean
  claimButtonLabel: string
  resolverDisabledReason: string | null
  resolverActionReason: string | null
  closeDisabledReason: string | null
  resolveDisabledReason: string | null
  cancelDisabledReason: string | null
  isPending: boolean
  actionState: LifecycleActionState
  onClaim: () => void
  onCloseMarket: () => void
  onResolveYes: () => void
  onResolveNo: () => void
  onCancelMarket: () => void
}) {
  const dialogState = getLifecycleDialogState(actionState)
  const actionAmount = actionState.status !== "idle" && actionState.label.startsWith("Claim")
    ? formatUsdc(data.claimableAmount)
    : undefined

  return (
    <Card className="border-indigo-500/20">
      <CardHeader>
        <div className="flex flex-wrap items-center gap-2">
          <CardTitle>Onchain Status</CardTitle>
          <Badge
            variant="outline"
            className="border-indigo-500/30 bg-indigo-500/10 text-indigo-300"
          >
            Arc Testnet
          </Badge>
        </div>
        <CardDescription className="leading-6">
          Market-specific browser-wallet reads and lifecycle transactions. No real funds or production settlement.
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-6">
        <LifecycleStatusCard data={data} />

        {!isArcTestnet ? (
          <ChainStatusCard
            isSwitchingChain={isSwitchingChain}
            onSwitchNetwork={onSwitchNetwork}
          />
        ) : null}

        <div className="grid gap-3 rounded-lg border border-border bg-muted/20 p-4">
          <div>
            <h3 className="text-sm font-medium text-foreground">
              {claimButtonLabel}
            </h3>
            <p className="mt-1 text-xs text-muted-foreground">
              {claimButtonLabel === "Claim Refund"
                ? "Cancelled markets return eligible connected-wallet collateral."
                : "Resolved markets allow eligible connected wallets to claim payouts."}
            </p>
          </div>
          {claimDisabledReason ? (
            <p className="text-sm text-muted-foreground">{claimDisabledReason}</p>
          ) : null}
          <Button
            className="w-full"
            disabled={!canClaim || isPending}
            onClick={onClaim}
            type="button"
          >
            {isPending && actionState.status === "pending" && actionState.label === claimButtonLabel
              ? "Claiming..."
              : claimButtonLabel}
          </Button>
        </div>

        <div className="grid gap-3 rounded-lg border border-border bg-muted/20 p-4">
          <h3 className="text-sm font-medium text-foreground">Resolver Actions</h3>
          {resolverDisabledReason || resolverActionReason ? (
            <p className="rounded-md border border-yellow-500/30 bg-yellow-500/5 p-3 text-sm text-yellow-300">
              {resolverDisabledReason ?? resolverActionReason}
            </p>
          ) : null}
          <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-1 xl:grid-cols-2">
            <Button
              disabled={Boolean(closeDisabledReason)}
              onClick={onCloseMarket}
              type="button"
              variant="outline"
            >
              Close Market
            </Button>
            <Button
              disabled={Boolean(resolveDisabledReason)}
              onClick={onResolveYes}
              type="button"
              variant="outline"
            >
              Resolve YES
            </Button>
            <Button
              disabled={Boolean(resolveDisabledReason)}
              onClick={onResolveNo}
              type="button"
              variant="outline"
            >
              Resolve NO
            </Button>
            <Button
              disabled={Boolean(cancelDisabledReason)}
              onClick={onCancelMarket}
              type="button"
              variant="destructive"
            >
              Cancel Market
            </Button>
          </div>
        </div>

        <LifecycleActionStatus actionState={actionState} />
        <TransactionResultDialog
          eventId={getLifecycleDialogEventId(actionState)}
          state={dialogState}
          actionLabel={actionState.status === "idle" ? "Market Action" : actionState.label}
          marketLabel={marketTitle ?? marketId}
          amount={actionAmount}
          txHash={actionState.status === "idle" ? undefined : actionState.hash}
          message={getLifecycleDialogMessage(actionState)}
          nextStep={
            actionState.status === "success"
              ? "Onchain market reads refresh after confirmation. Review the status card for the updated state."
              : "The inline onchain status panel remains available behind this dialog."
          }
          primaryAction={
            actionState.status === "success"
              ? { label: "View market", href: `/markets/${encodeURIComponent(marketId)}` }
              : undefined
          }
          details={[
            {
              label: "Market ID",
              value: marketId,
              monospace: true,
            },
          ]}
        />
      </CardContent>
    </Card>
  )
}

function isWalletRejected(message: string) {
  return message.toLowerCase().includes("wallet transaction was rejected")
}

function getLifecycleDialogState(
  actionState: LifecycleActionState,
): TransactionResultDialogState | null {
  if (actionState.status === "pending") {
    return actionState.hash ? "pending" : "wallet_confirmation"
  }

  if (actionState.status === "success") return "success"

  if (actionState.status === "error") {
    return isWalletRejected(actionState.message) ? "rejected" : "error"
  }

  return null
}

function getLifecycleDialogEventId(actionState: LifecycleActionState) {
  if (actionState.status === "idle") return null

  if (actionState.status === "pending") {
    return `lifecycle-pending-${actionState.label}-${actionState.hash ?? "signature"}`
  }

  if (actionState.status === "success") {
    return `lifecycle-success-${actionState.label}-${actionState.hash}`
  }

  return `lifecycle-error-${actionState.label}-${actionState.hash ?? "no-hash"}-${actionState.message}`
}

function getLifecycleDialogMessage(actionState: LifecycleActionState) {
  if (actionState.status === "pending" && !actionState.hash) {
    return `Confirm ${actionState.label} in your wallet before the transaction is submitted.`
  }

  if (actionState.status === "pending") {
    return `${actionState.label} was submitted and is waiting for Arc Testnet confirmation.`
  }

  if (actionState.status === "success") {
    return `${actionState.label} confirmed on Arc Testnet.`
  }

  if (actionState.status === "error") return actionState.message

  return null
}
