"use client"

import { Button } from "@/components/ui/button"
import { MARKET_CATEGORIES, type MarketCategoryId } from "@/modules/categories"

export function MarketCategoryFilter({
  value,
  onChange,
}: {
  value: MarketCategoryId
  onChange: (value: MarketCategoryId) => void
}) {
  return (
    <div className="-mx-4 overflow-x-auto px-4 sm:mx-0 sm:px-0">
      <div className="flex min-w-max items-center gap-1.5">
        {MARKET_CATEGORIES.map((category) => (
          <Button
            key={category.id}
            size="sm"
            variant={value === category.id ? "default" : "outline"}
            className={
              value === category.id
                ? "h-7 rounded-full px-3 text-xs"
                : "h-7 rounded-full px-3 text-xs text-muted-foreground"
            }
            aria-pressed={value === category.id}
            onClick={() => onChange(category.id)}
          >
            {category.label}
          </Button>
        ))}
      </div>
    </div>
  )
}
