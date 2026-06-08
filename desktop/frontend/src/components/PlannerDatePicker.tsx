import {
  useCallback,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
} from 'react'
import { formatYYYYMMDD, parseYYYYMMDD } from '../util/timeFormat'

export type PlannerDatePickerProps = {
  value: string
  onChange: (isoYYYYMMDD: string) => void
  disabled?: boolean
  autoFocus?: boolean
  id?: string
  className?: string
}

type CalendarDay = {
  date: Date
  outside: boolean
}

const longDatePtBr = new Intl.DateTimeFormat('pt-BR', { dateStyle: 'long' })
const monthTitlePtBr = new Intl.DateTimeFormat('pt-BR', {
  month: 'long',
  year: 'numeric',
})
const weekdayShortPtBr = new Intl.DateTimeFormat('pt-BR', {
  weekday: 'short',
})

const WEEKDAYS = Array.from({ length: 7 }, (_, index) => {
  const sunday = new Date(2024, 0, 7 + index)
  return weekdayShortPtBr.format(sunday).replace('.', '')
})

function startOfMonth(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), 1)
}

function addDays(date: Date, days: number): Date {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate() + days)
}

function addMonths(date: Date, months: number): Date {
  return new Date(date.getFullYear(), date.getMonth() + months, 1)
}

function sameDay(a?: Date, b?: Date): boolean {
  return (
    !!a &&
    !!b &&
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  )
}

function formatMonthTitle(date: Date): string {
  const title = monthTitlePtBr.format(date)
  return title.charAt(0).toLocaleUpperCase('pt-BR') + title.slice(1)
}

function buildCalendarDays(month: Date): CalendarDay[] {
  const first = startOfMonth(month)
  const gridStart = addDays(first, -first.getDay())
  return Array.from({ length: 42 }, (_, index) => {
    const date = addDays(gridStart, index)
    return {
      date,
      outside: date.getMonth() !== first.getMonth(),
    }
  })
}

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
  const focusedDayRef = useRef<HTMLButtonElement>(null)
  const parsed = useMemo(() => parseYYYYMMDD(value), [value])
  const today = useMemo(() => new Date(), [])
  const [open, setOpen] = useState(false)
  const [month, setMonth] = useState(() => startOfMonth(parsed ?? today))
  const [focusedDate, setFocusedDate] = useState(() => parsed ?? today)

  const triggerLabel = useMemo(() => {
    if (!parsed) return value
    return longDatePtBr.format(parsed)
  }, [parsed, value])

  const days = useMemo(() => buildCalendarDays(month), [month])

  const close = useCallback(() => {
    setOpen(false)
    requestAnimationFrame(() => triggerRef.current?.focus())
  }, [])

  const openCalendar = useCallback(() => {
    const nextFocus = parsed ?? today
    setMonth(startOfMonth(nextFocus))
    setFocusedDate(nextFocus)
    setOpen(true)
  }, [parsed, today])

  const selectDate = useCallback(
    (date: Date) => {
      onChange(formatYYYYMMDD(date))
      close()
    },
    [close, onChange],
  )

  const moveFocus = useCallback((date: Date) => {
    setFocusedDate(date)
    setMonth(startOfMonth(date))
  }, [])

  useEffect(() => {
    if (autoFocus) triggerRef.current?.focus()
  }, [autoFocus])

  useEffect(() => {
    if (!open) return
    const frame = window.requestAnimationFrame(() => focusedDayRef.current?.focus())
    return () => window.cancelAnimationFrame(frame)
  }, [open, focusedDate, month])

  useEffect(() => {
    if (!open) return
    const handler = (event: KeyboardEvent) => {
      if (event.key === 'Escape') close()
    }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [open, close])

  useEffect(() => {
    if (!open) return
    const handler = (event: PointerEvent) => {
      const el = wrapRef.current
      if (el && !el.contains(event.target as Node)) close()
    }
    document.addEventListener('pointerdown', handler, true)
    return () => document.removeEventListener('pointerdown', handler, true)
  }, [open, close])

  useEffect(() => {
    if (!disabled) return
    const frame = window.requestAnimationFrame(() => close())
    return () => window.cancelAnimationFrame(frame)
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
          if (disabled) return
          if (open) close()
          else openCalendar()
        }}
        onKeyDown={(event) => {
          if (disabled || open) return
          if (event.key === 'ArrowDown' || event.key === 'Enter' || event.key === ' ') {
            event.preventDefault()
            openCalendar()
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
          <div className="calendar-header">
            <button
              type="button"
              className="icon-btn"
              aria-label="Mês anterior"
              title="Mês anterior"
              onClick={() => setMonth((current) => addMonths(current, -1))}
            >
              ‹
            </button>
            <div className="calendar-title">{formatMonthTitle(month)}</div>
            <button
              type="button"
              className="icon-btn"
              aria-label="Próximo mês"
              title="Próximo mês"
              onClick={() => setMonth((current) => addMonths(current, 1))}
            >
              ›
            </button>
          </div>
          <div
            className="calendar-grid"
            role="grid"
            aria-label={formatMonthTitle(month)}
            onKeyDown={(event) => {
              if (event.key === 'Escape') {
                event.preventDefault()
                close()
                return
              }
              if (event.key === 'Enter' || event.key === ' ') {
                event.preventDefault()
                selectDate(focusedDate)
                return
              }

              const moves: Record<string, Date> = {
                ArrowLeft: addDays(focusedDate, -1),
                ArrowRight: addDays(focusedDate, 1),
                ArrowUp: addDays(focusedDate, -7),
                ArrowDown: addDays(focusedDate, 7),
                Home: addDays(focusedDate, -focusedDate.getDay()),
                End: addDays(focusedDate, 6 - focusedDate.getDay()),
                PageUp: addMonths(focusedDate, -1),
                PageDown: addMonths(focusedDate, 1),
              }
              const next = moves[event.key]
              if (next) {
                event.preventDefault()
                moveFocus(next)
              }
            }}
          >
            {WEEKDAYS.map((weekday) => (
              <div key={weekday} className="calendar-weekday" role="columnheader">
                {weekday}
              </div>
            ))}
            {days.map(({ date, outside }) => {
              const selected = sameDay(date, parsed)
              const focused = sameDay(date, focusedDate)
              return (
                <button
                  key={formatYYYYMMDD(date)}
                  ref={focused ? focusedDayRef : undefined}
                  type="button"
                  className={[
                    'calendar-day',
                    outside && 'outside',
                    selected && 'selected',
                    sameDay(date, today) && 'today',
                  ]
                    .filter(Boolean)
                    .join(' ')}
                  role="gridcell"
                  aria-selected={selected}
                  tabIndex={focused ? 0 : -1}
                  onFocus={() => setFocusedDate(date)}
                  onClick={() => selectDate(date)}
                >
                  {date.getDate()}
                </button>
              )
            })}
          </div>
        </div>
      ) : null}
    </div>
  )
}
