import type { BudgetMonthCategory } from './api/budget'

export type CategoryGroupView = {
  name: string | null
  categories: BudgetMonthCategory[]
}

export function groupCategories(categories: BudgetMonthCategory[]): CategoryGroupView[] {
  const byName = new Map<string | null, CategoryGroupView>()
  const order: (string | null)[] = []
  for (const category of categories) {
    const key = category.categoryGroupName ?? null
    let group = byName.get(key)
    if (!group) {
      group = { name: key, categories: [] }
      byName.set(key, group)
      order.push(key)
    }
    group.categories.push(category)
  }
  return order.map((key) => byName.get(key)!)
}

export function availableTone(value: number): string {
  if (value < 0) return 'text-red-400'
  if (value === 0) return 'text-slate-500'
  return 'text-emerald-400'
}

export function spentTone(spent: number): string {
  if (spent > 0) return 'text-red-400'
  if (spent < 0) return 'text-emerald-400'
  return 'text-slate-500'
}
