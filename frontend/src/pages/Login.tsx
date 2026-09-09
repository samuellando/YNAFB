import AuthForm from '../components/AuthForm'

export default function Login() {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-8 bg-slate-950 px-6 text-slate-100">
      <header className="text-center">
        <p className="text-sm font-medium tracking-widest text-emerald-400 uppercase">
          You Need A F** Budget
        </p>
        <h1 className="mt-4 text-3xl font-bold tracking-tight">YNAFB</h1>
      </header>
      <AuthForm mode="login" />
    </div>
  )
}
