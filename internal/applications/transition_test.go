package applications

import "testing"

func TestCanTransition(t *testing.T) {
	tests := []struct {
		from string
		to   string
		ok   bool
	}{
		{"applied", "screening", true},
		{"screening", "interview", true},
		{"offer", "hired", true},

		{"applied", "offer", false},
		{"hired", "screening", false},
		{"rejected", "interview", false},
	}

	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			got := canTransition(tt.from, tt.to)
			if got != tt.ok {
				t.Fatalf("expected %v, got %v", tt.ok, got)
			}
		})
	}
}
