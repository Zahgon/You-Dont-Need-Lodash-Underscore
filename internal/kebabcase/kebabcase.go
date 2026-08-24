// Package kebabcase is a translation of the npm package kebab-case@1.0.0,
// which the original JavaScript repository consumed as a runtime dependency.
//
// The original implementation is a single regular expression replacement:
//
//	var KEBAB_REGEX = /[A-Z\u00C0-\u00D6\u00D8-\u00DE]/g;
//	function kebabCase(str) {
//	  return str.replace(KEBAB_REGEX, function (match) {
//	    return '-' + match.toLowerCase();
//	  });
//	}
//
// It is deliberately naive: it has no acronym awareness, so "isNaN" becomes
// "is-na-n" rather than "is-nan". That behaviour is load bearing, because
// lib/rules/rules.json only overrides the generated rule name for a handful of
// entries and every other rule name is derived from this exact algorithm.
package kebabcase

import "strings"

// isKebabTrigger reports whether r is matched by KEBAB_REGEX, that is the
// uppercase ASCII letters plus the two uppercase ranges of Latin-1 Supplement.
func isKebabTrigger(r rune) bool {
	switch {
	case r >= 'A' && r <= 'Z':
		return true
	case r >= 0x00C0 && r <= 0x00D6:
		return true
	case r >= 0x00D8 && r <= 0x00DE:
		return true
	default:
		return false
	}
}

// KebabCase converts str the same way kebab-case@1.0.0 does.
func KebabCase(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, r := range str {
		if isKebabTrigger(r) {
			b.WriteByte('-')
			b.WriteRune(toLowerJS(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// toLowerJS lowercases a rune the way String.prototype.toLowerCase does for the
// characters KEBAB_REGEX can match. Every such character has a single-rune
// lowercase mapping at a fixed offset of 0x20.
func toLowerJS(r rune) rune {
	return r + 0x20
}
