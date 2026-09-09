import { useEffect } from 'react'
import { Navigate, Outlet, useParams } from 'react-router'
import { saveSelectedBudget } from '../lib/budgetSelection'
import Sidebar from './Sidebar'

export default function AppLayout() {
  const { budgetId } = useParams()
  const id = Number(budgetId)
  const valid = budgetId !== undefined && !Number.isNaN(id)

  useEffect(() => {
    if (valid) saveSelectedBudget(id)
  }, [id, valid])

  if (!valid) {
    return <Navigate to="/budget" replace />
  }

  return (
    <div className="flex h-svh overflow-hidden bg-slate-950 text-slate-100">
      <Sidebar budgetId={id} />
      <main className="min-w-0 flex-1 overflow-y-auto">
        <Outlet />
      </main>
    </div>
  )
}
