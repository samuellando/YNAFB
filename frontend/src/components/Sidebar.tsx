import { useRef, useState, type FormEvent } from 'react'
import { NavLink, useLocation, useNavigate } from 'react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createAccount,
  createBudget,
  listAccounts,
  listBudgets,
  listExpenseShares,
  type Budget,
} from '../lib/api/budget'
import { deauthenticate } from '../lib/api/auth'
import { saveSelectedBudget } from '../lib/budgetSelection'
import { formatMoney } from '../lib/money'
import { useClickOutside } from '../lib/useClickOutside'
import { ApiError } from '../lib/api/errors'
import { CheckIcon, ChevronDownIcon, LogOutIcon, PlusIcon } from './icons'
import ExpenseShareDialog, { type ExpenseShareDialogState } from './ExpenseShareDialog'

type SidebarProps = {
  budgetId: number
}

export default function Sidebar({ budgetId }: SidebarProps) {
  const navigate = useNavigate()
  const location = useLocation()
  const queryClient = useQueryClient()
  const [budgetMenuOpen, setBudgetMenuOpen] = useState(false)
  const [newBudgetOpen, setNewBudgetOpen] = useState(false)
  const [name, setName] = useState('')
  const [newAccountOpen, setNewAccountOpen] = useState(false)
  const [accountName, setAccountName] = useState('')
  const [expenseShareDialog, setExpenseShareDialog] = useState<ExpenseShareDialogState | null>(null)
  const budgetMenuRef = useRef<HTMLDivElement>(null)

  useClickOutside(
    budgetMenuRef,
    () => {
      setBudgetMenuOpen(false)
      setNewBudgetOpen(false)
    },
    budgetMenuOpen,
  )

  const budgets = useQuery({ queryKey: ['budgets'], queryFn: listBudgets })
  const accounts = useQuery({
    queryKey: ['accounts', budgetId],
    queryFn: () => listAccounts(budgetId),
  })
  const expenseShares = useQuery({
    queryKey: ['expense-shares', budgetId],
    queryFn: () => listExpenseShares(budgetId),
  })

  const activeBudget = budgets.data?.find((b) => b.id === budgetId)

  const create = useMutation({
    mutationFn: createBudget,
    onSuccess: async (budget) => {
      await queryClient.invalidateQueries({ queryKey: ['budgets'] })
      setBudgetMenuOpen(false)
      setNewBudgetOpen(false)
      setName('')
      saveSelectedBudget(budget.id)
      navigate(`/budget/${budget.id}`)
    },
  })

  const addAccount = useMutation({
    mutationFn: (newAccountName: string) => createAccount(budgetId, newAccountName),
    onSuccess: async (account) => {
      await queryClient.invalidateQueries({ queryKey: ['accounts', budgetId] })
      setNewAccountOpen(false)
      setAccountName('')
      navigate(`/budget/${budgetId}/account/${account.id}`)
    },
  })

  const signOut = useMutation({
    mutationFn: deauthenticate,
    onSettled: () => {
      queryClient.clear()
      navigate('/')
    },
  })

  function selectBudget(budget: Budget) {
    saveSelectedBudget(budget.id)
    setBudgetMenuOpen(false)
    setNewBudgetOpen(false)
    navigate(`/budget/${budget.id}`)
  }

  function submitNewBudget(e: FormEvent) {
    e.preventDefault()
    if (name.trim()) {
      create.mutate(name.trim())
    }
  }

  function submitNewAccount(e: FormEvent) {
    e.preventDefault()
    if (accountName.trim()) {
      addAccount.mutate(accountName.trim())
    }
  }

  const accountItems = accounts.data ?? []
  const reconciled = new Set(accountItems.filter((a) => a.balance === a.reconciledBalance).map((a) => a.id))
  const expenseShareItems = expenseShares.data ?? []

  return (
    <aside className="flex h-full w-72 shrink-0 flex-col border-r border-slate-800 bg-slate-900">
      <div className="border-b border-slate-800 p-3">
        <h2 className="px-3 pb-2 text-xs font-semibold tracking-widest text-slate-500 uppercase">
          Budget
        </h2>
        <div className="relative" ref={budgetMenuRef}>
          <button
            type="button"
            onClick={() => setBudgetMenuOpen((open) => !open)}
            className="flex w-full items-center justify-between gap-2 rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-left text-sm font-medium text-slate-100 hover:border-emerald-400"
          >
            <span className="truncate">{activeBudget?.name ?? 'Select budget'}</span>
            <ChevronDownIcon className="h-4 w-4 shrink-0 text-slate-400" />
          </button>

          {budgetMenuOpen && (
            <div className="absolute inset-x-0 top-full z-20 mt-1 overflow-hidden rounded-lg border border-slate-700 bg-slate-900 shadow-xl">
            <ul className="max-h-64 overflow-y-auto py-1">
              {budgets.data?.map((budget) => (
                <li key={budget.id}>
                  <button
                    type="button"
                    onClick={() => selectBudget(budget)}
                    className={`w-full truncate px-3 py-2 text-left text-sm hover:bg-slate-800 ${
                      budget.id === budgetId ? 'font-semibold text-emerald-400' : 'text-slate-200'
                    }`}
                  >
                    {budget.name}
                  </button>
                </li>
              ))}
              {budgets.data?.length === 0 && (
                <li className="px-3 py-2 text-sm text-slate-500">No budgets yet</li>
              )}
            </ul>
            <div className="border-t border-slate-700 py-1">
              {newBudgetOpen ? (
                <form onSubmit={submitNewBudget} className="flex gap-2 p-2">
                  <input
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="Budget name"
                    autoFocus
                    className="min-w-0 flex-1 rounded border border-slate-700 bg-slate-950 px-2 py-1 text-sm text-slate-100 outline-none focus:border-emerald-400"
                  />
                  <button
                    type="submit"
                    disabled={create.isPending}
                    className="rounded bg-emerald-500 px-3 py-1 text-sm font-semibold text-slate-950 hover:bg-emerald-400 disabled:opacity-50"
                  >
                    {create.isPending ? '…' : 'Add'}
                  </button>
                </form>
              ) : (
                <button
                  type="button"
                  onClick={() => setNewBudgetOpen(true)}
                  className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm font-medium text-emerald-400 hover:bg-slate-800"
                >
                  <PlusIcon className="h-4 w-4" />
                  New budget
                </button>
              )}
            </div>
          </div>
        )}
        </div>
      </div>

      <nav className="border-b border-slate-800 p-3">
        <NavLink
          to={`/budget/${budgetId}`}
          className={({ isActive }) =>
            `flex items-center justify-between gap-2 rounded-lg px-3 py-2 text-sm font-medium transition ${
              isActive && !location.pathname.includes('/account/') && !location.pathname.includes('/expense-share/')
                ? 'bg-slate-800'
                : 'text-slate-200 hover:bg-slate-800/60'
            }`
          }
        >
          Monthly budget
        </NavLink>
      </nav>

      <div className="flex min-h-0 flex-1 flex-col overflow-y-auto p-3">
        <h2 className="px-3 pb-2 text-xs font-semibold tracking-widest text-slate-500 uppercase">
          Accounts
        </h2>
        <ul className="space-y-1">
          {accountItems.map((account) => (
            <li key={account.id}>
              <NavLink
                to={`/budget/${budgetId}/account/${account.id}`}
                className={({ isActive }) =>
                  `flex items-center justify-between gap-2 rounded-lg px-3 py-2 text-sm transition ${
                    isActive ? 'bg-slate-800' : 'hover:bg-slate-800/60'
                  }`
                }
              >
                <span className="flex min-w-0 items-center gap-2">
                  <span className="truncate text-slate-200">{account.name}</span>
                  {reconciled.has(account.id) && (
                    <CheckIcon className="h-4 w-4 shrink-0 text-emerald-400" />
                  )}
                </span>
                <span className="shrink-0 font-medium text-slate-100">
                  {formatMoney(account.balance)}
                </span>
              </NavLink>
            </li>
          ))}
          {!accounts.isPending && accountItems.length === 0 && (
            <li className="px-3 py-2 text-sm text-slate-500">No accounts yet</li>
          )}
          <li>
            {newAccountOpen ? (
              <form onSubmit={submitNewAccount} className="flex gap-2 px-3 py-1">
                <input
                  type="text"
                  value={accountName}
                  onChange={(e) => setAccountName(e.target.value)}
                  placeholder="Account name"
                  autoFocus
                  className="min-w-0 flex-1 rounded border border-slate-700 bg-slate-950 px-2 py-1 text-sm text-slate-100 outline-none focus:border-emerald-400"
                />
                <button
                  type="submit"
                  disabled={accountName.trim() === '' || addAccount.isPending}
                  className="rounded bg-emerald-500 px-3 py-1 text-sm font-semibold text-slate-950 hover:bg-emerald-400 disabled:opacity-50"
                >
                  {addAccount.isPending ? '…' : 'Add'}
                </button>
              </form>
            ) : (
              <button
                type="button"
                onClick={() => setNewAccountOpen(true)}
                className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm font-medium text-emerald-400 transition hover:bg-slate-800/60"
              >
                <PlusIcon className="h-4 w-4" />
                Add account
              </button>
            )}
          </li>
          {addAccount.isError && (
            <li className="px-3 text-sm text-red-400">
              {addAccount.error instanceof ApiError
                ? addAccount.error.message
                : 'Failed to create account'}
            </li>
          )}
        </ul>
        <div className="mx-3 mt-4 border-t border-slate-800" />
        <h2 className="px-3 pt-4 pb-2 text-xs font-semibold tracking-widest text-slate-500 uppercase">
          Expense shares
        </h2>
        <ul className="space-y-1">
          {expenseShareItems.map((share) => (
            <li key={share.id}>
              <NavLink
                to={`/budget/${budgetId}/expense-share/${share.id}`}
                className={({ isActive }) =>
                  `flex items-center justify-between gap-2 rounded-lg px-3 py-2 text-sm transition ${
                    isActive ? 'bg-slate-800' : 'hover:bg-slate-800/60'
                  }`
                }
              >
                <span className="truncate text-slate-200">{share.name}</span>
              </NavLink>
            </li>
          ))}
          {!expenseShares.isPending && expenseShareItems.length === 0 && (
            <li className="px-3 py-2 text-sm text-slate-500">No expense shares yet</li>
          )}
          <li>
            <button
              type="button"
              onClick={() => setExpenseShareDialog({ mode: 'create' })}
              className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm font-medium text-emerald-400 transition hover:bg-slate-800/60"
            >
              <PlusIcon className="h-4 w-4" />
              Add expense share
            </button>
          </li>
          {expenseShares.isError && (
            <li className="px-3 text-sm text-red-400">
              {expenseShares.error instanceof ApiError
                ? expenseShares.error.message
                : 'Failed to load expense shares'}
            </li>
          )}
        </ul>
      </div>

      <div className="border-t border-slate-800 p-3">
        <button
          type="button"
          onClick={() => signOut.mutate()}
          className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium text-slate-300 transition hover:bg-slate-800 hover:text-red-400"
        >
          <LogOutIcon className="h-4 w-4" />
          Sign out
        </button>
      </div>
      {expenseShareDialog && (
        <ExpenseShareDialog
          key={expenseShareDialog.mode}
          budgetId={budgetId}
          dialog={expenseShareDialog}
          onClose={() => setExpenseShareDialog(null)}
          onDone={(shareId) => {
            setExpenseShareDialog(null)
            navigate(`/budget/${budgetId}/expense-share/${shareId}`)
          }}
        />
      )}
    </aside>
  )
}
