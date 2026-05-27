import type { PmConfig, PlannerPayload, PlannerSummary } from '../types'

interface GoPmApp {
  GetConfig(): Promise<PmConfig>
  SaveConfig(c: PmConfig): Promise<void>
  PingAuth(): Promise<string>
  LoadPlanner(date: string): Promise<PlannerPayload>
  RecalculatePlanner(
    date: string,
    targetSecs: number,
    in1: string,
    out1: string,
    in2: string,
  ): Promise<PlannerSummary>
}

function goApp(): GoPmApp | undefined {
  return (globalThis as unknown as { go?: { main?: { App: GoPmApp } } }).go
    ?.main?.App
}

export function hasWailsRuntime(): boolean {
  return !!goApp()
}

function requireRuntime(): GoPmApp {
  const a = goApp()
  if (!a) throw new Error('This page must run inside the PM desktop shell.')
  return a
}

/** Safe for browser/dev preview — returns mock mode flag */
export async function getConfig(): Promise<PmConfig> {
  if (!hasWailsRuntime())
    return { email: '', password: '', cache_ttl_hours: undefined }
  return requireRuntime().GetConfig()
}

export async function saveConfig(c: PmConfig): Promise<void> {
  const payload: PmConfig = {
    email: c.email,
    password: c.password,
  }
  if (
    typeof c.cache_ttl_hours === 'number' &&
    Number.isFinite(c.cache_ttl_hours) &&
    c.cache_ttl_hours > 0
  ) {
    payload.cache_ttl_hours = Math.trunc(c.cache_ttl_hours)
  }
  return requireRuntime().SaveConfig(payload)
}

/** Empty string means success */
export async function pingAuth(): Promise<string> {
  return requireRuntime().PingAuth()
}

export async function loadPlanner(date: string): Promise<PlannerPayload> {
  return requireRuntime().LoadPlanner(date)
}

export async function recalculatePlanner(
  date: string,
  targetSecs: number,
  in1: string,
  out1: string,
  in2: string,
): Promise<PlannerSummary> {
  return requireRuntime().RecalculatePlanner(
    date,
    targetSecs,
    in1,
    out1,
    in2,
  )
}
