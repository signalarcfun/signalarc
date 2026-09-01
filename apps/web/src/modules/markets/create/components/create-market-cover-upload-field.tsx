"use client"

import { MarketImageUpload } from "@/components/markets/market-image-upload"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export function CreateMarketCoverUploadField({
  coverImageUrl,
  onCoverImageUrlChange,
  onUploadingChange,
  disabled,
  isUploadingImage,
}: {
  coverImageUrl: string
  onCoverImageUrlChange: (value: string) => void
  onUploadingChange: (value: boolean) => void
  disabled: boolean
  isUploadingImage: boolean
}) {
  return (
    <>
      <div className="grid gap-2">
        <Label>Market Image</Label>
        <MarketImageUpload
          coverImageUrl={coverImageUrl}
          onChangeUrl={(url) => onCoverImageUrlChange(url ?? "")}
          onUploadingChange={onUploadingChange}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Optional. Upload an image directly, or paste an HTTPS image URL below.
        </p>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="cover_image_url">Market Image URL (advanced)</Label>
        <Input
          id="cover_image_url"
          name="cover_image_url"
          type="url"
          inputMode="url"
          maxLength={2048}
          pattern="https://.*"
          placeholder="https://example.com/market-cover.png"
          value={coverImageUrl}
          onChange={(event) => onCoverImageUrlChange(event.target.value)}
          disabled={isUploadingImage || disabled}
        />
        <p className="text-xs text-muted-foreground">
          Paste a public HTTPS image URL as an alternative to uploading.
        </p>
      </div>
    </>
  )
}
