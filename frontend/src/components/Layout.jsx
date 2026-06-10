import { useAuth } from '../AuthContext'
import { Sidebar } from './Sidebar'

const hospitals = {
  1: 'Hospital Loja',
  2: 'Hospital Cuenca',
  3: 'Hospital Quito',
  4: 'Hospital Guayaquil',
}

const roles = {
  admin: 'Administrador',
  doctor: 'Médico',
}

export function Layout({ children }) {
  const { user, logout } = useAuth()

  return (
    <div className="layout">
      <header className="topbar">
        <h1>Red Hospitalaria Distribuida</h1>
        <div className="user-badge">
          <div className="user-avatar" aria-hidden>
            <svg width="36" height="36" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="12" cy="8" r="3" fill="#0ea5e9" />
              <path d="M4 20c1.5-4 6-6 8-6s6.5 2 8 6" stroke="#0369a1" strokeWidth="1.2" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
          </div>
          <div className="user-info">
            <div className="user-name">{user.display_name}</div>
            <div className="user-meta">
              <span className="role">{roles[user.role]}</span>
              <span className="hospital">{hospitals[user.hospital_id]}</span>
            </div>
          </div>
          <button className="btn-logout" onClick={logout} aria-label="Cerrar sesión">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden>
              <path d="M16 17l5-5-5-5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
              <path d="M21 12H9" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
            <span className="logout-text">Salir</span>
          </button>
        </div>
      </header>
      <div className="layout-body">
        <Sidebar />
        <main className="layout-content">
          {children}
        </main>
      </div>
    </div>
  )
}
