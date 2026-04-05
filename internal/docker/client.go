package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/client"
)

type Service struct {
	cli *client.Client
}

func NewService() (*Service, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := cli.Ping(ctx); err != nil {
		return nil, fmt.Errorf("docker ping: %w", err)
	}
	return &Service{cli: cli}, nil
}
