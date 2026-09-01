"use client"

import Link from "next/link"
import { CheckCircle2, Loader2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import type { Market } from "@/lib/api"
import {
  SIGNAL_ARC_MARKET_FACTORY_ADDRESS,
} from "@/lib/contracts"
import { ChainStatusCard, TransactionLink } from "@/modules/wallet"

import type { DeployState } from "../types"

export function CreateMarketDeployPanel({
  market,
  deployState,
  canDeploy,
  isConnected,
  isArcTestnet,
  isSwitchingChain,
  isDeploying,
  onSwitchChain,
  onDeploy,
}: {
  market: Market
  deployState: DeployState
  canDeploy: boolean
  isConnected: boolean
  isArcTestnet: boolean
  isSwitchingChain: boolean
  isDeploying: boolean
  onSwitchChain: () => void
  onDeploy: () => void
}) {
  return (
    <Card className="border-green-500/20">
      <CardHeader>
        <div className="flex items-center gap-2">
          <CheckCircle2 className="h-5 w-5 text-green-400" aria-hidden="true" />
          <CardTitle>Market Created</CardTitle>
        </div>
        <CardDescription>{market.title}</CardDescription>
      </CardHeader>
      <CardContent className="grid gap-4">
        <div className="rounded-lg border border-yellow-500/30 bg-yellow-500/5 p-4">
          <p className="text-sm font-medium text-yellow-300">
            Onchain contract not deployed.
          </p>
          <p className="mt-1 text-sm text-muted-foreground">
            Your connected wallet will be the resolver for this testnet market.
          </p>
          {SIGNAL_ARC_MARKET_FACTORY_ADDRESS ? (
            <p className="mt-2 font-mono text-xs text-muted-foreground">
              Factory: {SIGNAL_ARC_MARKET_FACTORY_ADDRESS}
            </p>
          ) : (
            <p className="mt-2 text-sm text-muted-foreground">
              Factory address not configured.
            </p>
          )}
        </div>

        {!isConnected ? (
          <p className="text-sm text-muted-foreground">
            Connect a wallet to deploy the Arc Testnet market contract.
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
            <div className="flex items-center gap-2">
              <Loader2
                className="h-4 w-4 animate-spin text-blue-300"
                aria-hidden="true"
              />
              <p className="text-sm font-medium text-blue-200">
                Deploying market contract
              </p>
            </div>
            {deployState.hash ? (
              <p className="mt-2 text-sm text-muted-foreground">
                Transaction: <TransactionLink hash={deployState.hash} />
              </p>
            ) : null}
          </div>
        ) : null}

        {deployState.status === "error" ? (
          <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-4">
            <p className="text-sm font-medium text-destructive">
              Unable to deploy onchain market
            </p>
            <p className="mt-1 text-sm text-muted-foreground">
              {deployState.message}
            </p>
            {deployState.hash ? (
              <p className="mt-2 text-sm text-muted-foreground">
                Transaction: <TransactionLink hash={deployState.hash} />
              </p>
            ) : null}
          </div>
        ) : null}

        {deployState.status === "success" ? (
          <div className="rounded-lg border border-green-500/20 bg-green-500/5 p-4">
            <p className="text-sm font-medium text-green-300">
              Market contract deployed on Arc Testnet.
            </p>
            <p className="mt-2 font-mono text-xs text-muted-foreground">
              {deployState.marketAddress}
            </p>
            <p className="mt-2 text-sm text-muted-foreground">
              Transaction: <TransactionLink hash={deployState.hash} />
            </p>
          </div>
        ) : null}

        <div className="flex flex-col gap-3 sm:flex-row">
          <Button disabled={!canDeploy} onClick={onDeploy} type="button">
            {isDeploying ? "Deploying..." : "Deploy on Arc Testnet"}
          </Button>
          <Button asChild variant="outline">
            <Link href={`/markets/${market.id}`}>View Market</Link>
          </Button>
          <Button asChild variant="outline">
            <Link href="/markets">Back to Markets</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
