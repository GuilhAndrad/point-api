package handler

import (
	"net/http"
	"time"

	"github.com/GuilhAndrad/point-api/internal/service"
)

type PunchHandler struct {
    svc service.PunchService
}

func NewPunchHandler(svc service.PunchService) *PunchHandler {
    return &PunchHandler{svc: svc}
}

// Punch registra entrada ou saída.
// O tipo (entry/exit) é decidido automaticamente pelo service.
func (h *PunchHandler) Punch(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r.Context()) // extraído do JWT

    record, err := h.svc.Punch(r.Context(), userID)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "erro ao registrar ponto")
        return
    }

    respondJSON(w, http.StatusCreated, record)
}

// MyRecords retorna registros do período (?from=2024-01-01&to=2024-01-31)
func (h *PunchHandler) MyRecords(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r.Context())
    from, to := parseDateRange(r)

    records, err := h.svc.GetMyRecords(r.Context(), userID, from, to)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "erro ao buscar registros")
        return
    }

    respondJSON(w, http.StatusOK, records)
}

// MyReport retorna relatório diário com horas calculadas
func (h *PunchHandler) MyReport(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r.Context())
    from, to := parseDateRange(r)

    report, err := h.svc.GetDailyReport(r.Context(), userID, from, to)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "erro ao gerar relatório")
        return
    }

    respondJSON(w, http.StatusOK, report)
}

func (h *PunchHandler) AllEmployees(w http.ResponseWriter, r *http.Request) {
    from, to := parseDateRange(r)

    records, err := h.svc.GetAllRecords(r.Context(), from, to)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "erro ao buscar registros")
        return
    }

    respondJSON(w, http.StatusOK, records)
}

func (h *PunchHandler) AdminReport(w http.ResponseWriter, r *http.Request) {
    from, to := parseDateRange(r)

    report, err := h.svc.GetAdminReport(r.Context(), from, to)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "erro ao gerar relatório")
        return
    }

    respondJSON(w, http.StatusOK, report)
}

// --- helpers ---

func parseDateRange(r *http.Request) (time.Time, time.Time) {
    layout := "2006-01-02"
    from, _ := time.Parse(layout, r.URL.Query().Get("from"))
    to, _ := time.Parse(layout, r.URL.Query().Get("to"))
    if to.IsZero() {
        to = time.Now()
    }
    if from.IsZero() {
        from = to.AddDate(0, -1, 0) // default: último mês
    }
    return from, to.Add(24*time.Hour - time.Second)
}