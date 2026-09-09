import { Navigate, Outlet } from 'react-router'
import { useSession } from '../lib/useSession'

export default function RequireSession() {
  const { data: authed } = useSession()

  if (authed === undefined) {
    return <div className="min-h-svh bg-slate-950" />
  }

  if (!authed) {
    return <Navigate to="/" replace />
  }

  return <Outlet />
}
