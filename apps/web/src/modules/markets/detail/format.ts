import { getArcscanAddressUrl, getArcscanTxUrl } from "@/lib/contracts"

export function formatMarketDate(value: string) {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return value
  }

  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date)
}

export function marketStatusBadgeClass(status: string) {
  switch (status.toLowerCase()) {
    case "open":
      return "border-green-500/30 bg-green-500/10 text-green-300"
    case "closed":
      return "border-yellow-500/30 bg-yellow-500/10 text-yellow-300"
    case "resolved":
      return "border-indigo-500/30 bg-indigo-500/10 text-indigo-300"
    case "cancelled":
      return "border-red-500/30 bg-red-500/10 text-red-300"
    default:
      return ""
  }
}

export function onchainDeploymentBadgeClass(status: string) {
  switch (status) {
    case "DEPLOYED":
      return "border-green-500/30 bg-green-500/10 text-green-300"
    case "FAILED":
      return "border-red-500/30 bg-red-500/10 text-red-300"
    case "NOT_DEPLOYED":
      return "border-yellow-500/30 bg-yellow-500/10 text-yellow-300"
    default:
      return ""
  }
}

export function formatDeploymentStatus(status: string) {
  return status
    .toLowerCase()
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ")
}

export function marketStatusContext(
  status: string,
  winningOutcome: string | null,
) {
  switch (status.toLowerCase()) {
    case "open":
      return "This market is currently accepting positions."
    case "closed":
      return "This market has closed and is pending resolution."
    case "resolved":
      return winningOutcome
        ? `This market has been resolved. Winning outcome: ${winningOutcome}.`
        : "This market has been resolved."
    case "cancelled":
      return "This market has been cancelled. Participants may be eligible for refunds."
    default:
      return null
  }
}

export function arcscanContractUrl(address: string) {
  return getArcscanAddressUrl(address)
}

export function arcscanTransactionUrl(hash: string) {
  return getArcscanTxUrl(hash as `0x${string}`)
}
