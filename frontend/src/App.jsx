import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Login } from './pages/Login'
import { Dashboard } from './pages/Dashboard'
import { Pacientes } from './pages/Pacientes'
import { Donantes } from './pages/Donantes'
import { Organos } from './pages/Organos'
import { Donaciones } from './pages/Donaciones'
import { Trasplantes } from './pages/Trasplantes'
import { Usuarios } from './pages/Usuarios'
import { ProtectedRoute } from './components/ProtectedRoute'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
        <Route path="/pacientes" element={<ProtectedRoute><Pacientes /></ProtectedRoute>} />
        <Route path="/donantes" element={<ProtectedRoute><Donantes /></ProtectedRoute>} />
        <Route path="/organos" element={<ProtectedRoute><Organos /></ProtectedRoute>} />
        <Route path="/donaciones" element={<ProtectedRoute><Donaciones /></ProtectedRoute>} />
        <Route path="/trasplantes" element={<ProtectedRoute><Trasplantes /></ProtectedRoute>} />
        <Route path="/usuarios" element={<ProtectedRoute><Usuarios /></ProtectedRoute>} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
