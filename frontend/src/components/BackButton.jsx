import { useNavigate } from 'react-router-dom'

export function BackButton({ to = '/', fixed = false }) {
  const navigate = useNavigate()
  return (
    <button
      className={`btn-back ${fixed ? 'btn-back--fixed' : ''}`}
      onClick={() => navigate(to)}
      aria-label="Volver"
      title="Volver"
    >
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        aria-hidden="true"
      >
        <path d="M15 18l-6-6 6-6" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
      <span>Volver</span>
    </button>
  )
}
