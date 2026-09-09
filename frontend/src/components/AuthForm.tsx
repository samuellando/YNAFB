import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router'
import { authenticate, signup, AuthError } from '../lib/api/auth'

type AuthFormProps = {
  mode: 'login' | 'signup'
}

const copy = {
  login: {
    heading: 'Welcome back',
    submit: 'Log in',
    footer: "Don't have an account?",
    action: 'Sign up',
    to: '/signup',
  },
  signup: {
    heading: 'Create your account',
    submit: 'Sign up',
    footer: 'Already have an account?',
    action: 'Log in',
    to: '/login',
  },
}

export default function AuthForm({ mode }: AuthFormProps) {
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const text = copy[mode]

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      if (mode === 'login') {
        await authenticate(username, password)
      } else {
        await signup(username, password)
      }
      navigate('/budget')
    } catch (err) {
      setError(err instanceof AuthError ? err.message : 'Something went wrong. Please try again.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="w-full max-w-sm rounded-2xl border border-slate-800 bg-slate-900 p-8 shadow-xl">
      <h2 className="text-2xl font-bold tracking-tight">{text.heading}</h2>
      <form onSubmit={handleSubmit} className="mt-6 flex flex-col gap-4">
        <label className="flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          Username
          <input
            type="text"
            name="username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            autoComplete="username"
            required
            minLength={3}
            className="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-slate-100 outline-none focus:border-emerald-400"
          />
        </label>
        <label className="flex flex-col gap-1.5 text-sm font-medium text-slate-300">
          Password
          <input
            type="password"
            name="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
            required
            minLength={mode === 'signup' ? 8 : undefined}
            className="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-slate-100 outline-none focus:border-emerald-400"
          />
        </label>
        {mode === 'signup' && (
          <p className="text-xs text-slate-500">
            Username must be at least 3 characters; password at least 8.
          </p>
        )}
        {error && <p className="text-sm text-red-400">{error}</p>}
        <button
          type="submit"
          disabled={submitting}
          className="mt-2 rounded-lg bg-emerald-500 px-4 py-2 font-semibold text-slate-950 transition hover:bg-emerald-400 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {submitting ? 'Please wait…' : text.submit}
        </button>
      </form>
      <p className="mt-6 text-center text-sm text-slate-400">
        {text.footer}{' '}
        <Link to={text.to} className="font-medium text-emerald-400 hover:underline">
          {text.action}
        </Link>
      </p>
    </div>
  )
}
