const BASE = ''

export async function login(username, password) {
  const res = await fetch(`${BASE}/api/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || 'Error al iniciar sesión')
  return data
}

export async function logout(token) {
  await fetch(`${BASE}/api/logout`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function me(token) {
  const res = await fetch(`${BASE}/api/me`, {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) throw new Error('No autorizado')
  return res.json()
}

function token() { return localStorage.getItem('token') }

function authHeaders() {
  return {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${token()}`,
  }
}

export async function apiFetch(url, options = {}) {
  const res = await fetch(url, { ...options, headers: { ...authHeaders(), ...options.headers } })
  if (res.status === 401) {
    localStorage.removeItem('token')
    window.location.href = '/login'
    throw new Error('Sesión expirada')
  }
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || 'Error de servidor')
  return data
}

export function entityAPI(basePath) {
  return {
    list:   ()          => apiFetch(`${BASE}${basePath}`),
    get:    (id)        => apiFetch(`${BASE}${basePath}/${id}`),
    create: (data)      => apiFetch(`${BASE}${basePath}`, { method: 'POST',  body: JSON.stringify(data) }),
    update: (id, data)  => apiFetch(`${BASE}${basePath}/${id}`, { method: 'PUT',   body: JSON.stringify(data) }),
    delete: (id)        => apiFetch(`${BASE}${basePath}/${id}`, { method: 'DELETE' }),
  }
}

export const pacientesAPI = entityAPI('/api/pacientes')
export const donantesAPI = entityAPI('/api/donantes')
export const organosAPI = entityAPI('/api/organos')
export const trasplantesAPI = entityAPI('/api/trasplantes')
export const usuariosAPI = entityAPI('/api/usuarios')

export function organosCompatibles(pacienteId) {
  return apiFetch(`${BASE}/api/organos/compatibles?paciente_id=${pacienteId}`)
}

export async function acquireLock(resource, ttl) {
  return apiFetch('/api/lock', {
    method: 'POST',
    body: JSON.stringify({ resource, ttl }),
  })
}

export async function releaseLock(resource) {
  return apiFetch('/api/unlock', {
    method: 'POST',
    body: JSON.stringify({ resource }),
  })
}

export async function listLocks() {
  return apiFetch('/api/locks')
}

export function fetchLogs(filters = {}) {
  const params = new URLSearchParams()
  if (filters.node_id) params.set('node_id', filters.node_id)
  if (filters.level) params.set('level', filters.level)
  if (filters.category) params.set('category', filters.category)
  const qs = params.toString()
  return apiFetch(`${BASE}/api/admin/logs${qs ? '?' + qs : ''}`)
}
