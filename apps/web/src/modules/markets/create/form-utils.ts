export function optionalText(formData: FormData, key: string) {
  const value = String(formData.get(key) ?? "").trim()
  return value === "" ? undefined : value
}

export function requiredText(formData: FormData, key: string) {
  return String(formData.get(key) ?? "").trim()
}

export function toRfc3339(value: string) {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    throw new Error("Dates must be valid.")
  }

  return date.toISOString()
}

export function defaultCloseValue() {
  const date = new Date()
  date.setDate(date.getDate() + 7)
  date.setMinutes(date.getMinutes() - date.getTimezoneOffset())

  return date.toISOString().slice(0, 16)
}
