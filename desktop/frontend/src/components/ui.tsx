import type {
  ButtonHTMLAttributes,
  HTMLAttributes,
  ReactNode,
} from 'react'

type BannerTone = 'info' | 'success' | 'error'
type ButtonVariant = 'primary' | 'secondary'
type ToastTone = 'success' | 'error'

function classNames(...values: Array<string | false | null | undefined>) {
  return values.filter(Boolean).join(' ')
}

export function Page({ children }: { children: ReactNode }) {
  return <section className="page">{children}</section>
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
  children,
}: {
  title?: string
  intro?: ReactNode
  stretch?: boolean
  children: ReactNode
}) {
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
          aria-label="Fechar aviso"
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
}: {
  label: string
  value: ReactNode
  accent?: boolean
}) {
  return (
    <div className={classNames('stat', accent && 'accent')}>
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
