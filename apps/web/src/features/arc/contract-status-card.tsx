"use client"

import { useEffect, useState } from "react"

import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { ApiError, getArcContract, type ArcContractStatus } from "@/lib/api"

type ContractState =
  | { status: "loading" }
  | { status: "loaded"; contract: ArcContractStatus }
  | { status: "error"; message: string; requestId: string | null }

function getErrorState(error: unknown): Extract<ContractState, { status: "error" }> {
  if (error instanceof ApiError) {
    return {
      status: "error",
      message: error.message,
      requestId: error.requestId,
    }
  }

  if (error instanceof Error) {
    return {
      status: "error",
      message: error.message,
      requestId: null,
    }
  }

  return {
    status: "error",
    message: "Unable to load Arc contract status.",
    requestId: null,
  }
}

function ContractField({ label, value, mono = false }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="min-w-0">
      <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">{label}</dt>
      <dd className={`mt-1 break-all text-sm text-foreground ${mono ? "font-mono text-xs" : ""}`}>{value}</dd>
    </div>
  )
}

export function ContractStatusCard() {
  const [state, setState] = useState<ContractState>({ status: "loading" })

  useEffect(() => {
    let isActive = true

    async function loadContract() {
      try {
        const response = await getArcContract()
        if (isActive) {
          setState({ status: "loaded", contract: response.data })
        }
      } catch (error) {
        if (isActive) {
          setState(getErrorState(error))
        }
      }
    }

    void loadContract()

    return () => {
      isActive = false
    }
  }, [])

  if (state.status === "loading") {
    return (
      <Card className="animate-pulse">
        <CardHeader>
          <div className="h-5 w-1/3 rounded bg-muted" />
          <div className="mt-2 h-4 w-2/3 rounded bg-muted" />
        </CardHeader>
      </Card>
    )
  }

  if (state.status === "error") {
    return (
      <Card className="border-destructive/30 bg-destructive/5">
        <CardHeader>
          <CardTitle className="text-destructive">Arc contract unavailable</CardTitle>
          <CardDescription>{state.message}</CardDescription>
        </CardHeader>
        {state.requestId ? (
          <CardContent>
            <p className="font-mono text-xs text-muted-foreground">
              Request ID: {state.requestId}
            </p>
          </CardContent>
        ) : null}
      </Card>
    )
  }

  const { contract } = state

  return (
    <Card className="border-indigo-500/20">
      <CardHeader>
        <div className="flex flex-wrap items-center gap-2">
          <CardTitle>Arc Testnet Contract</CardTitle>
          <Badge variant="outline" className="border-indigo-500/30 bg-indigo-500/10 text-indigo-300">
            Prototype
          </Badge>
          <Badge variant="secondary" className="text-xs">
            Not for live use
          </Badge>
        </div>
        <CardDescription>
          Deployed SignalArcMarket prototype reference on Arc Testnet.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <dl className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <ContractField label="Network" value={contract.network} />
          <ContractField label="Chain ID" value={String(contract.chain_id)} />
          <ContractField label="SignalArcMarket" value={contract.signal_arc_market} mono />
          <ContractField label="USDC ERC20" value={contract.usdc_erc20_interface} mono />
          <ContractField label="Explorer" value={contract.explorer} />
          <ContractField label="Status" value={contract.status} />
        </dl>
      </CardContent>
    </Card>
  )
}
