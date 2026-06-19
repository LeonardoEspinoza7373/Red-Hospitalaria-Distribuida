import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../AuthContext'
import { pacientesAPI, donantesAPI, organosAPI, trasplantesAPI } from '../api'
import { IconUser, IconHeart, IconActivity, IconSwap, IconClipboard, IconSettings, IconTerminal } from '../icons'

const modules = [
  { to: '/pacientes', icon: IconUser, title: 'Pacientes', desc: 'Registro de pacientes en lista de espera' },
  { to: '/donantes', icon: IconHeart, title: 'Donantes', desc: 'Gestión de donantes voluntarios' },
  { to: '/organos', icon: IconActivity, title: 'Órganos', desc: 'Inventario de órganos disponibles' },
  { to: '/donaciones', icon: IconClipboard, title: 'Donaciones', desc: 'Registro de extracciones' },
  { to: '/trasplantes', icon: IconSwap, title: 'Trasplantes', desc: 'Asignación y seguimiento' },
]

export function Dashboard() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const [metrics, setMetrics] = useState(null)

  useEffect(() => {
    Promise.all([
      pacientesAPI.list(),
      donantesAPI.list(),
      organosAPI.list(),
      trasplantesAPI.list(),
    ]).then(([pacientes, donantes, organos, trasplantes]) => {
      setMetrics({
        pacientesEspera: pacientes.length,
        donantesRegistrados: donantes.length,
        organosDisponibles: organos.filter(o => o.estado === 'DISPONIBLE').length,
        trasplantesRealizados: trasplantes.filter(t => t.estado === 'REALIZADO').length,
      })
    }).catch(() => setMetrics(null))
  }, [])

  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <div>
          <h2>Panel de Control</h2>
          <div className="life-pulse">
            <span className="life-pulse-dot" />
            <span>Red: 5 nodos conectados</span>
          </div>
        </div>
      </div>

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">
            <span className="metric-value">{metrics?.pacientesEspera ?? '—'}</span>
            <span className="metric-icon life"><IconUser /></span>
          </div>
          <span className="metric-label">Pacientes en espera</span>
        </div>
        <div className="metric">
          <div className="metric-top">
            <span className="metric-value">{metrics?.donantesRegistrados ?? '—'}</span>
            <span className="metric-icon gold"><IconHeart /></span>
          </div>
          <span className="metric-label">Donantes registrados</span>
        </div>
        <div className="metric">
          <div className="metric-top">
            <span className="metric-value">{metrics?.organosDisponibles ?? '—'}</span>
            <span className="metric-icon clinical"><IconActivity /></span>
          </div>
          <span className="metric-label">Órganos disponibles</span>
        </div>
        <div className="metric">
          <div className="metric-top">
            <span className="metric-value">{metrics?.trasplantesRealizados ?? '—'}</span>
            <span className="metric-icon success"><IconSwap /></span>
          </div>
          <span className="metric-label">Trasplantes realizados</span>
        </div>
      </div>

      <h3>Módulos del sistema</h3>
      <div className="cards">
        {modules.map(m => {
          const Icon = m.icon
          return (
            <div key={m.to} className="card" onClick={() => navigate(m.to)}>
              <span className="card-icon"><Icon /></span>
              <h3>{m.title}</h3>
              <p>{m.desc}</p>
            </div>
          )
        })}
        {user.role === 'admin' && (
          <div className="card admin" onClick={() => navigate('/usuarios')}>
            <span className="card-icon" style={{ color: 'var(--color-gold)' }}><IconSettings /></span>
            <h3>Administración</h3>
            <p>Usuarios y monitoreo</p>
          </div>
        )}
        {user.role === 'admin' && (
          <div className="card admin" onClick={() => navigate('/logs')}>
            <span className="card-icon" style={{ color: 'var(--color-gold)' }}><IconTerminal /></span>
            <h3>Logs del Sistema</h3>
            <p>Monitoreo de nodos seguidores</p>
          </div>
        )}
      </div>
    </div>
  )
}
