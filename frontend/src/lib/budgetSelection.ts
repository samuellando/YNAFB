const KEY = 'ynafb:selectedBudget'

export function loadSelectedBudget(): number | null {
  const raw = localStorage.getItem(KEY)
  if (raw === null) return null
  const id = Number(raw)
  return Number.isNaN(id) ? null : id
}

export function saveSelectedBudget(id: number) {
  localStorage.setItem(KEY, String(id))
}
