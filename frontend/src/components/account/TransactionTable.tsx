import { useMemo } from 'react'
import { useQueries } from '@tanstack/react-query'
import {
  getExpenseShareDetail,
  type AccountTransaction,
} from '../../lib/api/budget'
import { TRANSACTION_GRID } from './layout'
import TransactionRow from './TransactionRow'

type TransactionTableProps = {
  budgetId: number
  transactions: AccountTransaction[]
  onSelectTransaction: (transaction: AccountTransaction) => void
}

export default function TransactionTable({
  budgetId,
  transactions,
  onSelectTransaction,
}: TransactionTableProps) {
  // Share ids referenced by any line, to resolve published state. Shares the
  // ['expense-share', budget, id] cache with the expense share page.
  const shareIds = useMemo(() => {
    const ids = new Set<number>()
    for (const transaction of transactions) {
      for (const line of transaction.transactionLines) {
        if (line.expenseShareId !== undefined) ids.add(line.expenseShareId)
      }
    }
    return [...ids]
  }, [transactions])

  const details = useQueries({
    queries: shareIds.map((shareId) => ({
      queryKey: ['expense-share', budgetId, shareId],
      queryFn: () => getExpenseShareDetail(budgetId, shareId),
    })),
  })

  // Published (share, trx) pairs, derived during render. Shares the
  // ['expense-share', budget, id] cache with the expense share page.
  const published = new Map<number, Set<number>>()
  shareIds.forEach((shareId, index) => {
    const trxIds = new Set<number>()
    for (const trx of details[index]?.data?.transactions ?? []) {
      trxIds.add(trx.trxId)
    }
    published.set(shareId, trxIds)
  })

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
              budgetId={budgetId}
              transaction={transaction}
              published={published}
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
