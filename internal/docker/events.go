package docker

import (
	"context"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
)

// ContainerEvents returns channels for container lifecycle events
// (start, stop, die, destroy). The caller must cancel the context to stop.
func (s *Service) ContainerEvents(ctx context.Context) (<-chan events.Message, <-chan error) {
	f := filters.NewArgs()
	f.Add("type", string(events.ContainerEventType))
	f.Add("event", "start")
	f.Add("event", "stop")
	f.Add("event", "die")
	f.Add("event", "destroy")
	f.Add("event", "pause")
	f.Add("event", "unpause")

	return s.cli.Events(ctx, events.ListOptions{Filters: f})
}
