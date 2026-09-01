import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

import { defaultCloseValue } from "../form-utils"

export function CreateMarketScheduleFields() {
  return (
    <div className="grid gap-4 sm:grid-cols-2">
      <div className="grid gap-2">
        <Label htmlFor="opens_at">Opens At (optional)</Label>
        <Input id="opens_at" name="opens_at" type="datetime-local" />
      </div>
      <div className="grid gap-2">
        <Label htmlFor="closes_at">Closes At</Label>
        <Input
          id="closes_at"
          name="closes_at"
          defaultValue={defaultCloseValue()}
          required
          type="datetime-local"
        />
      </div>
    </div>
  )
}
