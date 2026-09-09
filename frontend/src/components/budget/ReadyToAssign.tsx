import type { BudgetMonthSummary } from '../../lib/api/budget'
import { formatMoney } from '../../lib/money'
import UncategorizedChip from '../ui/UncategorizedChip'

type ReadyToAssignProps = {
  summary: BudgetMonthSummary
}

export default function ReadyToAssign({ summary }: ReadyToAssignProps) {
  return (
    <div className="mt-14 flex items-center">
      <div className="flex-1" />
      <div className="flex flex-col items-center gap-1">
        <p className="text-xs font-semibold tracking-widest text-slate-500 uppercase">
          Ready to assign
        </p>
        <p
          className={`text-4xl font-bold tabular-nums ${
            summary.readyToAssign < 0 ? 'text-red-400' : 'text-emerald-400'
          }`}
        >
          {formatMoney(summary.readyToAssign)}
        </p>
      </div>
      <div className="flex flex-1 justify-start pl-10">
        {summary.uncategorized > 0 && <UncategorizedChip amount={summary.uncategorized} />}
      </div>
    </div>
  )
}
