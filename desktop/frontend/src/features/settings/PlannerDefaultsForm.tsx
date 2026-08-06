import { useTranslation } from 'react-i18next'
import PlannerTimeInput from '../../components/PlannerTimeInput'
import { Button, Card, Field } from '../../components/ui'
import {
  builtinPlannerAnchors,
  DEFAULT_BALANCE_CREDIT_MULTIPLIER,
  DEFAULT_MAX_DAILY_EXTRA_MINUTES,
} from '../../util/plannerDefaults'
import {
  adjustPlannerAnchor,
  anchorLabelKey,
  anchorLabelParams,
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

const ANCHOR_FIELDS: AnchorField[] = ['in1', 'out1', 'in2', 'out2']

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
  const { t } = useTranslation()
  const builtins = builtinPlannerAnchors()

  return (
    <Card title={t('settings.planner.title')} intro={t('settings.planner.intro')}>
      <form
        className="form-stack"
        onSubmit={(event) => {
          event.preventDefault()
          onSave()
        }}
      >
        <div className="planner-anchor-grid">
          {ANCHOR_FIELDS.map((field, index) => (
            <Field
              key={field}
              id={`settings-anchor-${field}`}
              label={t(anchorLabelKey(index), anchorLabelParams(index))}
            >
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
          {t('settings.planner.factoryDefault', {
            times: `${builtins.in1}, ${builtins.out1}, ${builtins.in2}, ${builtins.out2}`,
          })}
        </small>
        <Field
          id="settings-max-daily-extra"
          label={t('settings.planner.maxDailyExtraLabel')}
          hint={t('settings.planner.maxDailyExtraHint')}
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
          label={t('settings.planner.creditMultiplierLabel')}
          hint={t('settings.planner.creditMultiplierHint')}
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
            {t('settings.planner.restoreDefaults')}
          </Button>
          <Button
            type="submit"
            variant="primary"
            disabled={busy}
            aria-busy={busy}
          >
            {busy ? t('common.saving') : t('settings.planner.save')}
          </Button>
        </div>
      </form>
    </Card>
  )
}
