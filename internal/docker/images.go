package docker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/image"
)

// ImageInfo is the view model for an image row.
type ImageInfo struct {
	ID       string
	RepoTags []string
	Size     int64
	Created  int64
	InUse    int64
}

// ListImages returns all local Docker images.
func (s *Service) ListImages(ctx context.Context) ([]ImageInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	images, err := s.cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("image list: %w", err)
	}

	out := make([]ImageInfo, len(images))
	for i, img := range images {
		out[i] = ImageInfo{
			ID:       img.ID,
			RepoTags: img.RepoTags,
			Size:     img.Size,
			Created:  img.Created,
			InUse:    img.Containers,
		}
	}
	return out, nil
}

// PullImage pulls an image by reference (e.g., "nginx:latest").
func (s *Service) PullImage(ctx context.Context, refStr string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	reader, err := s.cli.ImagePull(ctx, refStr, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("pull %s: %w", refStr, err)
	}
	io.Copy(io.Discard, reader)
	reader.Close()
	return nil
}

// RemoveImage removes an image by ID.
func (s *Service) RemoveImage(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	_, err := s.cli.ImageRemove(ctx, id, image.RemoveOptions{PruneChildren: true})
	if err != nil {
		return fmt.Errorf("remove image: %w", err)
	}
	return nil
}
