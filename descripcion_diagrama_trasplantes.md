# Descripción del Diagrama de Clases – Sistema de Gestión de Donación y Trasplante de Órganos

## Visión general

El sistema modela el proceso de donación y trasplante de órganos dentro de hospitales. Los principales actores son:

- Hospitales
- Donantes
- Órganos
- Pacientes
- Trasplantes
- Usuarios del sistema (administradores y médicos)

Además, utiliza varios enumeradores para representar estados, prioridades y roles.

---

# 1. Clase Hospital

## Atributos

- `idHospital: string`
- `nombre: string`
- `ciudad: string`
- `direccion: string`

## Métodos

- `registrarDonante()`
- `buscarOrgano()`

## Relaciones

### Hospital → Donante

Relación: **registra**

Cardinalidad:

- Un `Hospital` registra **muchos Donantes**.
- Cada `Donante` pertenece a **un Hospital**.

Representación:

```text
Hospital (1) -------- (*) Donante
```

---

# 2. Clase Donante

## Atributos

- `idDonante: string`
- `nombre: string`
- `tipoSangre: string`
- `edad: int`
- `detalle: string`

> En el diagrama, `edad` aparece con una barra (`/ int edad`), lo que normalmente indica un atributo derivado o calculado.

## Métodos

- `donarOrgano()`
- `actualizarEstado()`

## Relaciones

### Donante → Órgano

Relación: **dona**

Cardinalidad:

- Un `Donante` puede donar múltiples `Órganos`.
- Cada `Órgano` pertenece a un único `Donante`.

Representación:

```text
Donante (1) -------- (*) Órgano
```

---

# 3. Clase Órgano

## Atributos

- `idOrgano: string`
- `tipo: string`
- `compatibilidad: string`

## Métodos

- `validarCompatibilidad()`

## Relaciones

### Donante → Órgano

```text
Donante (1) -------- (*) Órgano
```

### Órgano ↔ EstadoOrgano

Cada órgano posee un estado representado mediante el enum `EstadoOrgano`.

### Órgano ↔ Trasplante

Relación: **incluye**

Cardinalidad mostrada:

```text
Órgano (1) -------- (1) Trasplante
```

Interpretación:

- Cada trasplante utiliza un único órgano.
- Cada órgano participa en un único trasplante.

---

# 4. Enum EstadoOrgano

Representa la disponibilidad del órgano.

## Valores

- `DISPONIBLE`
- `NO DISPONIBLE`

Uso:

- Asociado a la entidad `Órgano`.

---

# 5. Clase Paciente

## Atributos

- `idPaciente: string`
- `nombre: string`
- `tipoSangre: string`

## Relaciones

### Paciente → Trasplante

Relación: **necesita**

Cardinalidad:

```text
Paciente (1) -------- (*) Trasplante
```

Interpretación:

- Un paciente puede tener múltiples registros o solicitudes de trasplante.
- Cada trasplante corresponde a un único paciente.

### Paciente ↔ PrioridadTrasplante

Cada paciente tiene una prioridad de trasplante.

---

# 6. Enum PrioridadTrasplante

Define el nivel de urgencia de un paciente.

## Valores

- `CRITICA`
- `ALTA`
- `MEDIA`
- `BAJA`

Uso:

- Asociado a `Paciente`.

---

# 7. Clase Trasplante

## Atributos

- `idTrasplante: string`
- `fecha: Date`

## Relaciones

### Paciente → Trasplante

```text
Paciente (1) -------- (*) Trasplante
```

### Trasplante ↔ Órgano

```text
Órgano (1) -------- (1) Trasplante
```

### Trasplante ↔ EstadoTrasplante

Cada trasplante tiene un estado representado mediante el enum `EstadoTrasplante`.

---

# 8. Enum EstadoTrasplante

Representa el ciclo de vida del trasplante.

## Valores

- `PENDIENTE`
- `EN CURSO`
- `REALIZADO`
- `CANCELADO`

Uso:

- Asociado a la entidad `Trasplante`.

---

# 9. Clase Usuario

Representa las cuentas que acceden al sistema.

## Atributos

- `id_usuario: string`
- `nombre_usuario: string`
- `contraseña: string`
- `nombre_visible: string`

## Relaciones

### Usuario ↔ Rol

Cada usuario posee un rol definido mediante el enum `Rol`.

---

# 10. Enum Rol

Define los permisos del usuario.

## Valores

- `ADMINISTRADOR`
- `MEDICO`

Uso:

- Asociado a `Usuario`.

---

# Resumen de entidades

## Clases

1. Hospital
2. Donante
3. Órgano
4. Paciente
5. Trasplante
6. Usuario

## Enumeraciones

1. EstadoOrgano
2. EstadoTrasplante
3. PrioridadTrasplante
4. Rol

---

# Modelo de dominio resumido

```text
Hospital
   │
   └── registra ──> Donante
                         │
                         └── dona ──> Órgano
                                           │
                                           ├── EstadoOrgano
                                           │
                                           └── Trasplante
                                                   │
                                                   ├── EstadoTrasplante
                                                   │
                                                   └── Paciente
                                                            │
                                                            └── PrioridadTrasplante

Usuario
   │
   └── Rol
```

# Reglas de negocio implícitas

1. Un hospital administra el registro de donantes.
2. Un donante puede aportar varios órganos.
3. Cada órgano posee un estado de disponibilidad.
4. Antes de un trasplante debe validarse la compatibilidad del órgano.
5. Los pacientes poseen una prioridad médica para recibir órganos.
6. Todo trasplante tiene un estado operacional.
7. El sistema distingue usuarios administradores y médicos.
8. El trasplante constituye la entidad que conecta un paciente con un órgano disponible.
