import { useState, type FormEvent } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  createShare,
  type ExpenseShareMembership,
} from '../../lib/api/expenseShare'
import DialogShell from '../ui/DialogShell'
import { CANCEL_BUTTON, PRIMARY_BUTTON } from '../ui/buttons'

type ShareCreateDialogProps = {
  budgetId: number
  onClose: () => void
  onCreated: (membership: ExpenseShareMembership) => void
}

const inputClass =
  'rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-slate-100 outline-none focus:border-emerald-400'

export default function ShareCreateDialog({
  budgetId,
  onClose,
  onCreated,
}: ShareCreateDialogProps) {
  const queryClient = useQueryClient()
  const [name, setName] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [defaultName, setDefaultName] = useState('')

  const valid = name.trim().length >= 3 && displayName.trim().length >= 3

  const create = useMutation({
    mutationFn: () =>
      createShare(budgetId, {
        name: name.trim(),
        displayName: displayName.trim(),
        ...(defaultName.trim() !== '' ? { defaultName: defaultName.trim() } : {}),
      }),
    onSuccess: async (membership) => {
      await queryClient.invalidateQueries({ queryKey: ['expense-shares', budgetId] })
      onCreated(membership)
    },
  })

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!valid || create.isPending) return
    create.mutate()
  }

  return (
    <DialogShell onClose={onClose}>
      <h3 className="text-lg font-bold tracking-tight">New expense share</h3>
      <form id="share-create-form" onSubmit={handleSubmit} className="mt-5 flex flex-col gap-4">
        <label className="flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          Name
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="At least 3 characters"
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
            placeholder="How others see you"
            className={inputClass}
          />
        </label>
        <label className="flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          Default name <span className="font-normal text-slate-500">(optional, defaults to name)</span>
          <input
            type="text"
            value={defaultName}
            onChange={(e) => setDefaultName(e.target.value)}
            className={inputClass}
          />
        </label>
      </form>
      {create.isError && (
        <p className="mt-3 text-sm text-red-400">{create.error.message}</p>
      )}
      <div className="mt-6 flex justify-end gap-2">
        <button type="button" onClick={onClose} className={CANCEL_BUTTON}>
          Cancel
        </button>
        <button
          type="submit"
          form="share-create-form"
          disabled={!valid || create.isPending}
          className={PRIMARY_BUTTON}
        >
          {create.isPending ? 'Creating…' : 'Create'}
        </button>
      </div>
    </DialogShell>
  )
}
