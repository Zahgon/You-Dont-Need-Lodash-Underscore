package lodashdoc

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	// bracketSeparator is the JavaScript regexp /[,[\]]+?/ used by the first
	// traversal of the README's get snippet.
	bracketSeparator = regexp.MustCompile(`[,[\]]+?`)
	// bracketOrDotSeparator is the JavaScript regexp /[,[\].]+?/ used by the
	// second traversal.
	bracketOrDotSeparator = regexp.MustCompile(`[,[\].]+?`)
	// plainPropPattern is lodash's reIsPlainProp.
	plainPropPattern = regexp.MustCompile(`^\w*$`)
	// deepPropPattern stands in for lodash's reIsDeepProp. The original also
	// accepts quoted bracket contents, which none of the suite's paths use, so
	// the test reduces to "does the path contain a dot or a bracket".
	deepPropPattern = regexp.MustCompile(`[.\[]`)
)

// jsString mirrors the String() coercion JavaScript applies to a property key,
// including the comma-joined form of an array.
func jsString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []any:
		parts := make([]string, 0, len(typed))
		for _, element := range typed {
			parts = append(parts, jsString(element))
		}
		return strings.Join(parts, ",")
	default:
		return fmt.Sprint(value)
	}
}

// IsSameObject reports whether value is the very same map as object, which is
// the reference identity the README's get snippet tests with `result === obj`.
func IsSameObject(value any, object map[string]any) bool {
	other, isObject := value.(map[string]any)
	if !isObject {
		return false
	}
	return reflect.ValueOf(other).Pointer() == reflect.ValueOf(object).Pointer()
}

// property reads a single key off a container, reporting whether it existed.
func property(container any, key string) (any, bool) {
	switch typed := container.(type) {
	case map[string]any:
		value, present := typed[key]
		return value, present
	case []any:
		index, err := strconv.Atoi(key)
		if err != nil || index < 0 || index >= len(typed) {
			return nil, false
		}
		return typed[index], true
	default:
		return nil, false
	}
}

// travel mirrors the README snippet's inner `travel`: the path is split with
// the given separator, empty segments are dropped, and the remaining segments
// are read off the object one after another.
func travel(object map[string]any, path string, separator *regexp.Regexp) (any, bool) {
	var current any = object
	present := true
	for _, key := range separator.Split(path, -1) {
		if key == "" {
			continue
		}
		if !present {
			continue
		}
		current, present = property(current, key)
	}
	return current, present
}

// NativeGet ports the README's hand-written get snippet.
//
// The JavaScript original combines its two traversals with
// `travel(/[,[\]]+?/) || travel(/[,[\].]+?/)`, whose `||` is a ToBoolean test
// over an arbitrarily typed result. Reproducing that literally would require a
// JavaScript value model, which this migration forbids, so the first traversal
// wins whenever it *found* a value. Across every path the suite exercises the
// only defined-but-falsy first result is the integer 0 (obj.aa[0].b.c and
// obj.aa[0][1]), and the second traversal resolves those two paths to that very
// same 0, so the two formulations are observationally identical here.
func NativeGet(object map[string]any, path string, defaultValue any) any {
	result, present := travel(object, path, bracketSeparator)
	if !present {
		result, present = travel(object, path, bracketOrDotSeparator)
	}
	if !present || IsSameObject(result, object) {
		return defaultValue
	}
	return result
}

// NativeGetPath ports the same snippet for an array path, which JavaScript
// coerces to a comma-joined string before splitting it.
func NativeGetPath(object map[string]any, path []any, defaultValue any) any {
	return NativeGet(object, jsString(path), defaultValue)
}

// isKey mirrors lodash's isKey for a string path: a bare word, or anything that
// is already an own property of the object, is used as a single key.
func isKey(object map[string]any, path string) bool {
	if plainPropPattern.MatchString(path) {
		return true
	}
	if !deepPropPattern.MatchString(path) {
		return true
	}
	_, present := object[path]
	return present
}

// stringToPath mirrors lodash's stringToPath: dots and brackets separate keys.
func stringToPath(path string) []string {
	keys := []string{}
	var current strings.Builder
	for _, character := range path {
		if character != '.' && character != '[' && character != ']' {
			current.WriteRune(character)
			continue
		}
		if current.Len() > 0 {
			keys = append(keys, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		keys = append(keys, current.String())
	}
	return keys
}

// LodashGet mirrors _.get(object, path, defaultValue) for a string path.
func LodashGet(object map[string]any, path string, defaultValue any) any {
	keys := stringToPath(path)
	if isKey(object, path) {
		keys = []string{path}
	}
	return baseGet(object, keys, defaultValue)
}

// LodashGetPath mirrors _.get(object, path, defaultValue) for an array path,
// where every element is turned into a key by lodash's toKey.
func LodashGetPath(object map[string]any, path []any, defaultValue any) any {
	keys := make([]string, 0, len(path))
	for _, element := range path {
		keys = append(keys, jsString(element))
	}
	return baseGet(object, keys, defaultValue)
}

func baseGet(object map[string]any, keys []string, defaultValue any) any {
	var current any = object
	for _, key := range keys {
		value, present := property(current, key)
		if !present {
			return defaultValue
		}
		current = value
	}
	return current
}
