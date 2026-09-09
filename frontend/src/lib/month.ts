const labelFormatter = new Intl.DateTimeFormat('en-US', {
  month: 'long',
  year: 'numeric',
  timeZone: 'UTC',
})

export function currentMonth(): string {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

export function isMonth(value: string): boolean {
  return /^\d{4}-\d{2}$/.test(value)
}

export function shiftMonth(month: string, delta: number): string {
  const [year, monthIndex] = month.split('-').map(Number)
  const total = year * 12 + (monthIndex - 1) + delta
  const shiftedYear = Math.floor(total / 12)
  const shiftedMonth = (total % 12) + 1
  return `${shiftedYear}-${String(shiftedMonth).padStart(2, '0')}`
}

export function formatMonthLabel(month: string): string {
  const [year, monthIndex] = month.split('-').map(Number)
  return labelFormatter.format(new Date(Date.UTC(year, monthIndex - 1, 1)))
}
