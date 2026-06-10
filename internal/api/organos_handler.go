package api

import (
	"net/http"
	"strconv"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/data"
)

type OrganoCompatiblesHandler struct {
	OrganoStore   *data.GenericStore[*data.Organo]
	PacienteStore *data.GenericStore[*data.Paciente]
}

func (h *OrganoCompatiblesHandler) ListCompatibles(w http.ResponseWriter, r *http.Request) {
	pacienteIDStr := r.URL.Query().Get("paciente_id")
	if pacienteIDStr == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "paciente_id is required"})
		return
	}

	pid, err := strconv.Atoi(pacienteIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid paciente_id"})
		return
	}

	paciente, err := h.PacienteStore.GetByID(pid)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "paciente not found"})
		return
	}

	tipoSangre := paciente.TipoSangre

	todos, err := h.OrganoStore.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	compatibles := make([]*data.Organo, 0)
	for _, o := range todos {
		if o.Compatibilidad == tipoSangre && o.Estado == "DISPONIBLE" {
			compatibles = append(compatibles, o)
		}
	}

	if compatibles == nil {
		compatibles = make([]*data.Organo, 0)
	}

	writeJSON(w, http.StatusOK, compatibles)
}
