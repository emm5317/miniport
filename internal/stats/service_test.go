package stats

import "testing"

func TestServiceSnapshotRingBuffer(t *testing.T) {
	buf := NewRingBuffer[ServiceSnapshot](3)

	buf.Push(ServiceSnapshot{Name: "test", CPUPercent: 1.0})
	buf.Push(ServiceSnapshot{Name: "test", CPUPercent: 2.0})
	buf.Push(ServiceSnapshot{Name: "test", CPUPercent: 3.0})

	s, ok := buf.Last()
	if !ok || s.CPUPercent != 3.0 {
		t.Errorf("Last() = %v, %v; want 3.0, true", s.CPUPercent, ok)
	}

	buf.Push(ServiceSnapshot{Name: "test", CPUPercent: 4.0})
	all := buf.Slice()
	if len(all) != 3 {
		t.Fatalf("Slice() len = %d, want 3", len(all))
	}
	if all[0].CPUPercent != 2.0 || all[2].CPUPercent != 4.0 {
		t.Errorf("Slice() = [%.1f, %.1f, %.1f], want [2.0, 3.0, 4.0]",
			all[0].CPUPercent, all[1].CPUPercent, all[2].CPUPercent)
	}
}
