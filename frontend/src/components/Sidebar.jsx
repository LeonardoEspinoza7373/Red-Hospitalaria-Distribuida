import { useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '../AuthContext'
import { IconHome, IconUser, IconHeart, IconActivity, IconClipboard, IconSwap, IconSettings, IconTerminal, IconLogOut } from '../icons'

const links = [
  { to: '/', icon: IconHome, label: 'Panel de Control' },
  { to: '/pacientes', icon: IconUser, label: 'Pacientes' },
  { to: '/donantes', icon: IconHeart, label: 'Donantes' },
  { to: '/organos', icon: IconActivity, label: 'Órganos' },
  { to: '/donaciones', icon: IconClipboard, label: 'Donaciones' },
  { to: '/trasplantes', icon: IconSwap, label: 'Trasplantes' },
]

const adminLinks = [
  { to: '/usuarios', icon: IconSettings, label: 'Administración' },
  { to: '/logs', icon: IconTerminal, label: 'Logs' },
]

export function Sidebar() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()

  const isActive = to => {
    if (to === '/') return location.pathname === '/'
    return location.pathname.startsWith(to)
  }

  return (
    <nav className="sidebar">
      <div className="sidebar-nav">
        {links.map(link => {
          const Icon = link.icon
          return (
            <button
              key={link.to}
              className={`sidebar-link${isActive(link.to) ? ' active' : ''}`}
              onClick={() => navigate(link.to)}
            >
              <Icon />
              <span>{link.label}</span>
            </button>
          )
        })}
        {user.role === 'admin' && (
          <>
            <div className="sidebar-divider" />
            {adminLinks.map(link => {
              const Icon = link.icon
              return (
                <button
                  key={link.to}
                  className={`sidebar-link${isActive(link.to) ? ' active' : ''}`}
                  onClick={() => navigate(link.to)}
                >
                  <Icon />
                  <span>{link.label}</span>
                </button>
              )
            })}
          </>
        )}
      </div>
      <div className="sidebar-divider" />
      <button className="sidebar-link sidebar-logout" onClick={logout}>
        <IconLogOut />
        <span>Salir</span>
      </button>
    </nav>
  )
}
