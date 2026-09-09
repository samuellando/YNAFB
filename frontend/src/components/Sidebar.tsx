import { useEffect, useRef, useState, type FormEvent } from 'react'
import { NavLink, useNavigate } from 'react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createBudget, listAccounts, listBudgets, type Budget } from '../lib/api/budget'
import { deauthenticate } from '../lib/api/auth'
import { saveSelectedBudget } from '../lib/budgetSelection'
import { formatMoney } from '../lib/money'
import { CheckIcon, ChevronDownIcon, LogOutIcon, PlusIcon } from './icons'

type SidebarProps = {
  budgetId: number
}

export default function Sidebar({ budgetId }: SidebarProps) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [budgetMenuOpen, setBudgetMenuOpen] = useState(false)
  const [newBudgetOpen, setNewBudgetOpen] = useState(false)
  const [name, setName] = useState('')
  const budgetMenuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!budgetMenuOpen) return
    function handleClickOutside(event: MouseEvent) {
      if (budgetMenuRef.current && !budgetMenuRef.current.contains(event.target as Node)) {
        setBudgetMenuOpen(false)
        setNewBudgetOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [budgetMenuOpen])

  const budgets = useQuery({ queryKey: ['budgets'], queryFn: listBudgets })
  const accounts = useQuery({
    queryKey: ['accounts', budgetId],
    queryFn: () => listAccounts(budgetId),
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

  const accountItems = accounts.data ?? []
  const reconciled = new Set(accountItems.filter((a) => a.balance === a.reconciledBalance).map((a) => a.id))

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
          end
          className={({ isActive }) =>
            `flex items-center justify-between gap-2 rounded-lg px-3 py-2 text-sm font-medium transition ${
              isActive ? 'bg-slate-800' : 'text-slate-200 hover:bg-slate-800/60'
            }`
          }
        >
          Monthly budget
        </NavLink>
      </nav>

      <div className="flex min-h-0 flex-1 flex-col p-3">
        <h2 className="px-3 pb-2 text-xs font-semibold tracking-widest text-slate-500 uppercase">
          Accounts
        </h2>
        <ul className="flex-1 space-y-1 overflow-y-auto">
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
          {!accounts.isLoading && accountItems.length === 0 && (
            <li className="px-3 py-2 text-sm text-slate-500">No accounts yet</li>
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
    </aside>
  )
}
