import PlannerTimeInput from '../../components/PlannerTimeInput'
import { Card, Field } from '../../components/ui'
import type { BalanceAdjustment } from '../../types'
import type { PlannerTimeField, PlannerTimes } from '../../util/plannerTimes'
import PlannerClockoutField from './PlannerClockoutField'

type PlannerTimeFieldsProps = {
  originals: string
  times: PlannerTimes
  disabled: boolean
  out2: string
  alternativeOut2?: string
  balance?: BalanceAdjustment
  balanceError?: string
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
  alternativeOut2,
  balance,
  balanceError,
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
        <PlannerClockoutField
          out2={out2}
          alternativeOut2={alternativeOut2}
          balance={balance}
          balanceError={balanceError}
        />
      </div>
    </Card>
  )
}
