// Package slug validates task slugs and writer IDs used in paths and shell hooks.
package slug

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// validSlug is a single path segment: letters, digits, dot, underscore, hyphen.
var validSlug = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

// validWriterID restricts resume-hook substitution to shell-safe characters.
var validWriterID = regexp.MustCompile(`^[A-Za-z0-9._:@/+-]{1,128}$`)

// Validate rejects empty, traversal, and separator-bearing slugs.
func Validate(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("slug is required")
	}
	if s != strings.TrimSpace(s) {
		return fmt.Errorf("invalid slug %q: leading/trailing space", s)
	}
	if strings.Contains(s, "..") {
		return fmt.Errorf("invalid slug %q: path traversal", s)
	}
	for _, r := range s {
		if r == '/' || r == '\\' || unicode.IsSpace(r) {
			return fmt.Errorf("invalid slug %q: path separators/spaces not allowed", s)
		}
	}
	if !validSlug.MatchString(s) {
		return fmt.Errorf("invalid slug %q: use letters, digits, '.', '_', '-'", s)
	}
	return nil
}

// ValidateWriterID rejects empty or shell-metacharacter writer IDs when set.
func ValidateWriterID(id string) error {
	if id == "" {
		return fmt.Errorf("writer id is empty")
	}
	if !validWriterID.MatchString(id) {
		return fmt.Errorf("invalid writer id %q: unsafe characters", id)
	}
	return nil
}
