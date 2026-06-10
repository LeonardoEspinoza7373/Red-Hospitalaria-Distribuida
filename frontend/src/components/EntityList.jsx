import { useState, useEffect } from 'react'

export function EntityList({ api, columns, title, Form }) {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [form, setForm] = useState(null)
  const [relatedData, setRelatedData] = useState({})

  const load = () => {
    setLoading(true)
    const resolveCols = columns.filter(c => c.resolve?.api)
    const apis = resolveCols.length > 0
      ? [...new Set(resolveCols.map(c => c.resolve.api))]
      : []

    Promise.all([api.list(), ...apis.map(a => a.list())])
      .then(([items, ...relatedResults]) => {
        setItems(items)
        if (resolveCols.length > 0) {
          const related = {}
          for (const col of resolveCols) {
            const apiIndex = apis.indexOf(col.resolve.api)
            const data = relatedResults[apiIndex]
            const map = {}
            for (const item of data) {
              map[item.id] = item[col.resolve.displayKey] || item.id
            }
            related[col.key] = map
          }
          setRelatedData(related)
        }
      })
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
      onCancel={() => { setForm(null); setError('') }}
      error={error}
    />
  )

  return (
    <div className="entity-page">
      <header className="entity-header">
        <div className="entity-header-left">
          <h2>{title}</h2>
        </div>
        <button className="btn-primary" onClick={() => { setForm({}); setError('') }}>+ Nuevo</button>
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
                {columns.map(c => {
                  let display
                  if (c.resolve?.format) {
                    display = c.resolve.format(item[c.key])
                  } else if (c.resolve?.api) {
                    display = relatedData[c.key]?.[item[c.key]]
                  } else {
                    display = item[c.key]
                  }
                  return <td key={c.key}>{display ?? '-'}</td>
                })}
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

function EntityForm({ title, fields, initial, onSave, onCancel, error: serverError }) {
  const [data, setData] = useState({ ...initial })
  const [busy, setBusy] = useState(false)
  const [formError, setFormError] = useState('')
  const [showPw, setShowPw] = useState({})
  const [asyncOptions, setAsyncOptions] = useState({})

  useEffect(() => {
    const asyncFields = fields.filter(f => f.type === 'async-select')
    for (const f of asyncFields) {
      f.api.list().then(items => {
        setAsyncOptions(prev => ({ ...prev, [f.key]: items }))
      })
    }
  }, [fields])

  const handleChange = (key, value) => {
    const field = fields.find(f => f.key === key)
    setData(prev => {
      const next = { ...prev, [key]: value }
      if (field?.sync && field.type === 'async-select') {
        const items = asyncOptions[key] || []
        const selected = items.find(i => String(i.id) === String(value))
        if (selected) next[field.sync.field] = selected[field.sync.source]
      }
      return next
    })
  }

  const handleSubmit = async e => {
    e.preventDefault()
    setFormError('')

    const hasPwConfirm = fields.some(f => f.key === 'password_confirm')
    if (hasPwConfirm && data.password && data.password !== data.password_confirm) {
      setFormError('Las contraseñas no coinciden')
      return
    }

    const payload = { ...data }
    delete payload.password_confirm
    for (const key of Object.keys(payload)) {
      const field = fields.find(f => f.key === key)
      if ((field?.type === 'select' || field?.type === 'async-select') && typeof payload[key] === 'string') {
        const num = Number(payload[key])
        if (!isNaN(num)) payload[key] = num
      }
    }
    setBusy(true)
    await onSave(payload)
    setBusy(false)
  }

  const displayError = formError || serverError

  return (
    <div className="entity-form-page">
      <h2>{title}</h2>
      {displayError && <div className="error">{displayError}</div>}
      <form className="entity-form" onSubmit={handleSubmit}>
        {fields.map(f => (
          <label key={f.key}>
            <span>{f.label}</span>
            {f.type === 'select' ? (
              <select value={data[f.key] || ''} onChange={e => handleChange(f.key, e.target.value)} required={f.required}>
                <option value="">Seleccionar...</option>
                {f.options.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
            ) : f.type === 'async-select' ? (
              <select value={data[f.key] || ''} onChange={e => handleChange(f.key, e.target.value)} required={f.required}>
                <option value="">Seleccionar {f.label.toLowerCase()}...</option>
                {(asyncOptions[f.key] || []).map(item => (
                  <option key={item.id} value={item.id}>
                    {f.format ? f.format(item) : item[f.displayKey] || item.id}
                  </option>
                ))}
              </select>
            ) : f.type === 'number' ? (
              <input type="number" value={data[f.key] || ''} onChange={e => handleChange(f.key, +e.target.value)} required={f.required} />
            ) : f.type === 'password' ? (
              <div className="password-wrapper">
                <input
                  type={showPw[f.key] ? 'text' : 'password'}
                  value={data[f.key] || ''}
                  onChange={e => handleChange(f.key, e.target.value)}
                  required={f.required}
                  placeholder={initial.id ? f.placeholder : ''}
                />
                <button type="button" className="btn-toggle-pw" onClick={() => setShowPw(prev => ({ ...prev, [f.key]: !prev[f.key] }))} tabIndex={-1}>
                  {showPw[f.key] ? '🙈' : '👁️'}
                </button>
              </div>
            ) : (
              <input type="text" value={data[f.key] || ''} onChange={e => handleChange(f.key, e.target.value)} required={!f.readOnly && f.required} placeholder={f.placeholder} readOnly={f.readOnly} />
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
