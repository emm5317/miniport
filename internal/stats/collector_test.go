package stats

import "testing"

func TestRingBuffer_PushAndSlice(t *testing.T) {
	tests := []struct {
		name     string
		cap      int
		pushVals []int
		want     []int
	}{
		{"empty", 3, nil, []int{}},
		{"partial", 3, []int{1, 2}, []int{1, 2}},
		{"full", 3, []int{1, 2, 3}, []int{1, 2, 3}},
		{"overflow", 3, []int{1, 2, 3, 4}, []int{2, 3, 4}},
		{"double overflow", 3, []int{1, 2, 3, 4, 5, 6, 7}, []int{5, 6, 7}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewRingBuffer[int](tt.cap)
			for _, v := range tt.pushVals {
				buf.Push(v)
			}
			got := buf.Slice()
			if len(got) != len(tt.want) {
				t.Fatalf("Slice() len = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Slice()[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRingBuffer_Last(t *testing.T) {
	buf := NewRingBuffer[int](3)
	if _, ok := buf.Last(); ok {
		t.Error("Last() on empty buffer should return false")
	}
	buf.Push(10)
	v, ok := buf.Last()
	if !ok || v != 10 {
		t.Errorf("Last() = %d, %v; want 10, true", v, ok)
	}
	buf.Push(20)
	buf.Push(30)
	buf.Push(40) // overflows
	v, ok = buf.Last()
	if !ok || v != 40 {
		t.Errorf("Last() after overflow = %d, %v; want 40, true", v, ok)
	}
}

func TestRingBuffer_Len(t *testing.T) {
	buf := NewRingBuffer[int](3)
	if buf.Len() != 0 {
		t.Errorf("Len() = %d, want 0", buf.Len())
	}
	buf.Push(1)
	buf.Push(2)
	if buf.Len() != 2 {
		t.Errorf("Len() = %d, want 2", buf.Len())
	}
	buf.Push(3)
	buf.Push(4)
	if buf.Len() != 3 {
		t.Errorf("Len() = %d, want 3", buf.Len())
	}
}
