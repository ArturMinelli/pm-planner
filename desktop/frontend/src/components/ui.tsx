import type {
  ButtonHTMLAttributes,
  HTMLAttributes,
  ReactNode,
} from 'react'
import { useTranslation } from 'react-i18next'

type BannerTone = 'info' | 'success' | 'error'
type ButtonVariant = 'primary' | 'secondary'
type ToastTone = 'success' | 'error'

function classNames(...values: Array<string | false | null | undefined>) {
  return values.filter(Boolean).join(' ')
}

export function Page({
  children,
  narrow = false,
}: {
  children: ReactNode
  narrow?: boolean
}) {
  return (
    <section className={classNames('page', narrow && 'narrow')}>
      {children}
    </section>
  )
}

export function PageHeader({
  title,
  description,
  actions,
}: {
  title: string
  description?: ReactNode
  actions?: ReactNode
}) {
  return (
    <header className="page-header">
      <div className="page-header-row">
        <div className="page-header-text">
          <h1>{title}</h1>
          {description ? <p className="muted">{description}</p> : null}
        </div>
        {actions ? <div className="page-header-actions">{actions}</div> : null}
      </div>
    </header>
  )
}

export function Card({
  title,
  intro,
  stretch = false,
  collapsible = false,
  defaultOpen = false,
  children,
}: {
  title?: string
  intro?: ReactNode
  stretch?: boolean
  /** Shows only the title with a chevron; the content opens on click. Needs a title. */
  collapsible?: boolean
  defaultOpen?: boolean
  children: ReactNode
}) {
  if (collapsible && title) {
    return (
      <details className={classNames('card', 'collapsible', stretch && 'stretch')} open={defaultOpen}>
        <summary className="card-summary">
          <h2 className="card-title">{title}</h2>
          <svg
            className="card-chevron"
            width="16"
            height="16"
            viewBox="0 0 16 16"
            aria-hidden="true"
            focusable="false"
          >
            <path
              d="M3.5 6l4.5 4.5L12.5 6"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.6"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </summary>
        <div className="card-body">
          {intro ? <p className="muted card-intro">{intro}</p> : null}
          {children}
        </div>
      </details>
    )
  }

  return (
    <section className={classNames('card', stretch && 'stretch')}>
      {title ? <h2 className="card-title">{title}</h2> : null}
      {intro ? <p className="muted card-intro">{intro}</p> : null}
      {children}
    </section>
  )
}

export function Banner({
  tone = 'info',
  children,
}: {
  tone?: BannerTone
  children: ReactNode
}) {
  const role = tone === 'error' ? 'alert' : 'status'
  return (
    <div
      className={classNames('banner', tone)}
      role={role}
      aria-live={tone === 'error' ? 'assertive' : 'polite'}
    >
      {children}
    </div>
  )
}

export function Toast({
  tone,
  children,
  onDismiss,
}: {
  tone: ToastTone
  children: ReactNode
  onDismiss?: () => void
}) {
  const { t } = useTranslation()
  const role = tone === 'error' ? 'alert' : 'status'
  return (
    <div
      className={classNames('toast', tone)}
      role={role}
      aria-live={tone === 'error' ? 'assertive' : 'polite'}
    >
      <div className="toast-body">{children}</div>
      {onDismiss ? (
        <button
          type="button"
          className="toast-dismiss"
          aria-label={t('common.closeNotice')}
          onClick={onDismiss}
        >
          x
        </button>
      ) : null}
    </div>
  )
}

export function Field({
  id,
  label,
  hint,
  className,
  children,
}: {
  id: string
  label: string
  hint?: ReactNode
  className?: string
  children: ReactNode
}) {
  return (
    <div className={classNames('field', className)}>
      <label htmlFor={id}>{label}</label>
      {children}
      {hint ? <small className="hint">{hint}</small> : null}
    </div>
  )
}

export function Button({
  variant = 'primary',
  className,
  children,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant
}) {
  return (
    <button
      {...props}
      type={props.type ?? 'button'}
      className={classNames('btn', variant, className)}
    >
      {children}
    </button>
  )
}

export function StatRow({
  label,
  value,
  accent = false,
  emphasis = false,
}: {
  label: string
  value: ReactNode
  accent?: boolean
  emphasis?: boolean
}) {
  return (
    <div className={classNames('stat', accent && 'accent', emphasis && 'emphasis')}>
      <span className="stat-label">{label}</span>
      <span className="stat-value">{value}</span>
    </div>
  )
}

export function Stack({
  className,
  children,
}: HTMLAttributes<HTMLDivElement>) {
  return <div className={classNames('stack', className)}>{children}</div>
}
