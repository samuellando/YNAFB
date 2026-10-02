import { useState, type FormEvent } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  joinShare,
  type ExpenseShareMembership,
} from '../../lib/api/expenseShare'
import DialogShell from '../ui/DialogShell'
import { CANCEL_BUTTON, PRIMARY_BUTTON } from '../ui/buttons'

type ShareJoinDialogProps = {
  budgetId: number
  onClose: () => void
  onJoined: (membership: ExpenseShareMembership) => void
}

const inputClass =
  'rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-slate-100 outline-none focus:border-emerald-400'

export default function ShareJoinDialog({
  budgetId,
  onClose,
  onJoined,
}: ShareJoinDialogProps) {
  const queryClient = useQueryClient()
  const [code, setCode] = useState('')
  const [name, setName] = useState('')
  const [displayName, setDisplayName] = useState('')

  const valid =
    code.trim() !== '' && name.trim().length >= 3 && displayName.trim().length >= 3

  const join = useMutation({
    mutationFn: () =>
      joinShare(budgetId, {
        code: code.trim(),
        name: name.trim(),
        displayName: displayName.trim(),
      }),
    onSuccess: async (membership) => {
      await queryClient.invalidateQueries({ queryKey: ['expense-shares', budgetId] })
      onJoined(membership)
    },
  })

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!valid || join.isPending) return
    join.mutate()
  }

  return (
    <DialogShell onClose={onClose}>
      <h3 className="text-lg font-bold tracking-tight">Join expense share</h3>
      <form id="share-join-form" onSubmit={handleSubmit} className="mt-5 flex flex-col gap-4">
        <label className="flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          Invite code
          <input
            type="text"
            value={code}
            onChange={(e) => setCode(e.target.value)}
            autoFocus
            className={inputClass}
          />
        </label>
        <label className="flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          Name
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="At least 3 characters"
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
      </form>
      {join.isError && <p className="mt-3 text-sm text-red-400">{join.error.message}</p>}
      <div className="mt-6 flex justify-end gap-2">
        <button type="button" onClick={onClose} className={CANCEL_BUTTON}>
          Cancel
        </button>
        <button
          type="submit"
          form="share-join-form"
          disabled={!valid || join.isPending}
          className={PRIMARY_BUTTON}
        >
          {join.isPending ? 'Joining…' : 'Join'}
        </button>
      </div>
    </DialogShell>
  )
}
