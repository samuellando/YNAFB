import { useState, type FormEvent } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { renameShare } from '../../lib/api/expenseShare'
import DialogShell from '../ui/DialogShell'
import { CANCEL_BUTTON, PRIMARY_BUTTON } from '../ui/buttons'

export type ShareRenameDialogState = {
  name: string
  displayName: string
}

type ShareRenameDialogProps = {
  budgetId: number
  shareId: number
  dialog: ShareRenameDialogState
  onClose: () => void
}

const inputClass =
  'rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-slate-100 outline-none focus:border-emerald-400'

export default function ShareRenameDialog({
  budgetId,
  shareId,
  dialog,
  onClose,
}: ShareRenameDialogProps) {
  const queryClient = useQueryClient()
  const [name, setName] = useState(dialog.name)
  const [displayName, setDisplayName] = useState(dialog.displayName)

  const valid = name.trim().length >= 3 && displayName.trim().length >= 3

  const save = useMutation({
    mutationFn: () =>
      renameShare(budgetId, shareId, {
        name: name.trim(),
        displayName: displayName.trim(),
      }),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['expense-share', budgetId, shareId] }),
        queryClient.invalidateQueries({ queryKey: ['expense-shares', budgetId] }),
      ])
      onClose()
    },
  })

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!valid || save.isPending) return
    save.mutate()
  }

  return (
    <DialogShell onClose={onClose}>
      <h3 className="text-lg font-bold tracking-tight">Rename expense share</h3>
      <form id="share-rename-form" onSubmit={handleSubmit} className="mt-5 flex flex-col gap-4">
        <label className="flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          Name
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoFocus
            className={inputClass}
          />
        </label>
        <label className="flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          Your display name
          <input
            type="text"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            className={inputClass}
          />
        </label>
      </form>
      {save.isError && <p className="mt-3 text-sm text-red-400">{save.error.message}</p>}
      <div className="mt-6 flex justify-end gap-2">
        <button type="button" onClick={onClose} className={CANCEL_BUTTON}>
          Cancel
        </button>
        <button
          type="submit"
          form="share-rename-form"
          disabled={!valid || save.isPending}
          className={PRIMARY_BUTTON}
        >
          {save.isPending ? 'Saving…' : 'Save'}
        </button>
      </div>
    </DialogShell>
  )
}
