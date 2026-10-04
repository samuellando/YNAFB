import { useState } from 'react'
import { Navigate, useNavigate, useParams } from 'react-router'
import { useMutation, useQuery } from '@tanstack/react-query'
import {
  getShareDetail,
  listCategories,
  mintShareCode,
  type ExpenseShareCode,
  type ExpenseShareTransaction,
  type SplitCategorization,
} from '../lib/api/expenseShare'
import { ApiError } from '../lib/api/errors'
import { formatDate } from '../lib/date'
import { formatMoney } from '../lib/money'
import { parseIdParam } from '../lib/params'
import BudgetMonthError from '../components/budget/BudgetMonthError'
import { PencilIcon, PlusIcon } from '../components/icons'
import CategorizationDialog, {
  type CategorizationDialogState,
} from '../components/shares/CategorizationDialog'
import ShareCodeDialog from '../components/shares/ShareCodeDialog'
import ShareLeaveDialog from '../components/shares/ShareLeaveDialog'
import ShareTransactionTable from '../components/shares/ShareTransactionTable'
import ShareRenameDialog, {
  type ShareRenameDialogState,
} from '../components/shares/ShareRenameDialog'
import { OUTLINE_BUTTON } from '../components/ui/buttons'

type CategorizationTarget = {
  trxId: number
  dialog: CategorizationDialogState
}

function balanceTone(balance: number): string {
  if (balance > 0) return 'text-emerald-400'
  if (balance < 0) return 'text-red-400'
  return 'text-slate-300'
}

export default function ShareDetail() {
  const { budgetId, shareId } = useParams()
  const navigate = useNavigate()
  const budget = parseIdParam(budgetId)
  const share = parseIdParam(shareId)
  const valid = budget !== null && share !== null

  const query = useQuery({
    queryKey: ['expense-share', budget, share],
    queryFn: () => getShareDetail(budget ?? 0, share ?? 0),
    enabled: valid,
  })

  const categoriesQuery = useQuery({
    queryKey: ['categories', budget],
    queryFn: () => listCategories(budget ?? 0),
    enabled: valid,
  })

  const [rename, setRename] = useState<ShareRenameDialogState | null>(null)
  const [code, setCode] = useState<ExpenseShareCode | null>(null)
  const [leaveOpen, setLeaveOpen] = useState(false)
  const [target, setTarget] = useState<CategorizationTarget | null>(null)

  const mint = useMutation({
    mutationFn: () => mintShareCode(budget ?? 0, share ?? 0),
    onSuccess: (minted) => setCode(minted),
  })

  if (!valid) {
    return <Navigate to={budget !== null ? `/budget/${budget}` : '/budget'} replace />
  }

  if (query.isError && query.error instanceof ApiError && query.error.status === 404) {
    return <Navigate to={`/budget/${budget}`} replace />
  }

  const detail = query.data
  const membership = detail?.membership

  return (
    <>
      <div className="mx-auto flex min-h-full max-w-5xl flex-col px-8 pb-6 pt-12">
        <header className="flex items-center justify-between gap-4">
          <div className="flex min-w-0 items-center gap-2">
            <h1 className="truncate text-3xl font-bold tracking-tight">
              {membership?.name ?? 'Expense share'}
            </h1>
            <button
              type="button"
              onClick={() =>
                membership &&
                setRename({ name: membership.name, displayName: membership.displayName })
              }
              disabled={!membership}
              aria-label="Rename expense share"
              title="Rename expense share"
              className="shrink-0 cursor-pointer rounded-lg p-1.5 text-slate-500 transition hover:bg-slate-800 hover:text-emerald-400 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent disabled:hover:text-slate-500"
            >
              <PencilIcon className="h-4 w-4" />
            </button>
          </div>
          <div className="flex shrink-0 gap-2">
            <button
              type="button"
              onClick={() => mint.mutate()}
              disabled={!membership || mint.isPending}
              className={OUTLINE_BUTTON}
            >
              {mint.isPending ? '…' : 'Invite code'}
            </button>
            <button
              type="button"
              onClick={() => setLeaveOpen(true)}
              disabled={!membership}
              className={OUTLINE_BUTTON}
            >
              Leave
            </button>
          </div>
        </header>

        {mint.isError && (
          <p className="mt-3 text-sm text-red-400">{mint.error.message}</p>
        )}

        {query.isPending && <div className="mt-6 min-h-80" />}

        {query.isError && (
          <BudgetMonthError message={query.error.message} onRetry={() => query.refetch()} />
        )}

        {query.isSuccess && detail && (
          <>
            <section className="mt-6 rounded-xl border border-slate-800 bg-slate-900 p-5">
              <h2 className="text-xs font-semibold tracking-widest text-slate-500 uppercase">
                Balance
              </h2>
              <p className={`mt-1 text-2xl font-bold tabular-nums ${balanceTone(detail.summary.balance)}`}>
                {formatMoney(detail.summary.balance)}
              </p>
              {detail.summary.memberBalances.length > 0 && (
                <ul className="mt-3 space-y-1">
                  {detail.summary.memberBalances.map((b) => (
                    <li
                      key={b.budgetId}
                      className="flex items-center justify-between gap-2 text-sm"
                    >
                      <span className="truncate text-slate-300">{b.displayName}</span>
                      <span className={`shrink-0 font-medium tabular-nums ${balanceTone(b.balance)}`}>
                        {formatMoney(b.balance)}
                      </span>
                    </li>
                  ))}
                </ul>
              )}
            </section>

            <section className="mt-10">
              <h2 className="text-l font-semibold tracking-widest text-slate-500 uppercase">
                Transactions
              </h2>
              <ShareTransactionTable 
                  budgetId={budget} 
                  transactions={detail.transactions} 
                  onAdd={(trxId) => setTarget({ trxId: trxId, dialog: { mode: 'create' } })}
                  onEdit={(trxId, line) => setTarget({ trxId: trxId, dialog: { mode: 'edit', line } })}
              />
            </section>
          </>
        )}
      </div>

      {rename && (
        <ShareRenameDialog
          budgetId={budget}
          shareId={share}
          dialog={rename}
          onClose={() => setRename(null)}
        />
      )}

      {code && <ShareCodeDialog code={code} onClose={() => setCode(null)} />}

      {leaveOpen && membership && (
        <ShareLeaveDialog
          budgetId={budget}
          shareId={share}
          name={membership.name}
          onClose={() => setLeaveOpen(false)}
          onLeft={() => navigate(`/budget/${budget}`)}
        />
      )}

      {target && (
        <CategorizationDialog
          key={`${target.trxId}-${target.dialog.mode === 'edit' ? target.dialog.line.id : 'new'}`}
          budgetId={budget}
          shareId={share}
          trxId={target.trxId}
          categories={categoriesQuery.data ?? []}
          dialog={target.dialog}
          onClose={() => setTarget(null)}
        />
      )}
    </>
  )
}
