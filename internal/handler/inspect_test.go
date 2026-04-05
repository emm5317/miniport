package handler

import "testing"

func TestMaskValue(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "***"},
		{"ab", "***"},
		{"abc", "***"},
		{"abcd", "abc*"},
		{"mysecretpassword", "mys************"},
		{"short1", "sho***"},
	}
	for _, tt := range tests {
		got := MaskValue(tt.input)
		if got != tt.want {
			t.Errorf("MaskValue(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
