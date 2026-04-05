//go:build !linux

package systemd

import "context"

func Show(_ context.Context, _ string) (*ServiceInfo, error)       { return nil, nil }
func Logs(_ context.Context, _ string, _ int) (string, error)      { return "", nil }
func LogsSince(_ context.Context, _, _ string) (string, error)     { return "", nil }
func Start(_ context.Context, _ string) error                      { return nil }
func Stop(_ context.Context, _ string) error                       { return nil }
func Restart(_ context.Context, _ string) error                    { return nil }
