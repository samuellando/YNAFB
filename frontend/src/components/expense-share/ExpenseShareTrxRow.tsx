import { Link } from 'react-router'
import type { ExpenseShareTrx, ExpenseShareTrxSplit } from '../../lib/api/budget'
import { formatDate } from '../../lib/date'
import { formatMoney } from '../../lib/money'
import { WarningIcon } from '../icons'
import { SHARE_TRX_GRID } from './layout'

type ExpenseShareTrxRowProps = {
  transaction: ExpenseShareTrx
  ownBudgetId: number
  onSelect: (transaction: ExpenseShareTrx) => void
  onCategorize: (transaction: ExpenseShareTrx) => void
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

function splitTarget(split: ExpenseShareTrxSplit, isPublisher: boolean): string {
  // The publisher categorizes through their own transaction lines, not split
  // lines, so their split never shows a categorization state.
  if (isPublisher || split.lines === undefined) {
    return ''
  }
  if (split.lines.length === 0) {
    return 'Uncategorized'
  }
  return split.lines
    .map((line) => line.categoryName ?? line.destAccountName ?? 'Uncategorized')
    .join(', ')
}

// Warns when this budget's categorized lines don't add up to its split —
// including no categorization at all. Only the requesting budget's lines are
// known (others' `lines` are undefined), and the publisher categorizes through
// the source transaction, so only own non-publisher splits warn.
function splitWarning(
  split: ExpenseShareTrxSplit,
  isPublisher: boolean,
  outflowMode: boolean,
): string | null {
  if (isPublisher || split.lines === undefined) {
    return null
  }
  const splitAmount = outflowMode ? split.splitOutflow : split.splitInflow
  const categorized = split.lines.reduce(
    (sum, line) => sum + (outflowMode ? line.outflow : line.inflow),
    0,
  )
  if (categorized === splitAmount) {
    return null
  }
  return `Categorized ${formatMoney(categorized)} of ${formatMoney(splitAmount)}`
}

export default function ExpenseShareTrxRow({
  transaction,
  ownBudgetId,
  onSelect,
  onCategorize,
}: ExpenseShareTrxRowProps) {
  const publisherBudgetId = transaction.publisherBudgetId
  const outflowMode = transaction.totalInflow === 0
  const ownPublished = publisherBudgetId !== undefined && publisherBudgetId === ownBudgetId
  const sourceLink =
    ownPublished && transaction.accountId !== undefined
      ? `/budget/${ownBudgetId}/account/${transaction.accountId}?trx=${transaction.trxId}`
      : null
  return (
    <div>
      <div
        className={`${SHARE_TRX_GRID} relative items-center border-t border-slate-800/60 px-4 py-1.5 cursor-pointer hover:bg-slate-800/30`}
      >
        <span className="text-sm whitespace-nowrap text-slate-400">
          {formatDate(transaction.date)}
        </span>
        <span className="flex min-h-9 min-w-0 flex-col justify-center">
          <button
            type="button"
            onClick={() => onSelect(transaction)}
            className="cursor-pointer truncate text-left text-sm text-slate-100 transition after:absolute after:inset-0 after:content-[''] hover:text-emerald-300"
          >
            {transaction.payeeName}
          </button>
          {transaction.note && (
            <span className="truncate text-xs text-slate-500">{transaction.note}</span>
          )}
        </span>
        <span className="truncate text-sm text-slate-300">
          {transaction.publisherDisplayName ?? 'Unknown'}
        </span>
        <OutflowCell value={transaction.totalOutflow} />
        <InflowCell value={transaction.totalInflow} />
      </div>

      {transaction.splits.map((split) => {
        const own = split.budgetId === ownBudgetId
        const isPublisher =
          publisherBudgetId !== undefined && split.budgetId === publisherBudgetId
        const target = splitTarget(split, isPublisher)
        const warning = splitWarning(split, isPublisher, outflowMode)
        const linkable = isPublisher && sourceLink !== null
        // Only non-publishers categorize through split lines; the publisher
        // categorizes through the source transaction instead. A zero split has
        // nothing to categorize, so it offers no categorize option either.
        const splitAmount = outflowMode ? split.splitOutflow : split.splitInflow
        const categorizable = own && !isPublisher && splitAmount > 0
        return (
          <div
            key={split.budgetId}
            className={`${SHARE_TRX_GRID} items-center border-t border-slate-800/30 px-4 py-1 text-slate-400 ${
              categorizable ? 'cursor-pointer hover:bg-slate-800/30' : ''
            }`}
            title={categorizable ? 'Categorize my split' : undefined}
          >
            <span />
            <span />
            <span
              className={`truncate pl-5 text-sm ${split.isDefault ? 'italic' : ''} ${own ? 'text-slate-200' : ''}`}
            >
              {warning !== null && (
                <span title={warning} className="mr-1 inline-flex align-[-2px]">
                  <WarningIcon
                    className="h-3.5 w-3.5 text-yellow-300"
                    role="img"
                    aria-label={warning}
                  />
                </span>
              )}
              {linkable ? (
                <Link
                  to={sourceLink}
                  className="cursor-pointer transition hover:text-emerald-300"
                >
                  {split.displayName}
                </Link>
              ) : categorizable ? (
                <button
                  type="button"
                  onClick={() => onCategorize(transaction)}
                  className="cursor-pointer transition hover:text-emerald-300"
                >
                  {split.displayName}
                </button>
              ) : (
                split.displayName
              )}
              {own ? ' · you' : ''}
              {target !== '' && <span className="text-slate-500"> · {target}</span>}
            </span>
            <OutflowCell value={split.splitOutflow} />
            <InflowCell value={split.splitInflow} />
          </div>
        )
      })}
    </div>
  )
}
