import { AlertCircle, CheckCircle2, Loader2 } from "lucide-react"

import { TransactionLink } from "@/modules/wallet"

import type { TradeSubmitState } from "../types"

function TradeTransactionRow({
  label,
  hash,
  status,
}: {
  label: string
  hash?: `0x${string}`
  status: "pending" | "success" | "error" | "not_available"
}) {
  return (
    <div className="min-w-0 rounded-lg border border-border bg-background/40 p-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <dt className="text-xs uppercase tracking-wider text-muted-foreground/70">
          {label}
        </dt>
        <span className="rounded-md border border-border bg-background/50 px-2 py-1 text-xs capitalize text-muted-foreground">
          {status.replace("_", " ")}
        </span>
      </div>
      <dd className="mt-2">
        {hash ? (
          <TransactionLink hash={hash} />
        ) : (
          <span className="text-sm text-muted-foreground">Not available</span>
        )}
      </dd>
    </div>
  )
}

function TradeExecutionResult({
  state,
}: {
  state: Extract<TradeSubmitState, { status: "success" }>
}) {
  return (
    <div className="rounded-lg border border-green-500/20 bg-green-500/5 p-4">
      <div className="flex items-center gap-2">
        <CheckCircle2 className="h-4 w-4 text-green-400" aria-hidden="true" />
        <p className="text-sm font-medium text-green-300">
          Position opened on Arc Testnet
        </p>
      </div>
      <p className="mt-1 text-sm text-muted-foreground">
        Your connected wallet opened this position on Arc Testnet. Track the
        transactions below before refreshing balances.
      </p>
      <dl className="mt-4 grid gap-3 text-sm sm:grid-cols-2">
        <TradeTransactionRow
          label="USDC Approval"
          hash={state.approveHash}
          status="success"
        />
        <TradeTransactionRow
          label="Open Position"
          hash={state.openHash}
          status="success"
        />
      </dl>
    </div>
  )
}

function PendingState({
  state,
}: {
  state: Extract<TradeSubmitState, { status: "approving" | "opening" }>
}) {
  const label = state.status === "approving" ? "Approving USDC" : "Opening position"

  return (
    <div className="rounded-lg border border-blue-500/20 bg-blue-500/5 p-4">
      <div className="flex items-center gap-2">
        <Loader2 className="h-4 w-4 animate-spin text-blue-300" aria-hidden="true" />
        <p className="text-sm font-medium text-blue-200">{label}</p>
      </div>
      <p className="mt-1 text-sm text-muted-foreground">
        Keep this page open. The ticket will update after wallet confirmation and
        Arc Testnet receipt finality.
      </p>
      {state.approveHash ? (
        <dl className="mt-3">
          <TradeTransactionRow
            label="Approval"
            hash={state.approveHash}
            status="pending"
          />
        </dl>
      ) : null}
      {state.status === "opening" && state.openHash ? (
        <dl className="mt-2">
          <TradeTransactionRow
            label="Market transaction"
            hash={state.openHash}
            status="pending"
          />
        </dl>
      ) : null}
    </div>
  )
}

export function TradeSubmitStatus({ state }: { state: TradeSubmitState }) {
  return (
    <>
      {state.status === "error" ? (
        <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-4">
          <div className="flex items-center gap-2">
            <AlertCircle className="h-4 w-4 text-destructive" aria-hidden="true" />
            <p className="text-sm font-medium text-destructive">
              Unable to execute trade
            </p>
          </div>
          <p className="mt-1 text-sm text-muted-foreground">{state.message}</p>
          <p className="mt-2 text-xs text-muted-foreground">
            No new position is recorded unless the market transaction confirms.
          </p>
          {state.approveHash ? (
            <dl className="mt-3">
              <TradeTransactionRow
                label="Approval"
                hash={state.approveHash}
                status="error"
              />
            </dl>
          ) : null}
          {state.openHash ? (
            <dl className="mt-2">
              <TradeTransactionRow
                label="Market transaction"
                hash={state.openHash}
                status="error"
              />
            </dl>
          ) : null}
        </div>
      ) : null}

      {state.status === "approving" || state.status === "opening" ? (
        <PendingState state={state} />
      ) : null}

      {state.status === "success" ? <TradeExecutionResult state={state} /> : null}
    </>
  )
}
