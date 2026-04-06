package docker

import (
	"strings"
	"testing"
)

func TestContainerNameStripping(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/mycontainer", "mycontainer"},
		{"mycontainer", "mycontainer"},
		{"/my/container", "my/container"},
		{"", ""},
	}
	for _, tt := range tests {
		got := strings.TrimPrefix(tt.input, "/")
		if got != tt.want {
			t.Errorf("TrimPrefix(%q, '/') = %q, want %q", tt.input, got, tt.want)
		}
	}
}
