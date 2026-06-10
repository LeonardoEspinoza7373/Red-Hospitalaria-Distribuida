import { useAuth } from '../AuthContext'
import { Sidebar } from './Sidebar'
import { IconHospital, IconUser } from '../icons'

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
  const { user } = useAuth()

  return (
    <div className="layout">
      <header className="topbar">
        <div className="topbar-brand">
          <IconHospital />
          <h1>Red Hospitalaria Distribuida</h1>
        </div>
        <div className="user-badge">
          <div className="user-info">
            <div className="user-name">{user.display_name}</div>
            <div className="user-meta">
              <span className="role">{roles[user.role]}</span>
              <span className="hospital">{hospitals[user.hospital_id]}</span>
            </div>
          </div>
          <div className="user-avatar" aria-hidden>
            <IconUser />
          </div>
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
