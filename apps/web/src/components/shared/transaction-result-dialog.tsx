"use client"

import Link from "next/link"
import { useState } from "react"
import { AlertCircle, CheckCircle2, Clock, Loader2, XCircle } from "lucide-react"
import type { Hash } from "viem"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { cn } from "@/lib/utils"
import { TransactionLink } from "@/modules/wallet"

export type TransactionResultDialogState =
  | "wallet_confirmation"
  | "pending"
  | "success"
  | "error"
  | "rejected"

type TransactionResultDetail = {
  label: string
  value: string | null | undefined
  monospace?: boolean
}

export type TransactionResultDialogProps = {
  eventId: string | null
  state: TransactionResultDialogState | null
  actionLabel: string
  marketLabel?: string | null
  outcome?: string | null
  amount?: string | null
  txHash?: Hash | null
  approvalTxHash?: Hash | null
  message?: string | null
  nextStep?: string | null
  primaryAction?: {
    label: string
    href: string
  }
  details?: TransactionResultDetail[]
}

const stateCopy: Record<
  TransactionResultDialogState,
  {
    title: string
    defaultMessage: string
    toneClass: string
    icon: "clock" | "loader" | "success" | "error" | "rejected"
  }
> = {
  wallet_confirmation: {
    title: "Confirm in Wallet",
    defaultMessage: "Your wallet is waiting for a signature before this transaction is submitted.",
    toneClass: "border-blue-500/30 bg-blue-500/10 text-blue-200",
    icon: "clock",
  },
  pending: {
    title: "Transaction Pending",
    defaultMessage: "The transaction was submitted and is waiting for Arc Testnet confirmation.",
    toneClass: "border-blue-500/30 bg-blue-500/10 text-blue-200",
    icon: "loader",
  },
  success: {
    title: "Transaction Confirmed",
    defaultMessage: "The transaction confirmed on Arc Testnet.",
    toneClass: "border-green-500/30 bg-green-500/10 text-green-200",
    icon: "success",
  },
  error: {
    title: "Transaction Failed",
    defaultMessage: "The transaction could not be completed.",
    toneClass: "border-destructive/30 bg-destructive/10 text-destructive",
    icon: "error",
  },
  rejected: {
    title: "Wallet Request Cancelled",
    defaultMessage: "The wallet request was rejected or cancelled before completion.",
    toneClass: "border-yellow-500/30 bg-yellow-500/10 text-yellow-200",
    icon: "rejected",
  },
}

function ResultIcon({ icon }: { icon: TransactionResultDialogState }) {
  const iconName = stateCopy[icon].icon

  if (iconName === "loader") {
    return <Loader2 className="h-5 w-5 animate-spin" aria-hidden="true" />
  }

  if (iconName === "success") {
    return <CheckCircle2 className="h-5 w-5" aria-hidden="true" />
  }

  if (iconName === "error") {
    return <AlertCircle className="h-5 w-5" aria-hidden="true" />
  }

  if (iconName === "rejected") {
    return <XCircle className="h-5 w-5" aria-hidden="true" />
  }

  return <Clock className="h-5 w-5" aria-hidden="true" />
}

function DetailRow({ label, value, monospace }: TransactionResultDetail) {
  if (!value) return null

  return (
    <div className="min-w-0 rounded-lg border border-border bg-background/40 p-3">
      <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
        {label}
      </dt>
      <dd
        className={cn(
          "mt-1 break-words text-sm text-foreground",
          monospace && "font-mono text-xs",
        )}
      >
        {value}
      </dd>
    </div>
  )
}

function HashRow({ label, hash }: { label: string; hash: Hash | null | undefined }) {
  if (!hash) return null

  return (
    <div className="min-w-0 rounded-lg border border-border bg-background/40 p-3">
      <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
        {label}
      </dt>
      <dd className="mt-1 min-w-0">
        <TransactionLink hash={hash} />
      </dd>
    </div>
  )
}

export function TransactionResultDialog({
  eventId,
  state,
  actionLabel,
  marketLabel,
  outcome,
  amount,
  txHash,
  approvalTxHash,
  message,
  nextStep,
  primaryAction,
  details = [],
}: TransactionResultDialogProps) {
  const [dismissedEventId, setDismissedEventId] = useState<string | null>(null)
  const open = Boolean(eventId && state && dismissedEventId !== eventId)

  if (!eventId || !state) return null

  const copy = stateCopy[state]
  const description = message || copy.defaultMessage

  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        if (!nextOpen) {
          setDismissedEventId(eventId)
        }
      }}
    >
      <DialogContent
        aria-describedby="transaction-result-description"
        className="max-h-[min(88vh,720px)] overflow-y-auto sm:max-w-lg"
      >
        <DialogHeader className="pr-8">
          <div
            className={cn(
              "mb-1 flex h-10 w-10 items-center justify-center rounded-full border",
              copy.toneClass,
            )}
          >
            <ResultIcon icon={state} />
          </div>
          <DialogTitle>{copy.title}</DialogTitle>
          <DialogDescription
            id="transaction-result-description"
            className="leading-6"
          >
            {description}
          </DialogDescription>
        </DialogHeader>

        <dl className="grid gap-3">
          <DetailRow label="Action" value={actionLabel} />
          <DetailRow label="Market" value={marketLabel} />
          <DetailRow label="Outcome" value={outcome} />
          <DetailRow label="Amount" value={amount} />
          {details.map((detail) => (
            <DetailRow key={`${detail.label}-${detail.value}`} {...detail} />
          ))}
          <HashRow label="Approval Transaction" hash={approvalTxHash} />
          <HashRow label="Transaction" hash={txHash} />
        </dl>

        {nextStep ? (
          <p className="rounded-lg border border-border bg-muted/20 p-3 text-sm leading-6 text-muted-foreground">
            {nextStep}
          </p>
        ) : null}

        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Close</Button>
          </DialogClose>
          {primaryAction ? (
            <Button asChild>
              <Link href={primaryAction.href}>{primaryAction.label}</Link>
            </Button>
          ) : null}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
