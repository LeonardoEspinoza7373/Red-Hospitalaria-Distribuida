import { EntityList } from '../components/EntityList'
import { donantesAPI } from '../api'

const columns = [
  { key: 'nombre', label: 'Nombre' },
  { key: 'tipo_sangre', label: 'Tipo Sangre' },
  { key: 'edad', label: 'Edad' },
  { key: 'detalle', label: 'Detalle' },
]

const Form = [
  { key: 'nombre', label: 'Nombre', required: true },
  { key: 'tipo_sangre', label: 'Tipo de Sangre', required: true, placeholder: 'Ej: O+, A-...' },
  { key: 'edad', label: 'Edad', type: 'number', required: true },
  { key: 'detalle', label: 'Detalle', required: false, placeholder: 'Información adicional' },
]

export function Donantes() {
  return <EntityList api={donantesAPI} columns={columns} title="Donantes" Form={Form} />
}
