import { useMemo, useState } from 'react'
import { Navigate, useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { getAccountDetail, type AccountTransaction } from '../lib/api/budget'
import { needsCategorize, uncategorizedAmount } from '../lib/accountView'
import AccountBalance from '../components/account/AccountBalance'
import ImportDialog from '../components/account/ImportDialog'
import ReconcileDialog from '../components/account/ReconcileDialog'
import TransactionDialog from '../components/account/TransactionDialog'
import TransactionTable from '../components/account/TransactionTable'
import BudgetMonthError from '../components/budget/BudgetMonthError'

const HEADER_BUTTON =
  'rounded-lg border border-slate-700 px-3 py-1.5 text-sm font-medium text-slate-200 transition hover:border-emerald-400 hover:text-emerald-400 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:border-slate-700 disabled:hover:text-slate-200'

function isIdParam(value: string | undefined): value is string {
  return value !== undefined && /^\d+$/.test(value)
}

export default function Account() {
  const { budgetId, accountId } = useParams()
  const valid = isIdParam(budgetId) && isIdParam(accountId)
  const budget = valid ? Number(budgetId) : NaN
  const id = valid ? Number(accountId) : NaN

  const query = useQuery({
    queryKey: ['account', budget, id],
    queryFn: () => getAccountDetail(budget, id),
    enabled: valid,
  })
  const transactions = useMemo(() => query.data?.transactions ?? [], [query.data])

  const uncategorized = useMemo(() => uncategorizedAmount(transactions), [transactions])

  const [selected, setSelected] = useState<AccountTransaction | null>(null)
  const [reconcileOpen, setReconcileOpen] = useState(false)
  const [importOpen, setImportOpen] = useState(false)
  const [queue, setQueue] = useState<number[] | null>(null)

  const needsWork = useMemo(() => transactions.filter(needsCategorize), [transactions])

  function startCategorize() {
    if (needsWork.length === 0) return
    setQueue(needsWork.map((t) => t.id))
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
    return <Navigate to={isIdParam(budgetId) ? `/budget/${budgetId}` : '/budget'} replace />
  }

  const summary = query.data?.summary

  return (
    <>
      <div className="mx-auto flex min-h-full max-w-5xl flex-col px-8 pb-6 pt-12">
        <header className="flex items-center justify-between gap-4">
          <h1 className="text-3xl font-bold tracking-tight">{summary?.name ?? 'Account'}</h1>
          <div className="flex shrink-0 gap-2">
            <button
              type="button"
              onClick={startCategorize}
              disabled={!summary || needsWork.length === 0}
              className={HEADER_BUTTON}
            >
              Categorize
            </button>
            <button
              type="button"
              onClick={() => setImportOpen(true)}
              disabled={!summary}
              className={HEADER_BUTTON}
            >
              Import
            </button>
            <button
              type="button"
              onClick={() => setReconcileOpen(true)}
              disabled={!summary}
              className={HEADER_BUTTON}
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
            <TransactionTable transactions={transactions} onSelectTransaction={setSelected} />
          </>
        )}
      </div>

      {selected && (
        <TransactionDialog
          key={selected.id}
          budgetId={budget}
          accountId={id}
          dialog={{ transaction: selected }}
          onCancel={() => setSelected(null)}
          onDone={() => setSelected(null)}
        />
      )}

      {flowTransaction && (
        <TransactionDialog
          key={flowTransaction.id}
          budgetId={budget}
          accountId={id}
          dialog={{ transaction: flowTransaction }}
          onCancel={() => setQueue(null)}
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
