import { Card, CardContent } from "@/components/ui/card"

import { formatWalletAddress } from "../format"

function WalletIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth={1.5}
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M21 12a2.25 2.25 0 00-2.25-2.25H15a3 3 0 11-6 0H5.25A2.25 2.25 0 003 12m18 0v6a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 18v-6m18 0V9M3 12V9m18 0a2.25 2.25 0 00-2.25-2.25H5.25A2.25 2.25 0 003 9m18 0V6a2.25 2.25 0 00-2.25-2.25H5.25A2.25 2.25 0 003 6v3"
      />
    </svg>
  )
}

export function WalletNotConnectedState() {
  return (
    <Card>
      <CardContent className="flex flex-col items-start gap-4 p-5 sm:flex-row sm:items-center sm:p-6">
        <div className="flex h-12 w-12 items-center justify-center rounded-full bg-muted">
          <WalletIcon className="h-6 w-6 text-muted-foreground" />
        </div>
        <div className="min-w-0">
          <h2 className="text-base font-semibold text-foreground">
            Connect wallet to anchor portfolio context
          </h2>
          <p className="mt-1 text-sm leading-6 text-muted-foreground">
            After connecting, this page can show the active wallet address and
            align future position, exposure, claim, refund, and activity views
            around that wallet.
          </p>
          <p className="mt-2 text-xs text-muted-foreground/70">
            Backend wallet-indexed portfolio loading is not implemented yet; use
            the API lookup below for existing records.
          </p>
        </div>
      </CardContent>
    </Card>
  )
}

export function WalletIdentityCard({ address }: { address: string }) {
  return (
    <Card className="border-indigo-500/20">
      <CardContent className="flex flex-col gap-3 p-5 sm:flex-row sm:items-center sm:p-6">
        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-indigo-500/10">
          <WalletIcon className="h-5 w-5 text-indigo-400" />
        </div>
        <div className="min-w-0">
          <p className="text-sm font-medium text-foreground">Connected Wallet</p>
          <p className="mt-1 break-all font-mono text-xs text-muted-foreground" title={address}>
            {formatWalletAddress(address)}
          </p>
        </div>
      </CardContent>
    </Card>
  )
}
