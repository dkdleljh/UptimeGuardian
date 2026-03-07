package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"healthcheckbot/internal/checks"
)

type Handler struct {
	store  *checks.Store
	runner *checks.Runner
	apiKey string
}

func NewHandler(store *checks.Store, runner *checks.Runner, apiKey string) *Handler {
	return &Handler{store: store, runner: runner, apiKey: strings.TrimSpace(apiKey)}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /checks", h.listChecks)
	mux.HandleFunc("POST /checks", h.createCheck)
	mux.HandleFunc("GET /checks/{id}", h.getCheck)
	mux.HandleFunc("PUT /checks/{id}", h.updateCheck)
	mux.HandleFunc("DELETE /checks/{id}", h.deleteCheck)
	mux.HandleFunc("POST /checks/{id}/run", h.runCheck)

	return withJSONContentType(h.withAuth(mux))
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *Handler) listChecks(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": h.store.List()})
}

func (h *Handler) createCheck(w http.ResponseWriter, r *http.Request) {
	var in checks.CreateCheckInput
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	item, err := h.store.Create(in)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) getCheck(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	item, err := h.store.Get(id)
	if err != nil {
		if errors.Is(err, checks.ErrNotFound) {
			writeError(w, http.StatusNotFound, "check not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to read check")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) updateCheck(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in checks.UpdateCheckInput
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	item, err := h.store.Update(id, in)
	if err != nil {
		if errors.Is(err, checks.ErrNotFound) {
			writeError(w, http.StatusNotFound, "check not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) deleteCheck(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.store.Delete(id); err != nil {
		if errors.Is(err, checks.ErrNotFound) {
			writeError(w, http.StatusNotFound, "check not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete check")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) runCheck(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	result, err := h.runner.RunCheckNow(r.Context(), id)
	if err != nil {
		if errors.Is(err, checks.ErrNotFound) {
			writeError(w, http.StatusNotFound, "check not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to run check")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]any{"error": message})
}
