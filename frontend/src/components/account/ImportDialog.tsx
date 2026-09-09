import { useRef, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { importStatement } from '../../lib/api/budget'
import { CheckIcon } from '../icons'
import DialogShell from '../ui/DialogShell'

type ImportDialogProps = {
  budgetId: number
  accountId: number
  onClose: () => void
}

export default function ImportDialog({ budgetId, accountId, onClose }: ImportDialogProps) {
  const queryClient = useQueryClient()
  const [file, setFile] = useState<File | null>(null)
  const [imported, setImported] = useState<number | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const upload = useMutation({
    mutationFn: () => {
      if (!file) throw new Error('Choose a file to import')
      return importStatement(budgetId, accountId, file)
    },
    onSuccess: async (count) => {
      await queryClient.invalidateQueries({ queryKey: ['account', budgetId, accountId] })
      await queryClient.invalidateQueries({ queryKey: ['accounts', budgetId] })
      await queryClient.invalidateQueries({ queryKey: ['budget-month', budgetId] })
      setImported(count)
    },
  })

  return (
    <DialogShell onClose={onClose}>
      <h3 className="text-lg font-bold tracking-tight">Import statement</h3>
      <p className="mt-1 text-sm text-slate-400">
        Upload a bank statement PDF (Desjardins, Wealthsimple).
      </p>

      {imported === null ? (
        <>
          <button
            type="button"
            onClick={() => inputRef.current?.click()}
            className="mt-5 flex w-full cursor-pointer items-center justify-center rounded-lg border border-dashed border-slate-700 px-3 py-6 text-sm font-medium text-slate-300 transition hover:border-emerald-400 hover:text-emerald-400"
          >
            <span className="truncate">{file ? file.name : 'Choose PDF file'}</span>
          </button>
          <input
            ref={inputRef}
            type="file"
            accept="application/pdf,.pdf"
            className="hidden"
            onChange={(e) => setFile(e.target.files?.[0] ?? null)}
          />

          {upload.isError && (
            <p className="mt-3 text-sm text-red-400">{upload.error.message}</p>
          )}
          <div className="mt-6 flex justify-end gap-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-lg border border-slate-700 px-4 py-2 text-sm font-semibold text-slate-200 transition hover:bg-slate-800"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={() => upload.mutate()}
              disabled={file === null || upload.isPending}
              className="rounded-lg bg-emerald-500 px-4 py-2 text-sm font-semibold text-slate-950 transition hover:bg-emerald-400 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {upload.isPending ? 'Importing…' : 'Import'}
            </button>
          </div>
        </>
      ) : (
        <>
          <p className="mt-5 flex items-center gap-1.5 text-sm font-medium text-emerald-400">
            <CheckIcon className="h-4 w-4 shrink-0" />
            Imported {imported} transaction{imported === 1 ? '' : 's'}
          </p>
          <div className="mt-6 flex justify-end">
            <button
              type="button"
              onClick={onClose}
              autoFocus
              className="rounded-lg bg-emerald-500 px-4 py-2 text-sm font-semibold text-slate-950 transition hover:bg-emerald-400"
            >
              Close
            </button>
          </div>
        </>
      )}
    </DialogShell>
  )
}
