import type { ReactNode } from 'react'
import { ConfigProvider } from '../context/ConfigContext'

export function renderWithConfig(children: ReactNode) {
  return <ConfigProvider>{children}</ConfigProvider>
}
