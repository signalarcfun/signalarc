import type { Hash } from "viem"

import {
  TransactionLink as SharedTransactionLink,
  TransactionStatusLine,
} from "@/components/shared"

export function TransactionLink({
  hash,
  className = "inline-flex items-center gap-1 font-mono text-xs text-green-200 underline underline-offset-4 hover:text-green-100",
}: {
  hash: Hash
  className?: string
}) {
  return <SharedTransactionLink hash={hash} className={className} />
}

export function TransactionStatusCard({
  label = "Transaction",
  hash,
}: {
  label?: string
  hash: Hash
}) {
  return <TransactionStatusLine label={label} hash={hash} />
}
