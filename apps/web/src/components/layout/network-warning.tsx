"use client"

import { useAccount, useChainId, useSwitchChain } from "wagmi"

import { Button } from "@/components/ui/button"
import { arcTestnet } from "@/lib/wagmi"

export function NetworkWarning() {
  const { isConnected } = useAccount()
  const chainId = useChainId()
  const { switchChain, isPending } = useSwitchChain()

  if (!isConnected) {
    return null
  }

  if (chainId === arcTestnet.id) {
    return null
  }

  return (
    <div className="border-b border-yellow-500/30 bg-yellow-500/10 px-4 py-2.5">
      <div className="mx-auto flex max-w-7xl flex-col items-start gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-2">
          <svg
            className="h-4 w-4 shrink-0 text-yellow-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"
            />
          </svg>
          <p className="text-sm font-medium text-yellow-200">
            Wrong network detected. Please switch to Arc Testnet (Chain ID: {arcTestnet.id}).
          </p>
        </div>
        <Button
          size="sm"
          variant="outline"
          className="border-yellow-500/50 text-yellow-200 hover:bg-yellow-500/20"
          disabled={isPending}
          onClick={() => switchChain({ chainId: arcTestnet.id })}
        >
          {isPending ? "Switching..." : "Switch to Arc Testnet"}
        </Button>
      </div>
    </div>
  )
}
