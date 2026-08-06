import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import en from '../locales/en.json'
import ptBR from '../locales/pt-BR.json'

export const supportedLocales = ['en', 'pt-BR'] as const
export type AppLocale = (typeof supportedLocales)[number]

export const defaultLocale: AppLocale = 'pt-BR'

function syncDocumentLanguage(locale: string) {
  document.documentElement.lang = locale
}

export async function initI18n(locale: AppLocale = defaultLocale) {
  if (i18n.isInitialized) {
    await i18n.changeLanguage(locale)
    syncDocumentLanguage(locale)
    return i18n
  }

  await i18n.use(initReactI18next).init({
    resources: {
      en: { translation: en },
      'pt-BR': { translation: ptBR },
    },
    lng: locale,
    fallbackLng: defaultLocale,
    interpolation: { escapeValue: false },
  })

  syncDocumentLanguage(locale)
  return i18n
}

export function isAppLocale(value: string): value is AppLocale {
  return supportedLocales.includes(value as AppLocale)
}

export default i18n
