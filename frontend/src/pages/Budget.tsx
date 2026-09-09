import { useState, type FormEvent } from 'react'
import { Navigate, useNavigate } from 'react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createBudget, listBudgets } from '../lib/api/budget'
import { loadSelectedBudget, saveSelectedBudget } from '../lib/budgetSelection'

export default function Budget() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [name, setName] = useState('')

  const budgets = useQuery({ queryKey: ['budgets'], queryFn: listBudgets, retry: false })

  const create = useMutation({
    mutationFn: createBudget,
    onSuccess: async (budget) => {
      await queryClient.invalidateQueries({ queryKey: ['budgets'] })
      saveSelectedBudget(budget.id)
      navigate(`/budget/${budget.id}`)
    },
  })

  if (budgets.isLoading) {
    return <div className="min-h-svh bg-slate-950" />
  }

  if (budgets.data && budgets.data.length > 0) {
    const stored = loadSelectedBudget()
    const selected =
      budgets.data.find((b) => b.id === stored)?.id ?? budgets.data[0].id
    if (selected !== stored) {
      saveSelectedBudget(selected)
    }
    return <Navigate to={`/budget/${selected}`} replace />
  }

  function handleCreate(e: FormEvent) {
    e.preventDefault()
    if (name.trim()) {
      create.mutate(name.trim())
    }
  }

  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-8 bg-slate-950 px-6 text-slate-100">
      <header className="text-center">
        <p className="text-sm font-medium tracking-widest text-emerald-400 uppercase">
          You Need A F** Budget
        </p>
        <h1 className="mt-4 text-3xl font-bold tracking-tight">YNAFB</h1>
      </header>
      <form onSubmit={handleCreate} className="w-full max-w-sm rounded-2xl border border-slate-800 bg-slate-900 p-8 shadow-xl">
        <h2 className="text-xl font-bold tracking-tight">Create your first budget</h2>
        <label className="mt-6 flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          Budget name
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            autoFocus
            className="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-slate-100 outline-none focus:border-emerald-400"
          />
        </label>
        {create.isError && <p className="mt-3 text-sm text-red-400">{create.error.message}</p>}
        <button
          type="submit"
          disabled={create.isPending}
          className="mt-6 w-full rounded-lg bg-emerald-500 px-4 py-2 font-semibold text-slate-950 transition hover:bg-emerald-400 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {create.isPending ? 'Creating…' : 'Create budget'}
        </button>
      </form>
    </div>
  )
}
