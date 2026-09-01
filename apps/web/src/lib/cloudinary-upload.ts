/**
 * Server-side Cloudinary upload helpers.
 *
 * This module is ONLY imported by the API route and must never be referenced
 * from client components or browser bundles.  It reads server-only env vars
 * (`CLOUDINARY_API_KEY`, `CLOUDINARY_API_SECRET`) that are never exposed to
 * the browser.
 */

import crypto from "node:crypto"

// ---------------------------------------------------------------------------
// Env configuration
// ---------------------------------------------------------------------------

function requireEnv(value: string | undefined, name: string): string {
  const trimmed = value?.trim()
  if (!trimmed) {
    throw new Error(`Missing required server env: ${name}`)
  }
  return trimmed
}

export function getCloudinaryConfig() {
  return {
    cloudName: requireEnv(process.env.CLOUDINARY_CLOUD_NAME, "CLOUDINARY_CLOUD_NAME"),
    apiKey: requireEnv(process.env.CLOUDINARY_API_KEY, "CLOUDINARY_API_KEY"),
    apiSecret: requireEnv(process.env.CLOUDINARY_API_SECRET, "CLOUDINARY_API_SECRET"),
  }
}

// ---------------------------------------------------------------------------
// Validation (mirrors client-side validation in lib/cloudinary.ts)
// ---------------------------------------------------------------------------

export const ACCEPTED_MIME_TYPES: ReadonlySet<string> = new Set([
  "image/jpeg",
  "image/png",
  "image/webp",
])

/** 5 MB. */
export const MAX_FILE_SIZE_BYTES = 5 * 1024 * 1024

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

// ---------------------------------------------------------------------------
// Signed upload
// ---------------------------------------------------------------------------

export const CLOUDINARY_MARKET_COVERS_FOLDER = "signalarc/market-covers"

/**
 * Generates a Cloudinary upload signature.
 *
 * SHA-1 hex digest of `timestamp + <param_key>=<param_value>... + api_secret`.
 *
 * See: https://cloudinary.com/documentation/upload_images#generating_authentication_signatures
 */
function generateSignature(
  paramsToSign: Record<string, string>,
  apiSecret: string,
): string {
  const sortedEntries = Object.entries(paramsToSign).sort(([a], [b]) =>
    a.localeCompare(b),
  )

  const signString =
    sortedEntries
      .map(([key, value]) => `${key}=${value}`)
      .join("&") + apiSecret

  return crypto.createHash("sha1").update(signString, "utf8").digest("hex")
}

type CloudinaryUploadResponse = {
  secure_url?: string
  error?: { message?: string }
}

/**
 * Uploads a cover image to Cloudinary using signed authentication.
 *
 * The caller provides a validated `File`.  This function builds the signed
 * upload payload, POSTs it to Cloudinary, and returns the `secure_url`.
 */
export async function uploadMarketCoverImage(file: File): Promise<string> {
  const { cloudName, apiKey, apiSecret } = getCloudinaryConfig()

  const timestamp = Math.round(Date.now() / 1000).toString()

  const signatureParams: Record<string, string> = {
    folder: CLOUDINARY_MARKET_COVERS_FOLDER,
    timestamp,
  }

  const signature = generateSignature(signatureParams, apiSecret)

  const formData = new FormData()
  formData.append("file", file)
  formData.append("folder", CLOUDINARY_MARKET_COVERS_FOLDER)
  formData.append("timestamp", timestamp)
  formData.append("api_key", apiKey)
  formData.append("signature", signature)

  const uploadUrl = `https://api.cloudinary.com/v1_1/${cloudName}/image/upload`

  let response: Response
  try {
    response = await fetch(uploadUrl, { method: "POST", body: formData })
  } catch {
    throw new Error("Unable to reach Cloudinary. Check server connectivity.")
  }

  let payload: CloudinaryUploadResponse | null = null
  try {
    payload = (await response.json()) as CloudinaryUploadResponse
  } catch {
    payload = null
  }

  if (!response.ok) {
    const message = payload?.error?.message
    throw new Error(
      message && message.trim() !== ""
        ? message
        : "Cloudinary upload failed.",
    )
  }

  const secureUrl = payload?.secure_url
  if (!secureUrl || !secureUrl.startsWith("https://")) {
    throw new Error("Cloudinary upload completed but no secure_url was returned.")
  }

  return secureUrl
}
