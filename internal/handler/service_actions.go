package handler

import (
	"encoding/json"
	"net/http"

	"github.com/emm5317/miniport/internal/systemd"
)

func setServiceTriggers(w http.ResponseWriter, msg, toastType string) {
	trigger := map[string]any{
		"refresh-services": "",
		"showToast":        map[string]string{"msg": msg, "type": toastType},
	}
	b, _ := json.Marshal(trigger)
	w.Header().Set("HX-Trigger", string(b))
}

func (h *Handler) ServiceStart(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := systemd.Start(r.Context(), name); err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	setServiceTriggers(w, "Service "+name+" started", "success")
	w.WriteHeader(200)
}

func (h *Handler) ServiceStop(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := systemd.Stop(r.Context(), name); err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	setServiceTriggers(w, "Service "+name+" stopped", "success")
	w.WriteHeader(200)
}

func (h *Handler) ServiceRestart(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := systemd.Restart(r.Context(), name); err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	setServiceTriggers(w, "Service "+name+" restarted", "success")
	w.WriteHeader(200)
}
