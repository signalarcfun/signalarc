"use client"

import { AlertTriangle, Ban, CircleDollarSign, Wallet } from "lucide-react"

import { Button } from "@/components/ui/button"
import { getWalletConnectionMessage } from "@/modules/wallet"

function getStateTone(disabledReason: string) {
  const lowerReason = disabledReason.toLowerCase()

  if (lowerReason.includes("resolved") || lowerReason.includes("cancelled")) {
    return {
      className: "border-indigo-500/30 bg-indigo-500/5",
      textClassName: "text-indigo-200",
      icon: CircleDollarSign,
      title: "Settlement state",
    }
  }

  if (lowerReason.includes("closed") || lowerReason.includes("not open")) {
    return {
      className: "border-yellow-500/30 bg-yellow-500/5",
      textClassName: "text-yellow-300",
      icon: Ban,
      title: "Trading unavailable",
    }
  }

  if (lowerReason.includes("wallet") || lowerReason.includes("network")) {
    return {
      className: "border-blue-500/30 bg-blue-500/5",
      textClassName: "text-blue-200",
      icon: Wallet,
      title: "Wallet required",
    }
  }

  return {
    className: "border-yellow-500/30 bg-yellow-500/5",
    textClassName: "text-yellow-300",
    icon: AlertTriangle,
    title: "Action needed",
  }
}

export function TradeWalletState({
  disabledReason,
  isConnected,
  isArcTestnet,
  isSwitchingChain,
  onSwitchNetwork,
}: {
  disabledReason: string | null
  isConnected: boolean
  isArcTestnet: boolean
  isSwitchingChain: boolean
  onSwitchNetwork: () => void
}) {
  if (!disabledReason) return null
  const tone = getStateTone(disabledReason)
  const Icon = tone.icon

  return (
    <div className={`rounded-lg border p-4 ${tone.className}`}>
      <div className="flex items-start gap-3">
        <Icon className={`mt-0.5 h-4 w-4 ${tone.textClassName}`} aria-hidden="true" />
        <div className="min-w-0">
          <p className={`text-sm font-medium ${tone.textClassName}`}>
            {tone.title}
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{disabledReason}</p>
        </div>
      </div>
      {!isConnected ? (
        <p className="mt-1 text-sm text-muted-foreground">
          {getWalletConnectionMessage(isConnected)}
        </p>
      ) : null}
      {isConnected && !isArcTestnet ? (
        <Button
          className="mt-3"
          disabled={isSwitchingChain}
          onClick={onSwitchNetwork}
          size="sm"
          type="button"
          variant="outline"
        >
          {isSwitchingChain ? "Switching..." : "Switch network"}
        </Button>
      ) : null}
    </div>
  )
}
