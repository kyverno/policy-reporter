package email

import (
	"fmt"
	"strings"
	"unicode"
)

// ValidateAttachmentFormat rejects unknown formats rather than sending a report
// without the details the recipient requested.
func ValidateAttachmentFormat(format string) error {
	if format != "" && format != "csv" {
		return fmt.Errorf("unsupported attachment format %q (expected empty or csv)", format)
	}
	return nil
}

// CSVText neutralizes spreadsheet formulas without changing the source data.
// An apostrophe is prepended before dangerous prefixes, including formulas
// hidden behind whitespace, control characters, or Unicode format characters.
// Numeric count fields must not pass through this function.
func CSVText(value string) string {
	trimmed := strings.TrimLeftFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r) || unicode.Is(unicode.Cf, r)
	})
	if strings.HasPrefix(value, "\t") || strings.HasPrefix(value, "\r") || strings.HasPrefix(value, "\n") ||
		(trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0]))) {
		return "'" + value
	}
	return value
}
