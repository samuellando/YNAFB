import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { deleteAccount, updateAccount } from '../../lib/api/budget'
import DialogShell from '../ui/DialogShell'
import { CANCEL_BUTTON, DANGER_BUTTON, PRIMARY_BUTTON } from '../ui/buttons'

export type AccountDialogState = {
  name: string
}

type AccountDialogProps = {
  budgetId: number
  accountId: number
  dialog: AccountDialogState
  onClose: () => void
  onDeleted: () => void
}

export default function AccountDialog({ budgetId, accountId, dialog, onClose, onDeleted }: AccountDialogProps) {
  const queryClient = useQueryClient()
  const [name, setName] = useState(dialog.name)
  const [confirmingDelete, setConfirmingDelete] = useState(false)

  const save = useMutation({
    mutationFn: () => updateAccount(budgetId, accountId, name.trim()),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['account', budgetId, accountId] }),
        queryClient.invalidateQueries({ queryKey: ['accounts', budgetId] }),
      ])
      onClose()
    },
  })

  const remove = useMutation({
    mutationFn: () => deleteAccount(budgetId, accountId),
    onSuccess: async () => {
      queryClient.removeQueries({ queryKey: ['account', budgetId, accountId] })
      await queryClient.invalidateQueries({ queryKey: ['accounts', budgetId] })
      onDeleted()
    },
  })

  return (
    <DialogShell onClose={onClose}>
        <h3 className="text-lg font-bold tracking-tight">Edit account</h3>
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
            disabled={name.trim() === '' || save.isPending}
            className={PRIMARY_BUTTON}
          >
            {save.isPending ? 'Saving…' : 'Save'}
          </button>
        </div>
    </DialogShell>
  )
}
