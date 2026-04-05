package handler

import "net/http"

func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	if err := h.docker.Start(r.Context(), r.PathValue("id")); err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	w.Header().Set("HX-Trigger", "refresh-containers")
	w.WriteHeader(200)
}

func (h *Handler) Stop(w http.ResponseWriter, r *http.Request) {
	if err := h.docker.Stop(r.Context(), r.PathValue("id")); err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	w.Header().Set("HX-Trigger", "refresh-containers")
	w.WriteHeader(200)
}

func (h *Handler) Restart(w http.ResponseWriter, r *http.Request) {
	if err := h.docker.Restart(r.Context(), r.PathValue("id")); err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	w.Header().Set("HX-Trigger", "refresh-containers")
	w.WriteHeader(200)
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	if err := h.docker.Remove(r.Context(), r.PathValue("id")); err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	w.Header().Set("HX-Trigger", "refresh-containers")
	w.WriteHeader(200)
}
