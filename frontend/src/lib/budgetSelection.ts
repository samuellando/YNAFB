const KEY = 'ynafb:selectedBudget'

export function loadSelectedBudget(): number | null {
  const raw = localStorage.getItem(KEY)
  if (raw === null) return null
  const id = Number(raw)
  return Number.isInteger(id) && id > 0 ? id : null
}

export function saveSelectedBudget(id: number) {
  localStorage.setItem(KEY, String(id))
}
