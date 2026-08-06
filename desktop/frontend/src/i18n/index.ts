import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import ptBR from '../locales/pt-BR.json'

export const appLocale = 'pt-BR' as const
export type AppLocale = typeof appLocale

function syncDocumentLanguage() {
  document.documentElement.lang = appLocale
}

export async function initI18n() {
  if (i18n.isInitialized) {
    await i18n.changeLanguage(appLocale)
    syncDocumentLanguage()
    return i18n
  }

  await i18n.use(initReactI18next).init({
    resources: {
      'pt-BR': { translation: ptBR },
    },
    lng: appLocale,
    fallbackLng: appLocale,
    interpolation: { escapeValue: false },
  })

  syncDocumentLanguage()
  return i18n
}

export default i18n
