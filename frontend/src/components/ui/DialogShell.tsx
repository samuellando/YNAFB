import { useEffect, type ReactNode } from 'react'

type DialogShellProps = {
  onClose: () => void
  wide?: boolean
  // Allow absolutely-positioned children (e.g. picker dropdowns) to overflow
  // the dialog instead of being clipped. Only use for short dialogs that fit
  // the viewport; tall dialogs still need internal scrolling.
  overflowVisible?: boolean
  children: ReactNode
}

export default function DialogShell({
  onClose,
  wide = false,
  overflowVisible = false,
  children,
}: DialogShellProps) {
  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key !== 'Escape') return
      const target = event.target as HTMLElement | null
      if (target?.closest?.('[data-escape-stop]')) return
      onClose()
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [onClose])

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onClick={onClose}
    >
      <div
        role="dialog"
        aria-modal="true"
        className={`max-h-[calc(100svh-2rem)] w-full rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl ${
          overflowVisible ? 'overflow-visible' : 'overflow-y-auto'
        } ${wide ? 'max-w-xl' : 'max-w-sm'}`}
        onClick={(e) => e.stopPropagation()}
      >
        {children}
      </div>
    </div>
  )
}
