import { useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '../AuthContext'

const links = [
  { to: '/', icon: '🏠', label: 'Dashboard' },
  { to: '/pacientes', icon: '👤', label: 'Pacientes' },
  { to: '/donantes', icon: '❤️', label: 'Donantes' },
  { to: '/organos', icon: '🫀', label: 'Órganos' },
  { to: '/donaciones', icon: '📋', label: 'Donaciones' },
  { to: '/trasplantes', icon: '🔄', label: 'Trasplantes' },
]

const adminLinks = [
  { to: '/usuarios', icon: '⚙️', label: 'Usuarios' },
  { to: '/logs', icon: '📊', label: 'Logs' },
]

export function Sidebar() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()

  const isActive = to => {
    if (to === '/') return location.pathname === '/'
    return location.pathname.startsWith(to)
  }

  return (
    <nav className="sidebar">
      {links.map(link => (
        <button
          key={link.to}
          className={`sidebar-link${isActive(link.to) ? ' active' : ''}`}
          onClick={() => navigate(link.to)}
        >
          <span className="sidebar-icon">{link.icon}</span>
          <span>{link.label}</span>
        </button>
      ))}
      {user.role === 'admin' && (
        <>
          <div className="sidebar-divider" />
          {adminLinks.map(link => (
            <button
              key={link.to}
              className={`sidebar-link${isActive(link.to) ? ' active' : ''}`}
              onClick={() => navigate(link.to)}
            >
              <span className="sidebar-icon">{link.icon}</span>
              <span>{link.label}</span>
            </button>
          ))}
        </>
      )}
    </nav>
  )
}
