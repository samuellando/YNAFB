import type { ExpenseShareTransaction } from '../../lib/api/expenseShare'
import { formatDate } from '../../lib/date'
import { formatMoney } from '../../lib/money'
import { CheckIcon, WarningIcon } from '../icons'
import { TRANSACTION_GRID } from './layout'
import ExpenseShareTransactionSplitLine from './ShareTransactionSplitLine'

type TransactionRowProps = {
  budgetId: number 
  transaction: ExpenseShareTransaction
  onAdd: (trxId: number) => void
  onEdit: (trxId: number, catId: number) => void
}

function OutflowCell({ value }: { value: number }) {
  if (value === 0) {
    return <span />
  }
  return (
    <span className="text-right text-sm tabular-nums text-red-400">
      {formatMoney(-value)}
    </span>
  )
}

function InflowCell({ value }: { value: number }) {
  if (value === 0) {
    return <span />
  }
  return (
    <span className="text-right text-sm font-medium tabular-nums text-emerald-400">
      {formatMoney(value)}
    </span>
  )
}

export default function ShareTransactionRow({ budgetId, transaction, onAdd, onEdit }: TransactionRowProps) {
  var payee = transaction.payeeName
  if (transaction.settlementLine?.destBudgetDisplayName) {
      payee = "Settlement with "+transaction.settlementLine.destBudgetDisplayName
  }
  const cells = (
    <>
      <span className="text-sm whitespace-nowrap text-slate-400">
        {formatDate(transaction.date)}
      </span>
      <span className="flex min-h-9 min-w-0 flex-col justify-center">
      <span className="truncate text-sm text-slate-100">{transaction.sourceBudgetDisplayName}</span>
        {transaction.note && (
          <span className="truncate text-xs text-slate-500">{transaction.note}</span>
        )}
      </span>
      <span className={`flex min-w-0 items-center gap-1.5 text-sm text-slate-300`} >
      {payee}
      </span>
      <OutflowCell value={transaction.totalOutflow} />
      <InflowCell value={transaction.totalInflow} />
      <span className="flex justify-center">
        {transaction.reconciled && <CheckIcon className="h-4 w-4 text-emerald-400" />}
      </span>
    </>
  )

  return (
    <div>
      <div
        className={`${TRANSACTION_GRID} relative items-center border-t border-slate-800/60 px-4 py-1.5 cursor-pointer hover:bg-slate-800/301`}
      >
        {cells}
      </div>
        {transaction.splitLines && transaction.splitLines.map((line) => 
            <ExpenseShareTransactionSplitLine 
                budgetId={budgetId}
                line={line} 
                transaction={transaction}
                onAdd={() => {onAdd(transaction.trxId)}}
                onEdit={(catId) => {onEdit(transaction.trxId, catId)}}
                categorizations={budgetId === line.splitBudgetId ? transaction.myCategorizations : undefined} />
        )}
    </div>
  )
}
