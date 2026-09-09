import { Link } from 'react-router'

export default function Budget() {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-8 bg-slate-950 px-6 text-center text-slate-100">
      <header>
        <p className="text-sm font-medium tracking-widest text-emerald-400 uppercase">
          You Need A F** Budget
        </p>
        <h1 className="mt-4 text-3xl font-bold tracking-tight">YNAFB</h1>
      </header>
      <p className="max-w-md text-slate-400">
        You're signed in. Your budgets will live here soon.
      </p>
    </div>
  )
}
