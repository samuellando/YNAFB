import type { ExpenseShareTrx } from '../../lib/api/budget'
import ExpenseShareTrxRow from './ExpenseShareTrxRow'
import { SHARE_TRX_GRID } from './layout'

type ExpenseShareTrxTableProps = {
  transactions: ExpenseShareTrx[]
  ownBudgetId: number
  onSelect: (transaction: ExpenseShareTrx) => void
  onCategorize: (transaction: ExpenseShareTrx) => void
}

export default function ExpenseShareTrxTable({
  transactions,
  ownBudgetId,
  onSelect,
  onCategorize,
}: ExpenseShareTrxTableProps) {
  return (
    <div className="mt-14">
      {transactions.length > 0 ? (
        <div>
          <div
            className={`${SHARE_TRX_GRID} border-b border-slate-800 px-4 py-2 text-xs font-semibold tracking-widest text-slate-500 uppercase`}
          >
            <span>Date</span>
            <span>Payee</span>
            <span>Published by</span>
            <span className="text-right">Outflow</span>
            <span className="text-right">Inflow</span>
          </div>

          {transactions.map((transaction) => (
            <ExpenseShareTrxRow
              key={transaction.id}
              transaction={transaction}
              ownBudgetId={ownBudgetId}
              onSelect={onSelect}
              onCategorize={onCategorize}
            />
          ))}
        </div>
      ) : (
        <div className="rounded-xl border border-slate-800 p-6 text-sm text-slate-400">
          No transactions yet.
        </div>
      )}
    </div>
  )
}
