import { useState, useEffect } from 'react'
import { trasplantesAPI, pacientesAPI, organosAPI, organosCompatibles } from '../api'

const estados = [
  { value: 'PENDIENTE', label: 'Pendiente' },
  { value: 'EN_CURSO', label: 'En Curso' },
  { value: 'REALIZADO', label: 'Realizado' },
  { value: 'CANCELADO', label: 'Cancelado' },
]

export function Trasplantes() {
  const [items, setItems] = useState([])
  const [pacientes, setPacientes] = useState([])
  const [organos, setOrganos] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [form, setForm] = useState(false)
  const [editing, setEditing] = useState(null)

  const [pacienteId, setPacienteId] = useState('')
  const [compatibles, setCompatibles] = useState([])
  const [loadingCompatibles, setLoadingCompatibles] = useState(false)
  const [organoId, setOrganoId] = useState('')
  const [fecha, setFecha] = useState('')
  const [estado, setEstado] = useState('')
  const [busy, setBusy] = useState(false)

  const load = () => {
    setLoading(true)
    Promise.all([trasplantesAPI.list(), pacientesAPI.list(), organosAPI.list()])
      .then(([trasps, pacs, orgs]) => { setItems(trasps); setPacientes(pacs); setOrganos(orgs) })
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handlePacienteChange = id => {
    setPacienteId(id)
    setOrganoId('')
    setCompatibles([])
    if (!id) return
    setLoadingCompatibles(true)
    organosCompatibles(id)
      .then(setCompatibles)
      .catch(e => setError(e.message))
      .finally(() => setLoadingCompatibles(false))
  }

  const handleSubmit = async e => {
    e.preventDefault()
    setBusy(true)
    try {
      const data = { paciente_id: +pacienteId, organo_id: +organoId, fecha, estado }
      if (editing) await trasplantesAPI.update(editing.id, data)
      else await trasplantesAPI.create(data)
      cancelForm()
      load()
    } catch (e) { setError(e.message) }
    setBusy(false)
  }

  const openEdit = item => {
    setEditing(item)
    setPacienteId(String(item.paciente_id))
    setOrganoId(String(item.organo_id))
    setFecha(item.fecha)
    setEstado(item.estado)
    setForm(true)
    setError('')
    organosCompatibles(item.paciente_id)
      .then(setCompatibles)
      .catch(e => setError(e.message))
  }

  const cancelForm = () => {
    setForm(false)
    setEditing(null)
    setPacienteId('')
    setOrganoId('')
    setFecha('')
    setEstado('')
    setError('')
  }

  const handleDelete = async id => {
    if (!confirm('¿Eliminar este registro?')) return
    try {
      await trasplantesAPI.delete(id)
      load()
    } catch (e) { setError(e.message) }
  }

  const pacienteNombre = id => pacientes.find(p => p.id === id)?.nombre || id
  const organoTipo = id => organos.find(o => o.id === id)?.tipo || id

  if (form) return (
    <div className="entity-form-page">
      <h2>{editing ? 'Editar Trasplante' : 'Registrar Trasplante'}</h2>
      <form className="entity-form" onSubmit={handleSubmit}>
        <label>
          <span>Paciente</span>
          <select value={pacienteId} onChange={e => handlePacienteChange(e.target.value)} required disabled={!!editing}>
            <option value="">Seleccionar paciente...</option>
            {pacientes.map(p => <option key={p.id} value={p.id}>{p.nombre}</option>)}
          </select>
        </label>
        <label>
          <span>Órgano Compatible</span>
          <select value={organoId} onChange={e => setOrganoId(e.target.value)} required disabled={!!editing || !pacienteId || loadingCompatibles}>
            {!pacienteId ? (
              <option value="">Seleccione un paciente primero</option>
            ) : loadingCompatibles ? (
              <option value="">Buscando órganos compatibles...</option>
            ) : compatibles.length === 0 ? (
              <option value="">No hay órganos compatibles disponibles</option>
            ) : (
              <>
                <option value="">Seleccionar órgano...</option>
                {compatibles.map(o => <option key={o.id} value={o.id}>{o.tipo}</option>)}
              </>
            )}
          </select>
        </label>
        <label>
          <span>Fecha</span>
          <input type="text" value={fecha} onChange={e => setFecha(e.target.value)} required placeholder="YYYY-MM-DD" />
        </label>
        <label>
          <span>Estado</span>
          <select value={estado} onChange={e => setEstado(e.target.value)} required>
            <option value="">Seleccionar...</option>
            {estados.map(e => <option key={e.value} value={e.value}>{e.label}</option>)}
          </select>
        </label>
        <div className="form-actions">
          <button type="submit" className="btn-primary" disabled={busy || !organoId}>
            {busy ? 'Guardando...' : editing ? 'Actualizar Trasplante' : 'Registrar Trasplante'}
          </button>
          <button type="button" className="btn-secondary" onClick={cancelForm}>Cancelar</button>
        </div>
      </form>
    </div>
  )

  return (
    <div className="entity-page">
      <header className="entity-header">
        <div className="entity-header-left">
          <h2>Trasplantes</h2>
        </div>
        <button className="btn-primary" onClick={() => { setEditing(null); setForm(true); setError(''); setPacienteId(''); setOrganoId(''); setFecha(''); setEstado(''); setCompatibles([]) }}>+ Nuevo</button>
      </header>
      {error && <div className="error">{error}</div>}
      {loading ? <div className="loading">Cargando...</div> : (
        <table className="entity-table">
          <thead>
            <tr>
              <th>Paciente</th>
              <th>Órgano</th>
              <th>Fecha</th>
              <th>Estado</th>
              <th>Acciones</th>
            </tr>
          </thead>
          <tbody>
            {items.length === 0 && (
              <tr><td colSpan="5" className="empty">Sin registros</td></tr>
            )}
            {items.map(item => (
              <tr key={item.id}>
                <td>{pacienteNombre(item.paciente_id)}</td>
                <td>{organoTipo(item.organo_id)}</td>
                <td>{item.fecha}</td>
                <td>{item.estado}</td>
                <td className="actions">
                  <button className="btn-sm" onClick={() => openEdit(item)}>Editar</button>
                  <button className="btn-sm danger" onClick={() => handleDelete(item.id)}>Eliminar</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
