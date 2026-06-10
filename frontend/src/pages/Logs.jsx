import { useState, useEffect, useCallback } from 'react'
import { fetchLogs } from '../api'

const hospitals = {
  1: 'Hospital Loja',
  2: 'Hospital Cuenca',
  3: 'Hospital Quito',
  4: 'Hospital Guayaquil',
}

const levels = ['', 'DEBUG', 'INFO', 'WARN', 'ERROR']

const levelColors = {
  DEBUG: 'var(--muted)',
  INFO: 'var(--accent)',
  WARN: '#f59e0b',
  ERROR: 'var(--danger)',
}

function formatTime(ts) {
  const d = new Date(ts * 1000)
  return d.toLocaleTimeString('es-EC', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

export function Logs() {
  const [logs, setLogs] = useState([])
  const [filterNode, setFilterNode] = useState('')
  const [filterLevel, setFilterLevel] = useState('')
  const [loading, setLoading] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchLogs({ node_id: filterNode, level: filterLevel })
      setLogs(Array.isArray(data) ? data.reverse() : [])
    } catch {
      // ignore
    } finally {
      setLoading(false)
    }
  }, [filterNode, filterLevel])

  useEffect(() => {
    load()
    const interval = setInterval(load, 5000)
    return () => clearInterval(interval)
  }, [load])

  return (
    <div className="entity-page">
      <div className="entity-header">
        <h2>Logs del Sistema</h2>
      </div>

      <div className="logs-filters">
        <select value={filterNode} onChange={e => setFilterNode(e.target.value)}>
          <option value="">Todos los nodos</option>
          {Object.entries(hospitals).map(([id, name]) => (
            <option key={id} value={id}>{name}</option>
          ))}
        </select>

        <select value={filterLevel} onChange={e => setFilterLevel(e.target.value)}>
          {levels.map(l => (
            <option key={l} value={l}>{l || 'Todos los niveles'}</option>
          ))}
        </select>

        <button className="btn-primary" onClick={load} disabled={loading}>
          {loading ? 'Cargando...' : 'Actualizar'}
        </button>
      </div>

      <div className="logs-table-wrapper">
        <table className="logs-table">
          <thead>
            <tr>
              <th>Hora</th>
              <th>Nodo</th>
              <th>Nivel</th>
              <th>Mensaje</th>
            </tr>
          </thead>
          <tbody>
            {logs.length === 0 && (
              <tr>
                <td colSpan={4} className="empty">Sin registros</td>
              </tr>
            )}
            {logs.map((log, i) => (
              <tr key={i}>
                <td className="log-time">{formatTime(log.timestamp)}</td>
                <td>{hospitals[log.node_id] || `Nodo ${log.node_id}`}</td>
                <td>
                  <span className="log-level" style={{ color: levelColors[log.level] || 'var(--text)' }}>
                    {log.level}
                  </span>
                </td>
                <td className="log-message">{log.message}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
