import { useNavigate } from 'react-router-dom'
import { useAuth } from '../AuthContext'

export function Dashboard() {
  const { user } = useAuth()
  const navigate = useNavigate()

  return (
    <div className="dashboard">
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
          {user.role === 'admin' && (
            <div className="card admin" onClick={() => navigate('/logs')}>
              <span className="card-icon">📊</span>
              <h3>Logs del Sistema</h3>
              <p>Monitoreo de nodos seguidores</p>
            </div>
          )}
        </div>
      </main>
    </div>
  )
}
