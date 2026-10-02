import { useQuery } from '@tanstack/react-query'
import { getShareDetail } from '../../lib/api/expenseShare'
import EntityPicker, { type EntityOption } from '../ui/EntityPicker'

type ShareLineTargetProps = {
  budgetId: number
  kind: 'split' | 'settlement'
  shareOptions: EntityOption[]
  shareId: number | null
  memberBudgetId: number | null
  invalid: boolean
  onShareChange: (shareId: number | null) => void
  onMemberChange: (budgetId: number | null) => void
}

// Share + member pickers for one split/settlement line draft. Members come
// from the selected share's detail (own budget excluded); wire IDs are member
// budget ids, matching the line create/update contract.
export default function ShareLineTarget({
  budgetId,
  kind,
  shareOptions,
  shareId,
  memberBudgetId,
  invalid,
  onShareChange,
  onMemberChange,
}: ShareLineTargetProps) {
  const detail = useQuery({
    queryKey: ['expense-share', budgetId, shareId],
    queryFn: () => getShareDetail(budgetId, shareId ?? 0),
    enabled: shareId !== null,
  })

  const memberOptions: EntityOption[] = (detail.data?.members ?? [])
    .filter((m) => m.budgetId !== budgetId)
    .map((m) => ({ id: m.budgetId, name: m.displayName }))

  return (
    <div className="flex min-w-0 flex-col gap-1.5">
      <EntityPicker
        hideLabel
        label="Share"
        options={shareOptions}
        value={shareId}
        onChange={onShareChange}
        placeholder="Select share"
        invalid={invalid && shareId === null}
      />
      <EntityPicker
        hideLabel
        label={kind === 'split' ? 'Tagged member' : 'Counterparty'}
        options={memberOptions}
        value={memberBudgetId}
        onChange={onMemberChange}
        placeholder={shareId === null ? 'Pick a share first' : 'Select member'}
        disabled={shareId === null}
        invalid={invalid && memberBudgetId === null}
      />
    </div>
  )
}
