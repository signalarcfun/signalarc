export function ResolutionOutcomeSelector({
  outcome,
}: {
  outcome: string | null
}) {
  return <>{outcome ?? "-"}</>
}
