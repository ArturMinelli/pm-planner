import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { Journey, PlannerPayload, PlannerSummary, SolvedSlot } from '../types'
import * as backend from '../services/backend'
import { translateMessage } from '../i18n/translateMessage'
import { localDateYYYYMMDD } from '../util/timeFormat'
import {
  applyInstantSuggestions,
  defaultSolvedSlot,
  NO_SOLVED_SLOT,
  seedEntryAfterExit,
} from '../util/plannerSuggestions'
import { enforceJourneyOrder } from '../util/plannerTimes'
import { Banner, Field, Page, PageHeader, Stack } from '../components/ui'
import PlannerDatePicker from '../components/PlannerDatePicker'
import PlannerSummaryPanel from '../features/planner/PlannerSummaryPanel'
import PlannerJourneyList from '../features/planner/PlannerJourneyList'

const DEFAULT_BREAK_MINUTES = 60

export default function PlannerPage() {
  const { t } = useTranslation()
  const [date, setDate] = useState(localDateYYYYMMDD)
  const [fetching, setFetching] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [loaded, setLoaded] = useState<PlannerPayload | null>(null)
  const [rawJourneys, setRawJourneys] = useState<Journey[]>([])
  const [journeys, setJourneys] = useState<Journey[]>([])
  const [solvedSlot, setSolvedSlot] = useState<SolvedSlot>(NO_SOLVED_SLOT)
  const [summary, setSummary] = useState<PlannerSummary | null>(null)

  const loadedRef = useRef<PlannerPayload | null>(null)
  loadedRef.current = loaded

  const summaryRequestRef = useRef(0)
  const fetchRequestRef = useRef(0)

  const recalculateSummary = useCallback(
    (journeysToCalc: Journey[], activeSlot: SolvedSlot) => {
      const currentLoaded = loadedRef.current
      if (!currentLoaded) return

      const requestId = ++summaryRequestRef.current
      void backend
        .recalculateDay({
          date: currentLoaded.date,
          baseTargetSecs: currentLoaded.baseTargetSecs,
          balance: currentLoaded.balance,
          journeys: journeysToCalc,
          solvedSlot: activeSlot,
        })
        .then((newSummary) => {
          if (requestId !== summaryRequestRef.current) return
          setSummary(newSummary)
        })
        .catch(() => {
          // summary errors are non-fatal
        })
    },
    [],
  )

  const commitPlannerState = useCallback(
    (
      rawJourneysInput: Journey[],
      slot: SolvedSlot,
      preserveSlot?: { journeyIndex: number; isEntry: boolean },
    ) => {
      const currentLoaded = loadedRef.current
      if (!currentLoaded) return

      const { journeys: solvedJourneys, solvedSlot: activeSlot } = applyInstantSuggestions(
        rawJourneysInput,
        slot,
        currentLoaded.baseTargetSecs,
        { preserveSlot },
      )
      setRawJourneys(rawJourneysInput)
      setJourneys(solvedJourneys)
      setSolvedSlot(activeSlot)
      recalculateSummary(solvedJourneys, activeSlot)
    },
    [recalculateSummary],
  )

  const handleUpdateJourney = useCallback(
    (journeyIndex: number, isEntry: boolean, time: string) => {
      const patched = rawJourneys.map((journey, index) => {
        if (index !== journeyIndex) return journey
        if (isEntry) {
          return {
            ...journey,
            entry: { time, registered: journey.entry.registered },
          }
        }
        return {
          ...journey,
          exit: { time, registered: journey.exit.registered },
        }
      })
      const reordered = enforceJourneyOrder(patched, journeyIndex, isEntry)
      commitPlannerState(reordered, solvedSlot, { journeyIndex, isEntry })
    },
    [rawJourneys, solvedSlot, commitPlannerState],
  )

  const handleToggleSolved = useCallback(
    (journeyIndex: number, isEntry: boolean) => {
      const isActive =
        solvedSlot.journeyIndex === journeyIndex && solvedSlot.isEntry === isEntry
      const newSlot: SolvedSlot = isActive
        ? NO_SOLVED_SLOT
        : { journeyIndex, isEntry }
      commitPlannerState(rawJourneys, newSlot)
    },
    [rawJourneys, solvedSlot, commitPlannerState],
  )

  const handleAddJourney = useCallback(() => {
    const lastJourney = rawJourneys[rawJourneys.length - 1]
    const previousExitTime = lastJourney?.exit.time ?? ''
    const newEntryTime = seedEntryAfterExit(previousExitTime, DEFAULT_BREAK_MINUTES)
    const newJourney: Journey = {
      entry: { time: newEntryTime, registered: false },
      exit: { time: '', registered: false },
    }
    const newJourneys = [...rawJourneys, newJourney]
    const newSlot: SolvedSlot = {
      journeyIndex: newJourneys.length - 1,
      isEntry: false,
    }
    commitPlannerState(newJourneys, newSlot)
  }, [rawJourneys, commitPlannerState])

  const handleRemoveJourney = useCallback(
    (journeyIndex: number) => {
      const newJourneys = rawJourneys.filter((_, index) => index !== journeyIndex)
      let newSlot = solvedSlot
      if (solvedSlot.journeyIndex === journeyIndex) {
        newSlot = NO_SOLVED_SLOT
      } else if (solvedSlot.journeyIndex > journeyIndex) {
        newSlot = {
          journeyIndex: solvedSlot.journeyIndex - 1,
          isEntry: solvedSlot.isEntry,
        }
      }
      if (newSlot.journeyIndex < 0) {
        newSlot = defaultSolvedSlot(newJourneys)
      }
      commitPlannerState(newJourneys, newSlot)
    },
    [rawJourneys, solvedSlot, commitPlannerState],
  )

  useEffect(() => {
    const requestId = ++fetchRequestRef.current
    let cancelled = false

    const fetchDay = async () => {
      setFetching(true)
      setError(null)
      setLoaded(null)
      setSummary(null)
      setRawJourneys([])
      setJourneys([])
      setSolvedSlot(NO_SOLVED_SLOT)

      try {
        const payload = await backend.loadPlanner(date)
        if (cancelled || requestId !== fetchRequestRef.current) return

        const activeSlot = payload.solvedSlot?.journeyIndex >= 0
          ? payload.solvedSlot
          : defaultSolvedSlot(payload.journeys)
        const { journeys: instantJourneys, solvedSlot: instantSlot } = applyInstantSuggestions(
          payload.journeys,
          activeSlot,
          payload.baseTargetSecs,
        )

        setLoaded(payload)
        setRawJourneys(payload.journeys)
        setJourneys(instantJourneys)
        setSolvedSlot(instantSlot)

        try {
          const initialSummary = await backend.recalculateDay({
            date: payload.date,
            baseTargetSecs: payload.baseTargetSecs,
            balance: payload.balance,
            journeys: instantJourneys,
            solvedSlot: instantSlot,
          })
          if (cancelled || requestId !== fetchRequestRef.current) return
          setSummary(initialSummary)
        } catch {
          // summary recalculation is optional on first load
        }
      } catch (caughtError) {
        if (cancelled || requestId !== fetchRequestRef.current) return
        const message = caughtError instanceof Error ? caughtError.message : String(caughtError)
        setError(message.startsWith('errors.') ? t(message) : message)
      } finally {
        if (cancelled || requestId !== fetchRequestRef.current) return
        setFetching(false)
      }
    }

    void fetchDay()

    return () => {
      cancelled = true
    }
  }, [date, t])

  const timesDisabled = fetching
  const originalsLine = loaded?.originalsLine || t('planner.none')
  const isToday = loaded?.date === localDateYYYYMMDD()
  const loadWarning = translateMessage(t, loaded?.loadWarning)

  return (
    <Page>
      <PageHeader
        title={t('planner.title')}
        description={t('planner.description')}
        actions={
          <Field id="planner-date" label={t('common.date')} className="page-header-date">
            <PlannerDatePicker
              id="planner-date"
              value={date}
              onChange={setDate}
              disabled={fetching}
            />
          </Field>
        }
      />

      <Stack>
        {loadWarning ? <Banner>{loadWarning}</Banner> : null}

        {error ? <Banner tone="error">{error}</Banner> : null}

        <div className="grid-2">
          <PlannerJourneyList
            loading={fetching}
            journeys={journeys}
            solvedSlot={solvedSlot}
            balance={loaded?.balance}
            balanceError={isToday ? loaded?.balanceError : undefined}
            summary={summary}
            originalsLine={originalsLine}
            disabled={timesDisabled}
            onAddJourney={handleAddJourney}
            onRemoveJourney={handleRemoveJourney}
            onUpdateJourney={handleUpdateJourney}
            onToggleSolved={handleToggleSolved}
          />
          <PlannerSummaryPanel
            loading={fetching}
            loaded={loaded}
            summary={summary}
          />
        </div>
      </Stack>
    </Page>
  )
}
