import { useMemo, useState } from 'react'
import { Navigate, useNavigate, useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import {
  getBudgetMonth,
  listCategoryGroups,
  type BudgetMonthCategory,
} from '../lib/api/budget'
import { groupCategories } from '../lib/budgetView'
import { currentMonth, isMonth } from '../lib/month'
import { parseIdParam } from '../lib/params'
import CategoryDialog, { type CategoryDialogState } from '../components/CategoryDialog'
import GroupDialog, { type GroupDialogState } from '../components/GroupDialog'
import GoalDialog, { type GoalDialogState } from '../components/GoalDialog'
import BudgetMonthError from '../components/budget/BudgetMonthError'
import CategoryTable from '../components/budget/CategoryTable'
import MonthNavigator from '../components/budget/MonthNavigator'
import MonthSummaryAside from '../components/budget/MonthSummaryAside'
import ReadyToAssign from '../components/budget/ReadyToAssign'

export default function MonthlyBudget() {
  const navigate = useNavigate()
  const { budgetId, month } = useParams()
  const id = parseIdParam(budgetId)

  const now = currentMonth()
  const invalidMonth = month !== undefined && !isMonth(month)
  const selected = invalidMonth || month === undefined ? now : month

  const query = useQuery({
    queryKey: ['budget-month', id, selected],
    queryFn: () => getBudgetMonth(id ?? 0, selected),
    enabled: id !== null,
  })

  const categoryGroups = useQuery({
    queryKey: ['category-groups', id],
    queryFn: () => listCategoryGroups(id ?? 0),
    enabled: id !== null,
  })

  const [dialog, setDialog] = useState<CategoryDialogState | null>(null)
  const [groupDialog, setGroupDialog] = useState<GroupDialogState | null>(null)
  const [goalDialog, setGoalDialog] = useState<GoalDialogState | null>(null)

  const groups = useMemo(
    () => groupCategories(query.data?.categories ?? []),
    [query.data],
  )

  const summary = query.data?.summary

  function groupIdForName(name: string | null): number | null {
    if (name === null) return null
    return categoryGroups.data?.find((g) => g.name === name)?.id ?? null
  }

  function goToMonth(target: string) {
    if (target === now) {
      navigate(`/budget/${id}`)
    } else {
      navigate(`/budget/${id}/${target}`)
    }
  }

  function handleEditCategory(category: BudgetMonthCategory, groupId: number | null) {
    setDialog({
      mode: 'edit',
      categoryId: category.categoryId,
      name: category.categoryName,
      groupId,
    })
  }

  function handleOpenGoal(category: BudgetMonthCategory) {
    setGoalDialog({
      categoryId: category.categoryId,
      categoryName: category.categoryName,
      goal: category.goal,
    })
  }

  // Unreachable: AppLayout rejects invalid ids, but the queries above must be
  // disabled until then and the dialog props need a narrowed type.
  if (id === null) {
    return <Navigate to="/budget" replace />
  }

  return (
    <>
      {invalidMonth && <Navigate to={`/budget/${id}`} replace />}
      <div className="flex min-h-full">
        <div className="mx-auto flex max-w-4xl flex-1 flex-col px-8 pb-6 pt-12">
          <MonthNavigator selected={selected} current={now} onSelect={goToMonth} />

          {query.isPending && <div className="mt-6 min-h-80" />}

          {query.isError && (
            <BudgetMonthError message={query.error.message} onRetry={() => query.refetch()} />
          )}

          {query.isSuccess && summary && (
            <>
              <ReadyToAssign summary={summary} />
              <CategoryTable
                groups={groups}
                budgetId={id}
                month={selected}
                groupIdForName={groupIdForName}
                onAddCategory={() => setDialog({ mode: 'create', groupId: null })}
                onEditGroup={(groupId, name) => setGroupDialog({ groupId, name })}
                onEditCategory={handleEditCategory}
                onOpenGoal={handleOpenGoal}
              />
            </>
          )}
        </div>

        {query.isSuccess && summary && <MonthSummaryAside summary={summary} />}
      </div>

      {dialog && (
        <CategoryDialog
          key={dialog.mode === 'edit' ? `edit-${dialog.categoryId}` : 'create'}
          budgetId={id}
          dialog={dialog}
          groups={categoryGroups.data ?? []}
          onClose={() => setDialog(null)}
        />
      )}

      {groupDialog && (
        <GroupDialog
          key={`group-${groupDialog.groupId}`}
          budgetId={id}
          dialog={groupDialog}
          onClose={() => setGroupDialog(null)}
        />
      )}

      {goalDialog && (
        <GoalDialog
          key={`goal-${goalDialog.categoryId}`}
          budgetId={id}
          month={selected}
          dialog={goalDialog}
          onClose={() => setGoalDialog(null)}
        />
      )}
    </>
  )
}
