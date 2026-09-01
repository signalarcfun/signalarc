"use client"

import { Button } from "@/components/ui/button"

import { MARKET_DISCOVERY_TABS } from "../discovery"
import type { DiscoveryTabId } from "../types"

export function DiscoveryTabs({
  value,
  onChange,
}: {
  value: DiscoveryTabId
  onChange: (value: DiscoveryTabId) => void
}) {
  return (
    <div className="-mx-4 overflow-x-auto px-4 sm:mx-0 sm:px-0">
      <div className="flex min-w-max items-center gap-1.5">
        {MARKET_DISCOVERY_TABS.map((tab) => (
          <Button
            key={tab.id}
            size="sm"
            variant={value === tab.id ? "default" : "outline"}
            className={
              value === tab.id
                ? "h-8 rounded-full px-4 text-xs"
                : "h-8 rounded-full px-4 text-xs text-muted-foreground"
            }
            aria-pressed={value === tab.id}
            onClick={() => onChange(tab.id)}
          >
            {tab.label}
          </Button>
        ))}
      </div>
    </div>
  )
}
