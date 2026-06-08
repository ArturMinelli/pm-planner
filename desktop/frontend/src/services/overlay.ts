import type { OverlayPayload } from '../types'

interface GoOverlayApp {
  GetOverlayPayload(): Promise<OverlayPayload>
  DockOverlay(): Promise<void>
  CloseOverlay(): Promise<void>
}

function overlayApp(): GoOverlayApp | undefined {
  return (globalThis as unknown as { go?: { main?: { OverlayApp?: GoOverlayApp } } })
    .go?.main?.OverlayApp
}

export function hasOverlayRuntime(): boolean {
  return !!overlayApp()
}

export async function getOverlayPayload(): Promise<OverlayPayload> {
  const app = overlayApp()
  if (!app) throw new Error('Overlay runtime unavailable.')
  return app.GetOverlayPayload()
}

export async function dockOverlay(): Promise<void> {
  await overlayApp()?.DockOverlay()
}

export async function closeOverlay(): Promise<void> {
  await overlayApp()?.CloseOverlay()
}
