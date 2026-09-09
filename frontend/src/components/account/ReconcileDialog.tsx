import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { reconcileAccount } from '../../lib/api/budget'
import { dateFromInput, isDateInput, todayInput } from '../../lib/date'
import { formatMoney, signedCentsFromInput } from '../../lib/money'
import AmountInput from '../ui/AmountInput'
import DialogShell from '../ui/DialogShell'
import { CANCEL_BUTTON, PRIMARY_BUTTON } from '../ui/buttons'

type ReconcileDialogProps = {
  budgetId: number
  accountId: number
  currentBalance: number
  onClose: () => void
}

export default function ReconcileDialog({
  budgetId,
  accountId,
  currentBalance,
  onClose,
}: ReconcileDialogProps) {
  const queryClient = useQueryClient()
  const [date, setDate] = useState(todayInput())
  const [balance, setBalance] = useState('')

  const balanceCents = signedCentsFromInput(balance)
  const canSave = isDateInput(date) && balanceCents !== null

  const reconcile = useMutation({
    mutationFn: () => {
      if (balanceCents === null) throw new Error('Enter a valid balance')
      return reconcileAccount(budgetId, accountId, {
        date: dateFromInput(date),
        balance: balanceCents,
      })
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['account', budgetId, accountId] })
      await queryClient.invalidateQueries({ queryKey: ['accounts', budgetId] })
      onClose()
    },
  })

  return (
    <DialogShell onClose={onClose}>
      <h3 className="text-lg font-bold tracking-tight">Reconcile account</h3>
      <p className="mt-1 text-sm text-slate-400">
        Current balance:{' '}
        <span className="font-semibold tabular-nums text-slate-100">
          {formatMoney(currentBalance)}
        </span>
      </p>

      <div className="mt-5 grid grid-cols-2 gap-3">
        <label className="flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          Statement date
          <input
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            className="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-slate-100 outline-none focus:border-emerald-400"
          />
        </label>
        <label className="flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          Statement balance
          <AmountInput
            signed
            value={balance}
            onChange={setBalance}
            className="px-3 py-2"
          />
        </label>
      </div>
      <p className="mt-1.5 text-xs text-slate-500">
        Marks every transaction up to and including this date reconciled. The statement balance
        must match the calculated balance or nothing is saved.
      </p>

      {reconcile.isError && (
        <p className="mt-3 text-sm text-red-400">{reconcile.error.message}</p>
      )}
      <div className="mt-6 flex justify-end gap-2">
        <button type="button" onClick={onClose} className={CANCEL_BUTTON}>
          Cancel
        </button>
        <button
          type="button"
          onClick={() => reconcile.mutate()}
          disabled={!canSave || reconcile.isPending}
          className={PRIMARY_BUTTON}
        >
          {reconcile.isPending ? 'Reconciling…' : 'Reconcile'}
        </button>
      </div>
    </DialogShell>
  )
}
