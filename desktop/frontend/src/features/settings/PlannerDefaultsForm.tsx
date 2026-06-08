import PlannerTimeInput from '../../components/PlannerTimeInput'
import { Button, Card, Field } from '../../components/ui'
import { builtinPlannerAnchors } from '../../util/plannerDefaults'
import {
  adjustPlannerAnchor,
  type PlannerAnchorTimes,
} from '../../util/plannerTimes'

type AnchorField = keyof PlannerAnchorTimes

type PlannerDefaultsFormProps = {
  anchors: PlannerAnchorTimes
  busy: boolean
  onAnchorsChange: (anchors: PlannerAnchorTimes) => void
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
  busy,
  onAnchorsChange,
  onSave,
}: PlannerDefaultsFormProps) {
  const builtins = builtinPlannerAnchors()

  return (
    <Card
      title="Horários Padrão do Planner"
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
        <div className="btn-row">
          <Button
            type="button"
            variant="secondary"
            onClick={() => onAnchorsChange(builtinPlannerAnchors())}
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
            {busy ? 'Salvando...' : 'Salvar Horários'}
          </Button>
        </div>
      </form>
    </Card>
  )
}
