package docker

import "testing"

func TestImageInfoMapping(t *testing.T) {
	info := ImageInfo{
		ID:       "sha256:abc123",
		RepoTags: []string{"nginx:latest", "nginx:1.25"},
		Size:     187654321,
		Created:  1700000000,
		InUse:    2,
	}

	if info.ID != "sha256:abc123" {
		t.Errorf("ID = %q", info.ID)
	}
	if len(info.RepoTags) != 2 {
		t.Errorf("RepoTags len = %d", len(info.RepoTags))
	}
	if info.InUse != 2 {
		t.Errorf("InUse = %d", info.InUse)
	}
}
