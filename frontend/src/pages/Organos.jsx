import { EntityList } from '../components/EntityList'
import { organosAPI } from '../api'

const columns = [
  { key: 'id', label: 'ID' },
  { key: 'tipo', label: 'Tipo' },
  { key: 'compatibilidad', label: 'Compatibilidad' },
  { key: 'estado', label: 'Estado' },
  { key: 'donante_id', label: 'Donante ID' },
]

const Form = [
  { key: 'tipo', label: 'Tipo de Órgano', required: true,
    options: [
      { value: 'CORAZON', label: 'Corazón' },
      { value: 'PULMON', label: 'Pulmón' },
      { value: 'HIGADO', label: 'Hígado' },
      { value: 'RIÑON', label: 'Riñón' },
      { value: 'PANCREAS', label: 'Páncreas' },
      { value: 'INTESTINO', label: 'Intestino' },
      { value: 'CORNEA', label: 'Córnea' },
      { value: 'MEDULA_OSEA', label: 'Médula Ósea' },
    ],
  },
  { key: 'compatibilidad', label: 'Compatibilidad', required: true, placeholder: 'Ej: O+, A-...' },
  { key: 'estado', label: 'Estado', type: 'select', required: true,
    options: [
      { value: 'DISPONIBLE', label: 'Disponible' },
      { value: 'NO_DISPONIBLE', label: 'No Disponible' },
    ],
  },
  { key: 'donante_id', label: 'ID del Donante', type: 'number', required: true },
]

export function Organos() {
  return <EntityList api={organosAPI} columns={columns} title="Órganos" Form={Form} />
}
