import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  deleteExpenseShareTrx,
  updateExpenseShareSplits,
  type ExpenseShareTrx,
} from '../../lib/api/budget'
import { centsFromInput, centsToInput, formatMoney } from '../../lib/money'
import { WarningIcon } from '../icons'
import AmountInput from '../ui/AmountInput'
import DialogShell from '../ui/DialogShell'
import { CANCEL_BUTTON, DANGER_BUTTON, PRIMARY_BUTTON } from '../ui/buttons'

export type ExpenseShareSplitDialogState = {
  transaction: ExpenseShareTrx
}

type ExpenseShareSplitDialogProps = {
  budgetId: number
  expenseShareId: number
  ownBudgetId: number
  dialog: ExpenseShareSplitDialogState
  onClose: () => void
  onDone: () => void
}

export default function ExpenseShareSplitDialog({
  budgetId,
  expenseShareId,
  ownBudgetId,
  dialog,
  onClose,
  onDone,
}: ExpenseShareSplitDialogProps) {
  const queryClient = useQueryClient()
  const { transaction } = dialog
  const outflowMode = transaction.totalInflow === 0
  const requested = outflowMode ? transaction.requestedOutflow : transaction.requestedInflow
  const kept = outflowMode
    ? transaction.totalOutflow - transaction.requestedOutflow
    : transaction.totalInflow - transaction.requestedInflow

  // The publisher keeps the unshared remainder automatically, so its split is
  // read-only. A departed publisher has no split and everything is editable.
  const publisherSplit = transaction.publisherBudgetId
    ? transaction.splits.find((split) => split.budgetId === transaction.publisherBudgetId)
    : undefined
  const editable = transaction.splits.filter((split) => split !== publisherSplit)

  const [amounts, setAmounts] = useState<Record<number, string>>(() => {
    const initial: Record<number, string> = {}
    for (const split of editable) {
      initial[split.budgetId] = centsToInput(
        outflowMode ? split.splitOutflow : split.splitInflow,
      )
    }
    return initial
  })

  const parsed = editable.map((split) => ({
    split,
    cents: centsFromInput(amounts[split.budgetId] ?? ''),
  }))
  const allValid = parsed.every(({ cents }) => cents !== null)
  const total = parsed.reduce((sum, { cents }) => sum + (cents ?? 0), 0)
  const balanced = total === requested
  const canSave = allValid && balanced

  const [confirmingDelete, setConfirmingDelete] = useState(false)

  const save = useMutation({
    mutationFn: () => {
      for (const { cents } of parsed) {
        if (cents === null) throw new Error('Enter valid amounts')
      }
      return updateExpenseShareSplits(
        budgetId,
        expenseShareId,
        transaction.id,
        parsed.map(({ split, cents }) => ({
          budgetId: split.budgetId,
          outflow: outflowMode ? (cents ?? 0) : 0,
          inflow: outflowMode ? 0 : (cents ?? 0),
        })),
      )
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ['expense-share', budgetId, expenseShareId],
      })
      onDone()
    },
  })

  const remove = useMutation({
    mutationFn: () => deleteExpenseShareTrx(budgetId, expenseShareId, transaction.id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ['expense-share', budgetId, expenseShareId],
      })
      onDone()
    },
  })

  return (
    <DialogShell onClose={onClose}>
      <h3 className="text-lg font-bold tracking-tight">Edit splits</h3>
      <p className="mt-1 truncate text-sm text-slate-400">{transaction.payeeName}</p>

      <div className="mt-5 flex flex-col gap-2">
        {publisherSplit && (
          <div className="flex items-center justify-between gap-3 text-sm">
            <span className="min-w-0 truncate text-slate-200">
              {publisherSplit.displayName}
              {publisherSplit.budgetId === ownBudgetId ? ' · you' : ''}
              <span className="text-slate-500"> · publisher, keeps the remainder</span>
            </span>
            <span className="w-28 shrink-0 px-2 py-2 text-right text-sm tabular-nums text-slate-400">
              {formatMoney(kept)}
            </span>
          </div>
        )}
        {parsed.map(({ split, cents }) => (
          <label
            key={split.budgetId}
            className="flex items-center justify-between gap-3 text-sm text-slate-200"
          >
            <span className="min-w-0 truncate">
              {split.displayName}
              {split.budgetId === ownBudgetId ? ' · you' : ''}
            </span>
            <AmountInput
              ariaLabel={`Split amount for ${split.displayName}`}
              value={amounts[split.budgetId] ?? ''}
              onChange={(value) =>
                setAmounts((prev) => ({ ...prev, [split.budgetId]: value }))
              }
              invalid={cents === null}
              className="w-28 px-2 py-2 text-right tabular-nums"
            />
          </label>
        ))}
      </div>

      {!balanced && (
        <p className="mt-3 flex items-center gap-1.5 text-xs font-semibold text-yellow-300">
          <WarningIcon className="h-3.5 w-3.5" />
          Splits ({formatMoney(total)}) don&apos;t add up to the requested{' '}
          {formatMoney(requested)}.
        </p>
      )}

      {(save.isError || remove.isError) && (
        <p className="mt-3 text-sm text-red-400">
          {save.isError ? save.error.message : remove.isError ? remove.error.message : ''}
        </p>
      )}

      <div className="mt-6 flex justify-end gap-2">
        <button
          type="button"
          onClick={() => (confirmingDelete ? remove.mutate() : setConfirmingDelete(true))}
          disabled={remove.isPending}
          className={DANGER_BUTTON}
        >
          {remove.isPending ? 'Deleting…' : confirmingDelete ? 'Confirm delete' : 'Delete'}
        </button>
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
    </DialogShell>
  )
}
