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

const cards = [
  { title: 'Pacientes', icon: '👤', desc: 'Registro de pacientes en lista de espera' },
  { title: 'Donantes', icon: '❤️', desc: 'Gestión de donantes voluntarios' },
  { title: 'Órganos', icon: '🫀', desc: 'Inventario de órganos disponibles' },
  { title: 'Donaciones', icon: '📋', desc: 'Registro de extracciones y donaciones' },
  { title: 'Trasplantes', icon: '🔄', desc: 'Asignación y seguimiento de trasplantes' },
]

export function Dashboard() {
  const { user, logout } = useAuth()

  return (
    <div className="dashboard">
      <header>
        <h1>Red Hospitalaria Distribuida</h1>
        <div className="user-badge">
          <span className="name">{user.display_name}</span>
          <span className="role">{roles[user.role]}</span>
          <span className="hospital">{hospitals[user.hospital_id]}</span>
          <button className="btn-logout" onClick={logout}>Cerrar Sesión</button>
        </div>
      </header>
      <main>
        <h2>Panel de Control</h2>
        <div className="cards">
          {cards.map(c => (
            <div key={c.title} className="card">
              <span className="card-icon">{c.icon}</span>
              <h3>{c.title}</h3>
              <p>{c.desc}</p>
            </div>
          ))}
          {user.role === 'admin' && (
            <div className="card admin">
              <span className="card-icon">⚙️</span>
              <h3>Administración</h3>
              <p>Gestión de usuarios y monitoreo del sistema</p>
            </div>
          )}
        </div>
      </main>
    </div>
  )
}
