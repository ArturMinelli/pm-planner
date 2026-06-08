import { Button, Card, Field, Stack } from '../../components/ui'

type AccountSettingsFormProps = {
  email: string
  password: string
  ttl: string
  busy: boolean
  canTest: boolean
  onEmailChange: (value: string) => void
  onPasswordChange: (value: string) => void
  onTTLChange: (value: string) => void
  onSave: () => void
  onTest: () => void
}

export default function AccountSettingsForm({
  email,
  password,
  ttl,
  busy,
  canTest,
  onEmailChange,
  onPasswordChange,
  onTTLChange,
  onSave,
  onTest,
}: AccountSettingsFormProps) {
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
          <Field id="settings-email" label="E-mail de Login">
            <input
              id="settings-email"
              name="email"
              type="email"
              autoComplete="username email"
              spellCheck={false}
              value={email}
              onChange={(event) => onEmailChange(event.target.value)}
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
