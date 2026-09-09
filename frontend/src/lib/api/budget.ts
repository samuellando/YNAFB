import { client } from './client'
import { ApiError } from './errors'

export type Budget = { id: number; name: string }

export type AccountSummary = {
  id: number
  name: string
  balance: number
  reconciledBalance: number
}

export async function listBudgets(): Promise<Budget[]> {
  const { data, response } = await client.GET('/budget')
  if (!response.ok) {
    throw new ApiError('Failed to load budgets', response.status)
  }
  return data
}

export async function listAccounts(budgetId: number): Promise<AccountSummary[]> {
  const { data, response } = await client.GET('/budget/{budgetId}/account', {
    params: { path: { budgetId: String(budgetId) } },
  })
  if (!response.ok) {
    throw new ApiError('Failed to load accounts', response.status)
  }
  return data
}

export async function createBudget(name: string): Promise<Budget> {
  const { data, response } = await client.POST('/budget', { body: { name } })
  if (!response.ok) {
    throw new ApiError('Failed to create budget', response.status)
  }
  return data
}
