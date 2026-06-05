# Red Hospitalaria Distribuida — Especificación del Sistema

## 1. Visión General

Sistema distribuido de simulación para la gestión de una red hospitalaria
enfocada en **donación y trasplante de órganos**. Cada nodo representa un
hospital en una zona geográfica distinta. Un coordinador (elegido por el
algoritmo Bully) mantiene un índice parcial de la información de todos los
nodos y coordina las búsquedas.

---

## 2. Arquitectura

### 2.1 Topología

| Nodo | ID | IP | Hospital | Prioridad |
|------|----|------|-----------|-----------|
| Nodo 1 | 1 | 192.168.1.13 | Hospital Loja | Baja |
| Nodo 2 | 2 | 192.168.1.12 | Hospital Cuenca | Media |
| Nodo 3 | 3 | 192.168.1.11 | Hospital Quito | Alta |
| Nodo 4 | 4 | 192.168.1.10 | Hospital Guayaquil | Más alta |
| Proxy | — | 192.168.1.14 | AP WiFi + Reverse Proxy | — |

### 2.2 Capas

```
┌──────────────────────────────────────────────────┐
│                  Frontend SPA                     │
│       (servido por el coordinador :8080)          │
├──────────────────────────────────────────────────┤
│              API REST (HTTP/JSON)                  │
│         coordinador auth + búsquedas              │
├──────────────────────────────────────────────────┤
│         Capa de Negocio (dominio hospital)        │
│   usuarios | pacientes | donantes | órganos       │
│   donaciones | trasplantes | logs                 │
├──────────────────────────────────────────────────┤
│        Almacenamiento Local (JSON por nodo)       │
│      + Índice Parcial en el coordinador           │
├──────────────────────────────────────────────────┤
│         Capa Distribuida (Bully + TCP)            │
│   heartbeat | elección | replicación entre nodos  │
├──────────────────────────────────────────────────┤
│          Transporte TCP :5000 + HTTP :8080        │
│            Proxy reverso :8080 → coordinador      │
└──────────────────────────────────────────────────┘
```

### 2.3 Flujo de una petición

```
Usuario → Proxy (:8080) → Coordinador (:8080)
                              ├── responde con índice local
                              └── consulta a nodo específico (TCP :5000)
```

---

## 3. Modelo de Datos

### 3.1 Usuarios

Cada nodo almacena sus propios usuarios localmente. El coordinador mantiene
una copia parcial (índice) de todos los usuarios para autenticación global.

```
Usuario {
    id:           number (único por nodo)
    username:     string (único global)
    password:     string (hash SHA-256)
    display_name: string
    role:         "admin" | "doctor"
    hospital_id:  number (ID del nodo al que pertenece)
    created_at:   string (ISO 8601)
    updated_at:   string (ISO 8601)
}
```

#### Roles

| Rol | Permisos |
|-----|----------|
| **admin** | CRUD usuarios, monitoreo salud del sistema, logs |
| **doctor** | CRUD pacientes, donantes, órganos, donaciones, trasplantes |

### 3.2 Entidades futuras (en orden de prioridad)

1. **Paciente** — datos del paciente en lista de espera
2. **Donante** — donante voluntario o fallecido
3. **Órgano** — tipo, estado, compatibilidad
4. **Donación** — registro de extracción
5. **Trasplante** — asignación órgano → paciente
6. **Log del sistema** — auditoría de operaciones

### 3.3 Índice Parcial del Coordinador

El coordinador mantiene un resumen de cada entidad de todos los nodos:

- **Usuarios**: `(id, username, display_name, role, hospital_id)`
- **Pacientes** (futuro): `(id, nombre, grupo_sanguineo, hospital_id)`
- **Órganos** (futuro): `(id, tipo, estado, hospital_id)`

Cuando el coordinador necesita más detalle, consulta al nodo específico
vía TCP :5000.

### 3.4 Replicación

Cuando un nodo crea/actualiza una entidad, notifica al coordinador mediante
un mensaje TCP para que actualice su índice parcial.

```
Nodo → TCP REPLICATE { entity, data } → Coordinador
```

*Nota: la replicación se implementará en una fase posterior.*

---

## 4. Almacenamiento

- **Formato**: JSON en archivos planos
- **Ubicación**: `data/{hospital_id}/` en cada nodo
- **Estructura**: un archivo por entidad (`users.json`, `patients.json`, etc.)
- **Sin dependencias externas**: solo `encoding/json` de la stdlib
- **Concurrencia**: acceso sincronizado con `sync.RWMutex`

---

## 5. Autenticación y Sesiones (Prioridad 1)

### 5.1 Endpoints

| Método | Ruta | Descripción |
|--------|------|-------------|
| POST | `/api/login` | Iniciar sesión |
| POST | `/api/logout` | Cerrar sesión |
| GET | `/api/me` | Obtener usuario actual |

### 5.2 Flujo de Login

```
1. Cliente → POST /api/login { username, password }
2. Proxy reenvía al coordinador
3. Coordinador busca username en su índice de usuarios
4. Verifica hash de password (SHA-256)
5. Si ok: crea sesión, devuelve { token, user }
   Si no: 401 Unauthorized
6. Cliente guarda token, lo envía en header Authorization: Bearer <token>
```

### 5.3 Sesiones

- Token: UUID v4
- Almacenamiento: mapa en memoria en el coordinador
- Expiración: 24 horas (configurable)
- Estado: `active` | `expired`

### 5.4 Middleware de Autenticación

Todas las rutas `/api/*` (excepto `/api/login`) requieren token válido.
El middleware extrae el token del header `Authorization`, lo valida y
agrega el usuario al contexto de la petición.

---

## 6. Roadmap

| Fase | Tarea | Prioridad |
|------|-------|-----------|
| 1 | Login + gestión de usuarios | **Ahora** |
| 2 | CRUD Pacientes | Próximo |
| 3 | CRUD Donantes | Próximo |
| 4 | CRUD Órganos | Próximo |
| 5 | Registro de Donaciones | Siguiente |
| 6 | Registro de Trasplantes | Siguiente |
| 7 | Replicación entre nodos | Siguiente |
| 8 | Índice parcial de búsqueda | Siguiente |
| 9 | Monitoreo y logs | Futuro |
| 10 | Frontend SPA | Futuro |

---

## 7. Convenciones

- **Lenguaje**: Go (solo stdlib)
- **Comunicación entre nodos**: TCP :5000, mensajes JSON delimitados por `\n`
- **API REST**: JSON, puerto 8080 en el coordinador
- **Estilo**: sin comentarios en código (excepto doc strings públicos si son necesarios)
- **Tests**: `go test ./internal/...` debe pasar siempre
