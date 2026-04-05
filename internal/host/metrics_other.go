//go:build !linux

package host

// Snapshot returns nil on non-Linux platforms (no /proc filesystem).
func Snapshot() (*HostSnapshot, error) {
	return nil, nil
}
