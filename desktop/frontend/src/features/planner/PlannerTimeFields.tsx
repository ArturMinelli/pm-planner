import PlannerTimeInput from '../../components/PlannerTimeInput'
import { Card, Field } from '../../components/ui'
import type { PlannerTimeField, PlannerTimes } from '../../util/plannerTimes'

type PlannerTimeFieldsProps = {
  originals: string
  times: PlannerTimes
  disabled: boolean
  out2: string
  onTimeChange: (field: PlannerTimeField, value: string) => void
  onStep: (field: PlannerTimeField, delta: number) => void
  onBlurNormalize: (field: PlannerTimeField) => void
}

const TIME_FIELDS: Array<{ field: PlannerTimeField; label: string }> = [
  { field: 'in1', label: 'Entrada 1' },
  { field: 'out1', label: 'Saída 1' },
  { field: 'in2', label: 'Entrada 2' },
]

export default function PlannerTimeFields({
  originals,
  times,
  disabled,
  out2,
  onTimeChange,
  onStep,
  onBlurNormalize,
}: PlannerTimeFieldsProps) {
  return (
    <Card title="Marcações Sugeridas" stretch>
      <p className="muted small originals-line">{originals}</p>
      <div className="time-grid">
        {TIME_FIELDS.map(({ field, label }) => (
          <Field key={field} id={`planner-${field}`} label={label}>
            <PlannerTimeInput
              id={`planner-${field}`}
              name={`planner-${field}`}
              value={times[field]}
              onChange={(value) => onTimeChange(field, value)}
              onStep={(delta) => onStep(field, delta)}
              onBlurNormalize={() => onBlurNormalize(field)}
              disabled={disabled}
            />
          </Field>
        ))}
        <Field id="planner-out2" label="Saída 2 Calculada">
          <input
            id="planner-out2"
            name="planner-out2"
            readOnly
            tabIndex={-1}
            value={out2}
          />
        </Field>
      </div>
    </Card>
  )
}
