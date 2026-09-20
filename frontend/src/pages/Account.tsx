import { useMemo, useState } from 'react'
import { Navigate, useNavigate, useParams, useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { getAccountDetail, type AccountTransaction } from '../lib/api/budget'
import { needsCategorize, uncategorizedAmount } from '../lib/accountView'
import { parseIdParam } from '../lib/params'
import AccountBalance from '../components/account/AccountBalance'
import AccountDialog from '../components/account/AccountDialog'
import ImportDialog from '../components/account/ImportDialog'
import { PencilIcon } from '../components/icons'
import ReconcileDialog from '../components/account/ReconcileDialog'
import TransactionDialog from '../components/account/TransactionDialog'
import TransactionTable from '../components/account/TransactionTable'
import BudgetMonthError from '../components/budget/BudgetMonthError'
import { OUTLINE_BUTTON } from '../components/ui/buttons'

export default function Account() {
  const { budgetId, accountId } = useParams()
  const navigate = useNavigate()
  const budget = parseIdParam(budgetId)
  const id = parseIdParam(accountId)
  const valid = budget !== null && id !== null

  const query = useQuery({
    queryKey: ['account', budget, id],
    queryFn: () => getAccountDetail(budget ?? 0, id ?? 0),
    enabled: valid,
  })
  const transactions = useMemo(() => query.data?.transactions ?? [], [query.data])

  const uncategorized = useMemo(() => uncategorizedAmount(transactions), [transactions])

  const [selected, setSelected] = useState<AccountTransaction | null>(null)
  const [searchParams, setSearchParams] = useSearchParams()
  const deepLinkId = parseIdParam(searchParams.get('trx') ?? undefined)
  const deepLinked =
    deepLinkId === null
      ? null
      : (transactions.find((t) => t.id === deepLinkId) ?? null)
  const shown = selected ?? deepLinked

  function closeSelected() {
    setSelected(null)
    if (deepLinkId !== null) {
      setSearchParams({})
    }
  }
  const [adding, setAdding] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [reconcileOpen, setReconcileOpen] = useState(false)
  const [importOpen, setImportOpen] = useState(false)
  const [queue, setQueue] = useState<number[] | null>(null)
  const [queueTotal, setQueueTotal] = useState(0)

  const needsWork = useMemo(() => transactions.filter(needsCategorize), [transactions])

  function startCategorize() {
    if (needsWork.length === 0) return
    // Walk oldest to newest, even though the table lists newest first.
    const ordered = [...needsWork].sort(
      (a, b) => Date.parse(a.date) - Date.parse(b.date) || a.id - b.id,
    )
    setQueue(ordered.map((t) => t.id))
    setQueueTotal(ordered.length)
  }

  function cancelQueue() {
    setQueue(null)
    setQueueTotal(0)
  }

  function advanceQueue() {
    setQueue((prev) => {
      if (!prev || prev.length <= 1) return null
      return prev.slice(1)
    })
  }

  const flowId = queue?.[0] ?? null
  const flowTransaction =
    flowId === null ? null : (transactions.find((t) => t.id === flowId) ?? null)

  if (!valid) {
    return <Navigate to={budget !== null ? `/budget/${budget}` : '/budget'} replace />
  }

  const summary = query.data?.summary

  return (
    <>
      <div className="mx-auto flex min-h-full max-w-5xl flex-col px-8 pb-6 pt-12">
        <header className="flex items-center justify-between gap-4">
          <div className="flex min-w-0 items-center gap-2">
            <h1 className="truncate text-3xl font-bold tracking-tight">
              {summary?.name ?? 'Account'}
            </h1>
            <button
              type="button"
              onClick={() => setEditOpen(true)}
              disabled={!summary}
              aria-label="Edit account"
              title="Edit account"
              className="shrink-0 cursor-pointer rounded-lg p-1.5 text-slate-500 transition hover:bg-slate-800 hover:text-emerald-400 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent disabled:hover:text-slate-500"
            >
              <PencilIcon className="h-4 w-4" />
            </button>
          </div>
          <div className="flex shrink-0 gap-2">
            <button
              type="button"
              onClick={() => setAdding(true)}
              disabled={!summary}
              className={OUTLINE_BUTTON}
            >
              Add transaction
            </button>
            <button
              type="button"
              onClick={startCategorize}
              disabled={!summary || needsWork.length === 0}
              className={OUTLINE_BUTTON}
            >
              Categorize
            </button>
            <button
              type="button"
              onClick={() => setImportOpen(true)}
              disabled={!summary}
              className={OUTLINE_BUTTON}
            >
              Import
            </button>
            <button
              type="button"
              onClick={() => setReconcileOpen(true)}
              disabled={!summary}
              className={OUTLINE_BUTTON}
            >
              Reconcile
            </button>
          </div>
        </header>

        {query.isPending && <div className="mt-6 min-h-80" />}

        {query.isError && (
          <BudgetMonthError message={query.error.message} onRetry={() => query.refetch()} />
        )}

        {query.isSuccess && summary && (
          <>
            <AccountBalance
              balance={summary.balance}
              reconciledBalance={summary.reconciledBalance}
              uncategorized={uncategorized}
            />
            <TransactionTable
              budgetId={budget}
              transactions={transactions}
              onSelectTransaction={setSelected}
            />
          </>
        )}
      </div>

      {shown && (
        <TransactionDialog
          key={shown.id}
          budgetId={budget}
          accountId={id}
          dialog={{ transaction: shown }}
          onCancel={closeSelected}
          onDone={closeSelected}
        />
      )}

      {adding && (
        <TransactionDialog
          key="new"
          budgetId={budget}
          accountId={id}
          dialog={{ transaction: null }}
          onCancel={() => setAdding(false)}
          onDone={() => setAdding(false)}
        />
      )}

      {editOpen && summary && (
        <AccountDialog
          budgetId={budget}
          accountId={id}
          dialog={{ name: summary.name }}
          onClose={() => setEditOpen(false)}
          onDeleted={() => navigate(`/budget/${budget}`)}
        />
      )}

      {flowTransaction && queue && (
        <TransactionDialog
          key={flowTransaction.id}
          budgetId={budget}
          accountId={id}
          dialog={{ transaction: flowTransaction }}
          progress={{ index: queueTotal - queue.length + 1, total: queueTotal }}
          onCancel={cancelQueue}
          onDone={advanceQueue}
          onSkip={advanceQueue}
        />
      )}

      {reconcileOpen && summary && (
        <ReconcileDialog
          budgetId={budget}
          accountId={id}
          currentBalance={summary.balance}
          onClose={() => setReconcileOpen(false)}
        />
      )}

      {importOpen && (
        <ImportDialog budgetId={budget} accountId={id} onClose={() => setImportOpen(false)} />
      )}
    </>
  )
}
