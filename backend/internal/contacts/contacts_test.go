package contacts

import (
	"testing"
	"was/backend/internal/utils"
)

func TestPhoneNormalization(t *testing.T) {
	tests := []struct {
		input       string
		expected    string
		expectError bool
	}{
		{input: "+1 (555) 123-4567", expected: "+15551234567", expectError: false},
		{input: "5551234567", expected: "+5551234567", expectError: false},
		{input: "+91 98765-43210", expected: "+919876543210", expectError: false},
		{input: "++44 20 7946 0958", expected: "+442079460958", expectError: false},
		{input: "", expected: "", expectError: true},
		{input: "123", expected: "", expectError: true}, // too short
		{input: "+123456789012345678", expected: "", expectError: true}, // too long
	}

	for _, tc := range tests {
		actual, err := utils.NormalizePhoneNumber(tc.input)
		if tc.expectError {
			if err == nil {
				t.Errorf("NormalizePhoneNumber(%q) expected error, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("NormalizePhoneNumber(%q) unexpected error: %v", tc.input, err)
			}
			if actual != tc.expected {
				t.Errorf("NormalizePhoneNumber(%q) = %q, expected %q", tc.input, actual, tc.expected)
			}
		}
	}
}
