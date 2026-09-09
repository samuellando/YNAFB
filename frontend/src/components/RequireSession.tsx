import { Navigate, Outlet } from 'react-router'
import { useSession } from '../lib/useSession'
import LoadingScreen from './ui/LoadingScreen'

export default function RequireSession() {
  const { data: authed } = useSession()

  if (authed === undefined) {
    return <LoadingScreen />
  }

  if (!authed) {
    return <Navigate to="/" replace />
  }

  return <Outlet />
}
