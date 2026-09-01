"use client"

import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export function CreateMarketAdvancedFields({
  showAdvanced,
  onToggleAdvanced,
}: {
  showAdvanced: boolean
  onToggleAdvanced: () => void
}) {
  return (
    <div>
      <button
        type="button"
        onClick={onToggleAdvanced}
        className="text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
      >
        {showAdvanced ? "Hide" : "Show"} advanced settings
      </button>

      {showAdvanced ? (
        <div className="mt-3 grid gap-4 rounded-lg border border-border/50 p-4 sm:grid-cols-2">
          <div className="grid gap-2">
            <Label htmlFor="collateral_asset" className="text-xs">
              Collateral Asset
            </Label>
            <Input
              id="collateral_asset"
              name="collateral_asset"
              defaultValue="USDC"
              className="text-sm"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="chain" className="text-xs">
              Chain
            </Label>
            <Input
              id="chain"
              name="chain"
              defaultValue="Arc Testnet"
              className="text-sm"
            />
          </div>
        </div>
      ) : (
        <>
          <input type="hidden" name="collateral_asset" value="USDC" />
          <input type="hidden" name="chain" value="Arc Testnet" />
        </>
      )}
    </div>
  )
}
