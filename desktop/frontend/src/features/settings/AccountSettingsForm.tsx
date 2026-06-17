import { Button, Card, Field, Stack } from '../../components/ui'
import { loginCredentialFieldMeta } from '../../util/loginCredential'

type AccountSettingsFormProps = {
  login: string
  password: string
  ttl: string
  busy: boolean
  canTest: boolean
  onLoginChange: (value: string) => void
  onPasswordChange: (value: string) => void
  onTTLChange: (value: string) => void
  onSave: () => void
  onTest: () => void
}

export default function AccountSettingsForm({
  login,
  password,
  ttl,
  busy,
  canTest,
  onLoginChange,
  onPasswordChange,
  onTTLChange,
  onSave,
  onTest,
}: AccountSettingsFormProps) {
  const loginMeta = loginCredentialFieldMeta(login)

  return (
    <Card title="Conta e Sessão">
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
            label="Login (e-mail ou CPF)"
            hint={loginMeta.hint}
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
          <Field id="settings-password" label="Senha">
            <input
              id="settings-password"
              name="password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(event) => onPasswordChange(event.target.value)}
            />
          </Field>
          <Field
            id="settings-cache-ttl"
            label="Cache TTL em Horas"
            hint={
              <>
                Quando a API não envia validade da sessão, limita há quantas
                horas o arquivo{' '}
                <code className="pill" translate="no">
                  session.json
                </code>{' '}
                é reutilizado.
              </>
            }
          >
            <input
              id="settings-cache-ttl"
              name="cache_ttl_hours"
              type="number"
              min="1"
              inputMode="numeric"
              autoComplete="off"
              placeholder="8..."
              value={ttl}
              onChange={(event) => onTTLChange(event.target.value)}
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
            {busy ? 'Salvando...' : 'Salvar Conta'}
          </Button>
          <Button
            type="button"
            variant="secondary"
            onClick={onTest}
            disabled={busy || !canTest}
            aria-busy={busy}
          >
            Testar Login
          </Button>
        </div>
      </form>
    </Card>
  )
}
