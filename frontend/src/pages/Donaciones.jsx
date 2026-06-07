import { useState, useEffect } from 'react'
import { organosAPI, donantesAPI } from '../api'

const tiposOrgano = [
  { value: 'CORAZON', label: 'Corazón' },
  { value: 'PULMON', label: 'Pulmón' },
  { value: 'HIGADO', label: 'Hígado' },
  { value: 'RIÑON', label: 'Riñón' },
  { value: 'PANCREAS', label: 'Páncreas' },
  { value: 'INTESTINO', label: 'Intestino' },
  { value: 'CORNEA', label: 'Córnea' },
  { value: 'MEDULA_OSEA', label: 'Médula Ósea' },
]

export function Donaciones() {
  const [extractions, setExtractions] = useState([])
  const [donantes, setDonantes] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [form, setForm] = useState(false)
  const [donanteId, setDonanteId] = useState('')
  const [tipo, setTipo] = useState('')
  const [compatibilidad, setCompatibilidad] = useState('')
  const [busy, setBusy] = useState(false)

  const load = () => {
    setLoading(true)
    Promise.all([organosAPI.list(), donantesAPI.list()])
      .then(([orgs, dons]) => { setExtractions(orgs); setDonantes(dons) })
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleExtract = async e => {
    e.preventDefault()
    setBusy(true)
    try {
      await organosAPI.create({ tipo, compatibilidad, estado: 'DISPONIBLE', donante_id: +donanteId })
      setForm(false)
      setTipo('')
      setCompatibilidad('')
      setDonanteId('')
      load()
    } catch (e) { setError(e.message) }
    setBusy(false)
  }

  if (form) return (
    <div className="entity-form-page">
      <h2>Registrar Extracción</h2>
      <form className="entity-form" onSubmit={handleExtract}>
        <label>
          <span>Donante</span>
          <select value={donanteId} onChange={e => setDonanteId(e.target.value)} required>
            <option value="">Seleccionar donante...</option>
            {donantes.map(d => <option key={d.id} value={d.id}>{d.nombre} (ID {d.id})</option>)}
          </select>
        </label>
        <label>
          <span>Tipo de Órgano</span>
          <select value={tipo} onChange={e => setTipo(e.target.value)} required>
            <option value="">Seleccionar...</option>
            {tiposOrgano.map(t => <option key={t.value} value={t.value}>{t.label}</option>)}
          </select>
        </label>
        <label>
          <span>Compatibilidad</span>
          <input type="text" value={compatibilidad} onChange={e => setCompatibilidad(e.target.value)} required placeholder="Ej: O+, A-..." />
        </label>
        <div className="form-actions">
          <button type="submit" className="btn-primary" disabled={busy}>{busy ? 'Registrando...' : 'Registrar Extracción'}</button>
          <button type="button" className="btn-secondary" onClick={() => setForm(false)}>Cancelar</button>
        </div>
      </form>
    </div>
  )

  return (
    <div className="entity-page">
      <header className="entity-header">
        <h2>Donaciones (Extracciones)</h2>
        <button className="btn-primary" onClick={() => setForm(true)}>+ Nueva Extracción</button>
      </header>
      {error && <div className="error">{error}</div>}
      {loading ? <div className="loading">Cargando...</div> : (
        <table className="entity-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Donante ID</th>
              <th>Tipo</th>
              <th>Compatibilidad</th>
              <th>Estado</th>
            </tr>
          </thead>
          <tbody>
            {extractions.length === 0 && (
              <tr><td colSpan="5" className="empty">Sin registros</td></tr>
            )}
            {extractions.map(org => (
              <tr key={org.id}>
                <td>{org.id}</td>
                <td>{org.donante_id}</td>
                <td>{tiposOrgano.find(t => t.value === org.tipo)?.label || org.tipo}</td>
                <td>{org.compatibilidad}</td>
                <td>{org.estado === 'DISPONIBLE' ? <span style={{color: '#4ade80'}}>Disponible</span> : <span style={{color: '#f87171'}}>No Disponible</span>}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
