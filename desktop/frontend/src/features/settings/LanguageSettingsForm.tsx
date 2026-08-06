import { useTranslation } from 'react-i18next'
import type { AppLocale } from '../../types'
import { Button, Card, Field } from '../../components/ui'
import * as backend from '../../services/backend'
import i18n, { isAppLocale } from '../../i18n'

type LanguageSettingsFormProps = {
  locale: AppLocale
  busy: boolean
  onLocaleChange: (locale: AppLocale) => void
  onSaved: () => void
  onError: (error: unknown) => void
}

export default function LanguageSettingsForm({
  locale,
  busy,
  onLocaleChange,
  onSaved,
  onError,
}: LanguageSettingsFormProps) {
  const { t } = useTranslation()

  const saveLocale = async () => {
    try {
      await backend.mergeAndSave({ locale })
      await i18n.changeLanguage(locale)
      document.documentElement.lang = locale
      onSaved()
    } catch (error) {
      onError(error)
    }
  }

  return (
    <Card title={t('settings.language.title')} intro={t('settings.language.description')}>
      <form
        className="form-stack"
        onSubmit={(event) => {
          event.preventDefault()
          void saveLocale()
        }}
      >
        <Field id="settings-locale" label={t('settings.language.title')}>
          <select
            id="settings-locale"
            name="locale"
            value={locale}
            onChange={(event) => {
              const nextLocale = event.target.value
              if (!isAppLocale(nextLocale)) return
              onLocaleChange(nextLocale)
            }}
          >
            <option value="en">{t('settings.language.english')}</option>
            <option value="pt-BR">{t('settings.language.portuguese')}</option>
          </select>
        </Field>
        <div className="btn-row">
          <Button type="submit" variant="primary" disabled={busy} aria-busy={busy}>
            {busy ? t('common.saving') : t('settings.language.save')}
          </Button>
        </div>
      </form>
    </Card>
  )
}
