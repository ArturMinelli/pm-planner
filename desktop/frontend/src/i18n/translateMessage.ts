import type { TFunction } from 'i18next'
import type { LocalizedMessage } from '../types'

export function translateMessage(
  translate: TFunction,
  message?: LocalizedMessage | null,
): string | undefined {
  if (!message?.key) return undefined
  return translate(message.key, message.params ?? {})
}

export function translateMessageOr(
  translate: TFunction,
  message: LocalizedMessage | null | undefined,
  fallback: string,
): string {
  return translateMessage(translate, message) ?? fallback
}

export function messageKey(message: LocalizedMessage): string {
  const params = message.params
    ? Object.entries(message.params)
        .sort(([left], [right]) => left.localeCompare(right))
        .map(([key, value]) => `${key}=${value}`)
        .join(',')
    : ''
  return params ? `${message.key}:${params}` : message.key
}
