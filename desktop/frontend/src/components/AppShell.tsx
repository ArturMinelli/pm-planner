import { NavLink, Outlet } from 'react-router-dom'

export default function AppShell() {
  return (
    <div className="app-root">
      <aside className="sidebar">
        <div className="brand">
          <span className="brand-mark">PM</span>
          <div>
            <div className="brand-title">Planner</div>
            <div className="brand-sub">PontoMais</div>
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
      <main className="main">
        <Outlet />
      </main>
    </div>
  )
}
