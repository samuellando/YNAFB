import type { ExpenseShareSummary as Summary } from '../../lib/api/budget'
import { availableTone } from '../../lib/budgetView'
import { formatMoney } from '../../lib/money'

type ExpenseShareSummaryProps = {
  summary: Summary
  ownBudgetId: number
}

export default function ExpenseShareSummary({ summary, ownBudgetId }: ExpenseShareSummaryProps) {
  return (
    <div className="mt-14 flex flex-col items-center gap-1">
      <p className="text-xs font-semibold tracking-widest text-slate-500 uppercase">
        Balances
      </p>
      <ul className="mt-2 w-full max-w-md">
        {summary.balances.map((entry) => (
          <li
            key={entry.budgetId}
            className="flex items-baseline justify-between gap-4 py-1"
          >
            <span className="truncate text-sm text-slate-200">
              {entry.displayName}
              {entry.budgetId === ownBudgetId ? ' · you' : ''}
            </span>
            <span className={`font-semibold tabular-nums ${availableTone(entry.balance)}`}>
              {formatMoney(entry.balance)}
            </span>
          </li>
        ))}
      </ul>
      <p className="mt-1 text-xs text-slate-500">Positive means others owe them.</p>
      <p className="mt-4 text-sm text-slate-500">
        {summary.memberCount} {summary.memberCount === 1 ? 'member' : 'members'} ·{' '}
        {summary.transactionCount}{' '}
        {summary.transactionCount === 1 ? 'transaction' : 'transactions'}
      </p>
    </div>
  )
}
