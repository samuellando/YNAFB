import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { leaveShare } from '../../lib/api/expenseShare'
import DialogShell from '../ui/DialogShell'
import { CANCEL_BUTTON, DANGER_BUTTON } from '../ui/buttons'

type ShareLeaveDialogProps = {
  budgetId: number
  shareId: number
  name: string
  onClose: () => void
  onLeft: () => void
}

export default function ShareLeaveDialog({
  budgetId,
  shareId,
  name,
  onClose,
  onLeft,
}: ShareLeaveDialogProps) {
  const queryClient = useQueryClient()
  const [armed, setArmed] = useState(false)

  const leave = useMutation({
    mutationFn: () => leaveShare(budgetId, shareId),
    onSuccess: async () => {
      queryClient.removeQueries({ queryKey: ['expense-share', budgetId, shareId] })
      await queryClient.invalidateQueries({ queryKey: ['expense-shares', budgetId] })
      onLeft()
    },
  })

  return (
    <DialogShell onClose={onClose}>
      <h3 className="text-lg font-bold tracking-tight">Leave expense share</h3>
      <p className="mt-2 text-sm text-slate-400">
        Leave &ldquo;{name}&rdquo;? You will no longer see its transactions.
      </p>
      {leave.isError && (
        <p className="mt-3 text-sm text-red-400">{leave.error.message}</p>
      )}
      <div className="mt-6 flex justify-end gap-2">
        <button type="button" onClick={onClose} className={CANCEL_BUTTON}>
          Cancel
        </button>
        <button
          type="button"
          onClick={() => (armed ? leave.mutate() : setArmed(true))}
          disabled={leave.isPending}
          className={DANGER_BUTTON}
        >
          {leave.isPending ? 'Leaving…' : armed ? 'Confirm leave' : 'Leave'}
        </button>
      </div>
    </DialogShell>
  )
}
