package handler

import (
	"encoding/json"
	"net/http"
)

// setActionTrigger sets HX-Trigger with a refresh key and a toast message.
func setActionTrigger(w http.ResponseWriter, refreshKey, msg, toastType string) {
	trigger := map[string]any{
		refreshKey:  "",
		"showToast": map[string]string{"msg": msg, "type": toastType},
	}
	b, _ := json.Marshal(trigger)
	w.Header().Set("HX-Trigger", string(b))
}

// setTriggers sets HX-Trigger with refresh-containers and a toast message.
func setTriggers(w http.ResponseWriter, msg, toastType string) {
	setActionTrigger(w, "refresh-containers", msg, toastType)
}

func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	if err := h.docker.Start(r.Context(), r.PathValue("id")); err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	setTriggers(w, "Container started", "success")
	w.WriteHeader(200)
}

func (h *Handler) Stop(w http.ResponseWriter, r *http.Request) {
	if err := h.docker.Stop(r.Context(), r.PathValue("id")); err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	setTriggers(w, "Container stopped", "success")
	w.WriteHeader(200)
}

func (h *Handler) Restart(w http.ResponseWriter, r *http.Request) {
	if err := h.docker.Restart(r.Context(), r.PathValue("id")); err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	setTriggers(w, "Container restarted", "success")
	w.WriteHeader(200)
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	if err := h.docker.Remove(r.Context(), r.PathValue("id")); err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	setTriggers(w, "Container removed", "success")
	w.WriteHeader(200)
}

func (h *Handler) Recreate(w http.ResponseWriter, r *http.Request) {
	if err := h.docker.Recreate(r.Context(), r.PathValue("id")); err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	setTriggers(w, "Container recreated with latest image", "success")
	w.WriteHeader(200)
}
