import { useId, useMemo, useRef, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createCategory,
  getBudgetMonth,
  listAccounts,
  updateExpenseShareSplitLines,
  type ExpenseShareTrx,
  type ExpenseShareTrxSplitLine,
} from '../../lib/api/budget'
import { categoryOptions } from '../../lib/accountView'
import { centsFromInput, centsToInput, formatMoney } from '../../lib/money'
import { currentMonth } from '../../lib/month'
import { PlusIcon, WarningIcon, XIcon } from '../icons'
import AmountInput from '../ui/AmountInput'
import DialogShell from '../ui/DialogShell'
import { CANCEL_BUTTON, PRIMARY_BUTTON } from '../ui/buttons'
import EntityPicker, { type EntityOption } from '../ui/EntityPicker'

export type ExpenseShareCategorizeDialogState = {
  transaction: ExpenseShareTrx
}

type ExpenseShareCategorizeDialogProps = {
  budgetId: number
  expenseShareId: number
  dialog: ExpenseShareCategorizeDialogState
  onClose: () => void
  onDone: () => void
}

type LineKind = 'spending' | 'transfer'

type LineDraft = {
  key: string
  kind: LineKind
  targetId: number | null
  outflow: string
  inflow: string
}

function draftFromLine(line: ExpenseShareTrxSplitLine, key: string): LineDraft {
  return {
    key,
    kind: line.destAccountId !== undefined ? 'transfer' : 'spending',
    targetId: line.destAccountId ?? line.categoryId ?? null,
    outflow: centsToInput(line.outflow),
    inflow: centsToInput(line.inflow),
  }
}

const LINE_GRID = 'grid grid-cols-[7rem_minmax(0,1fr)_6.5rem_6.5rem_1.5rem] gap-x-2'

export default function ExpenseShareCategorizeDialog({
  budgetId,
  expenseShareId,
  dialog,
  onClose,
  onDone,
}: ExpenseShareCategorizeDialogProps) {
  const queryClient = useQueryClient()
  const { transaction } = dialog
  const outflowMode = transaction.totalInflow === 0
  const own = transaction.splits.find((split) => split.budgetId === budgetId)
  const splitAmount = own
    ? outflowMode
      ? own.splitOutflow
      : own.splitInflow
    : 0

  const instanceId = useId()
  const keyCounter = useRef(0)
  const current = currentMonth()
  const newKey = () => {
    keyCounter.current += 1
    return `${instanceId}-line-${keyCounter.current}`
  }

  const accountsQuery = useQuery({
    queryKey: ['accounts', budgetId],
    queryFn: () => listAccounts(budgetId),
  })
  const monthQuery = useQuery({
    queryKey: ['budget-month', budgetId, current],
    queryFn: () => getBudgetMonth(budgetId, current),
  })
  const accountOptions: EntityOption[] = useMemo(
    () => (accountsQuery.data ?? []).map((account) => ({ id: account.id, name: account.name })),
    [accountsQuery.data],
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

  const [lines, setLines] = useState<LineDraft[]>(() =>
    (own?.lines ?? []).map((line, i) => draftFromLine(line, `${instanceId}-line-${i}`)),
  )

  const parsed = lines.map((draft) => {
    const out = centsFromInput(draft.outflow.trim() === '' ? '0' : draft.outflow)
    const inc = centsFromInput(draft.inflow.trim() === '' ? '0' : draft.inflow)
    const activeOk = outflowMode ? (out ?? 0) > 0 : (inc ?? 0) > 0
    const otherOk = outflowMode ? inc === 0 : out === 0
    return {
      draft,
      out: out ?? 0,
      in: inc ?? 0,
      valid: out !== null && inc !== null && draft.targetId !== null && activeOk && otherOk,
    }
  })
  const allValid = parsed.every(({ valid }) => valid)
  const lineTotal = parsed.reduce((sum, { out, in: inc }) => sum + (outflowMode ? out : inc), 0)
  const remaining = splitAmount - lineTotal

  function updateDraft(key: string, patch: Partial<LineDraft>) {
    setLines((prev) => prev.map((draft) => (draft.key === key ? { ...draft, ...patch } : draft)))
  }

  function setKind(key: string, kind: LineKind) {
    setLines((prev) =>
      prev.map((draft) => (draft.key === key ? { ...draft, kind, targetId: null } : draft)),
    )
  }

  function addLine() {
    const left = remaining > 0 ? remaining : 0
    setLines((prev) => [
      ...prev,
      {
        key: newKey(),
        kind: 'spending',
        targetId: null,
        outflow: outflowMode && left > 0 ? centsToInput(left) : '',
        inflow: !outflowMode && left > 0 ? centsToInput(left) : '',
      },
    ])
  }

  async function handleCreateCategory(name: string): Promise<EntityOption> {
    const category = await createCategory(budgetId, name)
    await queryClient.invalidateQueries({ queryKey: ['budget-month', budgetId] })
    return { id: category.id, name: category.name }
  }

  const save = useMutation({
    mutationFn: () => {
      for (const { valid } of parsed) {
        if (!valid) throw new Error('Fix the highlighted lines')
      }
      return updateExpenseShareSplitLines(
        budgetId,
        expenseShareId,
        transaction.id,
        parsed.map(({ draft, out, in: inc }) => ({
          ...(draft.kind === 'transfer' && draft.targetId !== null
            ? { destAccountId: draft.targetId }
            : {}),
          ...(draft.kind === 'spending' && draft.targetId !== null
            ? { categoryId: draft.targetId }
            : {}),
          outflow: out,
          inflow: inc,
        })),
      )
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['expense-share', budgetId, expenseShareId] }),
        queryClient.invalidateQueries({ queryKey: ['budget-month', budgetId] }),
      ])
      onDone()
    },
  })

  const loading = accountsQuery.isPending || monthQuery.isPending
  const loadError = accountsQuery.isError || monthQuery.isError
  const canSave = own !== undefined && allValid

  return (
    <DialogShell wide onClose={onClose}>
      <h3 className="text-lg font-bold tracking-tight">Categorize my split</h3>
      <p className="mt-1 truncate text-sm text-slate-400">
        {transaction.payeeName} · my split {formatMoney(splitAmount)}
      </p>

      {loading && <p className="py-6 text-center text-sm text-slate-400">Loading…</p>}

      {!loading && loadError && (
        <div className="mt-5">
          <p className="text-sm text-red-400">Failed to load pickers.</p>
          <button
            type="button"
            onClick={() => {
              accountsQuery.refetch()
              monthQuery.refetch()
            }}
            className="mt-4 rounded-lg border border-red-800 px-3 py-1.5 text-sm font-medium text-red-400 transition hover:bg-red-900/50"
          >
            Retry
          </button>
        </div>
      )}

      {!loading && !loadError && own === undefined && (
        <p className="mt-5 text-sm text-red-400">No split found for this budget.</p>
      )}

      {!loading && !loadError && own !== undefined && (
        <>
          <div
            className={`${LINE_GRID} mt-5 border-b border-slate-800 px-2 py-2 text-xs font-semibold tracking-widest text-slate-500 uppercase`}
          >
            <span>Type</span>
            <span>Target</span>
            <span className="text-right">Outflow</span>
            <span className="text-right">Inflow</span>
            <span />
          </div>

          {lines.length === 0 && (
            <p className="border-b border-slate-800/60 px-2 py-3 text-sm text-slate-400">
              Not categorized yet. The whole split counts as outstanding.
            </p>
          )}

          {parsed.map(({ draft, out, in: inc }, index) => {
            const missingTarget = draft.targetId === null
            const badActive = outflowMode ? out <= 0 : inc <= 0
            const badOther = outflowMode ? inc !== 0 : out !== 0
            return (
              <div key={draft.key} className={`${LINE_GRID} border-b border-slate-800/60 px-2 py-2`}>
                <select
                  aria-label="Line type"
                  value={draft.kind}
                  onChange={(e) => setKind(draft.key, e.target.value as LineKind)}
                  className="w-full cursor-pointer rounded-lg border border-slate-700 bg-slate-950 px-2 py-2 text-sm text-slate-100 outline-none focus:border-emerald-400"
                >
                  <option value="spending">Spending</option>
                  <option value="transfer">Transfer</option>
                </select>
                {draft.kind === 'spending' ? (
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
                  />
                )}
                <AmountInput
                  ariaLabel="Line outflow"
                  value={draft.outflow}
                  onChange={(outflow) => updateDraft(draft.key, { outflow })}
                  invalid={badActive || badOther}
                  className="px-2 py-2 text-right tabular-nums"
                />
                <AmountInput
                  ariaLabel="Line inflow"
                  value={draft.inflow}
                  onChange={(inflow) => updateDraft(draft.key, { inflow })}
                  invalid={badActive || badOther}
                  className="px-2 py-2 text-right tabular-nums"
                />
                <button
                  type="button"
                  aria-label={`Remove line ${index + 1}`}
                  title="Remove line"
                  onClick={() => setLines((prev) => prev.filter((other) => other.key !== draft.key))}
                  className="flex cursor-pointer justify-center text-slate-500 transition hover:text-red-400"
                >
                  <XIcon className="h-4 w-4" />
                </button>
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

          {remaining > 0 && (
            <p className="mt-3 text-xs font-semibold text-slate-400">
              Uncategorized: {formatMoney(remaining)} of {formatMoney(splitAmount)}.
            </p>
          )}
          {remaining < 0 && (
            <p className="mt-3 flex items-center gap-1.5 text-xs font-semibold text-yellow-300">
              <WarningIcon className="h-3.5 w-3.5" />
              Lines ({formatMoney(lineTotal)}) exceed the split ({formatMoney(splitAmount)}).
            </p>
          )}

          {save.isError && <p className="mt-3 text-sm text-red-400">{save.error.message}</p>}

          <div className="mt-6 flex justify-end gap-2">
            <button type="button" onClick={onClose} className={CANCEL_BUTTON}>
              Cancel
            </button>
            <button
              type="button"
              onClick={() => save.mutate()}
              disabled={!canSave || save.isPending}
              className={PRIMARY_BUTTON}
            >
              {save.isPending ? 'Saving…' : 'Save'}
            </button>
          </div>
        </>
      )}
    </DialogShell>
  )
}
