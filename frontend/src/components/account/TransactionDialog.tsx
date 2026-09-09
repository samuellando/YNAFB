import { useCallback, useEffect, useId, useMemo, useRef, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createCategory,
  createPayee,
  createPayeeDefaultLine,
  createTransaction,
  createTransactionLine,
  deletePayeeDefaultLine,
  deleteTransaction,
  deleteTransactionLine,
  getBudgetMonth,
  listAccounts,
  listPayeeDefaultLines,
  listPayees,
  updateTransaction,
  updateTransactionLine,
  type AccountTransaction,
  type AccountTransactionLine,
  type PayeeDefaultLine,
  type PayeeDefaultLineInput,
  type TransactionLineInput,
} from '../../lib/api/budget'
import { categoryOptions, defaultLineAmounts, linePercents } from '../../lib/accountView'
import { dateFromInput, dateInputValue, isDateInput, todayInput } from '../../lib/date'
import { centsFromInput, centsToInput, formatMoney } from '../../lib/money'
import { currentMonth } from '../../lib/month'
import { PlusIcon, WarningIcon, XIcon } from '../icons'
import AmountInput from '../ui/AmountInput'
import DialogShell from '../ui/DialogShell'
import { CANCEL_BUTTON, DANGER_BUTTON, PRIMARY_BUTTON } from '../ui/buttons'
import EntityPicker, { type EntityOption } from '../ui/EntityPicker'
import { TRANSACTION_LINE_GRID } from './layout'

export type TransactionDialogState = {
  transaction: AccountTransaction | null
}

type TransactionDialogProps = {
  budgetId: number
  accountId: number
  dialog: TransactionDialogState
  progress?: { index: number; total: number }
  onCancel: () => void
  onDone: () => void
  onSkip?: () => void
}

type LineKind = 'income' | 'spending' | 'transfer'

type LineDraft = {
  key: string
  lineId: number | null
  kind: LineKind
  targetId: number | null
  outflow: string
  inflow: string
}

function lineKindOf(line: AccountTransactionLine): LineKind {
  if (line.destAccountId !== undefined) return 'transfer'
  if (line.income) return 'income'
  return 'spending'
}

function draftFromLine(line: AccountTransactionLine): LineDraft {
  return {
    key: `line-${line.lineId}`,
    lineId: line.lineId,
    kind: lineKindOf(line),
    targetId: line.destAccountId ?? line.categoryId ?? null,
    outflow: centsToInput(line.outflow),
    inflow: centsToInput(line.inflow),
  }
}

function emptyDraft(key: string): LineDraft {
  return { key, lineId: null, kind: 'spending', targetId: null, outflow: '', inflow: '' }
}

function draftToInput(draft: LineDraft, outflow: number, inflow: number): TransactionLineInput {
  return {
    ...(draft.kind === 'transfer' && draft.targetId !== null
      ? { destAccountId: draft.targetId }
      : {}),
    ...(draft.kind === 'spending' && draft.targetId !== null
      ? { categoryId: draft.targetId }
      : {}),
    income: draft.kind === 'income',
    outflow,
    inflow,
  }
}

function parseAmount(input: string): number | null {
  if (input.trim() === '') return 0
  return centsFromInput(input)
}

function draftsFromDefaults(
  defaults: PayeeDefaultLine[],
  outflowTotal: number,
  inflowTotal: number,
  newKey: () => string,
): LineDraft[] | null {
  if (defaults.length === 0) return null
  const useOutflow = outflowTotal !== 0
  const total = useOutflow ? outflowTotal : inflowTotal
  if (total <= 0) return null
  const amounts = defaultLineAmounts(defaults, total)
  return defaults.map((d, i) => {
    const kind: LineKind =
      d.destAccountId !== undefined ? 'transfer' : d.income ? 'income' : 'spending'
    return {
      key: newKey(),
      lineId: null,
      kind,
      targetId: d.destAccountId ?? d.categoryId ?? null,
      outflow: useOutflow ? centsToInput(amounts[i]) : centsToInput(0),
      inflow: useOutflow ? centsToInput(0) : centsToInput(amounts[i]),
    }
  })
}

function defaultLineInput(draft: LineDraft, percent: number): PayeeDefaultLineInput {
  return {
    ...(draft.kind === 'transfer' && draft.targetId !== null
      ? { destAccountId: draft.targetId }
      : {}),
    ...(draft.kind === 'spending' && draft.targetId !== null
      ? { categoryId: draft.targetId }
      : {}),
    income: draft.kind === 'income',
    percent,
  }
}

export default function TransactionDialog({
  budgetId,
  accountId,
  dialog,
  progress,
  onCancel,
  onDone,
  onSkip,
}: TransactionDialogProps) {
  const queryClient = useQueryClient()
  const { transaction } = dialog
  const isCreate = transaction === null
  const instanceId = useId()
  const keyCounter = useRef(0)
  const current = currentMonth()

  const newKey = useCallback((): string => {
    keyCounter.current += 1
    return `${instanceId}-new-${keyCounter.current}`
  }, [instanceId])

  const payeesQuery = useQuery({
    queryKey: ['payees', budgetId],
    queryFn: () => listPayees(budgetId),
  })
  const accountsQuery = useQuery({
    queryKey: ['accounts', budgetId],
    queryFn: () => listAccounts(budgetId),
  })
  const monthQuery = useQuery({
    queryKey: ['budget-month', budgetId, current],
    queryFn: () => getBudgetMonth(budgetId, current),
  })
  const payeeOptions: EntityOption[] = useMemo(
    () => (payeesQuery.data ?? []).map((payee) => ({ id: payee.id, name: payee.name })),
    [payeesQuery.data],
  )
  const accountOptions: EntityOption[] = useMemo(
    () =>
      (accountsQuery.data ?? [])
        .filter((account) => account.id !== accountId)
        .map((account) => ({ id: account.id, name: account.name })),
    [accountsQuery.data, accountId],
  )
  const categoryOpts = useMemo(
    () => categoryOptions(monthQuery.data?.categories ?? []),
    [monthQuery.data],
  )
  const categoryPickerOptions: EntityOption[] = useMemo(
    () =>
      categoryOpts.map((category) => ({
        id: category.id,
        name: category.name,
        ...(category.groupName ? { subtext: category.groupName } : {}),
      })),
    [categoryOpts],
  )

  const [payeeId, setPayeeId] = useState<number | null>(transaction?.payeeId ?? null)
  const [date, setDate] = useState(
    transaction ? dateInputValue(transaction.date) : todayInput(),
  )
  const [outflow, setOutflow] = useState(
    transaction ? centsToInput(transaction.outflow) : '',
  )
  const [inflow, setInflow] = useState(transaction ? centsToInput(transaction.inflow) : '')
  const [note, setNote] = useState(transaction?.note ?? '')
  const [lines, setLines] = useState<LineDraft[]>(() =>
    transaction ? transaction.transactionLines.map(draftFromLine) : [],
  )
  const [confirmingDelete, setConfirmingDelete] = useState(false)
  const [saveAsDefault, setSaveAsDefault] = useState(true)
  const outflowRefs = useRef(new Map<string, HTMLInputElement>())
  const prefilledDefaults = useRef(false)
  const prefilledCreatePayee = useRef<number | null>(null)

  const defaultsQuery = useQuery({
    queryKey: ['payee-defaults', budgetId, payeeId],
    queryFn: () => {
      if (payeeId === null) throw new Error('Select a payee')
      return listPayeeDefaultLines(budgetId, payeeId)
    },
    enabled: payeeId !== null,
  })

  const existingLineCount = transaction?.transactionLines.length ?? 0
  const existingOutflow = transaction?.outflow ?? 0
  const existingInflow = transaction?.inflow ?? 0

  // One-shot prefill of async payee defaults into form state on open.
  // Fires at most once (ref guard); the set-state-in-effect lint warning is expected here.
  useEffect(() => {
    if (isCreate || transaction === null) return
    if (prefilledDefaults.current) return
    if (transaction.transactionLines.length > 0) return
    if (lines.length > 0) return
    const defaults = defaultsQuery.data
    if (!defaults || defaults.length === 0) return
    const drafts = draftsFromDefaults(defaults, transaction.outflow, transaction.inflow, newKey)
    if (!drafts) return
    prefilledDefaults.current = true
    setLines(drafts)
  }, [
    defaultsQuery.data,
    existingInflow,
    existingLineCount,
    existingOutflow,
    isCreate,
    lines.length,
    newKey,
    transaction,
  ])

  const outflowCents = parseAmount(outflow)
  const inflowCents = parseAmount(inflow)
  const validDate = isDateInput(date)
  const validAmounts = outflowCents !== null && inflowCents !== null

  // In create mode there are no stored totals, so prefill payee defaults once
  // per payee selection as soon as totals are entered. Skips when the user
  // already added lines.
  useEffect(() => {
    if (!isCreate) return
    if (payeeId === null) return
    if (lines.length > 0) return
    if (prefilledCreatePayee.current === payeeId) return
    const defaults = defaultsQuery.data
    if (!defaults || defaults.length === 0) return
    if (outflowCents === null || inflowCents === null) return
    const drafts = draftsFromDefaults(defaults, outflowCents, inflowCents, newKey)
    if (!drafts) return
    prefilledCreatePayee.current = payeeId
    setLines(drafts)
  }, [defaultsQuery.data, inflowCents, isCreate, lines.length, newKey, outflowCents, payeeId])

  const lineStates = lines.map((draft) => {
    const lineOut = parseAmount(draft.outflow)
    const lineIn = parseAmount(draft.inflow)
    const needsTarget = draft.kind === 'spending' || draft.kind === 'transfer'
    const valid = lineOut !== null && lineIn !== null && (!needsTarget || draft.targetId !== null)
    return { draft, out: lineOut ?? 0, in: lineIn ?? 0, valid }
  })
  const validLines = lineStates.every((state) => state.valid)

  const lineOutTotal = lineStates.reduce((sum, state) => sum + state.out, 0)
  const lineInTotal = lineStates.reduce((sum, state) => sum + state.in, 0)
  const unbalanced =
    validAmounts &&
    lines.length > 0 &&
    (lineOutTotal !== outflowCents || lineInTotal !== inflowCents)

  const canSave = payeeId !== null && validDate && validAmounts && validLines

  function updateDraft(key: string, patch: Partial<LineDraft>) {
    setLines((prev) => prev.map((draft) => (draft.key === key ? { ...draft, ...patch } : draft)))
  }

  function setKind(key: string, kind: LineKind) {
    setLines((prev) =>
      prev.map((draft) => (draft.key === key ? { ...draft, kind, targetId: null } : draft)),
    )
  }

  function focusLineOutflow(key: string) {
    outflowRefs.current.get(key)?.focus()
  }

  function addLine() {
    const outLeft = (outflowCents ?? 0) - lineOutTotal
    const inLeft = (inflowCents ?? 0) - lineInTotal
    const kind: LineKind = inLeft > 0 && outLeft <= 0 ? 'income' : 'spending'
    setLines((prev) => [
      ...prev,
      {
        ...emptyDraft(newKey()),
        kind,
        outflow: outLeft > 0 ? centsToInput(outLeft) : '',
        inflow: inLeft > 0 ? centsToInput(inLeft) : '',
      },
    ])
  }

  async function invalidate() {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['account', budgetId, accountId] }),
      queryClient.invalidateQueries({ queryKey: ['accounts', budgetId] }),
      queryClient.invalidateQueries({ queryKey: ['budget-month', budgetId] }),
      ...(payeeId !== null
        ? [queryClient.invalidateQueries({ queryKey: ['payee-defaults', budgetId, payeeId] })]
        : []),
    ])
  }

  async function replacePayeeDefaults() {
    if (payeeId === null || outflowCents === null || inflowCents === null) return
    const total = outflowCents !== 0 ? outflowCents : inflowCents
    if (total <= 0) return
    const useOutflow = outflowCents !== 0
    const amounts = lineStates.map((state) => (useOutflow ? state.out : state.in))
    const percents = linePercents(amounts, total)
    const existing = await listPayeeDefaultLines(budgetId, payeeId)
    for (const line of existing) {
      await deletePayeeDefaultLine(budgetId, payeeId, line.id)
    }
    for (let i = 0; i < lineStates.length; i++) {
      await createPayeeDefaultLine(budgetId, payeeId, defaultLineInput(lineStates[i].draft, percents[i]))
    }
  }

  const save = useMutation({
    mutationFn: async () => {
      if (payeeId === null) throw new Error('Select a payee')
      if (!validDate) throw new Error('Enter a valid date')
      if (outflowCents === null || inflowCents === null) {
        throw new Error('Enter valid amounts')
      }
      const input = {
        payeeId,
        date: dateFromInput(date),
        outflow: outflowCents,
        inflow: inflowCents,
        note: note.trim(),
      }
      if (isCreate || transaction === null) {
        const created = await createTransaction(budgetId, accountId, input)
        for (const state of lineStates) {
          await createTransactionLine(
            budgetId,
            accountId,
            created.id,
            draftToInput(state.draft, state.out, state.in),
          )
        }
        if (saveAsDefault && lineStates.length > 0) {
          await replacePayeeDefaults()
        }
        return
      }
      await updateTransaction(budgetId, accountId, transaction.id, input)
      const kept = new Set<number>()
      for (const state of lineStates) {
        const lineInput = draftToInput(state.draft, state.out, state.in)
        if (state.draft.lineId === null) {
          await createTransactionLine(budgetId, accountId, transaction.id, lineInput)
        } else {
          kept.add(state.draft.lineId)
          await updateTransactionLine(budgetId, accountId, transaction.id, state.draft.lineId, lineInput)
        }
      }
      for (const line of transaction.transactionLines) {
        if (!kept.has(line.lineId)) {
          await deleteTransactionLine(budgetId, accountId, transaction.id, line.lineId)
        }
      }
      if (saveAsDefault && lineStates.length > 0) {
        await replacePayeeDefaults()
      }
    },
    onSuccess: async () => {
      await invalidate()
      onDone()
    },
  })

  const remove = useMutation({
    mutationFn: () => {
      if (transaction === null) throw new Error('Nothing to delete')
      return deleteTransaction(budgetId, accountId, transaction.id)
    },
    onSuccess: async () => {
      await invalidate()
      onDone()
    },
  })

  async function handleCreatePayee(name: string): Promise<EntityOption> {
    const payee = await createPayee(budgetId, name)
    await queryClient.invalidateQueries({ queryKey: ['payees', budgetId] })
    return { id: payee.id, name: payee.name }
  }

  async function handleCreateCategory(name: string): Promise<EntityOption> {
    const category = await createCategory(budgetId, name)
    await queryClient.invalidateQueries({ queryKey: ['budget-month', budgetId] })
    return { id: category.id, name: category.name }
  }

  const loading = payeesQuery.isPending || accountsQuery.isPending || monthQuery.isPending
  const loadError = payeesQuery.isError || accountsQuery.isError || monthQuery.isError

  return (
    <DialogShell wide onClose={onCancel}>
      <div className="flex items-baseline justify-between gap-4">
        <h3 className="text-lg font-bold tracking-tight">
          {isCreate ? 'Add transaction' : 'Edit transaction'}
        </h3>
        {progress && (
          <p className="shrink-0 text-sm text-slate-400" aria-live="polite">
            transaction {progress.index}/{progress.total}
          </p>
        )}
      </div>

      {loading && <p className="py-6 text-center text-sm text-slate-400">Loading…</p>}

      {!loading && loadError && (
        <div className="mt-5">
          <p className="text-sm text-red-400">Failed to load pickers.</p>
          <button
            type="button"
            onClick={() => {
              payeesQuery.refetch()
              accountsQuery.refetch()
              monthQuery.refetch()
            }}
            className="mt-4 rounded-lg border border-red-800 px-3 py-1.5 text-sm font-medium text-red-400 transition hover:bg-red-900/50"
          >
            Retry
          </button>
        </div>
      )}

      {!loading && !loadError && (
        <>
          <div className="mt-5">
            <EntityPicker
              label="Payee"
              options={payeeOptions}
              value={payeeId}
              onChange={setPayeeId}
              placeholder="Select payee"
              onCreate={handleCreatePayee}
              createLabel="New payee"
            />
          </div>

          <div className="mt-4 grid grid-cols-3 gap-3">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-slate-300">
              Date
              <input
                type="date"
                value={date}
                onChange={(e) => setDate(e.target.value)}
                className="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-slate-100 outline-none focus:border-emerald-400"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-slate-300">
              Outflow
              <AmountInput value={outflow} onChange={setOutflow} className="px-3 py-2" />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-slate-300">
              Inflow
              <AmountInput value={inflow} onChange={setInflow} className="px-3 py-2" />
            </label>
          </div>

          <label className="mt-4 flex flex-col gap-1.5 text-sm font-medium text-slate-300">
            Note
            <input
              type="text"
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="Optional"
              className="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-slate-100 outline-none focus:border-emerald-400"
            />
          </label>

          <div className="mt-6">
            <h4 className="text-xs font-semibold tracking-widest text-slate-500 uppercase">
              Lines
            </h4>

            <div
              className={`${TRANSACTION_LINE_GRID} mt-2 border-b border-slate-800 px-2 py-2 text-xs font-semibold tracking-widest text-slate-500 uppercase`}
            >
              <span>Type</span>
              <span>Target</span>
              <span className="text-right">Outflow</span>
              <span className="text-right">Inflow</span>
              <span />
            </div>

            {lines.length === 0 && (
              <p className="border-b border-slate-800/60 px-2 py-3 text-sm text-slate-400">
                No lines yet. The whole amount counts as uncategorized.
              </p>
            )}

            {lines.map((draft, index) => {
              const needsTarget = draft.kind === 'spending' || draft.kind === 'transfer'
              const missingTarget = needsTarget && draft.targetId === null
              const badOut = parseAmount(draft.outflow) === null
              const badIn = parseAmount(draft.inflow) === null
              return (
                <div
                  key={draft.key}
                  className={`${TRANSACTION_LINE_GRID} border-b border-slate-800/60 px-2 py-2`}
                >
                  <select
                    aria-label="Line type"
                    value={draft.kind}
                    onChange={(e) => setKind(draft.key, e.target.value as LineKind)}
                    className="w-full cursor-pointer rounded-lg border border-slate-700 bg-slate-950 px-2 py-2 text-sm text-slate-100 outline-none focus:border-emerald-400 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    <option value="income">Income</option>
                    <option value="spending">Spending</option>
                    <option value="transfer">Transfer</option>
                  </select>
                  {draft.kind === 'income' ? (
                    <span className="truncate px-3 py-2 text-sm text-slate-500">Income</span>
                  ) : draft.kind === 'spending' ? (
                    <EntityPicker
                      hideLabel
                      label="Category"
                      options={categoryPickerOptions}
                      value={draft.targetId}
                      onChange={(id) => updateDraft(draft.key, { targetId: id })}
                      placeholder="Select category"
                      invalid={missingTarget}
                      onCreate={handleCreateCategory}
                      createLabel="New category"
                      onEnter={() => focusLineOutflow(draft.key)}
                    />
                  ) : (
                    <EntityPicker
                      hideLabel
                      label="To account"
                      options={accountOptions}
                      value={draft.targetId}
                      onChange={(id) => updateDraft(draft.key, { targetId: id })}
                      placeholder="Select account"
                      invalid={missingTarget}
                      onEnter={() => focusLineOutflow(draft.key)}
                    />
                  )}
                  <AmountInput
                    ariaLabel="Line outflow"
                    inputRef={(el) => {
                      if (el) outflowRefs.current.set(draft.key, el)
                      else outflowRefs.current.delete(draft.key)
                    }}
                    value={draft.outflow}
                    onChange={(outflow) => updateDraft(draft.key, { outflow })}
                    invalid={badOut}
                    className="px-2 py-2 text-right tabular-nums"
                  />
                  <AmountInput
                    ariaLabel="Line inflow"
                    value={draft.inflow}
                    onChange={(inflow) => updateDraft(draft.key, { inflow })}
                    invalid={badIn}
                    className="px-2 py-2 text-right tabular-nums"
                  />
                  <button
                    type="button"
                    aria-label={`Remove line ${index + 1}`}
                    title="Remove line"
                    onClick={() =>
                      setLines((prev) => prev.filter((other) => other.key !== draft.key))
                    }
                    className={`flex cursor-pointer justify-center transition ${
                      missingTarget ? 'text-red-400 hover:text-red-300' : 'text-slate-500 hover:text-red-400'
                    }`}
                  >
                    <XIcon className="h-4 w-4" />
                  </button>
                  {missingTarget && <span className="sr-only">Pick a target</span>}
                </div>
              )
            })}

            <button
              type="button"
              onClick={addLine}
              className="mt-2 flex w-full cursor-pointer items-center justify-center gap-1.5 rounded-lg border border-dashed border-slate-700 px-3 py-2 text-sm font-medium text-emerald-400 transition hover:border-emerald-400"
            >
              <PlusIcon className="h-4 w-4" /> Add line
            </button>

            {unbalanced && (
              <p className="mt-3 flex items-center gap-1.5 text-xs font-semibold text-yellow-300">
                <WarningIcon className="h-3.5 w-3.5" />
                Lines ({formatMoney(lineOutTotal)} out, {formatMoney(lineInTotal)} in) don&apos;t
                add up to the transaction total.
              </p>
            )}
          </div>

          {(save.isError || remove.isError) && (
            <p className="mt-3 text-sm text-red-400">
              {save.isError ? save.error.message : remove.isError ? remove.error.message : ''}
            </p>
          )}
          <div className="mt-6 flex justify-end">
            <label className="flex cursor-pointer items-center gap-2 text-sm text-slate-300">
              <input
                type="checkbox"
                checked={saveAsDefault}
                disabled={lines.length === 0}
                onChange={(e) => setSaveAsDefault(e.target.checked)}
                className="h-4 w-4 accent-emerald-500 disabled:cursor-not-allowed disabled:opacity-50"
              />
              Save as payee default
            </label>
          </div>
          <div className="mt-3 flex justify-end gap-2">
            {!isCreate && (
              <button
                type="button"
                onClick={() => (confirmingDelete ? remove.mutate() : setConfirmingDelete(true))}
                disabled={remove.isPending}
                className={DANGER_BUTTON}
              >
                {remove.isPending ? 'Deleting…' : confirmingDelete ? 'Confirm delete' : 'Delete'}
              </button>
            )}
            <button
              type="button"
              onClick={onCancel}
              className={CANCEL_BUTTON}
            >
              {onSkip ? 'Stop' : 'Cancel'}
            </button>
            {onSkip && (
              <button
                type="button"
                onClick={onSkip}
                className={CANCEL_BUTTON}
              >
                Skip
              </button>
            )}
            <button
              type="button"
              onClick={() => save.mutate()}
              disabled={!canSave || save.isPending}
              className={PRIMARY_BUTTON}
            >
              {save.isPending
                ? isCreate
                  ? 'Adding…'
                  : 'Saving…'
                : onSkip
                  ? 'Next'
                  : isCreate
                    ? 'Add'
                    : 'Save'}
            </button>
          </div>
        </>
      )}
    </DialogShell>
  )
}
