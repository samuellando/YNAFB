import { useEffect, useRef, useState } from 'react'
import type { AccountTransaction } from '../../lib/api/budget'
import { TRANSACTION_GRID } from './layout'
import TransactionRow from './TransactionRow'

type TransactionTableProps = {
  transactions: AccountTransaction[]
  onSelectTransaction: (transaction: AccountTransaction) => void
}

const PAGE_SIZE = 200

export default function TransactionTable({ transactions, onSelectTransaction }: TransactionTableProps) {
  const [visibleCount, setVisibleCount] = useState(PAGE_SIZE)
  const sentinelRef = useRef<HTMLDivElement | null>(null)

  // slice() clamps naturally when the list shrinks, so no reset effect here:
  // edits/refetches keep scroll position, remount via key on account switch
  // restores the initial 200.
  const visible = transactions.slice(0, visibleCount)
  const hasMore = visibleCount < transactions.length

  useEffect(() => {
    if (!hasMore) return
    const el = sentinelRef.current
    if (!el) return
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) {
          setVisibleCount((v) => Math.min(v + PAGE_SIZE, transactions.length))
        }
      },
      { rootMargin: '600px' },
    )
    observer.observe(el)
    return () => observer.disconnect()
  }, [hasMore, transactions.length])

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

          {visible.map((transaction) => (
            <TransactionRow
              key={transaction.id}
              transaction={transaction}
              onSelect={onSelectTransaction}
            />
          ))}

          {hasMore && (
            <div ref={sentinelRef} aria-hidden="true" className="h-1" />
          )}
          {transactions.length > PAGE_SIZE && (
            <div className="px-4 py-3 text-center text-xs text-slate-500">
              Showing {visible.length} of {transactions.length} transactions
            </div>
          )}
        </div>
      ) : (
        <div className="rounded-xl border border-slate-800 p-6 text-sm text-slate-400">
          No transactions yet.
        </div>
      )}
    </div>
  )
}
