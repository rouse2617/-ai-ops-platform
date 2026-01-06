package permission

import "testing"

func TestMatch(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		want    bool
	}{
		{"exact match", "check_disk", "check_disk", true},
		{"star prefix", "check_*", "check_disk", true},
		{"star suffix", "*_disk", "check_disk", true},
		{"star middle", "check_*_usage", "check_disk_usage", true},
		{"multiple stars", "*_*", "check_disk", true},
		{"double star", "**", "anything", true},
		{"no match", "check_*", "list_files", false},
		{"question mark", "check_?", "check_a", true},
		{"question mark no match", "check_?", "check_ab", false},
		{"empty pattern", "", "", true},
		{"empty input", "test", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Match(tt.pattern, tt.input)
			if got != tt.want {
				t.Errorf("Match(%q, %q) = %v; want %v", tt.pattern, tt.input, got, tt.want)
			}
		})
	}
}
