import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export function CreateMarketOutcomeFields() {
  return (
    <div className="grid gap-4 sm:grid-cols-2">
      <div className="grid gap-2">
        <Label htmlFor="outcome_yes_label">YES Label</Label>
        <Input
          id="outcome_yes_label"
          name="outcome_yes_label"
          placeholder="Yes"
        />
      </div>
      <div className="grid gap-2">
        <Label htmlFor="outcome_no_label">NO Label</Label>
        <Input id="outcome_no_label" name="outcome_no_label" placeholder="No" />
      </div>
    </div>
  )
}
