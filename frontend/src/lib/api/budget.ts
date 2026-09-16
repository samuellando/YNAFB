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

export type AccountDetail = components['schemas']['accountDetail']
export type AccountTransaction = components['schemas']['accountTransaction']
export type AccountTransactionLine = components['schemas']['accountTransactionLine']

// ---- Payees ----

export type Payee = components['schemas']['payee']

export async function listPayees(budgetId: number): Promise<Payee[]> {
  const { data, error, response } = await client.GET('/budget/{budgetId}/payee', {
    params: { path: { budgetId: String(budgetId) } },
  })
  if (!response.ok) {
    throw apiError(error, 'Failed to load payees', response.status)
  }
  return data
}

export async function createPayee(budgetId: number, name: string): Promise<Payee> {
  const { data, error, response } = await client.POST('/budget/{budgetId}/payee', {
    params: { path: { budgetId: String(budgetId) } },
    body: { name },
  })
  if (!response.ok) {
    throw apiError(error, 'Failed to create payee', response.status)
  }
  return data
}

function apiError(error: unknown, fallback: string, status: number): ApiError {
  const detail = typeof error === 'string' ? error.trim() : ''
  return new ApiError(detail !== '' ? detail : fallback, status)
}

export async function reconcileAccount(
  budgetId: number,
  accountId: number,
  input: { date: string; balance: number },
): Promise<void> {
  const { error, response } = await client.POST('/budget/{budgetId}/account/{id}/reconcile', {
    params: { path: { budgetId: String(budgetId), id: String(accountId) } },
    body: { date: input.date, balance: input.balance },
    parseAs: 'text',
  })
  if (!response.ok) {
    throw apiError(error, 'Failed to reconcile account', response.status)
  }
}

export async function importStatement(
  budgetId: number,
  accountId: number,
  file: File,
): Promise<number> {
  const form = new FormData()
  form.append('statement', file)
  // openapi-fetch types multipart bodies as their schema shape; the runtime
  // body must be FormData.
  const { data, error, response } = await client.POST(
    '/budget/{budgetId}/account/{id}/import',
    {
      params: { path: { budgetId: String(budgetId), id: String(accountId) } },
      body: form as unknown as { statement: string },
      parseAs: 'text',
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to import statement', response.status)
  }
  const count = Number(data)
  if (!Number.isInteger(count) || count < 0) {
    throw new ApiError('Failed to import statement', response.status)
  }
  return count
}

export type PayeeDefaultLine = components['schemas']['payeeDefaultLine']

export async function listPayeeDefaultLines(
  budgetId: number,
  payeeId: number,
): Promise<PayeeDefaultLine[]> {
  const { data, error, response } = await client.GET(
    '/budget/{budgetId}/payee/{payeeId}/default-line',
    {
      params: { path: { budgetId: String(budgetId), payeeId: String(payeeId) } },
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to load payee defaults', response.status)
  }
  return data
}

export type PayeeDefaultLineInput = {
  destAccountId?: number
  categoryId?: number
  income: boolean
  percent: number
}

export async function createPayeeDefaultLine(
  budgetId: number,
  payeeId: number,
  input: PayeeDefaultLineInput,
): Promise<void> {
  const { error, response } = await client.POST(
    '/budget/{budgetId}/payee/{payeeId}/default-line',
    {
      params: { path: { budgetId: String(budgetId), payeeId: String(payeeId) } },
      body: {
        ...(input.destAccountId !== undefined ? { destAccountId: input.destAccountId } : {}),
        ...(input.categoryId !== undefined ? { categoryId: input.categoryId } : {}),
        income: input.income,
        percent: input.percent,
      },
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to save payee default', response.status)
  }
}

export async function deletePayeeDefaultLine(
  budgetId: number,
  payeeId: number,
  id: number,
): Promise<void> {
  const { error, response } = await client.DELETE(
    '/budget/{budgetId}/payee/{payeeId}/default-line/{id}',
    {
      params: { path: { budgetId: String(budgetId), payeeId: String(payeeId), id: String(id) } },
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to save payee default', response.status)
  }
}

export type TransactionInput = {
  payeeId: number
  date: string
  outflow: number
  inflow: number
  note: string
}

export type Transaction = components['schemas']['transaction']

// ---- Transactions ----

export async function createTransaction(
  budgetId: number,
  accountId: number,
  input: TransactionInput,
): Promise<Transaction> {
  const { data, error, response } = await client.POST(
    '/budget/{budgetId}/account/{accountId}/transaction',
    {
      params: {
        path: { budgetId: String(budgetId), accountId: String(accountId) },
      },
      body: {
        payeeId: input.payeeId,
        date: input.date,
        outflow: input.outflow,
        inflow: input.inflow,
        note: input.note,
      },
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to create transaction', response.status)
  }
  return data
}

export async function updateTransaction(
  budgetId: number,
  accountId: number,
  trxId: number,
  input: TransactionInput,
): Promise<void> {
  const { error, response } = await client.PUT(
    '/budget/{budgetId}/account/{accountId}/transaction/{id}',
    {
      params: {
        path: { budgetId: String(budgetId), accountId: String(accountId), id: String(trxId) },
      },
      body: {
        payeeId: input.payeeId,
        date: input.date,
        outflow: input.outflow,
        inflow: input.inflow,
        note: input.note,
      },
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to update transaction', response.status)
  }
}

export async function deleteTransaction(
  budgetId: number,
  accountId: number,
  trxId: number,
): Promise<void> {
  const { error, response } = await client.DELETE(
    '/budget/{budgetId}/account/{accountId}/transaction/{id}',
    {
      params: {
        path: { budgetId: String(budgetId), accountId: String(accountId), id: String(trxId) },
      },
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to delete transaction', response.status)
  }
}

export type TransactionLineInput = {
  destAccountId?: number
  categoryId?: number
  income: boolean
  outflow: number
  inflow: number
}

function lineBody(input: TransactionLineInput) {
  return {
    ...(input.destAccountId !== undefined ? { destAccountId: input.destAccountId } : {}),
    ...(input.categoryId !== undefined ? { categoryId: input.categoryId } : {}),
    income: input.income,
    outflow: input.outflow,
    inflow: input.inflow,
  }
}

export async function createTransactionLine(
  budgetId: number,
  accountId: number,
  trxId: number,
  input: TransactionLineInput,
): Promise<void> {
  const { error, response } = await client.POST(
    '/budget/{budgetId}/account/{accountId}/transaction/{trxId}/line',
    {
      params: {
        path: {
          budgetId: String(budgetId),
          accountId: String(accountId),
          trxId: String(trxId),
        },
      },
      body: lineBody(input),
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to create transaction line', response.status)
  }
}

export async function updateTransactionLine(
  budgetId: number,
  accountId: number,
  trxId: number,
  lineId: number,
  input: TransactionLineInput,
): Promise<void> {
  const { error, response } = await client.PUT(
    '/budget/{budgetId}/account/{accountId}/transaction/{trxId}/line/{id}',
    {
      params: {
        path: {
          budgetId: String(budgetId),
          accountId: String(accountId),
          trxId: String(trxId),
          id: String(lineId),
        },
      },
      body: lineBody(input),
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to update transaction line', response.status)
  }
}

export async function deleteTransactionLine(
  budgetId: number,
  accountId: number,
  trxId: number,
  lineId: number,
): Promise<void> {
  const { error, response } = await client.DELETE(
    '/budget/{budgetId}/account/{accountId}/transaction/{trxId}/line/{id}',
    {
      params: {
        path: {
          budgetId: String(budgetId),
          accountId: String(accountId),
          trxId: String(trxId),
          id: String(lineId),
        },
      },
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to delete transaction line', response.status)
  }
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

// ---- Accounts ----

export async function getAccountDetail(
  budgetId: number,
  accountId: number,
): Promise<AccountDetail> {
  const { data, error, response } = await client.GET('/budget/{budgetId}/account/{id}', {
    params: { path: { budgetId: String(budgetId), id: String(accountId) } },
  })
  if (!response.ok) {
    throw apiError(error, 'Failed to load account', response.status)
  }
  return data
}

export type Account = components['schemas']['account']

export async function createAccount(budgetId: number, name: string): Promise<Account> {
  const { data, error, response } = await client.POST('/budget/{budgetId}/account', {
    params: { path: { budgetId: String(budgetId) } },
    body: { name },
  })
  if (!response.ok) {
    throw apiError(error, 'Failed to create account', response.status)
  }
  return data
}

export async function updateAccount(
  budgetId: number,
  accountId: number,
  name: string,
): Promise<Account> {
  const { data, error, response } = await client.PUT('/budget/{budgetId}/account/{id}', {
    params: { path: { budgetId: String(budgetId), id: String(accountId) } },
    body: { name },
  })
  if (!response.ok) {
    throw apiError(error, 'Failed to update account', response.status)
  }
  return data
}

export async function deleteAccount(budgetId: number, accountId: number): Promise<void> {
  const { error, response } = await client.DELETE('/budget/{budgetId}/account/{id}', {
    params: { path: { budgetId: String(budgetId), id: String(accountId) } },
  })
  if (!response.ok) {
    throw apiError(error, 'Failed to delete account', response.status)
  }
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

export type Category = components['schemas']['category']

export async function createCategory(
  budgetId: number,
  name: string,
  categoryGroupId?: number,
): Promise<Category> {
  const { data, response } = await client.POST('/budget/{budgetId}/category', {
    params: { path: { budgetId: String(budgetId) } },
    body: { name, ...(categoryGroupId !== undefined ? { categoryGroupId } : {}) },
  })
  if (!response.ok) {
    throw new ApiError('Failed to create category', response.status)
  }
  return data
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

export async function deleteCategory(budgetId: number, categoryId: number): Promise<void> {
  const { response } = await client.DELETE('/budget/{budgetId}/category/{id}', {
    params: { path: { budgetId: String(budgetId), id: String(categoryId) } },
  })
  if (!response.ok) {
    throw new ApiError('Failed to delete category', response.status)
  }
}

export async function deleteCategoryGroup(budgetId: number, groupId: number): Promise<void> {
  const { response } = await client.DELETE('/budget/{budgetId}/category-group/{id}', {
    params: { path: { budgetId: String(budgetId), id: String(groupId) } },
  })
  if (!response.ok) {
    throw new ApiError('Failed to delete category group', response.status)
  }
}

// ---- Expense shares ----

export type BudgetExpenseShare = components['schemas']['BudgetExpenseShare']
export type ExpenseShare = components['schemas']['ExpenseShare']

export async function listExpenseShares(budgetId: number): Promise<BudgetExpenseShare[]> {
  const { data, error, response } = await client.GET('/budget/{budgetId}/expense-share', {
    params: { path: { budgetId: String(budgetId) } },
  })
  if (!response.ok) {
    throw apiError(error, 'Failed to load expense shares', response.status)
  }
  return data
}

export async function createExpenseShare(budgetId: number, defaultName: string): Promise<ExpenseShare> {
  const { data, error, response } = await client.POST('/budget/{budgetId}/expense-share', {
    params: { path: { budgetId: String(budgetId) } },
    body: { defaultName },
  })
  if (!response.ok) {
    throw apiError(error, 'Failed to create expense share', response.status)
  }
  return data
}

export async function joinExpenseShare(budgetId: number, code: string): Promise<ExpenseShare> {
  const { data, error, response } = await client.POST('/budget/{budgetId}/expense-share', {
    params: { path: { budgetId: String(budgetId) } },
    body: { code },
  })
  if (!response.ok) {
    throw apiError(error, 'Failed to join expense share', response.status)
  }
  return data
}

export async function updateExpenseShare(
  budgetId: number,
  expenseShareId: number,
  name: string,
): Promise<BudgetExpenseShare> {
  const { data, error, response } = await client.PUT(
    '/budget/{budgetId}/expense-share/{expenseShareId}',
    {
      params: { path: { budgetId: String(budgetId), expenseShareId: String(expenseShareId) } },
      body: { name },
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to update expense share', response.status)
  }
  return data
}

export async function leaveExpenseShare(budgetId: number, expenseShareId: number): Promise<void> {
  const { error, response } = await client.DELETE(
    '/budget/{budgetId}/expense-share/{expenseShareId}',
    {
      params: { path: { budgetId: String(budgetId), expenseShareId: String(expenseShareId) } },
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to leave expense share', response.status)
  }
}

export async function getExpenseShareCode(budgetId: number, expenseShareId: number): Promise<string> {
  const { data, error, response } = await client.GET(
    '/budget/{budgetId}/expense-share/{expenseShareId}/code',
    {
      params: { path: { budgetId: String(budgetId), expenseShareId: String(expenseShareId) } },
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to get expense share code', response.status)
  }
  return data.code
}
