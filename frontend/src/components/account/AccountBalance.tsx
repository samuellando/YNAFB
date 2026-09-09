import { availableTone } from '../../lib/budgetView'
import { formatMoney } from '../../lib/money'
import { WarningIcon } from '../icons'

type AccountBalanceProps = {
  balance: number
  reconciledBalance: number
  uncategorized: number
}

export default function AccountBalance({
  balance,
  reconciledBalance,
  uncategorized,
}: AccountBalanceProps) {
  return (
    <div className="mt-14 flex flex-col items-center gap-1">
      <p className="text-xs font-semibold tracking-widest text-slate-500 uppercase">
        Balance
      </p>
      <p className={`text-4xl font-bold tabular-nums ${availableTone(balance)}`}>
        {formatMoney(balance)}
      </p>
      <p className="mt-4 text-sm">
        <span className="text-xs font-semibold tracking-widest text-slate-500 uppercase">
          Reconciled{' '}
        </span>
        <span className="font-semibold tabular-nums text-slate-100">
          {formatMoney(reconciledBalance)}
        </span>
      </p>
      {uncategorized > 0 && (
        <span className="mt-3 flex items-center gap-1.5 rounded-full border border-yellow-500/40 bg-yellow-500/10 px-3 py-1 text-xs font-semibold whitespace-nowrap text-yellow-300">
          <WarningIcon className="h-3.5 w-3.5" />
          {formatMoney(uncategorized)} uncategorized
        </span>
      )}
    </div>
  )
}
