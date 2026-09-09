type BudgetMonthErrorProps = {
  message: string
  onRetry: () => void
}

export default function BudgetMonthError({ message, onRetry }: BudgetMonthErrorProps) {
  return (
    <div className="mt-6 rounded-xl border border-red-900 bg-red-950/40 p-6 text-sm text-red-400">
      <p>{message}</p>
      <button
        type="button"
        onClick={onRetry}
        className="mt-4 rounded-lg border border-red-800 px-3 py-1.5 font-medium transition hover:bg-red-900/50"
      >
        Retry
      </button>
    </div>
  )
}
