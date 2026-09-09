import type { BudgetMonthCategory } from '../../lib/api/budget'
import type { CategoryGroupView } from '../../lib/budgetView'
import { PlusIcon } from '../icons'
import CategoryGroupSection from './CategoryGroupSection'
import { CATEGORY_GRID } from './layout'

type CategoryTableProps = {
  groups: CategoryGroupView[]
  budgetId: number
  month: string
  groupIdForName: (name: string | null) => number | null
  onAddCategory: () => void
  onEditGroup: (groupId: number, name: string) => void
  onEditCategory: (category: BudgetMonthCategory, groupId: number | null) => void
  onOpenGoal: (category: BudgetMonthCategory) => void
}

export default function CategoryTable({
  groups,
  budgetId,
  month,
  groupIdForName,
  onAddCategory,
  onEditGroup,
  onEditCategory,
  onOpenGoal,
}: CategoryTableProps) {
  return (
    <div className="mt-14">
      <div className="mb-2 flex justify-end">
        <button
          type="button"
          onClick={onAddCategory}
          className="flex items-center gap-1 rounded-lg border border-slate-700 px-3 py-1.5 text-xs font-medium text-emerald-400 transition hover:border-emerald-400"
        >
          <PlusIcon className="h-3.5 w-3.5" />
          Add category
        </button>
      </div>
      {groups.length > 0 ? (
        <div>
          <div
            className={`${CATEGORY_GRID} border-b border-slate-800 px-4 py-2 text-xs font-semibold tracking-widest text-slate-500 uppercase`}
          >
            <span>Category</span>
            <span className="text-right">Goal</span>
            <span className="text-right">Allocated</span>
            <span className="text-right">Activity</span>
            <span className="text-right">Available</span>
          </div>

          {groups.map((group, index) => (
            <CategoryGroupSection
              key={group.name ?? 'ungrouped'}
              group={group}
              groupId={groupIdForName(group.name)}
              isFirst={index === 0}
              budgetId={budgetId}
              month={month}
              onEditGroup={onEditGroup}
              onEditCategory={onEditCategory}
              onOpenGoal={onOpenGoal}
            />
          ))}
        </div>
      ) : (
        <div className="rounded-xl border border-slate-800 p-6 text-sm text-slate-400">
          No categories yet.
        </div>
      )}
    </div>
  )
}
