const formatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
})

export function formatMoney(cents: number): string {
  return formatter.format(cents / 100)
}
