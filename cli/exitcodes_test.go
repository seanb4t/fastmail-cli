package cli

import "testing"

// TestExitCodesCompile ensures the exit code constants are accessible.
// This is a constants-only file; the test verifies compilation and expected values.
func TestExitCodesCompile(t *testing.T) {
	// Verify constants have expected values matching design doc
	tests := []struct {
		name     string
		got      int
		expected int
	}{
		{"ExitSuccess", ExitSuccess, 0},
		{"ExitNoToken", ExitNoToken, 1},
		{"ExitInvalidToken", ExitInvalidToken, 2},
		{"ExitNetworkError", ExitNetworkError, 3},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s = %d; want %d", tt.name, tt.got, tt.expected)
		}
	}
}
