import Link from "next/link"

import { Button } from "@/components/ui/button"
import { CreateMarketForm } from "@/features/markets/create-market-form"

export default function NewMarketPage() {
  return (
    <div className="px-4 py-8 sm:px-6 lg:px-8">
      <div className="mx-auto flex w-full max-w-3xl flex-col gap-6">
        <div>
          <Button asChild size="sm" variant="ghost">
            <Link href="/markets">&larr; Back to markets</Link>
          </Button>
        </div>
        <CreateMarketForm />
      </div>
    </div>
  )
}
