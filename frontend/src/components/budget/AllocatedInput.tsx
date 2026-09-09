import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { setAllocation } from '../../lib/api/budget'
import { centsFromInput, formatMoney } from '../../lib/money'

type AllocatedInputProps = {
  budgetId: number
  categoryId: number
  month: string
  allocated: number
}

export default function AllocatedInput({ budgetId, categoryId, month, allocated }: AllocatedInputProps) {
  const queryClient = useQueryClient()
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState('')

  const update = useMutation({
    mutationFn: (amount: number) => setAllocation(budgetId, categoryId, month, amount),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['budget-month', budgetId, month] })
    },
  })

  function startEdit() {
    setDraft((allocated / 100).toFixed(2))
    setEditing(true)
  }

  function commit() {
    if (!editing) return
    setEditing(false)
    const cents = centsFromInput(draft)
    if (cents === null) return
    if (cents === allocated) return
    update.mutate(cents)
  }

  function cancel() {
    setEditing(false)
    setDraft('')
  }

  if (!editing) {
    return (
      <button
        type="button"
        onClick={startEdit}
        title="Edit allocated amount"
        aria-label={`Allocated ${formatMoney(allocated)}. Activate to edit.`}
        className="ml-auto w-28 cursor-pointer rounded px-1.5 py-0.5 text-right text-sm tabular-nums text-slate-100 transition hover:bg-slate-800/60 hover:text-emerald-300"
      >
        {formatMoney(allocated)}
      </button>
    )
  }

  return (
    <input
      type="text"
      inputMode="decimal"
      value={draft}
      autoFocus
      onFocus={(e) => e.currentTarget.select()}
      onChange={(e) => setDraft(e.target.value)}
      onBlur={commit}
      onKeyDown={(e) => {
        if (e.key === 'Enter') e.currentTarget.blur()
        if (e.key === 'Escape') cancel()
      }}
      aria-label="Allocated amount"
      className="ml-auto w-28 rounded border border-emerald-400 bg-slate-800/60 px-1.5 py-0.5 text-right text-sm tabular-nums text-slate-100 outline-none transition focus:ring-1 focus:ring-emerald-400"
    />
  )
}
