import type { AccountTransaction } from '../../lib/api/budget'
import { TRANSACTION_GRID } from './layout'
import TransactionRow from './TransactionRow'

type TransactionTableProps = {
  transactions: AccountTransaction[]
  onSelectTransaction: (transaction: AccountTransaction) => void
}

export default function TransactionTable({ transactions, onSelectTransaction }: TransactionTableProps) {
  return (
    <div className="mt-14">
      {transactions.length > 0 ? (
        <div>
          <div
            className={`${TRANSACTION_GRID} border-b border-slate-800 px-4 py-2 text-xs font-semibold tracking-widest text-slate-500 uppercase`}
          >
            <span>Date</span>
            <span>Payee</span>
            <span>Category</span>
            <span className="text-right">Outflow</span>
            <span className="text-right">Inflow</span>
            <span className="text-center">Reconciled</span>
          </div>

          {transactions.map((transaction) => (
            <TransactionRow
              key={transaction.id}
              transaction={transaction}
              onSelect={onSelectTransaction}
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
