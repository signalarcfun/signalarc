/**
 * Client-side Cloudinary helpers for market cover image uploads.
 *
 * This module is safe to import in browser code.  No API keys or secrets are
 * referenced here — the actual upload is proxied through the server-side
 * Next.js API route at /api/market-cover-upload.
 *
 * Only **client-side file validation** and the fetch to the API route live here.
 * Server-side signing and Cloudinary communication live in lib/cloudinary-upload.ts.
 */

/** Accepted MIME types for market cover images. Must match server validation. */
const ACCEPTED_MIME_TYPES: ReadonlySet<string> = new Set([
  "image/jpeg",
  "image/png",
  "image/webp",
])

/** 5 MB. Must match server validation. */
export const MAX_FILE_SIZE_BYTES = 5 * 1024 * 1024

export class MarketCoverUploadError extends Error {
  constructor(message: string) {
    super(message)
    this.name = "MarketCoverUploadError"
  }
}

/**
 * Validates a selected cover image file before any network request is made.
 * Returns a user-friendly error message, or `null` when the file is valid.
 */
export function validateMarketCoverFile(file: File): string | null {
  if (!ACCEPTED_MIME_TYPES.has(file.type)) {
    return "Image must be a JPEG, PNG, or WebP file."
  }

  if (file.size === 0) {
    return "The selected image file is empty."
  }

  if (file.size > MAX_FILE_SIZE_BYTES) {
    return "Image must be smaller than 5 MB."
  }

  return null
}

type ApiUploadResponse = {
  secure_url?: string
  error?: string
}

/**
 * Uploads a market cover image by POSTing it to the server-side API route.
 *
 * Client-side validation is performed first.  Then the file is forwarded to
 * `/api/market-cover-upload`, where it is validated again server-side and
 * uploaded to Cloudinary using signed authentication.
 *
 * Throws {@link MarketCoverUploadError} for validation, network, or
 * server-reported failures.
 */
export async function uploadMarketCoverImage(
  file: File,
): Promise<string> {
  const validationError = validateMarketCoverFile(file)
  if (validationError) {
    throw new MarketCoverUploadError(validationError)
  }

  const formData = new FormData()
  formData.append("file", file)

  let response: Response
  try {
    response = await fetch("/api/market-cover-upload", {
      method: "POST",
      body: formData,
    })
  } catch {
    throw new MarketCoverUploadError(
      "Unable to reach the image upload service. Check your connection and try again.",
    )
  }

  let payload: ApiUploadResponse | null = null
  try {
    payload = (await response.json()) as ApiUploadResponse
  } catch {
    payload = null
  }

  if (!response.ok) {
    const message = payload?.error
    throw new MarketCoverUploadError(
      message && message.trim() !== ""
        ? message
        : "Image upload failed. Try a different image.",
    )
  }

  const secureUrl = payload?.secure_url
  if (!secureUrl || !secureUrl.startsWith("https://")) {
    throw new MarketCoverUploadError(
      "Image upload completed but no image URL was returned.",
    )
  }

  return secureUrl
}
