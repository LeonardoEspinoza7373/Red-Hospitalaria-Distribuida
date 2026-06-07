import { EntityList } from '../components/EntityList'
import { trasplantesAPI } from '../api'

const columns = [
  { key: 'id', label: 'ID' },
  { key: 'paciente_id', label: 'Paciente ID' },
  { key: 'organo_id', label: 'Órgano ID' },
  { key: 'fecha', label: 'Fecha' },
  { key: 'estado', label: 'Estado' },
]

const Form = [
  { key: 'paciente_id', label: 'ID del Paciente', type: 'number', required: true },
  { key: 'organo_id', label: 'ID del Órgano', type: 'number', required: true },
  { key: 'fecha', label: 'Fecha', required: true, placeholder: 'YYYY-MM-DD' },
  { key: 'estado', label: 'Estado', type: 'select', required: true,
    options: [
      { value: 'PENDIENTE', label: 'Pendiente' },
      { value: 'EN_CURSO', label: 'En Curso' },
      { value: 'REALIZADO', label: 'Realizado' },
      { value: 'CANCELADO', label: 'Cancelado' },
    ],
  },
]

export function Trasplantes() {
  return <EntityList api={trasplantesAPI} columns={columns} title="Trasplantes" Form={Form} />
}
