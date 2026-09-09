import type { BudgetMonthCategory } from '../../lib/api/budget'
import { availableTone, spentTone } from '../../lib/budgetView'
import { formatMoney } from '../../lib/money'
import AllocatedInput from './AllocatedInput'
import GoalCell from './GoalCell'
import { CATEGORY_GRID } from './layout'

type CategoryRowProps = {
  budgetId: number
  month: string
  category: BudgetMonthCategory
  groupId: number | null
  onEditCategory: (category: BudgetMonthCategory, groupId: number | null) => void
  onOpenGoal: (category: BudgetMonthCategory) => void
}

export default function CategoryRow({
  budgetId,
  month,
  category,
  groupId,
  onEditCategory,
  onOpenGoal,
}: CategoryRowProps) {
  return (
    <div className={`${CATEGORY_GRID} items-center border-t border-slate-800/60 px-4 py-1.5 hover:bg-slate-800/30`}>
      <button
        type="button"
        onClick={() => onEditCategory(category, groupId)}
        className="cursor-pointer truncate pl-5 text-left text-sm text-slate-300 transition hover:text-slate-100"
      >
        {category.categoryName}
      </button>
      <GoalCell goal={category.goal} onOpen={() => onOpenGoal(category)} />
      <AllocatedInput
        budgetId={budgetId}
        categoryId={category.categoryId}
        month={month}
        allocated={category.allocated}
      />
      <span className={`text-right text-sm tabular-nums ${spentTone(category.spent)}`}>
        {formatMoney(-category.spent)}
      </span>
      <span className={`text-right text-sm tabular-nums ${availableTone(category.available)}`}>
        {formatMoney(category.available)}
      </span>
    </div>
  )
}
