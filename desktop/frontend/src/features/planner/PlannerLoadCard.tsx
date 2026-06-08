import PlannerDatePicker from '../../components/PlannerDatePicker'
import { Button, Card, Field } from '../../components/ui'

type PlannerLoadCardProps = {
  date: string
  busy: boolean
  onDateChange: (date: string) => void
  onLoad: () => void
}

export default function PlannerLoadCard({
  date,
  busy,
  onDateChange,
  onLoad,
}: PlannerLoadCardProps) {
  return (
    <Card>
      <div className="row wrap">
        <Field id="planner-date" label="Data">
          <PlannerDatePicker
            id="planner-date"
            value={date}
            onChange={onDateChange}
            disabled={busy}
            autoFocus
          />
        </Field>
        <Button
          variant="primary"
          onClick={onLoad}
          disabled={busy}
          aria-busy={busy}
        >
          {busy ? 'Carregando...' : 'Carregar Dia'}
        </Button>
      </div>
    </Card>
  )
}
