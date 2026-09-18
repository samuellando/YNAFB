import type { AccountTransaction } from '../../lib/api/budget'
import {
  isSplitTransaction,
  isUnbalanced,
  mainTarget,
  transactionTarget,
} from '../../lib/accountView'
import { formatDate } from '../../lib/date'
import { formatMoney } from '../../lib/money'
import { CheckIcon, WarningIcon } from '../icons'
import { TRANSACTION_GRID } from './layout'
import PublishShareButton, { type PublishShareTarget } from './PublishShareButton'

type TransactionRowProps = {
  budgetId: number
  transaction: AccountTransaction
  published: Map<number, Set<number>>
  onSelect: (transaction: AccountTransaction) => void
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

export default function TransactionRow({ budgetId, transaction, published, onSelect }: TransactionRowProps) {
  const split = isSplitTransaction(transaction)
  const unbalanced = isUnbalanced(transaction)
  const mirror = transaction.sourceAccountId !== undefined

  const shareTargets = (() => {
    const seen = new Map<number, PublishShareTarget>()
    for (const line of transaction.transactionLines) {
      if (line.expenseShareId === undefined || seen.has(line.expenseShareId)) continue
      seen.set(line.expenseShareId, {
        id: line.expenseShareId,
        name: line.expenseShareName ?? `Share ${line.expenseShareId}`,
        published: published.get(line.expenseShareId)?.has(transaction.id) ?? false,
      })
    }
    return [...seen.values()]
  })()

  const cells = (
    <>
      <span className="text-sm whitespace-nowrap text-slate-400">
        {formatDate(transaction.date)}
      </span>
      <span className="flex min-h-9 min-w-0 flex-col justify-center">
        {mirror ? (
          <span className="truncate text-sm text-slate-100">{transaction.payeeName}</span>
        ) : (
          <button
            type="button"
            onClick={() => onSelect(transaction)}
            className="cursor-pointer truncate text-left text-sm text-slate-100 transition after:absolute after:inset-0 after:content-[''] hover:text-emerald-300"
          >
            {transaction.payeeName}
          </button>
        )}
        {transaction.note && (
          <span className="truncate text-xs text-slate-500">{transaction.note}</span>
        )}
        {shareTargets.map((target) => (
          <PublishShareButton
            key={target.id}
            budgetId={budgetId}
            trxId={transaction.id}
            target={target}
          />
        ))}
      </span>
      <span
        className={`flex min-w-0 items-center gap-1.5 text-sm ${split ? 'text-slate-500 italic' : 'text-slate-300'}`}
      >
        {unbalanced && (
          <WarningIcon
            className="h-3.5 w-3.5 shrink-0 text-yellow-300"
            role="img"
            aria-label="Lines don't add up to the transaction total"
          />
        )}
        <span className="truncate">{split ? 'Split' : mainTarget(transaction)}</span>
      </span>
      <OutflowCell value={transaction.outflow} />
      <InflowCell value={transaction.inflow} />
      <span className="flex justify-center">
        {transaction.reconciled && <CheckIcon className="h-4 w-4 text-emerald-400" />}
      </span>
    </>
  )

  return (
    <div>
      <div
        className={`${TRANSACTION_GRID} relative items-center border-t border-slate-800/60 px-4 py-1.5 ${
          mirror ? '' : 'cursor-pointer hover:bg-slate-800/30'
        }`}
        title={mirror ? "Can't edit transfers from other accounts" : undefined}
      >
        {cells}
      </div>

      {split &&
        transaction.transactionLines.map((line) => (
          <div
            key={line.lineId}
            className={`${TRANSACTION_GRID} items-center border-t border-slate-800/30 px-4 py-1 text-slate-400 hover:bg-slate-800/30`}
          >
            <span />
            <span />
            <span className="truncate pl-5 text-sm text-slate-400">
              {transactionTarget(line)}
            </span>
            <OutflowCell value={line.outflow} />
            <InflowCell value={line.inflow} />
            <span />
          </div>
        ))}
    </div>
  )
}
