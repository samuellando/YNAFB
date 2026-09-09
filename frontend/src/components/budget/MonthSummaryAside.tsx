import type { BudgetMonthSummary } from '../../lib/api/budget'
import { availableTone, spentTone } from '../../lib/budgetView'
import { formatMoney } from '../../lib/money'

type MonthSummaryAsideProps = {
  summary: BudgetMonthSummary
}

export default function MonthSummaryAside({ summary }: MonthSummaryAsideProps) {
  return (
    <aside className="hidden w-60 shrink-0 border-l border-slate-800 bg-slate-900 lg:block">
      <h2 className="border-b border-slate-800 px-5 py-3 text-xs font-semibold tracking-widest text-slate-500 uppercase">
        Monthly summary
      </h2>
      <div className="flex flex-col gap-4 p-5">
        <div className="flex items-center justify-between">
          <span className="text-xs font-semibold tracking-widest text-slate-500 uppercase">
            Income
          </span>
          <span className="text-sm font-semibold tabular-nums text-emerald-400">
            {formatMoney(summary.income)}
          </span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-xs font-semibold tracking-widest text-slate-500 uppercase">
            Allocated
          </span>
          <span className="text-sm font-semibold tabular-nums text-slate-100">
            {formatMoney(summary.allocated)}
          </span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-xs font-semibold tracking-widest text-slate-500 uppercase">
            Spent
          </span>
          <span className={`text-sm font-semibold tabular-nums ${spentTone(summary.spent)}`}>
            {formatMoney(-summary.spent)}
          </span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-xs font-semibold tracking-widest text-slate-500 uppercase">
            Available
          </span>
          <span className={`text-sm font-semibold tabular-nums ${availableTone(summary.available)}`}>
            {formatMoney(summary.available)}
          </span>
        </div>
        {summary.uncategorized > 0 && (
          <div className="flex items-center justify-between border-t border-slate-800 pt-3">
            <span className="text-xs font-semibold tracking-widest text-slate-500 uppercase">
              Uncategorized
            </span>
            <span className="text-sm font-semibold tabular-nums text-red-400">
              {formatMoney(-summary.uncategorized)}
            </span>
          </div>
        )}
      </div>
    </aside>
  )
}
