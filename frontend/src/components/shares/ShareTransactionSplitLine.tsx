import type { ExpenseShareSplitLine, SplitCategorization } from '../../lib/api/expenseShare'
import { formatDate } from '../../lib/date'
import { formatMoney } from '../../lib/money'
import { CheckIcon, WarningIcon } from '../icons'
import { TRANSACTION_GRID } from './layout'
import { PlusIcon } from '../../components/icons'
import { ReactNode, MouseEventHandler, HTMLDivElement } from 'react'

type TransactionSplitLineProps = {
  line: ExpenseShareSplitLine
  categorizations: SplitCategorization[]
  onAdd: () => void
  onEdit: (cat: SplitCategorization) => void
}

function OutflowCell({ value }: { value: number }) {
  if (value === 0) {
    return <span />
  }
  return (
    <span className="text-right tabular-nums text-red-400">
      {formatMoney(-value)}
    </span>
  )
}

function InflowCell({ value }: { value: number }) {
  if (value === 0) {
    return <span />
  }
  return (
    <span className="text-right font-medium tabular-nums text-emerald-400">
      {formatMoney(value)}
    </span>
  )
}

export default function ShareTransactionSplitLine({ line, categorizations, onAdd, onEdit }: TransactionSplitLineProps) {
    return <div>
        <Row  outflow={line.outflow} inflow={line.inflow}>
            {line.splitBudgetDisplayName}
        </Row>
        {categorizations?.map((cat) => {
                return <Row className="text-xs" outflow={cat.outflow} inflow={cat.inflow} onClick={() => onEdit(cat)}>
                    <span className="ml-auto">
                    {cat.categoryName}
                    </span>
                </Row>
        }
        )}
        {categorizations && 
            <Row inflow={0} outflow={0} className="text-xs" onClick={onAdd}>
              <div
                className="ml-auto flex items-center gap-1 font-medium text-emerald-400 hover:text-emerald-300"
              >
                  <PlusIcon className="h-4 w-4" />
                  Add category
              </div>
            </Row>
        }
    </div>
}

type RowParams = {
  outflow: number
  inflow: number
  children: ReactNode
  className?: string
  onClick?: ouseEventHandler<HTMLDivElement>
}

function Row({outflow, inflow, children, className, onClick}: RowParams) {
    if (onClick) {
        className += " hover:cursor-pointer"
    }
    return <div onClick={onClick} className={`${TRANSACTION_GRID} text-sm items-center border-t border-slate-800/30 px-4 py-1 text-slate-400 hover:bg-slate-800/30 ${className}`}>
                <span />
                <span />
                    {children}
                <OutflowCell value={outflow} />
                <InflowCell value={inflow} />
                <span />
        </div>
}
