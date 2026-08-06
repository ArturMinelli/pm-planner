import PlannerTimeInput from '../../components/PlannerTimeInput'
import { Button, Card, Field } from '../../components/ui'
import {
  builtinPlannerAnchors,
  DEFAULT_BALANCE_CREDIT_MULTIPLIER,
  DEFAULT_MAX_DAILY_EXTRA_MINUTES,
} from '../../util/plannerDefaults'
import {
  adjustPlannerAnchor,
  type PlannerAnchorTimes,
} from '../../util/plannerTimes'

type AnchorField = keyof PlannerAnchorTimes

type PlannerDefaultsFormProps = {
  anchors: PlannerAnchorTimes
  maxDailyExtraHours: string
  balanceCreditMultiplier: string
  busy: boolean
  onAnchorsChange: (anchors: PlannerAnchorTimes) => void
  onMaxDailyExtraHoursChange: (value: string) => void
  onBalanceCreditMultiplierChange: (value: string) => void
  onSave: () => void
}

const ANCHOR_FIELDS: Array<{ field: AnchorField; label: string }> = [
  { field: 'in1', label: 'Entrada 1' },
  { field: 'out1', label: 'Saída 1' },
  { field: 'in2', label: 'Entrada 2' },
  { field: 'out2', label: 'Saída 2' },
]

export default function PlannerDefaultsForm({
  anchors,
  maxDailyExtraHours,
  balanceCreditMultiplier,
  busy,
  onAnchorsChange,
  onMaxDailyExtraHoursChange,
  onBalanceCreditMultiplierChange,
  onSave,
}: PlannerDefaultsFormProps) {
  const builtins = builtinPlannerAnchors()

  return (
    <Card
      title="Planner"
      intro={
        <>
          Âncoras usadas para atribuir batidas aos quatro campos e preencher
          Entrada 1 quando não há registro correspondente.
        </>
      }
    >
      <form
        className="form-stack"
        onSubmit={(event) => {
          event.preventDefault()
          onSave()
        }}
      >
        <div className="planner-anchor-grid">
          {ANCHOR_FIELDS.map(({ field, label }) => (
            <Field key={field} id={`settings-anchor-${field}`} label={label}>
              <PlannerTimeInput
                id={`settings-anchor-${field}`}
                name={`planner_${field}`}
                value={anchors[field]}
                onChange={(value) => onAnchorsChange({ ...anchors, [field]: value })}
                onStep={(delta) =>
                  onAnchorsChange(adjustPlannerAnchor(anchors, field, delta))
                }
                onBlurNormalize={() => onAnchorsChange(anchors)}
              />
            </Field>
          ))}
        </div>
        <small className="hint">
          Padrão de fábrica: {builtins.in1}, {builtins.out1}, {builtins.in2},{' '}
          {builtins.out2}. Intervalo mínimo de 15 minutos entre horários
          consecutivos.
        </small>
        <Field
          id="settings-max-daily-extra"
          label="Limite Diário de Trabalho Extra (Horas)"
          hint="Limita apenas a saída alternativa para pagar saldo negativo. O padrão é 3 horas."
        >
          <input
            id="settings-max-daily-extra"
            name="max_daily_extra_hours"
            type="number"
            min="0.02"
            max="24"
            step="any"
            inputMode="decimal"
            autoComplete="off"
            value={maxDailyExtraHours}
            onChange={(event) => onMaxDailyExtraHoursChange(event.target.value)}
          />
        </Field>
        <Field
          id="settings-balance-credit-multiplier"
          label="Multiplicador de Crédito do Banco de Horas"
          hint="Taxa de crédito para saldo negativo: 1 hora trabalhada extra gera este múltiplo em crédito. Padrão: 1,5x. Intervalo: 1,0 a 3,0."
        >
          <input
            id="settings-balance-credit-multiplier"
            name="balance_credit_multiplier"
            type="number"
            min="1"
            max="3"
            step="0.1"
            inputMode="decimal"
            autoComplete="off"
            value={balanceCreditMultiplier}
            onChange={(event) => onBalanceCreditMultiplierChange(event.target.value)}
          />
        </Field>
        <div className="btn-row">
          <Button
            type="button"
            variant="secondary"
            onClick={() => {
              onAnchorsChange(builtinPlannerAnchors())
              onMaxDailyExtraHoursChange(
                String(DEFAULT_MAX_DAILY_EXTRA_MINUTES / 60),
              )
              onBalanceCreditMultiplierChange(
                String(DEFAULT_BALANCE_CREDIT_MULTIPLIER),
              )
            }}
            disabled={busy}
          >
            Restaurar Padrões
          </Button>
          <Button
            type="submit"
            variant="primary"
            disabled={busy}
            aria-busy={busy}
          >
            {busy ? 'Salvando...' : 'Salvar Planner'}
          </Button>
        </div>
      </form>
    </Card>
  )
}
