"use client"

import type { FormEvent, ReactNode } from "react"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { localDemoUserId } from "@/lib/api"

export function PortfolioShell({
  children,
}: {
  children: ReactNode
}) {
  return <div className="grid gap-4 sm:gap-6">{children}</div>
}

export function PortfolioAdvancedLookup({
  showAdvanced,
  isLoading,
  onToggleAdvanced,
  onSubmit,
}: {
  showAdvanced: boolean
  isLoading: boolean
  onToggleAdvanced: () => void
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}) {
  return (
    <div>
      <button
        type="button"
        onClick={onToggleAdvanced}
        className="text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
      >
        {showAdvanced ? "Hide" : "Show"} API position lookup
      </button>

      {showAdvanced ? (
        <Card className="mt-3">
          <CardHeader>
            <CardTitle className="text-sm">API Position Lookup</CardTitle>
            <CardDescription className="text-xs leading-5">
              Load existing backend position and settlement records by user ID.
              Wallet-indexed portfolio loading is not implemented yet.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form
              className="flex flex-col gap-3 sm:flex-row sm:items-end"
              onSubmit={onSubmit}
            >
              <div className="grid flex-1 gap-2">
                <Label htmlFor="user_id" className="text-xs">
                  User ID
                </Label>
                <Input
                  id="user_id"
                  name="user_id"
                  defaultValue={localDemoUserId}
                  placeholder="UUID"
                  className="text-sm"
                />
              </div>
              <Button disabled={isLoading} type="submit" size="sm" className="w-full sm:w-auto">
                {isLoading ? "Loading..." : "Load"}
              </Button>
            </form>
          </CardContent>
        </Card>
      ) : null}
    </div>
  )
}
