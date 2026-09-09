import { useMemo, useRef, useState, type FormEvent } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createCategoryGroup, type CategoryGroup } from '../lib/api/budget'
import { ApiError } from '../lib/api/errors'
import { useClickOutside } from '../lib/useClickOutside'
import { ChevronDownIcon, PlusIcon } from './icons'

type GroupPickerProps = {
  budgetId: number
  groups: CategoryGroup[]
  value: number | null
  onChange: (id: number | null) => void
}

export default function GroupPicker({ budgetId, groups, value, onChange }: GroupPickerProps) {
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [creating, setCreating] = useState(false)
  const [newName, setNewName] = useState('')
  const [localGroups, setLocalGroups] = useState<CategoryGroup[]>([])
  const rootRef = useRef<HTMLDivElement>(null)

  useClickOutside(
    rootRef,
    () => {
      setOpen(false)
      setCreating(false)
    },
    open,
  )

  const visibleGroups = useMemo(() => {
    const known = new Set(groups.map((g) => g.id))
    return [...groups, ...localGroups.filter((g) => !known.has(g.id))]
  }, [groups, localGroups])

  const active = value === null ? undefined : visibleGroups.find((g) => g.id === value)

  const createGroup = useMutation({
    mutationFn: (groupName: string) => createCategoryGroup(budgetId, groupName),
    onSuccess: (group) => {
      queryClient.invalidateQueries({ queryKey: ['category-groups', budgetId] })
      setLocalGroups((prev) => (prev.some((g) => g.id === group.id) ? prev : [...prev, group]))
      onChange(group.id)
      setCreating(false)
      setNewName('')
    },
  })

  function pick(id: number | null) {
    onChange(id)
    setOpen(false)
    setCreating(false)
  }

  function submitNewGroup(e: FormEvent) {
    e.preventDefault()
    if (newName.trim()) {
      createGroup.mutate(newName.trim())
    }
  }

  return (
    <div className="mt-4 flex flex-col gap-1.5 text-sm font-medium text-slate-300">
      <span id="group-picker-label">Group</span>
      <div className="relative" ref={rootRef}>
        <button
          type="button"
          aria-haspopup="listbox"
          aria-expanded={open}
          aria-labelledby="group-picker-label group-picker-value"
          onClick={() => setOpen((isOpen) => !isOpen)}
          className="flex w-full items-center justify-between gap-2 rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-left text-sm font-medium text-slate-100 hover:border-emerald-400"
        >
          <span id="group-picker-value" className="truncate">
            {active?.name ?? (value === null ? 'No group' : 'Unknown group')}
          </span>
          <ChevronDownIcon className="h-4 w-4 shrink-0 text-slate-400" />
        </button>

        {open && (
          <div className="absolute inset-x-0 top-full z-20 mt-1 overflow-hidden rounded-lg border border-slate-700 bg-slate-900 shadow-xl">
            <ul role="listbox" aria-labelledby="group-picker-label" className="max-h-64 overflow-y-auto py-1">
              <li role="presentation">
                <button
                  type="button"
                  role="option"
                  aria-selected={value === null}
                  onClick={() => pick(null)}
                  className={`w-full truncate px-3 py-2 text-left text-sm hover:bg-slate-800 ${
                    value === null ? 'font-semibold text-emerald-400' : 'text-slate-200'
                  }`}
                >
                  No group
                </button>
              </li>
              {visibleGroups.map((group) => (
                <li key={group.id} role="presentation">
                  <button
                    type="button"
                    role="option"
                    aria-selected={group.id === value}
                    onClick={() => pick(group.id)}
                    className={`w-full truncate px-3 py-2 text-left text-sm hover:bg-slate-800 ${
                      group.id === value ? 'font-semibold text-emerald-400' : 'text-slate-200'
                    }`}
                  >
                    {group.name}
                  </button>
                </li>
              ))}
            </ul>
            <div className="border-t border-slate-700 py-1">
              {creating ? (
                <div data-escape-stop>
                  <form onSubmit={submitNewGroup} className="flex gap-2 p-2">
                    <input
                      type="text"
                      value={newName}
                      onChange={(e) => setNewName(e.target.value)}
                      placeholder="Group name"
                      autoFocus
                      onKeyDown={(e) => {
                        if (e.key === 'Escape') {
                          setCreating(false)
                          setNewName('')
                        }
                      }}
                      className="min-w-0 flex-1 rounded border border-slate-700 bg-slate-950 px-2 py-1 text-sm text-slate-100 outline-none focus:border-emerald-400"
                    />
                    <button
                      type="submit"
                      disabled={createGroup.isPending}
                      className="rounded bg-emerald-500 px-3 py-1 text-sm font-semibold text-slate-950 hover:bg-emerald-400 disabled:opacity-50"
                    >
                      {createGroup.isPending ? '…' : 'Add'}
                    </button>
                  </form>
                </div>
              ) : (
                <button
                  type="button"
                  onClick={() => setCreating(true)}
                  className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm font-medium text-emerald-400 hover:bg-slate-800"
                >
                  <PlusIcon className="h-4 w-4" />
                  New group
                </button>
              )}
            </div>
            {createGroup.isError && (
              <p className="border-t border-slate-700 px-3 py-2 text-sm text-red-400">
                {createGroup.error instanceof ApiError
                  ? createGroup.error.message
                  : 'Failed to create group'}
              </p>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
