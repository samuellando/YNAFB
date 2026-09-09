/** Parse a numeric route param (`:budgetId`, `:accountId`). Returns null unless
 * the whole segment is digits and fits in a safe integer. */
export function parseIdParam(value: string | undefined): number | null {
  if (value === undefined || !/^\d+$/.test(value)) return null
  const id = Number(value)
  return Number.isSafeInteger(id) ? id : null
}
