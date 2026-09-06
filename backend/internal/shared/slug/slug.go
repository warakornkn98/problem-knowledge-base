// Package slug produces URL/identifier-friendly slugs. It keeps Unicode letters
// (so Thai category names still yield a usable slug) and collapses everything
// else to single hyphens.
package slug

import (
	"strings"
	"unicode"
)

// Make converts s into a lowercase, hyphen-delimited slug.
func Make(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))

	var b strings.Builder
	lastHyphen := false
	for _, r := range s {
		switch {
		// Keep letters, digits and combining marks (Thai/Indic vowel signs are
		// category M, not L, but are part of the word).
		case unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsMark(r):
			b.WriteRune(r)
			lastHyphen = false
		default:
			if !lastHyphen && b.Len() > 0 {
				b.WriteRune('-')
				lastHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
