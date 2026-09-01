import type { Metadata } from "next"

import { AnalyticsShell } from "@/modules/analytics"

export const metadata: Metadata = {
  title: "SignalArc Analytics \u2014 Arc Testnet Proof-of-Activity",
  description:
    "Public Arc Testnet proof-of-activity for SignalArc's verified factory, created YES/NO markets, testnet USDC collateral activity, lifecycle events, and agent execution readiness.",
}

export default function AnalyticsPage() {
  return <AnalyticsShell />
}
