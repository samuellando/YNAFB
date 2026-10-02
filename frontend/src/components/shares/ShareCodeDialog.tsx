import { useState } from 'react'
import type { ExpenseShareCode } from '../../lib/api/expenseShare'
import DialogShell from '../ui/DialogShell'
import { CANCEL_BUTTON, PRIMARY_BUTTON } from '../ui/buttons'

type ShareCodeDialogProps = {
  code: ExpenseShareCode
  onClose: () => void
}

export default function ShareCodeDialog({ code, onClose }: ShareCodeDialogProps) {
  const [copied, setCopied] = useState(false)

  async function copy() {
    try {
      await navigator.clipboard.writeText(code.code)
      setCopied(true)
    } catch {
      setCopied(false)
    }
  }

  return (
    <DialogShell onClose={onClose}>
      <h3 className="text-lg font-bold tracking-tight">Invite code</h3>
      <p className="mt-2 text-sm text-slate-400">
        Share this code with the other member. It is shown once and expires{' '}
        {new Date(code.expires).toLocaleString()}.
      </p>
      <div className="mt-4 flex gap-2">
        <input
          type="text"
          value={code.code}
          readOnly
          onFocus={(e) => e.target.select()}
          className="min-w-0 flex-1 rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 font-mono text-sm text-slate-100 outline-none"
        />
        <button type="button" onClick={copy} className={PRIMARY_BUTTON}>
          {copied ? 'Copied' : 'Copy'}
        </button>
      </div>
      <div className="mt-6 flex justify-end">
        <button type="button" onClick={onClose} className={CANCEL_BUTTON}>
          Done
        </button>
      </div>
    </DialogShell>
  )
}
