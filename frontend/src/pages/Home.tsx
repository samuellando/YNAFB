import { Link, Navigate } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { hasSession } from '../lib/api/auth'

export default function Home() {
  const { data: authed } = useQuery({
    queryKey: ['session'],
    queryFn: hasSession,
    retry: false,
    refetchOnWindowFocus: false,
  })

  if (authed === undefined) {
    return <div className="min-h-svh bg-slate-950" />
  }

  if (authed) {
    return <Navigate to="/budget" replace />
  }

  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-8 bg-slate-950 px-6 text-slate-100">
      <header className="text-center">
        <p className="text-sm font-medium tracking-widest text-emerald-400 uppercase">
          You Need A F** Budget
        </p>
        <h1 className="mt-4 text-5xl font-bold tracking-tight">YNAFB</h1>
        <nav className="mt-8 flex justify-center gap-4">
          <Link
            to="/login"
            className="rounded-lg border border-emerald-500 px-5 py-2 font-semibold text-emerald-400 transition hover:bg-emerald-500 hover:text-slate-950"
          >
            Log in
          </Link>
          <Link
            to="/signup"
            className="rounded-lg bg-emerald-500 px-5 py-2 font-semibold text-slate-950 transition hover:bg-emerald-400"
          >
            Sign up
          </Link>
        </nav>
      </header>

    </div>
  )
}
