import { useTranslation } from 'react-i18next'
import { Button, Card, Field, Stack } from '../../components/ui'
import { loginCredentialFieldMeta } from '../../util/loginCredential'

type AccountSettingsFormProps = {
  login: string
  password: string
  busy: boolean
  canTest: boolean
  onLoginChange: (value: string) => void
  onPasswordChange: (value: string) => void
  onSave: () => void
  onTest: () => void
}

export default function AccountSettingsForm({
  login,
  password,
  busy,
  canTest,
  onLoginChange,
  onPasswordChange,
  onSave,
  onTest,
}: AccountSettingsFormProps) {
  const { t } = useTranslation()
  const loginMeta = loginCredentialFieldMeta(login)

  return (
    <Card title={t('settings.account.title')}>
      <form
        className="form-stack"
        onSubmit={(event) => {
          event.preventDefault()
          onSave()
        }}
      >
        <Stack>
          <Field
            id="settings-login"
            label={t('settings.account.loginLabel')}
            hint={loginMeta.hintKey ? t(loginMeta.hintKey) : undefined}
          >
            <input
              id="settings-login"
              name="login"
              type="text"
              inputMode={loginMeta.inputMode}
              autoComplete="username"
              spellCheck={false}
              value={login}
              onChange={(event) => onLoginChange(event.target.value)}
            />
          </Field>
          <Field id="settings-password" label={t('settings.account.passwordLabel')}>
            <input
              id="settings-password"
              name="password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(event) => onPasswordChange(event.target.value)}
            />
          </Field>
        </Stack>
        <div className="btn-row">
          <Button
            type="submit"
            variant="primary"
            disabled={busy}
            aria-busy={busy}
          >
            {busy ? t('common.saving') : t('settings.account.save')}
          </Button>
          <Button
            type="button"
            variant="secondary"
            onClick={onTest}
            disabled={busy || !canTest}
            aria-busy={busy}
          >
            {t('settings.account.test')}
          </Button>
        </div>
      </form>
    </Card>
  )
}
