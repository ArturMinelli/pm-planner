import { NavLink, Outlet } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export default function AppShell() {
  const { t } = useTranslation()

  return (
    <div className="app-root">
      <a className="skip-link" href="#main-content">
        {t('common.skipToContent')}
      </a>
      <aside className="sidebar">
        <div className="brand">
          <span className="brand-mark" translate="no">
            PM
          </span>
          <div>
            <div className="brand-title">Planner</div>
            <div className="brand-sub" translate="no">
              PontoMais
            </div>
          </div>
        </div>
        <nav className="nav">
          <NavLink end className="nav-link" to="/">
            {t('nav.journey')}
          </NavLink>
          <NavLink className="nav-link" to="/settings">
            {t('nav.settings')}
          </NavLink>
        </nav>
      </aside>
      <main id="main-content" className="main" tabIndex={-1}>
        <Outlet />
      </main>
    </div>
  )
}
