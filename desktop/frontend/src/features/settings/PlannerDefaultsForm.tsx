import PlannerTimeInput from '../../components/PlannerTimeInput'
import { Button, Card, Field } from '../../components/ui'
import {
  builtinPlannerAnchors,
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
  busy: boolean
  onAnchorsChange: (anchors: PlannerAnchorTimes) => void
  onMaxDailyExtraHoursChange: (value: string) => void
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
  busy,
  onAnchorsChange,
  onMaxDailyExtraHoursChange,
  onSave,
}: PlannerDefaultsFormProps) {
  const builtins = builtinPlannerAnchors()

  return (
    <Card
      title="Planner"
      intro={
        <>
          Âncoras usadas para atribuir batidas aos quatro campos e preencher
          Entrada 1 quando não há registro correspondente. A CLI{' '}
          <code className="pill" translate="no">
            pm plan
          </code>{' '}
          usa os mesmos valores.
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
          hint="Limita apenas a saída alternativa para pagar saldo negativo. O padrão é 3 horas; o crédito continua sendo calculado em 1,5x."
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
        <div className="btn-row">
          <Button
            type="button"
            variant="secondary"
            onClick={() => {
              onAnchorsChange(builtinPlannerAnchors())
              onMaxDailyExtraHoursChange(
                String(DEFAULT_MAX_DAILY_EXTRA_MINUTES / 60),
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
