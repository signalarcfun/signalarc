import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import type { Resolution, Settlement } from "@/lib/api"

import { formatResolutionDate, truncateResolutionId } from "../format"
import type { ResolutionState } from "../types"
import { ResolutionActionStatus } from "./resolution-action-status"
import { ResolutionOutcomeSelector } from "./resolution-outcome-selector"

function ResolutionDetails({ resolution }: { resolution: Resolution }) {
  return (
    <dl className="grid gap-3 sm:grid-cols-2 lg:grid-cols-1 xl:grid-cols-2">
      <div className="min-w-0 rounded-lg border border-border bg-muted/20 p-3">
        <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
          Status
        </dt>
        <dd className="mt-1 text-sm font-medium text-foreground">
          {resolution.status}
        </dd>
      </div>
      <div className="min-w-0 rounded-lg border border-border bg-muted/20 p-3">
        <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
          Winning Outcome
        </dt>
        <dd className="mt-1 text-sm font-medium text-foreground">
          <ResolutionOutcomeSelector outcome={resolution.winning_outcome} />
        </dd>
      </div>
      <div className="min-w-0 rounded-lg border border-border bg-muted/20 p-3">
        <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
          Resolver Type
        </dt>
        <dd className="mt-1 text-sm font-medium text-foreground">
          {resolution.resolver_type ?? "-"}
        </dd>
      </div>
      <div className="min-w-0 rounded-lg border border-border bg-muted/20 p-3">
        <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
          Evidence
        </dt>
        <dd className="mt-1 break-words text-sm font-medium text-foreground">
          {resolution.evidence_reference ?? "-"}
        </dd>
      </div>
      <div className="min-w-0 rounded-lg border border-border bg-muted/20 p-3">
        <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground/70">
          Resolved
        </dt>
        <dd className="mt-1 text-sm font-medium text-foreground">
          {formatResolutionDate(resolution.resolved_at)}
        </dd>
      </div>
    </dl>
  )
}

function SettlementsTable({ settlements }: { settlements: Settlement[] }) {
  if (settlements.length === 0) {
    return <p className="text-sm text-muted-foreground">No settlements yet.</p>
  }

  return (
    <div className="overflow-x-auto rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>User</TableHead>
            <TableHead>Outcome</TableHead>
            <TableHead>Amount</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Settled</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {settlements.map((settlement) => (
            <TableRow key={settlement.id}>
              <TableCell
                className="font-mono text-xs"
                title={settlement.user_id ?? undefined}
              >
                {settlement.user_id
                  ? truncateResolutionId(settlement.user_id)
                  : "-"}
              </TableCell>
              <TableCell>{settlement.outcome ?? "-"}</TableCell>
              <TableCell>{settlement.amount}</TableCell>
              <TableCell>{settlement.status}</TableCell>
              <TableCell className="text-muted-foreground">
                {formatResolutionDate(settlement.settled_at)}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

export function ResolutionPanel({ state }: { state: ResolutionState }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Resolution & Settlements</CardTitle>
        <CardDescription>
          Resolution outcome and settlement records for this market.
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-6">
        {state.status === "loading" || state.status === "error" ? (
          <ResolutionActionStatus state={state} />
        ) : null}

        {state.status === "empty" ? (
          <div className="grid gap-6">
            <p className="text-sm text-muted-foreground">
              This market has not been resolved yet.
            </p>
            <SettlementsTable settlements={state.settlements} />
          </div>
        ) : null}

        {state.status === "loaded" ? (
          <div className="grid gap-6">
            <ResolutionDetails resolution={state.resolution} />
            <SettlementsTable settlements={state.settlements} />
          </div>
        ) : null}
      </CardContent>
    </Card>
  )
}
