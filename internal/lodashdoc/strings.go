package lodashdoc

import "strings"

// LodashSplit mirrors _.split(string, separator, limit).
func LodashSplit(text string, separator string, limit ...int) []string {
	maximum := -1
	if len(limit) > 0 {
		maximum = limit[0]
	}
	result := []string{}
	remaining := text
	for separator != "" {
		if maximum >= 0 && len(result) >= maximum {
			return result
		}
		index := strings.Index(remaining, separator)
		if index < 0 {
			break
		}
		result = append(result, remaining[:index])
		remaining = remaining[index+len(separator):]
	}
	if maximum >= 0 && len(result) >= maximum {
		return result
	}
	return append(result, remaining)
}

// NativeSplit mirrors String.prototype.split(separator, limit), where limit
// truncates the fully split result rather than stopping the scan early.
func NativeSplit(text string, separator string, limit ...int) []string {
	parts := strings.Split(text, separator)
	if len(limit) > 0 && limit[0] >= 0 && limit[0] < len(parts) {
		parts = parts[:limit[0]]
	}
	return parts
}

// LodashPadStart mirrors _.padStart(string, length, chars).
func LodashPadStart(text string, length int, pad string) string {
	missing := length - len(text)
	if missing <= 0 || pad == "" {
		return text
	}
	var builder strings.Builder
	for builder.Len() < missing {
		builder.WriteString(pad)
	}
	return builder.String()[:missing] + text
}

// NativePadStart mirrors String.prototype.padStart(length, pad).
func NativePadStart(text string, length int, pad string) string {
	if len(text) >= length || pad == "" {
		return text
	}
	missing := length - len(text)
	filler := strings.Repeat(pad, missing/len(pad)+1)
	return filler[:missing] + text
}

// LodashPadEnd mirrors _.padEnd(string, length, chars).
func LodashPadEnd(text string, length int, pad string) string {
	missing := length - len(text)
	if missing <= 0 || pad == "" {
		return text
	}
	var builder strings.Builder
	for builder.Len() < missing {
		builder.WriteString(pad)
	}
	return text + builder.String()[:missing]
}

// NativePadEnd mirrors String.prototype.padEnd(length, pad).
func NativePadEnd(text string, length int, pad string) string {
	if len(text) >= length || pad == "" {
		return text
	}
	missing := length - len(text)
	filler := strings.Repeat(pad, missing/len(pad)+1)
	return text + filler[:missing]
}

// LodashUpperFirst mirrors _.upperFirst.
func LodashUpperFirst(text string) string {
	if text == "" {
		return ""
	}
	runes := []rune(text)
	return strings.ToUpper(string(runes[0])) + string(runes[1:])
}

// NativeUpperFirst mirrors the README's upperFirst snippet. The JavaScript
// original guards with `string ? ... : ”`; for a string argument the guard is
// only false for the empty string, whose two branches both produce "", so the
// coercion never changes the result.
func NativeUpperFirst(text string) string {
	if text == "" {
		return ""
	}
	first := []rune(text)[:1]
	return strings.ToUpper(string(first)) + text[len(string(first)):]
}

// LodashLowerFirst mirrors _.lowerFirst.
func LodashLowerFirst(text string) string {
	if text == "" {
		return ""
	}
	runes := []rune(text)
	return strings.ToLower(string(runes[0])) + string(runes[1:])
}

// NativeLowerFirst mirrors the README's lowerFirst snippet; see
// NativeUpperFirst for the treatment of the truthiness guard.
func NativeLowerFirst(text string) string {
	if text == "" {
		return ""
	}
	first := []rune(text)[:1]
	return strings.ToLower(string(first)) + text[len(string(first)):]
}

// LodashCapitalize mirrors _.capitalize.
func LodashCapitalize(text string) string {
	if text == "" {
		return ""
	}
	runes := []rune(text)
	return strings.ToUpper(string(runes[0])) + strings.ToLower(string(runes[1:]))
}

// NativeCapitalize mirrors the README's capitalize snippet; see
// NativeUpperFirst for the treatment of the truthiness guard.
func NativeCapitalize(text string) string {
	if text == "" {
		return ""
	}
	first := []rune(text)[:1]
	rest := text[len(string(first)):]
	return strings.ToUpper(string(first)) + strings.ToLower(rest)
}

// LodashStartsWith mirrors _.startsWith(string, target, position).
func LodashStartsWith(text string, target string, position ...int) bool {
	start := 0
	if len(position) > 0 {
		start = position[0]
	}
	if start < 0 {
		start = 0
	}
	if start > len(text) {
		start = len(text)
	}
	end := start + len(target)
	if end > len(text) {
		return false
	}
	return text[start:end] == target
}

// NativeStartsWith mirrors String.prototype.startsWith(target, position).
func NativeStartsWith(text string, target string, position ...int) bool {
	start := 0
	if len(position) > 0 {
		start = position[0]
	}
	if start < 0 {
		start = 0
	}
	if start > len(text) {
		start = len(text)
	}
	return strings.HasPrefix(text[start:], target)
}

// LodashEndsWith mirrors _.endsWith(string, target, position).
func LodashEndsWith(text string, target string, position ...int) bool {
	end := len(text)
	if len(position) > 0 {
		end = position[0]
	}
	if end > len(text) {
		end = len(text)
	}
	start := end - len(target)
	if start < 0 || end < 0 {
		return false
	}
	return text[start:end] == target
}

// NativeEndsWith mirrors String.prototype.endsWith(target, endPosition).
func NativeEndsWith(text string, target string, position ...int) bool {
	end := len(text)
	if len(position) > 0 {
		end = position[0]
	}
	if end < 0 {
		end = 0
	}
	if end > len(text) {
		end = len(text)
	}
	return strings.HasSuffix(text[:end], target)
}
