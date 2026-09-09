import { useState, type FormEvent } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  createCategory,
  deleteCategory,
  updateCategory,
  type CategoryGroup,
} from '../lib/api/budget'
import GroupPicker from './GroupPicker'
import DialogShell from './ui/DialogShell'
import { CANCEL_BUTTON, DANGER_BUTTON, PRIMARY_BUTTON } from './ui/buttons'

export type CategoryDialogState =
  | { mode: 'edit'; categoryId: number; name: string; groupId: number | null }
  | { mode: 'create'; groupId: number | null }

type CategoryDialogProps = {
  budgetId: number
  dialog: CategoryDialogState
  groups: CategoryGroup[]
  onClose: () => void
}

export default function CategoryDialog({ budgetId, dialog, groups, onClose }: CategoryDialogProps) {
  const queryClient = useQueryClient()
  const [name, setName] = useState(dialog.mode === 'edit' ? dialog.name : '')
  const [groupId, setGroupId] = useState<number | null>(dialog.groupId)
  const [confirmingDelete, setConfirmingDelete] = useState(false)

  const save = useMutation({
    mutationFn: async () => {
      if (dialog.mode === 'edit') {
        await updateCategory(budgetId, dialog.categoryId, name.trim(), groupId ?? undefined)
      } else {
        await createCategory(budgetId, name.trim(), groupId ?? undefined)
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['budget-month', budgetId] })
      queryClient.invalidateQueries({ queryKey: ['category-groups', budgetId] })
      onClose()
    },
  })

  const remove = useMutation({
    mutationFn: () => {
      if (dialog.mode !== 'edit') throw new Error('Nothing to delete')
      return deleteCategory(budgetId, dialog.categoryId)
    },
    onSuccess: async () => {
      if (dialog.mode === 'edit') {
        queryClient.removeQueries({ queryKey: ['goal', budgetId, dialog.categoryId] })
      }
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['budget-month', budgetId] }),
        queryClient.invalidateQueries({ queryKey: ['category-groups', budgetId] }),
      ])
      onClose()
    },
  })

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (name.trim() === '' || save.isPending) return
    save.mutate()
  }

  return (
    <DialogShell onClose={onClose} overflowVisible>
      <h3 className="text-lg font-bold tracking-tight">
        {dialog.mode === 'edit' ? 'Edit category' : 'Add category'}
      </h3>
      <form id="category-form" onSubmit={handleSubmit} className="mt-5">
        <label className="flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          Name
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoFocus
            className="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-slate-100 outline-none focus:border-emerald-400"
          />
        </label>
      </form>
      <GroupPicker budgetId={budgetId} groups={groups} value={groupId} onChange={setGroupId} />
      {(save.isError || remove.isError) && (
        <p className="mt-3 text-sm text-red-400">
          {save.isError ? save.error.message : remove.isError ? remove.error.message : ''}
        </p>
      )}
      <div className="mt-6 flex justify-end gap-2">
        {dialog.mode === 'edit' && (
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
          form="category-form"
          disabled={name.trim() === '' || save.isPending}
          className={PRIMARY_BUTTON}
        >
          {save.isPending ? 'Saving…' : 'Save'}
        </button>
      </div>
    </DialogShell>
  )
}
