import { ExternalLink } from "lucide-react"
import type { Hash } from "viem"

import { formatShortHash, getArcscanTxUrl } from "@/lib/contracts"

export function TransactionLink({
  hash,
  className = "inline-flex items-center gap-1 font-mono text-xs text-green-200 underline underline-offset-4 hover:text-green-100",
}: {
  hash: Hash
  className?: string
}) {
  return (
    <a
      className={className}
      href={getArcscanTxUrl(hash)}
      rel="noreferrer"
      target="_blank"
    >
      {formatShortHash(hash)}
      <ExternalLink className="h-3 w-3" aria-hidden="true" />
    </a>
  )
}

export function TransactionStatusLine({
  label = "Transaction",
  hash,
}: {
  label?: string
  hash: Hash
}) {
  return (
    <>
      {label}: <TransactionLink hash={hash} />
    </>
  )
}
