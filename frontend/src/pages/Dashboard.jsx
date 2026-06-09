import { useNavigate } from 'react-router-dom'
import { useAuth } from '../AuthContext'

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

export function Dashboard() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  return (
    <div className="dashboard">
      <header>
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
      <main>
        <h2>Panel de Control</h2>
        <div className="cards">
          <div className="card" onClick={() => navigate('/pacientes')}>
            <span className="card-icon">👤</span>
            <h3>Pacientes</h3>
            <p>Registro de pacientes en lista de espera</p>
          </div>
          <div className="card" onClick={() => navigate('/donantes')}>
            <span className="card-icon">❤️</span>
            <h3>Donantes</h3>
            <p>Gestión de donantes voluntarios</p>
          </div>
          <div className="card" onClick={() => navigate('/organos')}>
            <span className="card-icon">🫀</span>
            <h3>Órganos</h3>
            <p>Inventario de órganos disponibles</p>
          </div>
          <div className="card" onClick={() => navigate('/donaciones')}>
            <span className="card-icon">📋</span>
            <h3>Donaciones</h3>
            <p>Registro de extracciones</p>
          </div>
          <div className="card" onClick={() => navigate('/trasplantes')}>
            <span className="card-icon">🔄</span>
            <h3>Trasplantes</h3>
            <p>Asignación y seguimiento</p>
          </div>
          {user.role === 'admin' && (
            <div className="card admin" onClick={() => navigate('/usuarios')}>
              <span className="card-icon">⚙️</span>
              <h3>Administración</h3>
              <p>Usuarios y monitoreo</p>
            </div>
          )}
        </div>
      </main>
    </div>
  )
}
