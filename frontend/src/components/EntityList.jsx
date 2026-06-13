import { useState, useEffect, useRef } from 'react'
import { acquireLock, releaseLock } from '../api'
import { IconEye, IconEyeOff } from '../icons'
import { useAuth } from '../AuthContext'

const hospitals = {
  1: 'Hospital Loja',
  2: 'Hospital Cuenca',
  3: 'Hospital Quito',
  4: 'Hospital Guayaquil',
}

export function EntityList({ api, columns, title, Form, onFormChange, defaults }) {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [form, setForm] = useState(null)
  const [relatedData, setRelatedData] = useState({})
  const [formLocked, setFormLocked] = useState(null)
  const formResourceRef = useRef(null)
  const renewRef = useRef(null)
  const { user } = useAuth()

  const openForm = async (data) => {
    const resource = data.id ? `${title.toLowerCase()}:${data.id}` : `${title.toLowerCase()}:new`
    try {
      const result = await acquireLock(resource)
      if (result.acquired) {
        formResourceRef.current = resource
        setForm(data)
        setFormLocked(null)
        renewRef.current = setInterval(async () => {
          try { await acquireLock(resource) } catch {}
        }, 30000)
      } else if (result.user_id === user?.username) {
        formResourceRef.current = resource
        setForm(data)
        setFormLocked(null)
      } else {
        setFormLocked({ userName: result.user_name, resource })
      }
    } catch {
      setForm(data)
    }
    onFormChange?.(true)
  }

  const closeForm = () => {
    if (renewRef.current) {
      clearInterval(renewRef.current)
      renewRef.current = null
    }
    if (formResourceRef.current) {
      releaseLock(formResourceRef.current).catch(() => {})
      formResourceRef.current = null
    }
    setForm(null)
    setFormLocked(null)
    setError('')
    onFormChange?.(false)
  }

  useEffect(() => () => {
    if (renewRef.current) clearInterval(renewRef.current)
    if (formResourceRef.current) releaseLock(formResourceRef.current).catch(() => {})
  }, [])

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
    if (!confirm('¿Dar de baja este registro?')) return
    try {
      await api.delete(id)
      load()
    } catch (e) { setError(e.message) }
  }

  const handleSave = async data => {
    try {
      if (data.id) await api.update(data.id, data)
      else await api.create(data)
      closeForm()
      load()
    } catch (e) { setError(e.message) }
  }

  if (formLocked) return (
    <div className="entity-form-page">
      <h2>{title}</h2>
      <div className="lock-notice">
        <p>Este recurso está siendo editado por <strong>{formLocked.userName}</strong>.</p>
        <button className="btn-secondary" onClick={() => setFormLocked(null)}>Volver</button>
      </div>
    </div>
  )

  if (form) return (
    <EntityForm
      title={form.id ? `Editar ${title}` : `Nuevo ${title}`}
      fields={Form}
      initial={form}
      onSave={handleSave}
      onCancel={closeForm}
      error={error}
    />
  )

  return (
    <div className="entity-page">
      <header className="entity-header">
        <div className="entity-header-left">
          <h2>{title}</h2>
        </div>
        <button className="btn-primary" onClick={() => openForm(defaults ? { ...defaults } : {})}>+ Nuevo</button>
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
                  <button className="btn-sm" onClick={() => openForm(item)}>Editar</button>
                  <button className="btn-sm danger" onClick={() => handleDelete(item.id)}>Dar de baja</button>
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
  const [data, setData] = useState(() => {
    const converted = {}
    for (const key of Object.keys(initial)) {
      const field = fields.find(f => f.key === key)
      if (field?.type === 'select' && typeof initial[key] === 'number') {
        converted[key] = String(initial[key])
      } else {
        converted[key] = initial[key]
      }
    }
    return converted
  })
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
              <select value={data[f.key] || ''} onChange={e => handleChange(f.key, e.target.value)} required={f.required} disabled={f.readOnly}>
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
            ) : f.type === 'date' ? (
              <input type="date" value={data[f.key] || ''} onChange={e => handleChange(f.key, e.target.value)} required={f.required} />
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
                  {showPw[f.key] ? <IconEyeOff /> : <IconEye />}
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
