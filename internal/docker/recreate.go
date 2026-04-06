package docker

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
)

const recreateTimeout = 120 * time.Second

// Recreate pulls the latest image for a container and recreates it with the
// same configuration. The old container is stopped and removed, and a new one
// is created and started with the original name, config, and network settings.
func (s *Service) Recreate(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, recreateTimeout)
	defer cancel()

	// 1. Inspect the existing container to capture its full config.
	info, err := s.cli.ContainerInspect(ctx, id)
	if err != nil {
		return fmt.Errorf("inspect: %w", err)
	}

	name := strings.TrimPrefix(info.Name, "/")
	imageName := info.Config.Image

	// 2. Pull the latest version of the image.
	reader, err := s.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("pull %s: %w", imageName, err)
	}
	// Must drain the reader for the pull to complete.
	io.Copy(io.Discard, reader)
	reader.Close()

	// 3. Stop the container if running (ignore "not running" errors).
	_ = s.cli.ContainerStop(ctx, id, container.StopOptions{})

	// 4. Remove the old container so the name can be reused.
	if err := s.cli.ContainerRemove(ctx, id, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("remove old container: %w", err)
	}

	// 5. Build networking config from the old container's networks.
	var netConfig *network.NetworkingConfig
	if info.NetworkSettings != nil && len(info.NetworkSettings.Networks) > 0 {
		endpoints := make(map[string]*network.EndpointSettings)
		for netName, ep := range info.NetworkSettings.Networks {
			endpoints[netName] = ep
		}
		netConfig = &network.NetworkingConfig{EndpointsConfig: endpoints}
	}

	// 6. Create a new container with the same config.
	resp, err := s.cli.ContainerCreate(ctx, info.Config, info.HostConfig, netConfig, nil, name)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	// 7. Start the new container.
	if err := s.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	return nil
}
