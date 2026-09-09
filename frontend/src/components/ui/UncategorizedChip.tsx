import { formatMoney } from '../../lib/money'
import { WarningIcon } from '../icons'

type UncategorizedChipProps = {
  amount: number
}

export default function UncategorizedChip({ amount }: UncategorizedChipProps) {
  return (
    <span className="flex items-center gap-1.5 rounded-full border border-yellow-500/40 bg-yellow-500/10 px-3 py-1 text-xs font-semibold whitespace-nowrap text-yellow-300">
      <WarningIcon className="h-3.5 w-3.5" />
      {formatMoney(amount)} uncategorized
    </span>
  )
}
