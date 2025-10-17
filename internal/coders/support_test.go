package coders

import "testing"

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic cases
		{"simple semantic version", "v1.0.2", "v1_0_2"},
		{"numeric only", "1.2.3", "v1_2_3"},
		{"beta with dash", "v1-beta.2", "v1_beta_2"},
		{"release path", "release/1.0.0", "release_1_0_0"},
		{"mixed symbols", "RC_2024-05", "rc_2024_05"},

		// Edge cases
		{"empty string", "", ""},
		{"starts with non-alphanumeric", "-v1.0.0", "v1_0_0"},
		{"multiple separators", "v1..0--2", "v1_0_2"},
		{"trailing symbols", "v1.0.0-", "v1_0_0"},
		{"leading digit", "9alpha", "v9alpha"},
		{"only digits", "123", "v123"},
		{"uppercase letters", "V2.BETA", "v2_beta"},
		{"multiple slashes", "release/v2/1.0", "release_v2_1_0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeVersion(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeVersion(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
