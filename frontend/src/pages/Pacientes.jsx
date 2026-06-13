import { EntityList } from '../components/EntityList'
import { pacientesAPI } from '../api'
import { useAuth } from '../AuthContext'

const hospitals = {
  1: 'Hospital Loja',
  2: 'Hospital Cuenca',
  3: 'Hospital Quito',
  4: 'Hospital Guayaquil',
}

const columns = [
  { key: 'nombre', label: 'Nombre' },
  { key: 'tipo_sangre', label: 'Tipo Sangre' },
  { key: 'prioridad', label: 'Prioridad' },
  { key: 'hospital_id', label: 'Hospital', resolve: { format: id => hospitals[id] || '—' } },
]

const Form = [
  { key: 'nombre', label: 'Nombre', required: true },
  { key: 'tipo_sangre', label: 'Tipo de Sangre', type: 'select', required: true,
    options: [
      { value: 'O+', label: 'O+' },
      { value: 'O-', label: 'O-' },
      { value: 'A+', label: 'A+' },
      { value: 'A-', label: 'A-' },
      { value: 'B+', label: 'B+' },
      { value: 'B-', label: 'B-' },
      { value: 'AB+', label: 'AB+' },
      { value: 'AB-', label: 'AB-' },
    ],
  },
  { key: 'prioridad', label: 'Prioridad', type: 'select', required: true,
    options: [
      { value: 'CRITICA', label: 'Crítica' },
      { value: 'ALTA', label: 'Alta' },
      { value: 'MEDIA', label: 'Media' },
      { value: 'BAJA', label: 'Baja' },
    ],
  },
  { key: 'hospital_id', label: 'Hospital', type: 'select', required: true,
    options: [
      { value: '1', label: 'Hospital Loja' },
      { value: '2', label: 'Hospital Cuenca' },
      { value: '3', label: 'Hospital Quito' },
      { value: '4', label: 'Hospital Guayaquil' },
    ],
  },
]

export function Pacientes() {
  const { user } = useAuth()
  return <EntityList api={pacientesAPI} columns={columns} title="Pacientes" Form={Form} defaults={{ hospital_id: String(user?.hospital_id || '') }} />
}
