import type { ReactNode } from "react"

import { Badge } from "@/components/ui/badge"

export function IntelligenceShell({ children }: { children: ReactNode }) {
  return (
    <div className="px-4 py-6 sm:px-6 lg:px-8">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-6">
        <header className="space-y-1">
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight sm:text-2xl">Market Intelligence</h1>
            <Badge variant="outline" className="border-indigo-500/20 bg-indigo-500/5 text-indigo-300 text-[10px]">
              Agent API
            </Badge>
          </div>
          <p className="text-xs text-muted-foreground">
            Structured probability signals and market data for agents, developers, and institutional workflows.
          </p>
        </header>
        {children}
      </div>
    </div>
  )
}
