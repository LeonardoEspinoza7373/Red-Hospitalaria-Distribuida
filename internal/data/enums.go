package data

type EstadoOrgano string

const (
	OrganoDisponible    EstadoOrgano = "DISPONIBLE"
	OrganoNoDisponible  EstadoOrgano = "NO_DISPONIBLE"
)

type EstadoTrasplante string

const (
	TrasplantePendiente  EstadoTrasplante = "PENDIENTE"
	TrasplanteEnCurso    EstadoTrasplante = "EN_CURSO"
	TrasplanteRealizado  EstadoTrasplante = "REALIZADO"
	TrasplanteCancelado  EstadoTrasplante = "CANCELADO"
)

type PrioridadTrasplante string

const (
	PrioridadCritica PrioridadTrasplante = "CRITICA"
	PrioridadAlta    PrioridadTrasplante = "ALTA"
	PrioridadMedia   PrioridadTrasplante = "MEDIA"
	PrioridadBaja    PrioridadTrasplante = "BAJA"
)
