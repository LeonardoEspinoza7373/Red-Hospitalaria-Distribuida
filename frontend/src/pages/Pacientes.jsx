import { EntityList } from '../components/EntityList'
import { pacientesAPI } from '../api'

const columns = [
  { key: 'nombre', label: 'Nombre' },
  { key: 'tipo_sangre', label: 'Tipo Sangre' },
  { key: 'prioridad', label: 'Prioridad' },
]

const Form = [
  { key: 'nombre', label: 'Nombre', required: true },
  { key: 'tipo_sangre', label: 'Tipo de Sangre', required: true, placeholder: 'Ej: O+, A-...' },
  { key: 'prioridad', label: 'Prioridad', type: 'select', required: true,
    options: [
      { value: 'CRITICA', label: 'Crítica' },
      { value: 'ALTA', label: 'Alta' },
      { value: 'MEDIA', label: 'Media' },
      { value: 'BAJA', label: 'Baja' },
    ],
  },
]

export function Pacientes() {
  return <EntityList api={pacientesAPI} columns={columns} title="Pacientes" Form={Form} />
}
