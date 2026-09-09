import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { updateCategoryGroup } from '../lib/api/budget'
import DialogShell from './ui/DialogShell'

export type GroupDialogState = {
  groupId: number
  name: string
}

type GroupDialogProps = {
  budgetId: number
  dialog: GroupDialogState
  onClose: () => void
}

export default function GroupDialog({ budgetId, dialog, onClose }: GroupDialogProps) {
  const queryClient = useQueryClient()
  const [name, setName] = useState(dialog.name)

  const save = useMutation({
    mutationFn: () => updateCategoryGroup(budgetId, dialog.groupId, name.trim()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['budget-month', budgetId] })
      queryClient.invalidateQueries({ queryKey: ['category-groups', budgetId] })
      onClose()
    },
  })

  return (
    <DialogShell onClose={onClose}>
        <h3 className="text-lg font-bold tracking-tight">Edit group</h3>
        <label className="mt-5 flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          Name
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoFocus
            className="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-slate-100 outline-none focus:border-emerald-400"
          />
        </label>
        {save.isError && <p className="mt-3 text-sm text-red-400">{save.error.message}</p>}
        <div className="mt-6 flex justify-end gap-2">
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg border border-slate-700 px-4 py-2 text-sm font-semibold text-slate-200 transition hover:bg-slate-800"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={() => save.mutate()}
            disabled={name.trim() === '' || save.isPending}
            className="rounded-lg bg-emerald-500 px-4 py-2 text-sm font-semibold text-slate-950 transition hover:bg-emerald-400 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {save.isPending ? 'Saving…' : 'Save'}
          </button>
        </div>
    </DialogShell>
  )
}
