import {
  useCallback,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
} from 'react'
import { ptBR } from 'date-fns/locale/pt-BR'
import { DayPicker } from 'react-day-picker'
import { formatYYYYMMDD, parseYYYYMMDD } from '../util/timeFormat'

export type PlannerDatePickerProps = {
  value: string
  onChange: (isoYYYYMMDD: string) => void
  disabled?: boolean
  autoFocus?: boolean
  id?: string
  className?: string
}

const longDatePtBr = new Intl.DateTimeFormat('pt-BR', { dateStyle: 'long' })

export default function PlannerDatePicker({
  value,
  onChange,
  disabled,
  autoFocus,
  id: idProp,
  className,
}: PlannerDatePickerProps) {
  const genId = useId()
  const baseId = idProp ?? genId
  const dialogId = `${baseId}-calendar`

  const wrapRef = useRef<HTMLDivElement>(null)
  const triggerRef = useRef<HTMLButtonElement>(null)
  const [open, setOpen] = useState(false)

  const parsed = useMemo(() => parseYYYYMMDD(value), [value])
  const triggerLabel = useMemo(() => {
    if (!parsed) return value
    return longDatePtBr.format(parsed)
  }, [parsed, value])

  const close = useCallback(() => {
    setOpen(false)
    requestAnimationFrame(() => triggerRef.current?.focus())
  }, [])

  useEffect(() => {
    if (autoFocus) triggerRef.current?.focus()
  }, [autoFocus])

  useEffect(() => {
    if (!open) return
    const handler = (e: PointerEvent) => {
      const el = wrapRef.current
      if (el && !el.contains(e.target as Node)) close()
    }
    document.addEventListener('pointerdown', handler, true)
    return () => document.removeEventListener('pointerdown', handler, true)
  }, [open, close])

  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') close()
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [open, close])

  useEffect(() => {
    if (disabled) close()
  }, [disabled, close])

  return (
    <div
      ref={wrapRef}
      className={['planner-date-picker-field', className].filter(Boolean).join(' ')}
    >
      <button
        ref={triggerRef}
        id={baseId}
        type="button"
        className="planner-date-picker-trigger"
        aria-haspopup="dialog"
        aria-expanded={open}
        aria-controls={dialogId}
        disabled={disabled}
        onClick={() => {
          if (!disabled) setOpen((o) => !o)
        }}
        onKeyDown={(e) => {
          if (disabled || open) return
          if (e.key === 'ArrowDown') {
            e.preventDefault()
            setOpen(true)
          }
        }}
      >
        {triggerLabel}
      </button>
      {open ? (
        <div
          id={dialogId}
          className="planner-datepicker-popover"
          role="dialog"
          aria-label="Escolher data"
        >
          <DayPicker
            mode="single"
            locale={ptBR}
            autoFocus
            selected={parsed}
            defaultMonth={parsed ?? new Date()}
            onSelect={(d) => {
              if (d) {
                onChange(formatYYYYMMDD(d))
                close()
              }
            }}
          />
        </div>
      ) : null}
    </div>
  )
}
