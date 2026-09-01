import Link from "next/link"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

export function MarketDetailEmptyState() {
  return (
    <div className="mx-auto max-w-lg px-4 py-10 sm:py-12">
      <Card>
        <CardHeader className="space-y-2 text-center">
          <CardTitle className="text-lg">Market not found</CardTitle>
          <CardDescription className="text-sm text-muted-foreground">
            This market is unavailable or no longer exists.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex justify-center">
          <Button asChild size="sm" variant="outline" className="w-full sm:w-auto">
            <Link href="/markets">Back to markets</Link>
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
