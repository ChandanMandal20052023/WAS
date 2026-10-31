package utils

import (
	"errors"
	"regexp"
	"strings"
)

var nonDigitRegex = regexp.MustCompile(`[^\d+]`)

// NormalizePhoneNumber normalizes an international phone number to E.164 format.
// Examples:
// "(555) 123-4567" -> "+5551234567" (if country code included)
// "+1 (555) 019-2834" -> "+15550192834"
// "919876543210" -> "+919876543210"
func NormalizePhoneNumber(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("phone number cannot be empty")
	}

	// Remove all formatting chars except digits and initial plus
	cleaned := nonDigitRegex.ReplaceAllString(trimmed, "")

	if !strings.HasPrefix(cleaned, "+") {
		// If no plus prefix, prepend it
		cleaned = "+" + cleaned
	}

	// Remove any duplicate '+' in between
	plusCount := strings.Count(cleaned, "+")
	if plusCount > 1 {
		cleaned = "+" + strings.ReplaceAll(cleaned, "+", "")
	}

	// An E.164 phone number consists of '+' followed by 7 to 15 digits
	digitsOnly := strings.TrimPrefix(cleaned, "+")
	if len(digitsOnly) < 7 || len(digitsOnly) > 15 {
		return "", errors.New("invalid phone number length: must be between 7 and 15 digits")
	}

	return cleaned, nil
}
