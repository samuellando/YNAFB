const dateFormatter = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: 'numeric',
  year: 'numeric',
  timeZone: 'UTC',
})

export function formatDate(iso: string): string {
  const time = Date.parse(iso)
  if (Number.isNaN(time)) return iso
  return dateFormatter.format(new Date(time))
}

export function dateInputValue(iso: string): string {
  const time = Date.parse(iso)
  if (Number.isNaN(time)) return ''
  return new Date(time).toISOString().slice(0, 10)
}

export function dateFromInput(value: string): string {
  return `${value}T00:00:00Z`
}

export function isDateInput(value: string): boolean {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value)
  if (!match) return false
  const [, year, month, day] = match
  const date = new Date(`${value}T00:00:00Z`)
  if (Number.isNaN(date.getTime())) return false
  return (
    date.getUTCFullYear() === Number(year) &&
    date.getUTCMonth() + 1 === Number(month) &&
    date.getUTCDate() === Number(day)
  )
}

export function todayInput(): string {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
