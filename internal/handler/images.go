package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/emm5317/miniport/internal/docker"
)

// ImageRow is the view model for a single image row in the template.
type ImageRow struct {
	ID      string
	ShortID string
	Repo    string
	Tag     string
	Size    uint64
	Created string
	InUse   int64
}

// buildImageRows converts ImageInfo to template-friendly ImageRows.
// Each RepoTag gets its own row; untagged images get a single row.
func buildImageRows(images []docker.ImageInfo) []ImageRow {
	var rows []ImageRow
	for _, img := range images {
		shortID := img.ID
		if strings.HasPrefix(shortID, "sha256:") && len(shortID) > 19 {
			shortID = shortID[7:19]
		}
		created := time.Unix(img.Created, 0).Format("2006-01-02 15:04")

		if len(img.RepoTags) == 0 {
			rows = append(rows, ImageRow{
				ID:      img.ID,
				ShortID: shortID,
				Repo:    "<none>",
				Tag:     "<none>",
				Size:    uint64(img.Size),
				Created: created,
				InUse:   img.InUse,
			})
			continue
		}
		for _, tag := range img.RepoTags {
			repo, tagName, _ := strings.Cut(tag, ":")
			rows = append(rows, ImageRow{
				ID:      img.ID,
				ShortID: shortID,
				Repo:    repo,
				Tag:     tagName,
				Size:    uint64(img.Size),
				Created: created,
				InUse:   img.InUse,
			})
		}
	}
	return rows
}

func (h *Handler) ImageList(w http.ResponseWriter, r *http.Request) {
	images, err := h.docker.ListImages(r.Context())
	if err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	renderPartial(w, "image-table.html", map[string]any{
		"Images": buildImageRows(images),
	})
}

func (h *Handler) ImagePull(w http.ResponseWriter, r *http.Request) {
	ref := strings.TrimSpace(r.FormValue("ref"))
	if ref == "" {
		httpError(w, "Image reference required", 400)
		return
	}
	if err := h.docker.PullImage(r.Context(), ref); err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	setActionTrigger(w, "refresh-images", fmt.Sprintf("Pulled %s", ref), "success")
	w.WriteHeader(200)
}

func (h *Handler) ImageRemove(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.docker.RemoveImage(r.Context(), id); err != nil {
		httpError(w, err.Error(), 500)
		return
	}
	setActionTrigger(w, "refresh-images", "Image removed", "success")
	w.WriteHeader(200)
}
