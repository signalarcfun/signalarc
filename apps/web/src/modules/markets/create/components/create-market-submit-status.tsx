import Link from "next/link"

import {
  InlineErrorState,
  TransactionResultDialog,
  type TransactionResultDialogState,
} from "@/components/shared"
import { Button } from "@/components/ui/button"
import { ChainStatusCard, TransactionLink } from "@/modules/wallet"

import type { DeployState, SubmitState } from "../types"

function isWalletRejected(message: string) {
  return message.toLowerCase().includes("wallet transaction was rejected")
}

function getCreateDialogState(
  state: SubmitState,
  deployState: DeployState,
): TransactionResultDialogState | null {
  if (deployState.status === "deploying") {
    return deployState.hash ? "pending" : "wallet_confirmation"
  }

  if (deployState.status === "success" && state.status === "success") {
    return "success"
  }

  if (deployState.status === "error") {
    return isWalletRejected(deployState.message) ? "rejected" : "error"
  }

  if (state.status === "error") {
    return isWalletRejected(state.message) ? "rejected" : "error"
  }

  return null
}

function getCreateDialogEventId(state: SubmitState, deployState: DeployState) {
  if (deployState.status === "deploying") {
    return `create-deploying-${deployState.marketId ?? "market"}-${deployState.hash ?? "signature"}`
  }

  if (deployState.status === "success" && state.status === "success") {
    return `create-success-${state.market.id}-${deployState.hash}`
  }

  if (deployState.status === "error") {
    return `create-error-${deployState.marketId ?? "market"}-${deployState.hash ?? "no-hash"}-${deployState.message}`
  }

  if (state.status === "error") {
    return `create-error-${state.message}-${state.requestId ?? "no-request"}`
  }

  return null
}

function getCreateDialogMessage(state: SubmitState, deployState: DeployState) {
  if (deployState.status === "deploying" && !deployState.hash) {
    return "Confirm deployment in your wallet. The backend save happens only after deployment succeeds."
  }

  if (deployState.status === "deploying") {
    return "The market contract deployment was submitted and is waiting for Arc Testnet confirmation."
  }

  if (deployState.status === "success" && state.status === "success") {
    return "The market contract deployed and the market was created after receipt confirmation."
  }

  if (deployState.status === "error") return deployState.message
  if (state.status === "error") return state.message

  return null
}

export function CreateMarketSubmitStatus({
  state,
  deployState,
  isUploadingImage,
  canSubmit,
  isConnected,
  isArcTestnet,
  isSwitchingChain,
  onSwitchChain,
}: {
  state: SubmitState
  deployState: DeployState
  isUploadingImage: boolean
  canSubmit: boolean
  isConnected: boolean
  isArcTestnet: boolean
  isSwitchingChain: boolean
  onSwitchChain: () => void
}) {
  const dialogState = getCreateDialogState(state, deployState)
  const marketId =
    deployState.status === "idle"
      ? state.status === "success"
        ? state.market.id
        : undefined
      : deployState.marketId
  const marketTitle =
    deployState.status === "idle"
      ? state.status === "success"
        ? state.market.title
        : undefined
      : deployState.marketTitle

  return (
    <>
      {!isConnected ? (
        <p className="text-sm text-muted-foreground">
          Connect a wallet to deploy the market contract before it becomes public.
        </p>
      ) : null}

      {isConnected && !isArcTestnet ? (
        <ChainStatusCard
          message="Wallet is not on Arc Testnet."
          switchLabel="Switch to Arc Testnet"
          isSwitchingChain={isSwitchingChain}
          onSwitchNetwork={onSwitchChain}
        />
      ) : null}

      {deployState.status === "deploying" ? (
        <div className="rounded-lg border border-blue-500/20 bg-blue-500/5 p-4">
          <p className="text-sm font-medium text-blue-200">
            Deploying market contract on Arc Testnet
          </p>
          {deployState.hash ? (
            <p className="mt-2 text-sm text-muted-foreground">
              Transaction: <TransactionLink hash={deployState.hash} />
            </p>
          ) : (
            <p className="mt-2 text-sm text-muted-foreground">
              Confirm the wallet transaction to continue. The backend save happens only after deployment succeeds.
            </p>
          )}
        </div>
      ) : null}

      {state.status === "error" ? (
        <InlineErrorState
          title="Unable to create market"
          message={state.message}
          requestId={state.requestId}
        />
      ) : null}

      <div className="flex flex-col gap-3 sm:flex-row">
        <Button
          disabled={!canSubmit || state.status === "submitting" || isUploadingImage}
          type="submit"
        >
          {state.status === "submitting"
            ? deployState.status === "deploying"
              ? "Deploying..."
              : "Creating..."
            : isUploadingImage
              ? "Uploading image..."
              : "Deploy and Create Market"}
        </Button>
        <Button asChild variant="outline">
          <Link href="/markets">Cancel</Link>
        </Button>
      </div>
      <TransactionResultDialog
        eventId={getCreateDialogEventId(state, deployState)}
        state={dialogState}
        actionLabel="Deploy and Create Market"
        marketLabel={marketTitle ?? marketId}
        txHash={deployState.status === "idle" ? undefined : deployState.hash}
        message={getCreateDialogMessage(state, deployState)}
        nextStep={
          dialogState === "success"
            ? "The market is public because deployment confirmed before the backend market row was created."
            : "No backend market is created unless the deployment transaction succeeds."
        }
        primaryAction={
          dialogState === "success" && marketId
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
    </>
  )
}
