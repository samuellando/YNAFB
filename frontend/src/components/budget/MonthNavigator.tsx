import { formatMonthLabel, shiftMonth } from '../../lib/month'
import { ChevronLeftIcon, ChevronRightIcon } from '../icons'
import { OUTLINE_BUTTON } from '../ui/buttons'

type MonthNavigatorProps = {
  selected: string
  current: string
  onSelect: (month: string) => void
}

export default function MonthNavigator({ selected, current, onSelect }: MonthNavigatorProps) {
  return (
    <header className="flex items-center">
      <div className="flex items-center gap-0">
        <button
          type="button"
          onClick={() => onSelect(shiftMonth(selected, -1))}
          aria-label="Previous month"
          className="rounded-lg p-2 text-slate-400 transition hover:bg-slate-800 hover:text-slate-100"
        >
          <ChevronLeftIcon className="h-5 w-5" />
        </button>
        <h1 className="min-w-40 text-center text-lg font-bold tracking-tight">
          {formatMonthLabel(selected)}
        </h1>
        <button
          type="button"
          onClick={() => onSelect(shiftMonth(selected, 1))}
          aria-label="Next month"
          className="rounded-lg p-2 text-slate-400 transition hover:bg-slate-800 hover:text-slate-100"
        >
          <ChevronRightIcon className="h-5 w-5" />
        </button>
      </div>
      <button
        type="button"
        onClick={() => onSelect(current)}
        disabled={selected === current}
        className={`ml-3 ${OUTLINE_BUTTON}`}
      >
        This month
      </button>
    </header>
  )
}
