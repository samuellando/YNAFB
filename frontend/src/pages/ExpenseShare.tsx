import { useState } from 'react'
import { Navigate, useNavigate, useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { getExpenseShareCode, listExpenseShares } from '../lib/api/budget'
import { parseIdParam } from '../lib/params'
import ExpenseShareDialog, {
  type ExpenseShareDialogState,
} from '../components/ExpenseShareDialog'
import BudgetMonthError from '../components/budget/BudgetMonthError'
import { PencilIcon } from '../components/icons'
import { OUTLINE_BUTTON } from '../components/ui/buttons'

export default function ExpenseShare() {
  const { budgetId, expenseShareId } = useParams()
  const navigate = useNavigate()
  const budget = parseIdParam(budgetId)
  const id = parseIdParam(expenseShareId)
  const valid = budget !== null && id !== null

  const shares = useQuery({
    queryKey: ['expense-shares', budget],
    queryFn: () => listExpenseShares(budget ?? 0),
    enabled: valid,
  })

  const [rename, setRename] = useState<ExpenseShareDialogState | null>(null)
  const [codeOpen, setCodeOpen] = useState(false)
  const [copied, setCopied] = useState(false)

  const code = useQuery({
    queryKey: ['expense-share-code', budget, id],
    queryFn: () => getExpenseShareCode(budget ?? 0, id ?? 0),
    enabled: valid && codeOpen,
  })

  if (!valid) {
    return <Navigate to={budget !== null ? `/budget/${budget}` : '/budget'} replace />
  }

  const share = shares.data?.find((s) => s.id === id)

  async function copyCode(value: string) {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
    } catch {
      setCopied(false)
    }
  }

  return (
    <>
      <div className="mx-auto flex min-h-full max-w-5xl flex-col px-8 pb-6 pt-12">
        <header className="flex items-center justify-between gap-4">
          <div className="flex min-w-0 items-center gap-2">
            <h1 className="truncate text-3xl font-bold tracking-tight">
              {share?.name ?? 'Expense share'}
            </h1>
            <button
              type="button"
              onClick={() =>
                share &&                 setRename({ mode: 'rename', shareId: share.id, name: share.name, displayName: share.displayName })
              }
              disabled={!share}
              aria-label="Edit expense share"
              title="Edit expense share"
              className="shrink-0 cursor-pointer rounded-lg p-1.5 text-slate-500 transition hover:bg-slate-800 hover:text-emerald-400 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent disabled:hover:text-slate-500"
            >
              <PencilIcon className="h-4 w-4" />
            </button>
          </div>
          <div className="flex shrink-0 gap-2">
            <button
              type="button"
              onClick={() => {
                setCodeOpen((open) => !open)
                setCopied(false)
              }}
              disabled={!share}
              className={OUTLINE_BUTTON}
            >
              {codeOpen ? 'Hide code' : 'Share'}
            </button>
          </div>
        </header>

        {shares.isPending && <div className="mt-6 min-h-80" />}

        {shares.isError && (
          <BudgetMonthError message={shares.error.message} onRetry={() => shares.refetch()} />
        )}

        {shares.isSuccess && !share && (
          <p className="mt-6 text-sm text-slate-400">
            This expense share is no longer linked to this budget.
          </p>
        )}

        {codeOpen && share && (
          <div className="mt-6 rounded-lg border border-slate-800 bg-slate-900 p-4">
            <h2 className="text-sm font-semibold tracking-widest text-slate-500 uppercase">
              Invite code
            </h2>
            {code.isPending && (
              <p className="mt-3 text-sm text-slate-400">Generating code…</p>
            )}
            {code.isError && (
              <p className="mt-3 text-sm text-red-400">{code.error.message}</p>
            )}
            {code.data && (
              <div className="mt-3 flex items-center gap-2">
                <code className="min-w-0 flex-1 truncate rounded-lg bg-slate-950 px-3 py-2 font-mono text-sm text-emerald-400">
                  {code.data}
                </code>
                <button
                  type="button"
                  onClick={() => copyCode(code.data)}
                  className={OUTLINE_BUTTON}
                >
                  {copied ? 'Copied' : 'Copy'}
                </button>
              </div>
            )}
            <p className="mt-3 text-sm text-slate-500">
              Share this code with someone to let them join this expense share. Each
              code can only be used while it is valid.
            </p>
          </div>
        )}
      </div>

      {rename && (
        <ExpenseShareDialog
          key={`rename-${id}`}
          budgetId={budget}
          dialog={rename}
          onClose={() => setRename(null)}
          onDone={() => setRename(null)}
          onDeleted={() => navigate(`/budget/${budget}`)}
        />
      )}
    </>
  )
}
