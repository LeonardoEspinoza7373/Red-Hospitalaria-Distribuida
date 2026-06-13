import { useState, useEffect } from 'react'
import { EntityList } from '../components/EntityList'
import { usuariosAPI, apiFetch } from '../api'
import { IconActivity } from '../icons'

const roles = {
  admin: 'Administrador',
  doctor: 'Médico',
}

const hospitals = {
  1: 'Hospital Loja',
  2: 'Hospital Cuenca',
  3: 'Hospital Quito',
  4: 'Hospital Guayaquil',
}

const columns = [
  { key: 'username', label: 'Usuario' },
  { key: 'display_name', label: 'Nombre' },
  { key: 'role', label: 'Rol', resolve: { format: r => roles[r] } },
  { key: 'hospital_id', label: 'Hospital', resolve: { format: id => hospitals[id] } },
]

const Form = [
  { key: 'username', label: 'Nombre de Usuario', required: true, placeholder: 'Ingrese el nombre de usuario (login)' },
  { key: 'password', label: 'Contraseña', type: 'password', required: false, placeholder: 'Solo si desea cambiarla' },
  { key: 'password_confirm', label: 'Confirmar Contraseña', type: 'password', required: false, placeholder: 'Repita la contraseña' },
  { key: 'display_name', label: 'Nombre Visible', required: true, placeholder: 'Ingrese el nombre visible del usuario' },
  { key: 'role', label: 'Rol', type: 'select', required: true,
    options: [
      { value: 'admin', label: 'Administrador' },
      { value: 'doctor', label: 'Médico' },
    ],
  },
  { key: 'hospital_id', label: 'Hospital', type: 'select', required: true,
    options: [
      { value: '1', label: 'Hospital Loja (1)' },
      { value: '2', label: 'Hospital Cuenca (2)' },
      { value: '3', label: 'Hospital Quito (3)' },
      { value: '4', label: 'Hospital Guayaquil (4)' },
    ],
  },
]

function BullyToggle() {
  const [enabled, setEnabled] = useState(true)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false

    apiFetch('/api/admin/bully')
      .then(d => {
        if (!cancelled) setEnabled(Boolean(d.enabled))
      })
      .catch(() => {
        if (!cancelled) setError('No se pudo cargar el estado del algoritmo.')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [])

  const toggle = () => {
    const next = !enabled
    setEnabled(next)
    setSaving(true)
    setError('')

    apiFetch('/api/admin/bully', { method: 'POST', body: JSON.stringify({ enabled: next }) })
      .then(d => setEnabled(Boolean(d.enabled)))
      .catch(() => {
        setEnabled(!next)
        setError('No se pudo actualizar el estado del algoritmo.')
      })
      .finally(() => setSaving(false))
  }

  if (loading) {
    return (
      <div className="bully-card" aria-busy="true">
        <div className="bully-card-inner">
          <div className="bully-card-left">
            <div className="bully-card-icon">
              <IconActivity />
            </div>
            <div className="bully-card-text">
              <h3>Algoritmo Bully</h3>
              <p>Cargando estado actual del algoritmo...</p>
            </div>
          </div>
          <div className="bully-card-right">
            <span className="bully-status">Cargando</span>
            <label className="toggle">
              <input type="checkbox" checked={false} disabled />
              <span className="toggle-track">
                <span className="toggle-thumb" />
              </span>
            </label>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="bully-card">
      <div className="bully-card-inner">
        <div className="bully-card-left">
          <div className="bully-card-icon">
            <IconActivity />
          </div>
          <div className="bully-card-text">
            <h3>Algoritmo Bully</h3>
            <p>
              {enabled
                ? 'Si el coordinador cae, los nodos elegirán automáticamente uno nuevo.'
                : 'Si el coordinador cae, el sistema quedará no disponible hasta que se reactive el algoritmo.'}
            </p>
          </div>
        </div>
        <div className="bully-card-right">
          <span className={`bully-status ${enabled ? 'on' : 'off'}`}>
            {enabled ? 'Activado' : 'Desactivado'}
          </span>
          <label className="toggle">
            <input type="checkbox" checked={enabled} onChange={toggle} disabled={saving} />
            <span className="toggle-track">
              <span className="toggle-thumb" />
            </span>
          </label>
        </div>
      </div>
      {error ? <p className="bully-error">{error}</p> : null}
    </div>
  )
}

export function Usuarios() {
  const [formOpen, setFormOpen] = useState(false)

  return (
    <div>
      {!formOpen && <BullyToggle />}
      <EntityList api={usuariosAPI} columns={columns} title="Usuarios" Form={Form} onFormChange={setFormOpen} />
    </div>
  )
}
