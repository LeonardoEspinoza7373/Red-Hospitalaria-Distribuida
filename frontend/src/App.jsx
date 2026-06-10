import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Login } from './pages/Login'
import { Dashboard } from './pages/Dashboard'
import { Pacientes } from './pages/Pacientes'
import { Donantes } from './pages/Donantes'
import { Organos } from './pages/Organos'
import { Donaciones } from './pages/Donaciones'
import { Trasplantes } from './pages/Trasplantes'
import { Usuarios } from './pages/Usuarios'
import { Logs } from './pages/Logs'
import { Layout } from './components/Layout'
import { ProtectedRoute } from './components/ProtectedRoute'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<ProtectedRoute><Layout><Dashboard /></Layout></ProtectedRoute>} />
        <Route path="/pacientes" element={<ProtectedRoute><Layout><Pacientes /></Layout></ProtectedRoute>} />
        <Route path="/donantes" element={<ProtectedRoute><Layout><Donantes /></Layout></ProtectedRoute>} />
        <Route path="/organos" element={<ProtectedRoute><Layout><Organos /></Layout></ProtectedRoute>} />
        <Route path="/donaciones" element={<ProtectedRoute><Layout><Donaciones /></Layout></ProtectedRoute>} />
        <Route path="/trasplantes" element={<ProtectedRoute><Layout><Trasplantes /></Layout></ProtectedRoute>} />
        <Route path="/usuarios" element={<ProtectedRoute><Layout><Usuarios /></Layout></ProtectedRoute>} />
        <Route path="/logs" element={<ProtectedRoute><Layout><Logs /></Layout></ProtectedRoute>} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
