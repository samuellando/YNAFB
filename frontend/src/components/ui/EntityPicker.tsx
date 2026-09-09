import { useId, useMemo, useRef, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { ApiError } from '../../lib/api/errors'
import { useClickOutside } from '../../lib/useClickOutside'
import { ChevronDownIcon, PlusIcon } from '../icons'

export type EntityOption = {
  id: number
  name: string
  subtext?: string
}

type EntityPickerProps = {
  label: string
  options: EntityOption[]
  value: number | null
  onChange: (id: number | null) => void
  placeholder?: string
  onCreate?: (name: string) => Promise<EntityOption>
  createLabel?: string
  disabled?: boolean
  hideLabel?: boolean
  invalid?: boolean
  onEnter?: () => void
}

export default function EntityPicker({
  label,
  options,
  value,
  onChange,
  placeholder = 'Select…',
  onCreate,
  createLabel = 'New',
  disabled = false,
  hideLabel = false,
  invalid = false,
  onEnter,
}: EntityPickerProps) {
  const labelId = useId()
  const listId = useId()
  const [open, setOpen] = useState(false)
  const [draft, setDraft] = useState<string | null>(null)
  const [highlight, setHighlight] = useState(0)
  const [localOptions, setLocalOptions] = useState<EntityOption[]>([])
  const rootRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const visibleOptions = useMemo(() => {
    const known = new Set(options.map((o) => o.id))
    return [...options, ...localOptions.filter((o) => !known.has(o.id))]
  }, [options, localOptions])

  const selected = value === null ? undefined : visibleOptions.find((o) => o.id === value)
  const selectedName = selected?.name ?? ''
  const query = open && draft !== null ? draft : selectedName

  useClickOutside(
    rootRef,
    () => {
      setOpen(false)
      setDraft(null)
    },
    open,
  )

  const normalizedQuery = query.trim().toLowerCase()

  const filtered = useMemo(() => {
    if (normalizedQuery === '') return visibleOptions
    return visibleOptions.filter(
      (o) =>
        o.name.toLowerCase().includes(normalizedQuery) ||
        o.subtext?.toLowerCase().includes(normalizedQuery) === true,
    )
  }, [visibleOptions, normalizedQuery])

  const showCreate =
    onCreate !== undefined &&
    query.trim() !== '' &&
    !visibleOptions.some((o) => o.name.toLowerCase() === normalizedQuery)

  const activeIndex = filtered.length === 0 ? -1 : Math.min(highlight, filtered.length - 1)

  const create = useMutation({
    mutationFn: (entityName: string) => {
      if (!onCreate) throw new Error('Creating a new option is not supported')
      return onCreate(entityName)
    },
    onSuccess: (option) => {
      setLocalOptions((prev) =>
        prev.some((o) => o.id === option.id) ? prev : [...prev, option],
      )
      onChange(option.id)
      setDraft(null)
      setOpen(false)
    },
  })

  function pick(option: EntityOption | null) {
    onChange(option?.id ?? null)
    setDraft(null)
    setOpen(false)
  }

  function submitCreate() {
    const name = query.trim()
    if (name !== '' && !create.isPending) {
      create.mutate(name)
    }
  }

  function closeAndRevert() {
    setOpen(false)
    setDraft(null)
  }

  function toggle() {
    if (open) {
      closeAndRevert()
    } else {
      setHighlight(0)
      setOpen(true)
      inputRef.current?.focus()
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Escape') {
      closeAndRevert()
      return
    }
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      if (!open) {
        setOpen(true)
      } else if (filtered.length > 0) {
        setHighlight((h) => Math.min(h + 1, filtered.length - 1))
      }
      return
    }
    if (e.key === 'ArrowUp') {
      e.preventDefault()
      if (open) {
        setHighlight((h) => Math.max(h - 1, 0))
      }
      return
    }
    if (e.key === 'Enter') {
      e.preventDefault()
      const target = activeIndex >= 0 ? filtered[activeIndex] : undefined
      if (target) {
        pick(target)
        onEnter?.()
      } else if (showCreate) {
        submitCreate()
      }
    }
  }

  return (
    <div
      className={
        hideLabel
          ? 'text-sm font-medium text-slate-300'
          : 'flex flex-col gap-1.5 text-sm font-medium text-slate-300'
      }
    >
      {!hideLabel && <span id={labelId}>{label}</span>}
      <div data-escape-stop>
        <div className="relative" ref={rootRef}>
          <input
            ref={inputRef}
            type="text"
            role="combobox"
            aria-expanded={open}
            aria-controls={listId}
            aria-activedescendant={
              activeIndex >= 0 ? `${listId}-option-${filtered[activeIndex].id}` : undefined
            }
            aria-label={hideLabel ? label : undefined}
            aria-labelledby={hideLabel ? undefined : labelId}
            value={query}
            disabled={disabled}
            placeholder={placeholder}
            autoComplete="off"
            onFocus={() => {
              setDraft(selectedName)
              setHighlight(0)
              setOpen(true)
              inputRef.current?.select()
            }}
            onChange={(e) => {
              setDraft(e.target.value)
              setOpen(true)
              setHighlight(0)
            }}
            onKeyDown={handleKeyDown}
            className={`w-full rounded-lg border bg-slate-950 px-3 py-2 pr-9 text-sm text-slate-100 outline-none placeholder:text-slate-600 focus:border-emerald-400 disabled:cursor-not-allowed disabled:opacity-50 ${
              invalid ? 'border-red-800' : 'border-slate-700'
            }`}
          />
          <button
            type="button"
            tabIndex={-1}
            aria-label={`Toggle ${label} options`}
            disabled={disabled}
            onClick={toggle}
            className="absolute top-1/2 right-2 -translate-y-1/2 text-slate-400 transition hover:text-slate-100 disabled:opacity-50"
          >
            <ChevronDownIcon className="h-4 w-4" />
          </button>

          {open && (
            <div className="absolute inset-x-0 top-full z-20 mt-1 overflow-hidden rounded-lg border border-slate-700 bg-slate-900 shadow-xl">
              <ul
                id={listId}
                role="listbox"
                aria-label={hideLabel ? label : undefined}
                aria-labelledby={hideLabel ? undefined : labelId}
                className="max-h-64 overflow-y-auto py-1"
              >
                {filtered.map((option, index) => (
                  <li key={option.id} role="presentation">
                    <button
                      type="button"
                      id={`${listId}-option-${option.id}`}
                      role="option"
                      aria-selected={option.id === value}
                      onMouseEnter={() => setHighlight(index)}
                      onClick={() => pick(option)}
                      className={`flex w-full items-baseline justify-between gap-2 truncate px-3 py-2 text-left text-sm ${
                        index === activeIndex ? 'bg-slate-800' : ''
                      } ${
                        option.id === value
                          ? 'font-semibold text-emerald-400'
                          : 'text-slate-200'
                      }`}
                    >
                      <span className="truncate">{option.name}</span>
                      {option.subtext && (
                        <span className="shrink-0 text-xs text-slate-500">{option.subtext}</span>
                      )}
                    </button>
                  </li>
                ))}
                {filtered.length === 0 && !showCreate && (
                  <li className="px-3 py-2 text-sm text-slate-500">No matches</li>
                )}
              </ul>
              {showCreate && (
                <div className="border-t border-slate-700 py-1">
                  <button
                    type="button"
                    onClick={submitCreate}
                    disabled={create.isPending}
                    className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm font-medium text-emerald-400 hover:bg-slate-800 disabled:opacity-50"
                  >
                    <PlusIcon className="h-4 w-4 shrink-0" />
                    <span className="truncate">
                      {createLabel} &ldquo;{query.trim()}&rdquo;
                    </span>
                  </button>
                </div>
              )}
              {create.isError && (
                <p className="border-t border-slate-700 px-3 py-2 text-sm text-red-400">
                  {create.error instanceof ApiError
                    ? create.error.message
                    : 'Failed to create'}
                </p>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
