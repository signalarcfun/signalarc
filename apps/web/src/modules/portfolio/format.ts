export function formatPortfolioDate(value: string | null) {
  if (!value) {
    return "-"
  }

  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return value
  }

  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date)
}

export function truncatePortfolioId(id: string) {
  if (id.length <= 12) return id
  return `${id.slice(0, 6)}...${id.slice(-4)}`
}

export function formatWalletAddress(address: string) {
  return truncatePortfolioId(address)
}

export function formatPortfolioAmount(value: number) {
  if (!Number.isFinite(value)) return "-"

  return new Intl.NumberFormat("en", {
    maximumFractionDigits: 6,
  }).format(value)
}

export function formatMarketStatus(status: string | null | undefined) {
  if (!status) return "-"
  return status.charAt(0).toUpperCase() + status.slice(1).toLowerCase()
}
