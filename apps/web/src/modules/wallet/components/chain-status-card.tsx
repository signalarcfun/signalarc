"use client"

import { Button } from "@/components/ui/button"

import { getArcTestnetSwitchLabel } from "../wallet-utils"

export function ChainStatusCard({
  message = "Wallet is not on Arc Testnet.",
  switchLabel,
  isSwitchingChain,
  onSwitchNetwork,
}: {
  message?: string
  switchLabel?: string
  isSwitchingChain: boolean
  onSwitchNetwork: () => void
}) {
  return (
    <div className="rounded-lg border border-yellow-500/30 bg-yellow-500/5 p-4">
      <p className="text-sm font-medium text-yellow-300">{message}</p>
      <Button
        className="mt-3"
        disabled={isSwitchingChain}
        onClick={onSwitchNetwork}
        size="sm"
        type="button"
        variant="outline"
      >
        {isSwitchingChain ? "Switching..." : (switchLabel ?? getArcTestnetSwitchLabel(isSwitchingChain))}
      </Button>
    </div>
  )
}
