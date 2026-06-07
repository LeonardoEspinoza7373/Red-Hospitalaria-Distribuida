import { useState, useEffect } from 'react'

export function EntityList({ api, columns, title, Form }) {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [form, setForm] = useState(null)

  const load = () => {
    setLoading(true)
    api.list()
      .then(setItems)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleDelete = async id => {
    if (!confirm('¿Eliminar este registro?')) return
    try {
      await api.delete(id)
      load()
    } catch (e) { setError(e.message) }
  }

  const handleSave = async data => {
    try {
      if (data.id) await api.update(data.id, data)
      else await api.create(data)
      setForm(null)
      load()
    } catch (e) { setError(e.message) }
  }

  if (form) return (
    <EntityForm
      title={form.id ? `Editar ${title}` : `Nuevo ${title}`}
      fields={Form}
      initial={form}
      onSave={handleSave}
      onCancel={() => setForm(null)}
    />
  )

  return (
    <div className="entity-page">
      <header className="entity-header">
        <h2>{title}</h2>
        <button className="btn-primary" onClick={() => setForm({})}>+ Nuevo</button>
      </header>
      {error && <div className="error">{error}</div>}
      {loading ? <div className="loading">Cargando...</div> : (
        <table className="entity-table">
          <thead>
            <tr>
              {columns.map(c => <th key={c.key}>{c.label}</th>)}
              <th>Acciones</th>
            </tr>
          </thead>
          <tbody>
            {items.length === 0 && (
              <tr><td colSpan={columns.length + 1} className="empty">Sin registros</td></tr>
            )}
            {items.map(item => (
              <tr key={item.id}>
                {columns.map(c => <td key={c.key}>{item[c.key] ?? '-'}</td>)}
                <td className="actions">
                  <button className="btn-sm" onClick={() => setForm(item)}>Editar</button>
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

function EntityForm({ title, fields, initial, onSave, onCancel }) {
  const [data, setData] = useState({ ...initial })
  const [busy, setBusy] = useState(false)

  const handleChange = (key, value) => setData(prev => ({ ...prev, [key]: value }))

  const handleSubmit = async e => {
    e.preventDefault()
    setBusy(true)
    await onSave(data)
    setBusy(false)
  }

  return (
    <div className="entity-form-page">
      <h2>{title}</h2>
      <form className="entity-form" onSubmit={handleSubmit}>
        {fields.map(f => (
          <label key={f.key}>
            <span>{f.label}</span>
            {f.type === 'select' ? (
              <select value={data[f.key] || ''} onChange={e => handleChange(f.key, e.target.value)} required={f.required}>
                <option value="">Seleccionar...</option>
                {f.options.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
            ) : f.type === 'number' ? (
              <input type="number" value={data[f.key] || ''} onChange={e => handleChange(f.key, +e.target.value)} required={f.required} />
            ) : (
              <input type="text" value={data[f.key] || ''} onChange={e => handleChange(f.key, e.target.value)} required={f.required} placeholder={f.placeholder} />
            )}
          </label>
        ))}
        <div className="form-actions">
          <button type="submit" className="btn-primary" disabled={busy}>{busy ? 'Guardando...' : 'Guardar'}</button>
          <button type="button" className="btn-secondary" onClick={onCancel}>Cancelar</button>
        </div>
      </form>
    </div>
  )
}
