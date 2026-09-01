"use client"

import { useRef, useState, type ChangeEvent } from "react"
import { Loader2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import {
  MarketCoverUploadError,
  uploadMarketCoverImage,
  validateMarketCoverFile,
} from "@/lib/cloudinary"

function UploadIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.75}
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-13.5-9L12 3m0 0l4.5 4.5M12 3v13.5"
      />
    </svg>
  )
}

function RemoveIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.75}
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M6 18L18 6M6 6l12 12"
      />
    </svg>
  )
}

type MarketImageUploadProps = {
  /** Current cover image URL, if any (from upload or manual paste). */
  coverImageUrl: string
  /** Called whenever the cover image URL changes (set or cleared). */
  onChangeUrl: (url: string | undefined) => void
  /** Called whenever the upload-in-progress state changes. */
  onUploadingChange: (uploading: boolean) => void
  disabled?: boolean
}

export function MarketImageUpload({
  coverImageUrl,
  onChangeUrl,
  onUploadingChange,
  disabled = false,
}: MarketImageUploadProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [isUploading, setIsUploading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function setUploading(value: boolean) {
    setIsUploading(value)
    onUploadingChange(value)
  }

  function resetInput() {
    if (inputRef.current) {
      inputRef.current.value = ""
    }
  }

  async function handleFileChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    // Always reset the native input so the same file can be re-selected later.
    resetInput()

    if (!file) return

    const validationError = validateMarketCoverFile(file)
    if (validationError) {
      setError(validationError)
      return
    }

    setError(null)
    setUploading(true)

    try {
      const secureUrl = await uploadMarketCoverImage(file)
      onChangeUrl(secureUrl)
    } catch (uploadError) {
      const message =
        uploadError instanceof MarketCoverUploadError
          ? uploadError.message
          : "Image upload failed. Try a different image."
      setError(message)
    } finally {
      setUploading(false)
    }
  }

  function handleRemove() {
    setError(null)
    onChangeUrl(undefined)
    resetInput()
  }

  const isDisabled = disabled || isUploading

  return (
    <div className="grid gap-2">
      <div className="flex flex-wrap items-center gap-2">
        <input
          ref={inputRef}
          type="file"
          accept="image/jpeg,image/png,image/webp"
          className="sr-only"
          onChange={handleFileChange}
          disabled={isDisabled}
          aria-label="Upload market image"
        />
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => inputRef.current?.click()}
          disabled={isDisabled}
        >
          {isUploading ? (
            <>
              <Loader2 className="animate-spin" aria-hidden="true" />
              Uploading...
            </>
          ) : (
            <>
              <UploadIcon />
              Upload image
            </>
          )}
        </Button>

        {coverImageUrl ? (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={handleRemove}
            disabled={isDisabled}
          >
            <RemoveIcon />
            Remove
          </Button>
        ) : null}

        <p className="text-xs text-muted-foreground">
          JPEG, PNG, or WebP. Max 5&nbsp;MB.
        </p>
      </div>

      {coverImageUrl ? (
        <div className="relative w-full overflow-hidden rounded-lg border border-border/60">
          {/* Plain img is intentional: the src is a user-provided remote URL. */}
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={coverImageUrl}
            alt="Market cover preview"
            className={cn(
              "h-40 w-full object-cover",
              isUploading && "opacity-60",
            )}
          />
        </div>
      ) : null}

      {error ? (
        <p role="alert" className="text-xs text-destructive">
          {error}
        </p>
      ) : null}
    </div>
  )
}
