import { IntelligenceDashboard } from "@/features/intelligence/intelligence-dashboard"
import { IntelligenceShell } from "@/modules/intelligence"

export default function IntelligencePage() {
  return (
    <IntelligenceShell>
      <IntelligenceDashboard />
    </IntelligenceShell>
  )
}
