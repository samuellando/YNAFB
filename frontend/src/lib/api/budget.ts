import type { components } from './schema'
import { client } from './client'
import { ApiError } from './errors'

export type Budget = { id: number; name: string }

export type BudgetMonth = components['schemas']['budgetMonth']
export type BudgetMonthCategory = components['schemas']['budgetMonthCategory']
export type BudgetMonthSummary = components['schemas']['budgetMonthSummary']
export type BudgetMonthGoal = components['schemas']['budgetMonthGoal']
export type Goal = components['schemas']['goal']
export type GoalType = Goal['type']

export type GoalInput = {
  type: GoalType
  startMonth: string
  endMonth?: string
  amount: number
}

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

export async function getBudgetMonth(budgetId: number, month: string): Promise<BudgetMonth> {
  const { data, response } = await client.GET('/budget/{budgetId}/{month}', {
    params: { path: { budgetId: String(budgetId), month } },
  })
  if (!response.ok) {
    throw new ApiError('Failed to load budget month', response.status)
  }
  return data
}

export async function setAllocation(
  budgetId: number,
  categoryId: number,
  month: string,
  amount: number,
): Promise<void> {
  const { response } = await client.PUT('/budget/{budgetId}/category/{id}/allocation/{month}', {
    params: {
      path: { budgetId: String(budgetId), id: String(categoryId), month },
    },
    body: { amount },
  })
  if (!response.ok) {
    throw new ApiError('Failed to update allocation', response.status)
  }
}

export type CategoryGroup = components['schemas']['categoryGroup']

export async function listCategoryGroups(budgetId: number): Promise<CategoryGroup[]> {
  const { data, response } = await client.GET('/budget/{budgetId}/category-group', {
    params: { path: { budgetId: String(budgetId) } },
  })
  if (!response.ok) {
    throw new ApiError('Failed to load category groups', response.status)
  }
  return data
}

export async function createCategoryGroup(budgetId: number, name: string): Promise<CategoryGroup> {
  const { data, response } = await client.POST('/budget/{budgetId}/category-group', {
    params: { path: { budgetId: String(budgetId) } },
    body: { name },
  })
  if (!response.ok) {
    throw new ApiError('Failed to create category group', response.status)
  }
  return data
}

export async function updateCategoryGroup(
  budgetId: number,
  groupId: number,
  name: string,
): Promise<void> {
  const { response } = await client.PUT('/budget/{budgetId}/category-group/{id}', {
    params: { path: { budgetId: String(budgetId), id: String(groupId) } },
    body: { name },
  })
  if (!response.ok) {
    throw new ApiError('Failed to update category group', response.status)
  }
}

export async function createCategory(
  budgetId: number,
  name: string,
  categoryGroupId?: number,
): Promise<void> {
  const { response } = await client.POST('/budget/{budgetId}/category', {
    params: { path: { budgetId: String(budgetId) } },
    body: { name, ...(categoryGroupId !== undefined ? { categoryGroupId } : {}) },
  })
  if (!response.ok) {
    throw new ApiError('Failed to create category', response.status)
  }
}

export async function getGoal(budgetId: number, categoryId: number): Promise<Goal> {
  const { data, response } = await client.GET('/budget/{budgetId}/category/{categoryId}/goal', {
    params: { path: { budgetId: String(budgetId), categoryId: String(categoryId) } },
  })
  if (!response.ok) {
    throw new ApiError('Failed to load goal', response.status)
  }
  return data
}

export async function createGoal(
  budgetId: number,
  categoryId: number,
  input: GoalInput,
): Promise<void> {
  const { response } = await client.POST('/budget/{budgetId}/category/{categoryId}/goal', {
    params: { path: { budgetId: String(budgetId), categoryId: String(categoryId) } },
    body: {
      type: input.type,
      startMonth: input.startMonth,
      ...(input.endMonth !== undefined ? { endMonth: input.endMonth } : {}),
      amount: input.amount,
    },
  })
  if (!response.ok) {
    throw new ApiError('Failed to create goal', response.status)
  }
}

export async function updateGoal(
  budgetId: number,
  categoryId: number,
  input: GoalInput,
): Promise<void> {
  const { response } = await client.PUT('/budget/{budgetId}/category/{categoryId}/goal', {
    params: { path: { budgetId: String(budgetId), categoryId: String(categoryId) } },
    body: {
      type: input.type,
      startMonth: input.startMonth,
      ...(input.endMonth !== undefined ? { endMonth: input.endMonth } : {}),
      amount: input.amount,
    },
  })
  if (!response.ok) {
    throw new ApiError('Failed to update goal', response.status)
  }
}

export async function deleteGoal(budgetId: number, categoryId: number): Promise<void> {
  const { response } = await client.DELETE('/budget/{budgetId}/category/{categoryId}/goal', {
    params: { path: { budgetId: String(budgetId), categoryId: String(categoryId) } },
  })
  if (!response.ok) {
    throw new ApiError('Failed to delete goal', response.status)
  }
}

export async function updateCategory(
  budgetId: number,
  categoryId: number,
  name: string,
  categoryGroupId?: number,
): Promise<void> {
  const { response } = await client.PUT('/budget/{budgetId}/category/{id}', {
    params: { path: { budgetId: String(budgetId), id: String(categoryId) } },
    body: { name, ...(categoryGroupId !== undefined ? { categoryGroupId } : {}) },
  })
  if (!response.ok) {
    throw new ApiError('Failed to update category', response.status)
  }
}
