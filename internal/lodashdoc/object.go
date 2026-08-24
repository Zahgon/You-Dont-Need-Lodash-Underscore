package lodashdoc

import (
	"fmt"
	"sort"
)

// LodashInvert mirrors _.invert(object). JavaScript property keys are strings,
// so the inverted keys are the values run through the same String() coercion
// the original relies on ({ a: 1 } inverts to { "1": "a" }).
func LodashInvert(object map[string]any) map[string]string {
	result := make(map[string]string, len(object))
	for key, value := range object {
		result[fmt.Sprint(value)] = key
	}
	return result
}

// NativeInvert mirrors the README's for-in + hasOwnProperty invert snippet.
func NativeInvert(object map[string]any) map[string]string {
	result := map[string]string{}
	for _, key := range sortedAnyKeys(object) {
		result[fmt.Sprint(object[key])] = key
	}
	return result
}

// LodashMapKeys mirrors _.mapKeys(object, iteratee).
func LodashMapKeys(object map[string]int, iteratee func(value int, key string) string) map[string]int {
	result := make(map[string]int, len(object))
	for key, value := range object {
		result[iteratee(value, key)] = value
	}
	return result
}

// NativeMapKeys mirrors the README's for-in + hasOwnProperty mapKeys snippet.
func NativeMapKeys(object map[string]int, iteratee func(value int, key string) string) map[string]int {
	result := map[string]int{}
	for _, key := range sortedIntKeys(object) {
		result[iteratee(object[key], key)] = object[key]
	}
	return result
}

// LodashPick mirrors _.pick(object, keys).
func LodashPick(object map[string]any, keys []string) map[string]any {
	result := make(map[string]any, len(keys))
	for _, key := range keys {
		value, present := object[key]
		if !present {
			continue
		}
		result[key] = value
	}
	return result
}

// NativePick mirrors the README's reduce-based pick snippet.
func NativePick(object map[string]any, keys []string) map[string]any {
	result := map[string]any{}
	for _, key := range keys {
		if object == nil {
			continue
		}
		if _, present := object[key]; present {
			result[key] = object[key]
		}
	}
	return result
}

// LodashDefaults mirrors _.defaults(object, source): keys missing from object
// are taken from source, and object itself is mutated and returned.
func LodashDefaults(object map[string]any, source map[string]any) map[string]any {
	for _, key := range sortedAnyKeys(source) {
		if _, present := object[key]; present {
			continue
		}
		object[key] = source[key]
	}
	return object
}

// NativeAssign mirrors Object.assign(target, ...sources).
func NativeAssign(target map[string]any, sources ...map[string]any) map[string]any {
	for _, source := range sources {
		for _, key := range sortedAnyKeys(source) {
			target[key] = source[key]
		}
	}
	return target
}

// LodashForEach mirrors _.forEach over an array.
func LodashForEach[T any](array []T, iteratee func(element T)) {
	for _, element := range array {
		iteratee(element)
	}
}

// NativeForEach mirrors Array.prototype.forEach.
func NativeForEach[T any](array []T, iteratee func(element T)) {
	for index := 0; index < len(array); index++ {
		iteratee(array[index])
	}
}

// LodashForEachObject mirrors _.forEach over an object. JavaScript visits
// string keys in insertion order; Go maps have no order at all, so this helper
// and NativeObjectEntriesForEach both visit the keys in sorted order. The
// property under test - that the two traversals yield the same values in the
// same order - is preserved exactly.
func LodashForEachObject(object map[string]int, iteratee func(value int)) {
	for _, key := range sortedIntKeys(object) {
		iteratee(object[key])
	}
}

// NativeObjectEntriesForEach mirrors Object.entries(object).forEach.
func NativeObjectEntriesForEach(object map[string]int, iteratee func(key string, value int)) {
	for _, key := range sortedIntKeys(object) {
		iteratee(key, object[key])
	}
}

func sortedAnyKeys(object map[string]any) []string {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedIntKeys(object map[string]int) []string {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
