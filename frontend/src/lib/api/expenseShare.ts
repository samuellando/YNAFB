import type { components } from './schema'
import { client } from './client'
import { ApiError } from './errors'

export type ExpenseShareMembership = components['schemas']['expenseShareMembership']
export type ExpenseShareDetail = components['schemas']['expenseShareDetail']
export type ExpenseShareSummary = components['schemas']['expenseShareSummary']
export type ExpenseShareMemberBalance =
  components['schemas']['expenseShareMemberBalance']
export type ShareMember = components['schemas']['shareMember']
export type ExpenseShareCode = components['schemas']['expenseShareCode']
export type ExpenseShareTransaction = components['schemas']['expenseShareTransaction']
export type SplitCategorization = components['schemas']['splitCategorization']
export type Category = components['schemas']['category']

export type ShareCreateInput = {
  name: string
  displayName: string
  defaultName?: string
}

export type CategorizationInput = {
  categoryId: number
  outflow: number
  inflow: number
}

function apiError(error: unknown, fallback: string, status: number): ApiError {
  const detail = typeof error === 'string' ? error.trim() : ''
  return new ApiError(detail !== '' ? detail : fallback, status)
}

export async function listShares(budgetId: number): Promise<ExpenseShareMembership[]> {
  const { data, error, response } = await client.GET('/budget/{budgetId}/expense-share', {
    params: { path: { budgetId: String(budgetId) } },
  })
  if (!response.ok) {
    throw apiError(error, 'Failed to load expense shares', response.status)
  }
  return data
}

export async function createShare(
  budgetId: number,
  input: ShareCreateInput,
): Promise<ExpenseShareMembership> {
  const { data, error, response } = await client.POST('/budget/{budgetId}/expense-share', {
    params: { path: { budgetId: String(budgetId) } },
    body: input,
  })
  if (!response.ok) {
    throw apiError(error, 'Failed to create expense share', response.status)
  }
  return data
}

export async function joinShare(
  budgetId: number,
  input: { code: string; name: string; displayName: string },
): Promise<ExpenseShareMembership> {
  const { data, error, response } = await client.POST(
    '/budget/{budgetId}/expense-share/join',
    {
      params: { path: { budgetId: String(budgetId) } },
      body: input,
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to join expense share', response.status)
  }
  return data
}

export async function getShareDetail(
  budgetId: number,
  shareId: number,
): Promise<ExpenseShareDetail> {
  const { data, error, response } = await client.GET(
    '/budget/{budgetId}/expense-share/{shareId}',
    {
      params: { path: { budgetId: String(budgetId), shareId: String(shareId) } },
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to load expense share', response.status)
  }
  return data
}

export async function renameShare(
  budgetId: number,
  shareId: number,
  input: { name?: string; displayName?: string },
): Promise<ExpenseShareMembership> {
  const { data, error, response } = await client.PUT(
    '/budget/{budgetId}/expense-share/{shareId}',
    {
      params: { path: { budgetId: String(budgetId), shareId: String(shareId) } },
      body: input,
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to rename expense share', response.status)
  }
  return data
}

export async function leaveShare(budgetId: number, shareId: number): Promise<void> {
  const { error, response } = await client.DELETE(
    '/budget/{budgetId}/expense-share/{shareId}',
    {
      params: { path: { budgetId: String(budgetId), shareId: String(shareId) } },
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to leave expense share', response.status)
  }
}

export async function mintShareCode(
  budgetId: number,
  shareId: number,
): Promise<ExpenseShareCode> {
  const { data, error, response } = await client.POST(
    '/budget/{budgetId}/expense-share/{shareId}/code',
    {
      params: { path: { budgetId: String(budgetId), shareId: String(shareId) } },
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to mint invite code', response.status)
  }
  return data
}

export async function addCategorization(
  budgetId: number,
  shareId: number,
  trxId: number,
  input: CategorizationInput,
): Promise<SplitCategorization> {
  const { data, error, response } = await client.POST(
    '/budget/{budgetId}/expense-share/{shareId}/transaction/{trxId}/line',
    {
      params: {
        path: {
          budgetId: String(budgetId),
          shareId: String(shareId),
          trxId: String(trxId),
        },
      },
      body: input,
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to add categorization', response.status)
  }
  return data
}

export async function updateCategorization(
  budgetId: number,
  shareId: number,
  trxId: number,
  lineId: number,
  input: CategorizationInput,
): Promise<SplitCategorization> {
  const { data, error, response } = await client.PUT(
    '/budget/{budgetId}/expense-share/{shareId}/transaction/{trxId}/line/{lineId}',
    {
      params: {
        path: {
          budgetId: String(budgetId),
          shareId: String(shareId),
          trxId: String(trxId),
          lineId: String(lineId),
        },
      },
      body: input,
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to update categorization', response.status)
  }
  return data
}

export async function deleteCategorization(
  budgetId: number,
  shareId: number,
  trxId: number,
  lineId: number,
): Promise<void> {
  const { error, response } = await client.DELETE(
    '/budget/{budgetId}/expense-share/{shareId}/transaction/{trxId}/line/{lineId}',
    {
      params: {
        path: {
          budgetId: String(budgetId),
          shareId: String(shareId),
          trxId: String(trxId),
          lineId: String(lineId),
        },
      },
    },
  )
  if (!response.ok) {
    throw apiError(error, 'Failed to delete categorization', response.status)
  }
}

export async function listCategories(budgetId: number): Promise<Category[]> {
  const { data, error, response } = await client.GET('/budget/{budgetId}/category', {
    params: { path: { budgetId: String(budgetId) } },
  })
  if (!response.ok) {
    throw apiError(error, 'Failed to load categories', response.status)
  }
  return data
}
