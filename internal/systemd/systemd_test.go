package systemd

import "testing"

func TestValidName(t *testing.T) {
	valid := []string{"voicetask", "caddy", "postgresql", "my-service", "nginx.service"}
	for _, n := range valid {
		if err := ValidName(n); err != nil {
			t.Errorf("ValidName(%q) = %v, want nil", n, err)
		}
	}

	invalid := []string{"", "../etc", "foo bar", "a;b", "a|b", "a`b", "a$b", "a&b", "a\\b"}
	for _, n := range invalid {
		if err := ValidName(n); err == nil {
			t.Errorf("ValidName(%q) = nil, want error", n)
		}
	}
}

func TestParseShow(t *testing.T) {
	output := `ActiveState=active
SubState=running
MainPID=1170
MemoryCurrent=18874368
CPUUsageNSec=13385000000
ExecMainStartTimestamp=Sat 2026-04-05 01:20:37 CDT
NRestarts=0
UnitFileState=enabled
Description=VoiceTask - Voice-to-Task Capture
`
	info := ParseShow("voicetask", output)

	if info.Name != "voicetask" {
		t.Errorf("Name = %q, want voicetask", info.Name)
	}
	if info.ActiveState != "active" {
		t.Errorf("ActiveState = %q, want active", info.ActiveState)
	}
	if info.SubState != "running" {
		t.Errorf("SubState = %q, want running", info.SubState)
	}
	if info.MainPID != 1170 {
		t.Errorf("MainPID = %d, want 1170", info.MainPID)
	}
	if info.MemCurrent != 18874368 {
		t.Errorf("MemCurrent = %d, want 18874368", info.MemCurrent)
	}
	if info.CPUNanos != 13385000000 {
		t.Errorf("CPUNanos = %d, want 13385000000", info.CPUNanos)
	}
	if info.NRestarts != 0 {
		t.Errorf("NRestarts = %d, want 0", info.NRestarts)
	}
	if info.UnitEnabled != "enabled" {
		t.Errorf("UnitEnabled = %q, want enabled", info.UnitEnabled)
	}
	if info.Description != "VoiceTask - Voice-to-Task Capture" {
		t.Errorf("Description = %q", info.Description)
	}
}

func TestParseShowNotSet(t *testing.T) {
	output := `ActiveState=inactive
SubState=dead
MainPID=0
MemoryCurrent=[not set]
CPUUsageNSec=[not set]
ExecMainStartTimestamp=
NRestarts=0
UnitFileState=disabled
Description=Some Service
`
	info := ParseShow("test", output)

	if info.MemCurrent != 0 {
		t.Errorf("MemCurrent = %d, want 0 for [not set]", info.MemCurrent)
	}
	if info.CPUNanos != 0 {
		t.Errorf("CPUNanos = %d, want 0 for [not set]", info.CPUNanos)
	}
	if info.ActiveState != "inactive" {
		t.Errorf("ActiveState = %q, want inactive", info.ActiveState)
	}
}
