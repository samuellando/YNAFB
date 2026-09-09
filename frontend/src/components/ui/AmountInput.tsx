import { centsFromInput, centsToInput, signedCentsFromInput } from '../../lib/money'

type AmountInputProps = {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  ariaLabel?: string
  signed?: boolean
  invalid?: boolean
  className?: string
  inputRef?: React.Ref<HTMLInputElement>
}

export default function AmountInput({
  value,
  onChange,
  placeholder = '0.00',
  ariaLabel,
  signed = false,
  invalid = false,
  className = '',
  inputRef,
}: AmountInputProps) {
  function resolve() {
    const cents = signed ? signedCentsFromInput(value) : centsFromInput(value)
    if (cents === null) return
    const resolved = centsToInput(cents)
    if (resolved !== value) onChange(resolved)
  }

  return (
    <input
      ref={inputRef}
      type="text"
      inputMode="decimal"
      value={value}
      aria-label={ariaLabel}
      placeholder={placeholder}
      autoComplete="off"
      spellCheck={false}
      onChange={(e) => onChange(e.target.value)}
      onBlur={resolve}
      onKeyDown={(e) => {
        if (e.key === 'Enter') {
          e.preventDefault()
          resolve()
        }
      }}
      className={`rounded-lg border bg-slate-950 text-sm text-slate-100 outline-none focus:border-emerald-400 ${
        invalid ? 'border-red-800' : 'border-slate-700'
      } ${className}`}
    />
  )
}
