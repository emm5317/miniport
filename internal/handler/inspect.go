package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// InspectData is the curated view model for container inspection.
type InspectData struct {
	ID            string
	Name          string
	Image         string
	Command       string
	Created       string
	Env           []EnvVar
	Mounts        []MountInfo
	Ports         []PortBinding
	Networks      []NetworkInfo
	Labels        map[string]string
	RestartPolicy string
	RevealEnv     bool
}

type EnvVar struct {
	Key   string
	Value string
}

type MountInfo struct {
	Type        string
	Source      string
	Destination string
	Mode        string
}

type PortBinding struct {
	Container string
	Host      string
	Protocol  string
}

type NetworkInfo struct {
	Name      string
	IPAddress string
	Gateway   string
	MacAddr   string
}

// MaskValue masks an environment variable value, showing only first 3 chars.
func MaskValue(v string) string {
	if len(v) <= 3 {
		return "***"
	}
	return v[:3] + strings.Repeat("*", min(len(v)-3, 12))
}

func (h *Handler) Inspect(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	reveal := r.URL.Query().Get("reveal") == "true"

	info, err := h.docker.Inspect(r.Context(), id)
	if err != nil {
		httpError(w, "Failed to inspect container: "+err.Error(), 500)
		return
	}

	data := InspectData{
		ID:        id,
		Name:      strings.TrimPrefix(info.Name, "/"),
		Image:     info.Config.Image,
		Command:   strings.Join(info.Config.Cmd, " "),
		Created:   formatCreated(info.Created),
		RevealEnv: reveal,
	}

	// Env
	for _, e := range info.Config.Env {
		k, v, _ := strings.Cut(e, "=")
		if !reveal {
			v = MaskValue(v)
		}
		data.Env = append(data.Env, EnvVar{Key: k, Value: v})
	}

	// Mounts
	for _, m := range info.Mounts {
		data.Mounts = append(data.Mounts, MountInfo{
			Type:        string(m.Type),
			Source:      m.Source,
			Destination: m.Destination,
			Mode:        m.Mode,
		})
	}

	// Ports
	if info.NetworkSettings != nil {
		for port, bindings := range info.NetworkSettings.Ports {
			for _, b := range bindings {
				data.Ports = append(data.Ports, PortBinding{
					Container: string(port),
					Host:      fmt.Sprintf("%s:%s", b.HostIP, b.HostPort),
					Protocol:  port.Proto(),
				})
			}
			if len(bindings) == 0 {
				data.Ports = append(data.Ports, PortBinding{
					Container: string(port),
					Host:      "—",
					Protocol:  port.Proto(),
				})
			}
		}
	}

	// Networks
	if info.NetworkSettings != nil {
		for name, net := range info.NetworkSettings.Networks {
			data.Networks = append(data.Networks, NetworkInfo{
				Name:      name,
				IPAddress: net.IPAddress,
				Gateway:   net.Gateway,
				MacAddr:   net.MacAddress,
			})
		}
	}

	// Labels
	data.Labels = info.Config.Labels

	// Restart policy
	if info.HostConfig != nil {
		rp := info.HostConfig.RestartPolicy
		data.RestartPolicy = string(rp.Name)
		if rp.MaximumRetryCount > 0 {
			data.RestartPolicy += fmt.Sprintf(" (max %d)", rp.MaximumRetryCount)
		}
	}

	renderPartial(w, "inspect-panel.html", data)
}

func formatCreated(s string) string {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return s
	}
	return t.Format("2006-01-02 15:04:05 MST")
}
