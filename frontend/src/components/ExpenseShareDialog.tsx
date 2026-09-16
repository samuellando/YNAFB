import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  createExpenseShare,
  joinExpenseShare,
  leaveExpenseShare,
  updateExpenseShare,
} from '../lib/api/budget'
import DialogShell from './ui/DialogShell'
import { CANCEL_BUTTON, DANGER_BUTTON, PRIMARY_BUTTON } from './ui/buttons'

export type ExpenseShareDialogState =
  | { mode: 'create' }
  | { mode: 'join' }
  | { mode: 'rename'; shareId: number; name: string }

type ExpenseShareDialogProps = {
  budgetId: number
  dialog: ExpenseShareDialogState
  onClose: () => void
  onDone: (shareId: number) => void
  onDeleted?: () => void
}

export default function ExpenseShareDialog({
  budgetId,
  dialog,
  onClose,
  onDone,
  onDeleted,
}: ExpenseShareDialogProps) {
  const queryClient = useQueryClient()
  const [tab, setTab] = useState<'create' | 'join'>(
    dialog.mode === 'join' ? 'join' : 'create',
  )
  const [name, setName] = useState(dialog.mode === 'rename' ? dialog.name : '')
  const [code, setCode] = useState('')

  async function invalidate() {
    await queryClient.invalidateQueries({ queryKey: ['expense-shares', budgetId] })
  }

  const create = useMutation({
    mutationFn: () => createExpenseShare(budgetId, name.trim()),
    onSuccess: async (share) => {
      await invalidate()
      onDone(share.id)
    },
  })

  const join = useMutation({
    mutationFn: () => joinExpenseShare(budgetId, code.trim()),
    onSuccess: async (share) => {
      await invalidate()
      onDone(share.id)
    },
  })

  if (dialog.mode === 'rename') {
    return <RenameDialog budgetId={budgetId} dialog={dialog} onClose={onClose} onDone={onDone} onDeleted={onDeleted} />
  }

  const active = tab === 'create' ? create : join
  const valid = tab === 'create' ? name.trim() !== '' : code.trim() !== ''

  return (
    <DialogShell onClose={onClose}>
      <h3 className="text-lg font-bold tracking-tight">Add expense share</h3>
      <div className="mt-4 flex gap-1 rounded-lg bg-slate-950 p-1">
        {(['create', 'join'] as const).map((t) => (
          <button
            key={t}
            type="button"
            onClick={() => setTab(t)}
            className={`flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition ${
              tab === t ? 'bg-slate-800 text-emerald-400' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            {t === 'create' ? 'New share' : 'Join with code'}
          </button>
        ))}
      </div>
      {tab === 'create' ? (
        <label className="mt-5 flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          Name
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Expense share name"
            autoFocus
            className="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-slate-100 outline-none focus:border-emerald-400"
          />
        </label>
      ) : (
        <label className="mt-5 flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          Invite code
          <input
            type="text"
            value={code}
            onChange={(e) => setCode(e.target.value)}
            placeholder="Paste the share code"
            autoFocus
            className="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-slate-100 outline-none focus:border-emerald-400"
          />
        </label>
      )}
      {active.isError && (
        <p className="mt-3 text-sm text-red-400">{active.error.message}</p>
      )}
      <div className="mt-6 flex justify-end gap-2">
        <button type="button" onClick={onClose} className={CANCEL_BUTTON}>
          Cancel
        </button>
        <button
          type="button"
          onClick={() => active.mutate()}
          disabled={!valid || active.isPending}
          className={PRIMARY_BUTTON}
        >
          {active.isPending ? 'Saving…' : tab === 'create' ? 'Create' : 'Join'}
        </button>
      </div>
    </DialogShell>
  )
}

type RenameDialogProps = {
  budgetId: number
  dialog: Extract<ExpenseShareDialogState, { mode: 'rename' }>
  onClose: () => void
  onDone: (shareId: number) => void
  onDeleted?: () => void
}

function RenameDialog({ budgetId, dialog, onClose, onDone, onDeleted }: RenameDialogProps) {
  const queryClient = useQueryClient()
  const [name, setName] = useState(dialog.name)
  const [confirmingLeave, setConfirmingLeave] = useState(false)

  const save = useMutation({
    mutationFn: () => updateExpenseShare(budgetId, dialog.shareId, name.trim()),
    onSuccess: async (share) => {
      await queryClient.invalidateQueries({ queryKey: ['expense-shares', budgetId] })
      onDone(share.id)
    },
  })

  const leave = useMutation({
    mutationFn: () => leaveExpenseShare(budgetId, dialog.shareId),
    onSuccess: async () => {
      queryClient.removeQueries({ queryKey: ['expense-share-code', budgetId, dialog.shareId] })
      await queryClient.invalidateQueries({ queryKey: ['expense-shares', budgetId] })
      onDeleted?.()
    },
  })

  return (
    <DialogShell onClose={onClose}>
      <h3 className="text-lg font-bold tracking-tight">Edit expense share</h3>
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
      {(save.isError || leave.isError) && (
        <p className="mt-3 text-sm text-red-400">
          {save.isError ? save.error.message : leave.isError ? leave.error.message : ''}
        </p>
      )}
      <div className="mt-6 flex justify-end gap-2">
        <button
          type="button"
          onClick={() => (confirmingLeave ? leave.mutate() : setConfirmingLeave(true))}
          disabled={leave.isPending}
          className={DANGER_BUTTON}
        >
          {leave.isPending ? 'Leaving…' : confirmingLeave ? 'Confirm leave' : 'Leave'}
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
