import { NavLink, Outlet } from 'react-router-dom'

export default function AppShell() {
  return (
    <div className="app-root">
      <a className="skip-link" href="#main-content">
        Pular para conteúdo
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
            Jornada
          </NavLink>
          <NavLink className="nav-link" to="/settings">
            Configurações
          </NavLink>
        </nav>
      </aside>
      <main id="main-content" className="main" tabIndex={-1}>
        <Outlet />
      </main>
    </div>
  )
}
