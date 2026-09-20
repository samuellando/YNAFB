import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { publishExpenseShareTrx } from '../../lib/api/budget'

export type PublishShareTarget = {
  id: number
  name: string
  published: boolean
}

type PublishShareButtonProps = {
  budgetId: number
  trxId: number
  target: PublishShareTarget
}

/** Publishes a saved transaction to one expense share. Sits above the row's
 * stretched-link overlay, so it stays clickable without disturbing row clicks. */
export default function PublishShareButton({ budgetId, trxId, target }: PublishShareButtonProps) {
  const queryClient = useQueryClient()
  const [failed, setFailed] = useState(false)

  const publish = useMutation({
    mutationFn: () => publishExpenseShareTrx(budgetId, target.id, trxId),
    onSuccess: async () => {
      setFailed(false)
      await queryClient.invalidateQueries({ queryKey: ['expense-share', budgetId, target.id] })
    },
    onError: async (err) => {
      // Republishing fails unique — refetch flips the button to Published.
      if (/unique constraint failed/i.test(err.message)) {
        await queryClient.invalidateQueries({ queryKey: ['expense-share', budgetId, target.id] })
        return
      }
      setFailed(true)
    },
  })

  if (target.published) {
    return (
      <span className="text-xs font-medium text-emerald-400">Published to {target.name}</span>
    )
  }
  return (
    <button
      type="button"
      title={`Publish the saved transaction to ${target.name}`}
      onClick={() => {
        setFailed(false)
        publish.mutate()
      }}
      disabled={publish.isPending}
      className="relative z-10 cursor-pointer truncate text-left text-xs text-emerald-400/80 transition hover:text-emerald-300 disabled:cursor-not-allowed disabled:opacity-50"
    >
      {publish.isPending ? 'Publishing…' : failed ? 'Retry publish' : `Publish to ${target.name}`}
    </button>
  )
}
