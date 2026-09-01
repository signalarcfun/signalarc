import { NextResponse } from "next/server"

import {
  validateMarketCoverFile,
  uploadMarketCoverImage,
} from "@/lib/cloudinary-upload"

export async function POST(request: Request) {
  const formData = await request.formData()
  const file = formData.get("file")

  if (!(file instanceof File)) {
    return NextResponse.json(
      { error: "No file provided." },
      { status: 400 },
    )
  }

  const validationError = validateMarketCoverFile(file)
  if (validationError) {
    return NextResponse.json(
      { error: validationError },
      { status: 422 },
    )
  }

  try {
    const secureUrl = await uploadMarketCoverImage(file)
    return NextResponse.json({ secure_url: secureUrl })
  } catch (err) {
    const message =
      err instanceof Error ? err.message : "Upload failed."

    console.error("[market-cover-upload]", message)

    // Cloudinary config errors (missing env vars) → 500.
    if (message.includes("Missing required server env")) {
      return NextResponse.json({ error: message }, { status: 500 })
    }

    return NextResponse.json({ error: message }, { status: 502 })
  }
}
