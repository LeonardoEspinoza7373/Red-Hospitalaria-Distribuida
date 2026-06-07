import { EntityList } from '../components/EntityList'
import { usuariosAPI } from '../api'

const columns = [
  { key: 'id', label: 'ID' },
  { key: 'username', label: 'Usuario' },
  { key: 'display_name', label: 'Nombre' },
  { key: 'role', label: 'Rol' },
  { key: 'hospital_id', label: 'Hospital ID' },
]

const Form = [
  { key: 'username', label: 'Usuario', required: true },
  { key: 'password', label: 'Contraseña', required: false, placeholder: 'Solo si desea cambiarla' },
  { key: 'display_name', label: 'Nombre Completo', required: true },
  { key: 'role', label: 'Rol', type: 'select', required: true,
    options: [
      { value: 'admin', label: 'Administrador' },
      { value: 'doctor', label: 'Médico' },
    ],
  },
  { key: 'hospital_id', label: 'Hospital', type: 'select', required: true,
    options: [
      { value: '1', label: 'Hospital Loja (1)' },
      { value: '2', label: 'Hospital Cuenca (2)' },
      { value: '3', label: 'Hospital Quito (3)' },
      { value: '4', label: 'Hospital Guayaquil (4)' },
    ],
  },
]

export function Usuarios() {
  return <EntityList api={usuariosAPI} columns={columns} title="Usuarios" Form={Form} />
}
