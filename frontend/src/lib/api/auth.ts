import { client } from './client'

export class AuthError extends Error {}

export async function hasSession(): Promise<boolean> {
  try {
    const { response } = await client.GET('/budget')
    return response.ok
  } catch {
    return false
  }
}

async function postForm(path: string, body: { username: string; password: string }) {
  let res: Response
  try {
    res = await fetch(path, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(body),
    })
  } catch {
    throw new AuthError('Could not reach the server. Please try again.')
  }

  if (res.ok) return

  const text = await res.text()
  if (res.status === 401 && path.endsWith('/authenticate')) {
    throw new AuthError('Invalid username or password.')
  }
  throw new AuthError(text || 'Something went wrong. Please try again.')
}

export function authenticate(username: string, password: string) {
  return postForm('/auth/authenticate', { username, password })
}

export function signup(username: string, password: string) {
  return postForm('/auth/signup', { username, password })
}

export async function deauthenticate(): Promise<void> {
  try {
    await fetch('/auth/deauthenticate', { method: 'POST', credentials: 'include' })
  } catch {
    throw new AuthError('Could not reach the server. Please try again.')
  }
}
