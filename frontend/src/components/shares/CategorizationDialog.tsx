import { useMemo, useState, type FormEvent } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  addCategorization,
  deleteCategorization,
  updateCategorization,
  type Category,
  type SplitCategorization,
} from '../../lib/api/expenseShare'
import { centsFromInput, centsToInput } from '../../lib/money'
import DialogShell from '../ui/DialogShell'
import AmountInput from '../ui/AmountInput'
import EntityPicker, { type EntityOption } from '../ui/EntityPicker'
import { CANCEL_BUTTON, DANGER_BUTTON, PRIMARY_BUTTON } from '../ui/buttons'

export type CategorizationDialogState =
  | { mode: 'create' }
  | { mode: 'edit'; line: SplitCategorization }

type CategorizationDialogProps = {
  budgetId: number
  shareId: number
  trxId: number
  categories: Category[]
  dialog: CategorizationDialogState
  onClose: () => void
}

export default function CategorizationDialog({
  budgetId,
  shareId,
  trxId,
  categories,
  dialog,
  onClose,
}: CategorizationDialogProps) {
  const queryClient = useQueryClient()
  const editing = dialog.mode === 'edit' ? dialog.line : null

  const [categoryId, setCategoryId] = useState<number | null>(
    editing?.categoryId ?? null,
  )
  const [outflow, setOutflow] = useState(
    editing && editing.outflow > 0 ? centsToInput(editing.outflow) : '',
  )
  const [inflow, setInflow] = useState(
    editing && editing.inflow > 0 ? centsToInput(editing.inflow) : '',
  )
  const [confirmingDelete, setConfirmingDelete] = useState(false)

  const options: EntityOption[] = useMemo(
    () => categories.map((c) => ({ id: c.id, name: c.name })),
    [categories],
  )

  const outflowCents = centsFromInput(outflow) ?? 0
  const inflowCents = centsFromInput(inflow) ?? 0
  const valid =
    categoryId !== null &&
    ((outflowCents > 0 && inflowCents === 0) || (inflowCents > 0 && outflowCents === 0))

  function input() {
    return {
      categoryId: categoryId ?? 0,
      outflow: outflowCents,
      inflow: inflowCents,
    }
  }

  const save = useMutation({
    mutationFn: () => {
      if (editing) {
        return updateCategorization(budgetId, shareId, trxId, editing.id, input())
      }
      return addCategorization(budgetId, shareId, trxId, input())
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['expense-share', budgetId, shareId] })
      onClose()
    },
  })

  const remove = useMutation({
    mutationFn: () => {
      if (!editing) throw new Error('Nothing to delete')
      return deleteCategorization(budgetId, shareId, trxId, editing.id)
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['expense-share', budgetId, shareId] })
      onClose()
    },
  })

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!valid || save.isPending) return
    save.mutate()
  }

  return (
    <DialogShell onClose={onClose} overflowVisible>
      <h3 className="text-lg font-bold tracking-tight">
        {editing ? 'Edit categorization' : 'Categorize my portion'}
      </h3>
      <form id="categorization-form" onSubmit={handleSubmit} className="mt-5 flex flex-col gap-4">
        <EntityPicker
          label="Category"
          options={options}
          value={categoryId}
          onChange={setCategoryId}
          placeholder="Select a category…"
        />
        <div className="flex gap-4">
          <label className="flex min-w-0 flex-1 flex-col gap-1.5 text-sm font-medium text-slate-300">
            Outflow
            <AmountInput
              value={outflow}
              onChange={setOutflow}
              ariaLabel="Outflow"
              className="w-full px-3 py-2"
            />
          </label>
          <label className="flex min-w-0 flex-1 flex-col gap-1.5 text-sm font-medium text-slate-300">
            Inflow
            <AmountInput
              value={inflow}
              onChange={setInflow}
              ariaLabel="Inflow"
              className="w-full px-3 py-2"
            />
          </label>
        </div>
      </form>
      {(save.isError || remove.isError) && (
        <p className="mt-3 text-sm text-red-400">
          {save.isError ? save.error.message : remove.isError ? remove.error.message : ''}
        </p>
      )}
      <div className="mt-6 flex justify-end gap-2">
        {editing && (
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
          type="submit"
          form="categorization-form"
          disabled={!valid || save.isPending}
          className={PRIMARY_BUTTON}
        >
          {save.isPending ? 'Saving…' : 'Save'}
        </button>
      </div>
    </DialogShell>
  )
}
