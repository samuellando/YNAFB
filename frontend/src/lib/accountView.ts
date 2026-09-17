import type {
  AccountTransaction,
  AccountTransactionLine,
  BudgetMonthCategory,
} from './api/budget'

export type CategoryOption = {
  id: number
  name: string
  groupName: string | null
}

export function categoryOptions(categories: BudgetMonthCategory[]): CategoryOption[] {
  return categories.map((category) => ({
    id: category.categoryId,
    name: category.categoryName,
    groupName: category.categoryGroupName ?? null,
  }))
}

export function transactionTarget(line: AccountTransactionLine): string {
  if (line.destAccountName) {
    return `@${line.destAccountName}`
  }
  if (line.income) {
    return 'Income'
  }
  if (line.expenseShareName) {
    return line.expenseShareName
  }
  return line.categoryName ?? '—'
}

export function isSplitTransaction(transaction: AccountTransaction): boolean {
  return transaction.transactionLines.length > 1
}

export function mainTarget(transaction: AccountTransaction): string {
  if (transaction.sourceAccountId !== undefined) {
    return transaction.sourceAccountName ? `@${transaction.sourceAccountName}` : 'Transfer'
  }
  const [first] = transaction.transactionLines
  if (!first) {
    return '—'
  }
  return transactionTarget(first)
}

export function isUnbalanced(transaction: AccountTransaction): boolean {
  // Mirror transfer rows carry no lines; the transfer itself is the assignment.
  if (transaction.sourceAccountId !== undefined) return false
  // Fully uncategorized transactions (no lines) are covered by the aggregate chip.
  if (transaction.transactionLines.length === 0) return false
  let outflow = 0
  let inflow = 0
  for (const line of transaction.transactionLines) {
    outflow += line.outflow
    inflow += line.inflow
  }
  return outflow !== transaction.outflow || inflow !== transaction.inflow
}

export function defaultLineAmounts(
  defaults: { percent: number }[],
  total: number,
): number[] {
  const amounts = new Array<number>(defaults.length).fill(0)
  let allocated = 0
  defaults.forEach((d, i) => {
    if (i === defaults.length - 1) {
      amounts[i] = total - allocated
    } else {
      amounts[i] = Math.floor((total * d.percent) / 100)
      allocated += amounts[i]
    }
  })
  return amounts
}

export function linePercents(amounts: number[], total: number): number[] {
  const percents = new Array<number>(amounts.length).fill(0)
  if (amounts.length === 0 || total <= 0) return percents
  let allocated = 0
  amounts.forEach((amount, i) => {
    if (i === amounts.length - 1) {
      percents[i] = 100 - allocated
    } else {
      const p = Math.floor((amount * 100) / total)
      allocated += p
      percents[i] = p
    }
  })
  return percents
}

export function needsCategorize(transaction: AccountTransaction): boolean {
  if (transaction.sourceAccountId !== undefined) return false
  if (transaction.transactionLines.length === 0) return true
  return isUnbalanced(transaction)
}

export function uncategorizedAmount(transactions: AccountTransaction[]): number {
  let expected = 0
  let assigned = 0
  for (const transaction of transactions) {
    // Mirror transfer rows carry no lines; the transfer itself is the assignment.
    if (transaction.sourceAccountId !== undefined) {
      continue
    }
    expected += transaction.outflow + transaction.inflow
    for (const line of transaction.transactionLines) {
      assigned += line.outflow + line.inflow
    }
  }
  return expected - assigned
}
