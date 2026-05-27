export function formatDurationSecs(totalSecs: number): string {
  const s =
    typeof totalSecs === 'number' && Number.isFinite(totalSecs)
      ? Math.max(0, Math.floor(totalSecs))
      : 0
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const sec = s % 60
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(h)}:${pad(m)}:${pad(sec)}`
}

export function localDateYYYYMMDD(d = new Date()): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

const ymdParts = /^(\d{4})-(\d{2})-(\d{2})$/

/** Parses a local calendar date from `YYYY-MM-DD` (avoids UTC interpretation issues). */
export function parseYYYYMMDD(iso: string): Date | undefined {
  const m = ymdParts.exec(iso.trim())
  if (!m) return undefined
  const y = Number(m[1])
  const mo = Number(m[2])
  const d = Number(m[3])
  if (mo < 1 || mo > 12 || d < 1 || d > 31) return undefined
  const dt = new Date(y, mo - 1, d)
  if (
    dt.getFullYear() !== y ||
    dt.getMonth() !== mo - 1 ||
    dt.getDate() !== d
  ) {
    return undefined
  }
  return dt
}

/** Same shape as {@link localDateYYYYMMDD}; use for any `Date` from the picker. */
export function formatYYYYMMDD(d: Date): string {
  return localDateYYYYMMDD(d)
}
