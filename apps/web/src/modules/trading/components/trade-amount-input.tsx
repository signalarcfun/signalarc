"use client"

import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

const quickAmounts = ["1", "5", "10", "25"]

export function TradeAmountInput({
  value,
  onChange,
}: {
  value: string
  onChange: (value: string) => void
}) {
  return (
    <div className="grid gap-3">
      <Label htmlFor="amount">Amount (USDC)</Label>
      <Input
        className="h-12 text-base sm:text-sm"
        id="amount"
        inputMode="decimal"
        min="0"
        name="amount"
        onChange={(event) => onChange(event.target.value)}
        required
        step="0.000001"
        value={value}
      />
      <div className="grid grid-cols-4 gap-2">
        {quickAmounts.map((quickAmount) => (
          <button
            key={quickAmount}
            type="button"
            onClick={() => onChange(quickAmount)}
            className="min-h-10 rounded-md border border-border bg-muted/30 px-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent/50 hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
          >
            {quickAmount}
          </button>
        ))}
      </div>
    </div>
  )
}
