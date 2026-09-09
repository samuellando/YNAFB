import type { BudgetMonthCategory } from '../../lib/api/budget'
import type { CategoryGroupView } from '../../lib/budgetView'
import CategoryRow from './CategoryRow'

type CategoryGroupSectionProps = {
  group: CategoryGroupView
  groupId: number | null
  isFirst: boolean
  budgetId: number
  month: string
  onEditGroup: (groupId: number, name: string) => void
  onEditCategory: (category: BudgetMonthCategory, groupId: number | null) => void
  onOpenGoal: (category: BudgetMonthCategory) => void
}

export default function CategoryGroupSection({
  group,
  groupId,
  isFirst,
  budgetId,
  month,
  onEditGroup,
  onEditCategory,
  onOpenGoal,
}: CategoryGroupSectionProps) {
  const editable = groupId !== null

  return (
    <section>
      <button
        type="button"
        disabled={!editable}
        onClick={editable ? () => onEditGroup(groupId, group.name ?? 'No group') : undefined}
        className={`block w-full bg-slate-900/80 px-4 py-2 text-left ${isFirst ? '' : 'border-t border-slate-800'} ${
          editable ? 'cursor-pointer transition hover:bg-slate-800/50' : 'cursor-default'
        }`}
      >
        <h2 className="text-sm font-semibold tracking-wide text-slate-100">
          {group.name ?? 'No group'}
        </h2>
      </button>
      {group.categories.map((category) => (
        <CategoryRow
          key={category.categoryId}
          budgetId={budgetId}
          month={month}
          category={category}
          groupId={groupId}
          onEditCategory={onEditCategory}
          onOpenGoal={onOpenGoal}
        />
      ))}
    </section>
  )
}
