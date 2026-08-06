import i18n from '../i18n'
import * as backend from './backend'
import type { ScheduledReminder } from '../types'
import { localDateYYYYMMDD } from '../util/timeFormat'

const DELIVERED_STORAGE_KEY = 'pm.browserReminders.delivered'
const POLL_INTERVAL_MS = 60_000
const DELIVERY_WINDOW_MS = 10 * 60 * 1000

let pollTimer: ReturnType<typeof setInterval> | null = null
let running = false

function readDeliveredIds(): Set<string> {
  try {
    const raw = localStorage.getItem(DELIVERED_STORAGE_KEY)
    if (!raw) return new Set()
    const parsed = JSON.parse(raw) as string[]
    return new Set(parsed)
  } catch {
    return new Set()
  }
}

function writeDeliveredIds(ids: Set<string>) {
  localStorage.setItem(DELIVERED_STORAGE_KEY, JSON.stringify([...ids]))
}

function formatSlotTime(slotTime: string): string {
  const date = new Date(slotTime)
  if (Number.isNaN(date.getTime())) return slotTime
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function buildNotificationContent(reminder: ScheduledReminder) {
  const title = i18n.t('reminders.slot_in_minutes', {
    slot: reminder.slotLabel,
    minutes: reminder.leadMinutes,
  })
  const body = i18n.t('reminders.recommended_time', {
    time: formatSlotTime(reminder.slotTime),
  })
  return { title, body }
}

function isDue(reminder: ScheduledReminder, now: Date): boolean {
  const fireAt = new Date(reminder.fireAt)
  if (Number.isNaN(fireAt.getTime())) return false
  if (fireAt > now) return false
  return now.getTime() - fireAt.getTime() <= DELIVERY_WINDOW_MS
}

async function pollReminders() {
  if (!backend.usesHttpTransport()) return
  if (!('Notification' in window) || Notification.permission !== 'granted') return

  try {
    const status = await backend.getReminderStatus()
    if (!status.settings.enabled) return

    const plan = await backend.getReminderPlan(localDateYYYYMMDD())
    const delivered = readDeliveredIds()
    const now = new Date()

    for (const reminder of plan.schedule) {
      if (delivered.has(reminder.id)) continue
      if (!isDue(reminder, now)) continue

      const { title, body } = buildNotificationContent(reminder)
      new Notification(title, { body })
      delivered.add(reminder.id)
    }

    writeDeliveredIds(delivered)
  } catch {
    // polling errors are non-fatal
  }
}

export function startBrowserReminderScheduler() {
  if (!backend.usesHttpTransport()) return
  if (running) return
  running = true

  void pollReminders()
  pollTimer = setInterval(() => {
    void pollReminders()
  }, POLL_INTERVAL_MS)
}

export function stopBrowserReminderScheduler() {
  running = false
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

export function notifyConfigChanged() {
  void pollReminders()
}

export async function requestBrowserNotificationPermission(): Promise<boolean> {
  if (!backend.usesHttpTransport()) return false
  const status = await backend.requestNotificationPermission()
  return status.authorized
}
