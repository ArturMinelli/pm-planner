import type { JSX } from 'react'

export type PlannerTimeInputProps = {
  value: string
  onChange: (next: string) => void
  onStep?: (deltaMinutes: number) => void
  onBlurNormalize?: () => void
  disabled?: boolean
  id?: string
  name?: string
  placeholder?: string
  className?: string
}

/** Normalize to HH:MM display from pasted or typed raw text (max 4 digits). */
function digitsToMaskedTime(raw: string): string {
  const digits = raw.replace(/\D/g, '').slice(0, 4)
  if (digits.length === 0) return ''
  if (digits.length <= 2) return digits
  return `${digits.slice(0, 2)}:${digits.slice(2)}`
}

export default function PlannerTimeInput({
  value,
  onChange,
  onStep,
  onBlurNormalize,
  disabled,
  id,
  name,
  placeholder = '08:00…',
  className,
}: PlannerTimeInputProps): JSX.Element {
  return (
    <input
      id={id}
      name={name}
      type="text"
      inputMode="numeric"
      autoComplete="off"
      spellCheck={false}
      maxLength={5}
      className={className}
      placeholder={placeholder}
      disabled={disabled}
      value={value}
      onChange={(e) => onChange(digitsToMaskedTime(e.target.value))}
      onBlur={() => onBlurNormalize?.()}
      onKeyDown={(e) => {
        if (!onStep) return
        if (e.key === 'ArrowUp') {
          e.preventDefault()
          onStep(15)
        } else if (e.key === 'ArrowDown') {
          e.preventDefault()
          onStep(-15)
        }
      }}
    />
  )
}
