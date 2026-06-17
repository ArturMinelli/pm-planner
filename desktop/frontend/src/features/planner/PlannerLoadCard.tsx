import PlannerDatePicker from '../../components/PlannerDatePicker'
import { Button, Card, Field } from '../../components/ui'

type PlannerLoadCardProps = {
  date: string
  busy: boolean
  embedded?: boolean
  onDateChange: (date: string) => void
  onLoad: () => void
}

export default function PlannerLoadCard({
  date,
  busy,
  embedded = false,
  onDateChange,
  onLoad,
}: PlannerLoadCardProps) {
  const content = (
    <div className={embedded ? 'planner-load-panel' : 'row wrap'}>
      <Field id="planner-date" label="Data">
        <PlannerDatePicker
          id="planner-date"
          value={date}
          onChange={onDateChange}
          disabled={busy}
          autoFocus={!embedded}
        />
      </Field>
      <Button
        variant={embedded ? 'secondary' : 'primary'}
        onClick={onLoad}
        disabled={busy}
        aria-busy={busy}
      >
        {busy ? 'Carregando...' : 'Carregar dia'}
      </Button>
    </div>
  )

  if (embedded) {
    return content
  }

  return <Card>{content}</Card>
}
