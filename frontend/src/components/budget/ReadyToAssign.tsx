import type { BudgetMonthSummary } from '../../lib/api/budget'
import { formatMoney } from '../../lib/money'
import { WarningIcon } from '../icons'

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
        {summary.uncategorized > 0 && (
          <span className="flex items-center gap-1.5 rounded-full border border-yellow-500/40 bg-yellow-500/10 px-3 py-1 text-xs font-semibold whitespace-nowrap text-yellow-300">
            <WarningIcon className="h-3.5 w-3.5" />
            {formatMoney(summary.uncategorized)} uncategorized
          </span>
        )}
      </div>
    </div>
  )
}
