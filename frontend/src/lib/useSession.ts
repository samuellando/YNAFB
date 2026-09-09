import { useQuery } from '@tanstack/react-query'
import { hasSession } from './api/auth'

export function useSession() {
  return useQuery({
    queryKey: ['session'],
    queryFn: hasSession,
    retry: false,
    refetchOnWindowFocus: false,
  })
}
