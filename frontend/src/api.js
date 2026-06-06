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
