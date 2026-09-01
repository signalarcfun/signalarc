"use client"

import {
  CREATE_MARKET_CATEGORIES,
  DEFAULT_CATEGORY_ID,
} from "@/modules/categories"

export function CreateMarketCategoryField({
  disabled,
}: {
  disabled: boolean
}) {
  return (
    <fieldset className="grid gap-2" disabled={disabled}>
      <legend className="text-sm font-medium leading-none">Category</legend>
      <div
        role="radiogroup"
        aria-label="Category"
        className="grid grid-cols-2 gap-2 sm:grid-cols-4"
      >
        {CREATE_MARKET_CATEGORIES.map((category) => (
          <label
            key={category.id}
            className="relative flex cursor-pointer items-center justify-center has-[:disabled]:cursor-not-allowed"
          >
            <input
              type="radio"
              name="category"
              value={category.id}
              defaultChecked={category.id === DEFAULT_CATEGORY_ID}
              disabled={disabled}
              className="peer sr-only"
            />
            <span className="flex h-9 w-full items-center justify-center rounded-md border border-input bg-transparent px-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted/40 peer-checked:border-indigo-500 peer-checked:bg-indigo-500/10 peer-checked:text-indigo-300 peer-focus-visible:outline-none peer-focus-visible:ring-1 peer-focus-visible:ring-ring peer-disabled:opacity-50">
              {category.label}
            </span>
          </label>
        ))}
      </div>
    </fieldset>
  )
}
