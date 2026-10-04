import { useEffect, useRef, useState } from 'react'
import type { ExpenseShareTransaction } from '../../lib/api/expenseShare'
import { TRANSACTION_GRID } from './layout'
import ShareTransactionRow from './ShareTransactionRow'

type ShareTransactionTableProps = {
  budgetId: number 
  transactions: ExpenseShareTransaction[]
  onAdd: (trxId: number) => void
  onEdit: (trxId: number, catId: number) => void
}

const PAGE_SIZE = 200

export default function ShareDetailTable({budgetId, transactions, onAdd, onEdit}: ShareDetailTableProps) {
    return <div className="mt-5">
      {transactions.length > 0 ? (
        <div>
          <div
            className={`${TRANSACTION_GRID} border-b border-slate-800 px-4 py-2 text-xs font-semibold tracking-widest text-slate-500 uppercase`}
          >
            <span>Date</span>
            <span>Member</span>
            <span>Payee</span>
            <span className="text-right">Outflow</span>
            <span className="text-right">Inflow</span>
            <span className="text-center">Reconciled</span>
          </div>
          {transactions.map((transaction) => (
            <ShareTransactionRow
              transaction={transaction}
              budgetId={budgetId}
              onAdd={onAdd}
              onEdit={onEdit}
            />
          ))}
        </div>
      ) : (
        <div className="rounded-xl border border-slate-800 p-6 text-sm text-slate-400">
          No transactions yet.
        </div>
      )}
    </div>
}
