import { EntityList } from '../components/EntityList'
import { organosAPI, donantesAPI } from '../api'
import { useAuth } from '../AuthContext'

const hospitals = {
  1: 'Hospital Loja',
  2: 'Hospital Cuenca',
  3: 'Hospital Quito',
  4: 'Hospital Guayaquil',
}

const columns = [
  { key: 'tipo', label: 'Tipo' },
  { key: 'compatibilidad', label: 'Compatibilidad' },
  { key: 'estado', label: 'Estado' },
  { key: 'donante_id', label: 'Donante', resolve: { api: donantesAPI, displayKey: 'nombre' } },
  { key: 'hospital_id', label: 'Hospital', resolve: { format: id => hospitals[id] || '—' } },
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
  { key: 'compatibilidad', label: 'Compatibilidad', required: true, placeholder: 'Se asigna del donante', readOnly: true },
  { key: 'estado', label: 'Estado', type: 'select', required: true,
    options: [
      { value: 'DISPONIBLE', label: 'Disponible' },
      { value: 'NO_DISPONIBLE', label: 'No Disponible' },
    ],
  },
  { key: 'donante_id', label: 'Donante', type: 'async-select', required: true,
    api: donantesAPI, displayKey: 'nombre',
    sync: { field: 'compatibilidad', source: 'tipo_sangre' } },
  { key: 'hospital_id', label: 'Hospital', type: 'select', required: true,
    options: [
      { value: '1', label: 'Hospital Loja' },
      { value: '2', label: 'Hospital Cuenca' },
      { value: '3', label: 'Hospital Quito' },
      { value: '4', label: 'Hospital Guayaquil' },
    ],
  },
]

export function Organos() {
  const { user } = useAuth()
  return <EntityList api={organosAPI} columns={columns} title="Órganos" Form={Form} defaults={{ hospital_id: String(user?.hospital_id || '') }} />
}
