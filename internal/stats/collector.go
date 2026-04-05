package stats

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/emm5317/miniport/internal/docker"
	"github.com/emm5317/miniport/internal/host"
	"github.com/emm5317/miniport/internal/systemd"
)

// ServiceSnapshot holds a point-in-time reading of a systemd service.
type ServiceSnapshot struct {
	Name        string
	ActiveState string
	SubState    string
	MemCurrent  uint64
	CPUPercent  float64
	StartedAt   string
	NRestarts   int
	Description string
	UnitEnabled string
}

// RingBuffer is a fixed-size circular buffer.
type RingBuffer[T any] struct {
	data []T
	pos  int
	full bool
}

// NewRingBuffer creates a ring buffer with the given capacity.
func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	return &RingBuffer[T]{data: make([]T, capacity)}
}

// Push adds an item, overwriting the oldest if full.
func (r *RingBuffer[T]) Push(item T) {
	r.data[r.pos] = item
	r.pos = (r.pos + 1) % len(r.data)
	if r.pos == 0 {
		r.full = true
	}
}

// Slice returns all items in chronological order.
func (r *RingBuffer[T]) Slice() []T {
	if !r.full {
		return append([]T(nil), r.data[:r.pos]...)
	}
	out := make([]T, len(r.data))
	copy(out, r.data[r.pos:])
	copy(out[len(r.data)-r.pos:], r.data[:r.pos])
	return out
}

// Last returns the most recent item and true, or zero value and false if empty.
func (r *RingBuffer[T]) Last() (T, bool) {
	var zero T
	if !r.full && r.pos == 0 {
		return zero, false
	}
	idx := r.pos - 1
	if idx < 0 {
		idx = len(r.data) - 1
	}
	return r.data[idx], true
}

// Len returns the number of items currently stored.
func (r *RingBuffer[T]) Len() int {
	if r.full {
		return len(r.data)
	}
	return r.pos
}

// Collector periodically fetches container stats and stores them in ring buffers.
type Collector struct {
	docker   *docker.Service
	interval time.Duration
	capacity int

	mu         sync.RWMutex
	containers map[string]*RingBuffer[docker.StatsSnapshot]
	host       *RingBuffer[host.HostSnapshot]

	serviceNames []string
	services     map[string]*RingBuffer[ServiceSnapshot]
	prevCPU      map[string]uint64
	prevTime     map[string]time.Time
}

// NewCollector creates a new stats collector.
func NewCollector(dockerSvc *docker.Service, interval time.Duration, capacity int, serviceNames []string) *Collector {
	if capacity <= 0 {
		capacity = 60
	}
	if interval <= 0 {
		interval = 15 * time.Second
	}
	return &Collector{
		docker:       dockerSvc,
		interval:     interval,
		capacity:     capacity,
		containers:   make(map[string]*RingBuffer[docker.StatsSnapshot]),
		host:         NewRingBuffer[host.HostSnapshot](capacity),
		serviceNames: serviceNames,
		services:     make(map[string]*RingBuffer[ServiceSnapshot]),
		prevCPU:      make(map[string]uint64),
		prevTime:     make(map[string]time.Time),
	}
}

// Start begins the background collection loop. Blocks until ctx is cancelled.
func (c *Collector) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Collect immediately on start
	c.collect(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.collect(ctx)
		}
	}
}

func (c *Collector) collect(ctx context.Context) {
	// Collect host metrics
	if snap, err := host.Snapshot(); err != nil {
		log.Printf("collector: host snapshot error: %v", err)
	} else if snap != nil {
		c.mu.Lock()
		c.host.Push(*snap)
		c.mu.Unlock()
	}

	containers, err := c.docker.List(ctx)
	if err != nil {
		log.Printf("collector: list error: %v", err)
		return
	}

	// Find running containers
	running := make(map[string]bool)
	var wg sync.WaitGroup
	var mu sync.Mutex
	type result struct {
		id    string
		stats *docker.StatsSnapshot
	}
	results := make([]result, 0)

	for _, cont := range containers {
		if cont.State != "running" {
			continue
		}
		running[cont.ID] = true
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			stats, err := c.docker.Stats(ctx, id)
			if err != nil {
				return
			}
			mu.Lock()
			results = append(results, result{id: id, stats: stats})
			mu.Unlock()
		}(cont.ID)
	}
	wg.Wait()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Push new stats
	for _, r := range results {
		buf, ok := c.containers[r.id]
		if !ok {
			buf = NewRingBuffer[docker.StatsSnapshot](c.capacity)
			c.containers[r.id] = buf
		}
		buf.Push(*r.stats)
	}

	// Clean up buffers for removed containers
	for id := range c.containers {
		if !running[id] {
			delete(c.containers, id)
		}
	}

	// Collect systemd service metrics
	now := time.Now()
	for _, name := range c.serviceNames {
		info, err := systemd.Show(ctx, name)
		if err != nil {
			log.Printf("collector: service %s: %v", name, err)
			continue
		}
		if info == nil {
			continue
		}

		cpuPct := 0.0
		if prev, ok := c.prevCPU[name]; ok && info.CPUNanos > prev {
			dt := now.Sub(c.prevTime[name]).Seconds()
			if dt > 0 {
				cpuPct = float64(info.CPUNanos-prev) / (dt * 1e9) * 100
			}
		}
		c.prevCPU[name] = info.CPUNanos
		c.prevTime[name] = now

		snap := ServiceSnapshot{
			Name:        name,
			ActiveState: info.ActiveState,
			SubState:    info.SubState,
			MemCurrent:  info.MemCurrent,
			CPUPercent:  cpuPct,
			StartedAt:   info.StartedAt,
			NRestarts:   info.NRestarts,
			Description: info.Description,
			UnitEnabled: info.UnitEnabled,
		}

		buf, ok := c.services[name]
		if !ok {
			buf = NewRingBuffer[ServiceSnapshot](c.capacity)
			c.services[name] = buf
		}
		buf.Push(snap)
	}
}

// ContainerHistory returns the stats history for a container.
func (c *Collector) ContainerHistory(id string) []docker.StatsSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	buf, ok := c.containers[id]
	if !ok {
		return nil
	}
	return buf.Slice()
}

// ContainerLatest returns the most recent stats for a container.
func (c *Collector) ContainerLatest(id string) *docker.StatsSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	buf, ok := c.containers[id]
	if !ok {
		return nil
	}
	s, ok := buf.Last()
	if !ok {
		return nil
	}
	return &s
}

// HostLatest returns the most recent host snapshot, or nil if none.
func (c *Collector) HostLatest() *host.HostSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.host.Last()
	if !ok {
		return nil
	}
	return &s
}

// HostHistory returns the host snapshot history.
func (c *Collector) HostHistory() []host.HostSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.host.Slice()
}

// AllServiceLatest returns the latest snapshot for all tracked services.
func (c *Collector) AllServiceLatest() map[string]*ServiceSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]*ServiceSnapshot, len(c.services))
	for name, buf := range c.services {
		if s, ok := buf.Last(); ok {
			out[name] = &s
		}
	}
	return out
}

// ServiceHistory returns the snapshot history for a service.
func (c *Collector) ServiceHistory(name string) []ServiceSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	buf, ok := c.services[name]
	if !ok {
		return nil
	}
	return buf.Slice()
}

// ServiceNames returns the configured service names.
func (c *Collector) ServiceNames() []string {
	return c.serviceNames
}

// AllLatest returns the latest stats for all tracked containers.
func (c *Collector) AllLatest() map[string]*docker.StatsSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]*docker.StatsSnapshot, len(c.containers))
	for id, buf := range c.containers {
		if s, ok := buf.Last(); ok {
			out[id] = &s
		}
	}
	return out
}
