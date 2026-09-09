import type { BudgetMonthGoal } from '../../lib/api/budget'
import { formatMoney } from '../../lib/money'
import { WarningIcon } from '../icons'

type GoalCellProps = {
  goal?: BudgetMonthGoal
  onOpen: () => void
}

export default function GoalCell({ goal, onOpen }: GoalCellProps) {
  if (!goal) {
    return (
      <button
        type="button"
        onClick={onOpen}
        title="Set goal"
        aria-label="Set goal"
        className="ml-auto w-28 cursor-pointer rounded px-1.5 py-0.5 text-right text-sm tabular-nums text-slate-600 transition hover:bg-slate-800/60 hover:text-slate-300"
      >
        Set goal
      </button>
    )
  }

  const behind = goal.gap < 0
  const over = goal.gap > 0
  const offTrack = behind || over
  let message = 'On track for this month'
  if (behind) message = `Allocate ${formatMoney(-goal.gap)} more this month`
  else if (over) message = `${formatMoney(goal.gap)} over this month's target`

  return (
    <button
      type="button"
      onClick={onOpen}
      title={message}
      aria-label={`Goal ${formatMoney(goal.amountForMonth)}. ${message}`}
      className={`ml-auto flex w-28 cursor-pointer items-center justify-end gap-1 rounded px-1.5 py-0.5 text-sm tabular-nums transition ${
        offTrack
          ? 'text-yellow-300 hover:bg-yellow-500/10'
          : 'text-slate-100 hover:bg-slate-800/60 hover:text-emerald-300'
      }`}
    >
      {offTrack && <WarningIcon className="h-3.5 w-3.5 shrink-0" />}
      <span className="text-right">{formatMoney(goal.amountForMonth)}</span>
    </button>
  )
}
