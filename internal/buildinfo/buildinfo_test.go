package buildinfo

import "testing"

func TestDetailsString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		details Details
		want    string
	}{
		{
			name:    "release",
			details: Details{Version: "0.1.0"},
			want:    "maisternia 0.1.0",
		},
		{
			name:    "development",
			details: Details{},
			want:    "maisternia dev",
		},
		{
			name:    "dirty",
			details: Details{Version: "dev", Dirty: true},
			want:    "maisternia dev (dirty)",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.details.String(); got != tt.want {
				t.Fatalf("Details.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
