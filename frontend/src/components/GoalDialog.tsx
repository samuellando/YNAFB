import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createGoal,
  deleteGoal,
  getGoal,
  updateGoal,
  type BudgetMonthGoal,
  type Goal,
  type GoalType,
} from '../lib/api/budget'
import { centsFromInput, centsToInput } from '../lib/money'
import { isMonth } from '../lib/month'
import AmountInput from './ui/AmountInput'
import DialogShell from './ui/DialogShell'
import { CANCEL_BUTTON, DANGER_BUTTON, PRIMARY_BUTTON } from './ui/buttons'

export type GoalDialogState = {
  categoryId: number
  categoryName: string
  goal?: BudgetMonthGoal
}

type GoalDialogProps = {
  budgetId: number
  month: string
  dialog: GoalDialogState
  onClose: () => void
}

const GOAL_TYPES: { value: GoalType; label: string; description: string }[] = [
  {
    value: 'monthly',
    label: 'Monthly',
    description: 'Allocate a fixed amount every month.',
  },
  {
    value: 'refill',
    label: 'Refill',
    description: 'Top the category back up to the target when it is spent.',
  },
  {
    value: 'save',
    label: 'Save',
    description: 'Spread the target across the months until the end month.',
  },
]

function toYearMonth(value: string): string {
  return value.slice(0, 7)
}

type GoalFormProps = {
  budgetId: number
  month: string
  dialog: GoalDialogState
  existing: Goal | undefined
  onClose: () => void
}

function GoalForm({ budgetId, month, dialog, existing, onClose }: GoalFormProps) {
  const queryClient = useQueryClient()
  const [type, setType] = useState<GoalType>(existing?.type ?? 'monthly')
  const [amount, setAmount] = useState(existing ? centsToInput(existing.amount) : '')
  const [startMonth, setStartMonth] = useState(
    existing ? toYearMonth(existing.startMonth) : month,
  )
  const [endMonth, setEndMonth] = useState(
    existing?.endMonth ? toYearMonth(existing.endMonth) : '',
  )
  const [confirmingDelete, setConfirmingDelete] = useState(false)

  const amountCents = centsFromInput(amount)
  const validAmount = amountCents !== null && amountCents > 0
  const validStart = isMonth(startMonth.trim())
  const end = endMonth.trim()
  const validEnd = end === '' ? type !== 'save' : isMonth(end)
  const canSave = validAmount && validStart && validEnd

  async function invalidate() {
    await queryClient.invalidateQueries({ queryKey: ['budget-month', budgetId] })
    await queryClient.invalidateQueries({ queryKey: ['goal', budgetId, dialog.categoryId] })
  }

  const save = useMutation({
    mutationFn: () => {
      if (amountCents === null) throw new Error('Enter a valid amount')
      const input = {
        type,
        startMonth: startMonth.trim(),
        ...(end !== '' ? { endMonth: end } : {}),
        amount: amountCents,
      }
      return existing
        ? updateGoal(budgetId, dialog.categoryId, input)
        : createGoal(budgetId, dialog.categoryId, input)
    },
    onSuccess: async () => {
      await invalidate()
      onClose()
    },
  })

  const remove = useMutation({
    mutationFn: () => deleteGoal(budgetId, dialog.categoryId),
    onSuccess: async () => {
      queryClient.removeQueries({ queryKey: ['goal', budgetId, dialog.categoryId] })
      await queryClient.invalidateQueries({ queryKey: ['budget-month', budgetId] })
      onClose()
    },
  })

  return (
    <>
      <h3 className="text-lg font-bold tracking-tight">{existing ? 'Edit goal' : 'New goal'}</h3>
      <p className="mt-1 text-sm text-slate-400">{dialog.categoryName}</p>

      <div className="mt-5 flex flex-col gap-1.5 text-sm font-medium text-slate-300">
        <span>Type</span>
        <div className="grid grid-cols-3 gap-2">
          {GOAL_TYPES.map((option) => (
            <button
              key={option.value}
              type="button"
              onClick={() => setType(option.value)}
              className={`rounded-lg border px-3 py-2 text-sm font-semibold transition ${
                type === option.value
                  ? 'border-emerald-400 bg-emerald-500/10 text-emerald-300'
                  : 'border-slate-700 text-slate-300 hover:border-slate-500'
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>
        <p className="text-xs font-normal text-slate-500">
          {GOAL_TYPES.find((option) => option.value === type)?.description ?? ''}
        </p>
      </div>

      <label className="mt-4 flex flex-col gap-1.5 text-sm font-medium text-slate-300">
        Amount
        <AmountInput value={amount} onChange={setAmount} className="px-3 py-2" />
      </label>

      <div className="mt-4 grid grid-cols-2 gap-3">
        <label className="flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          Start month
          <input
            type="text"
            value={startMonth}
            onChange={(e) => setStartMonth(e.target.value)}
            placeholder="YYYY-MM"
            maxLength={7}
            className="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-slate-100 outline-none focus:border-emerald-400"
          />
        </label>
        <label className="flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          End month{type === 'save' ? ' *' : ''}
          <input
            type="text"
            value={endMonth}
            onChange={(e) => setEndMonth(e.target.value)}
            placeholder="YYYY-MM"
            maxLength={7}
            className="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-slate-100 outline-none focus:border-emerald-400"
          />
        </label>
      </div>
      {type === 'save' && (
        <p className="mt-1.5 text-xs text-slate-500">Save goals require an end month.</p>
      )}

      {(save.isError || remove.isError) && (
        <p className="mt-3 text-sm text-red-400">
          {save.isError ? save.error.message : remove.isError ? remove.error.message : ''}
        </p>
      )}
      <div className="mt-6 flex justify-end gap-2">
        {existing && (
            <button
              type="button"
              onClick={() => (confirmingDelete ? remove.mutate() : setConfirmingDelete(true))}
              disabled={remove.isPending}
              className={DANGER_BUTTON}
            >
            {remove.isPending ? 'Deleting…' : confirmingDelete ? 'Confirm delete' : 'Delete'}
          </button>
        )}
        <button type="button" onClick={onClose} className={CANCEL_BUTTON}>
          Cancel
        </button>
        <button
          type="button"
          onClick={() => save.mutate()}
          disabled={!canSave || save.isPending}
          className={PRIMARY_BUTTON}
        >
          {save.isPending ? 'Saving…' : 'Save'}
        </button>
      </div>
    </>
  )
}

export default function GoalDialog({ budgetId, month, dialog, onClose }: GoalDialogProps) {
  const goalQuery = useQuery({
    queryKey: ['goal', budgetId, dialog.categoryId],
    queryFn: () => getGoal(budgetId, dialog.categoryId),
    retry: false,
  })

  return (
    <DialogShell onClose={onClose}>
        {goalQuery.isPending ? (
          <p className="py-6 text-center text-sm text-slate-400">Loading goal…</p>
        ) : (
          <GoalForm
            key={goalQuery.data ? `edit-${goalQuery.data.id}` : 'create'}
            budgetId={budgetId}
            month={month}
            dialog={dialog}
            existing={goalQuery.data}
            onClose={onClose}
          />
        )}
    </DialogShell>
  )
}
