import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"

export function CreateMarketFields() {
  return (
    <>
      <div className="grid gap-2">
        <Label htmlFor="title">Market Question</Label>
        <Input
          id="title"
          name="title"
          placeholder="Will ETH reach $5,000 by end of 2026?"
          required
        />
        <p className="text-xs text-muted-foreground">
          Frame as a yes/no question that can be resolved objectively.
        </p>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="description">Description (optional)</Label>
        <Textarea
          id="description"
          name="description"
          rows={3}
          placeholder="Additional context, resolution criteria, or relevant links..."
        />
      </div>

      <div className="grid gap-2">
        <Label htmlFor="resolution_source">Resolution Source</Label>
        <Input
          id="resolution_source"
          name="resolution_source"
          placeholder="CoinGecko, AP News, Official announcement..."
        />
      </div>
    </>
  )
}
