// Package lodashdoc holds the Go translation of the code snippets exercised by
// tests/unit/all.js in the original JavaScript repository.
//
// The original suite pairs a lodash function with the hand-written native
// JavaScript snippet documented in the project's README and asserts that the
// two agree. Both sides of every pair are reproduced here as independent Go
// functions - Lodash* for the library behaviour, Native* for the README
// snippet - so the tests can keep asserting the very same equivalence instead
// of comparing against a value copied out of a JavaScript run.
package lodashdoc

// LodashConcat mirrors _.concat(array, ...values): the receiver is copied and
// every argument is appended, with array arguments spread one level deep.
func LodashConcat(array []any, values ...any) []any {
	result := make([]any, 0, len(array)+len(values))
	result = append(result, array...)
	for _, value := range values {
		nested, isArray := value.([]any)
		if !isArray {
			result = append(result, value)
			continue
		}
		result = append(result, nested...)
	}
	return result
}

// NativeConcat mirrors Array.prototype.concat(...values).
func NativeConcat(array []any, values ...any) []any {
	result := append([]any{}, array...)
	for _, value := range values {
		switch typed := value.(type) {
		case []any:
			for _, element := range typed {
				result = append(result, element)
			}
		default:
			result = append(result, typed)
		}
	}
	return result
}

// Bounds is the [Start, End) window used by the four-argument form of fill.
type Bounds struct {
	Start int
	End   int
}

// LodashFill mirrors _.fill(array, value), which overwrites every element of
// array in place and returns it.
func LodashFill(array []any, value any) []any {
	for index := range array {
		array[index] = value
	}
	return array
}

// NativeFill mirrors Array.prototype.fill(value).
func NativeFill(array []any, value any) []any {
	for index := 0; index < len(array); index++ {
		array[index] = value
	}
	return array
}

// LodashFillRange mirrors _.fill(array, value, start, end).
func LodashFillRange(array []any, value any, bounds Bounds) []any {
	for index := bounds.Start; index < bounds.End && index < len(array); index++ {
		if index < 0 {
			continue
		}
		array[index] = value
	}
	return array
}

// NativeFillRange mirrors Array.prototype.fill(value, start, end).
func NativeFillRange(array []any, value any, bounds Bounds) []any {
	start, end := bounds.Start, bounds.End
	if start < 0 {
		start = 0
	}
	if end > len(array) {
		end = len(array)
	}
	for index := start; index < end; index++ {
		array[index] = value
	}
	return array
}

// LodashChunk mirrors _.chunk(input, size).
func LodashChunk[T any](input []T, size int) [][]T {
	result := [][]T{}
	if size < 1 {
		return result
	}
	for start := 0; start < len(input); start += size {
		end := start + size
		if end > len(input) {
			end = len(input)
		}
		group := make([]T, 0, end-start)
		group = append(group, input[start:end]...)
		result = append(result, group)
	}
	return result
}

// NativeChunk mirrors the README's reduce-based chunk snippet, which opens a
// new group whenever the index is a multiple of size.
func NativeChunk[T any](input []T, size int) [][]T {
	result := [][]T{}
	if size < 1 {
		return result
	}
	for index, item := range input {
		if index%size == 0 {
			result = append(result, []T{item})
			continue
		}
		last := len(result) - 1
		result[last] = append(result[last], item)
	}
	return result
}

// LodashTimes mirrors _.times(n), which yields the indices 0..n-1.
func LodashTimes(n int) []int {
	result := make([]int, 0, n)
	for index := 0; index < n; index++ {
		result = append(result, index)
	}
	return result
}

// NativeTimes mirrors Array.from(Array(n), (_, x) => x).
func NativeTimes(n int) []int {
	result := []int{}
	for index := range make([]struct{}, n) {
		result = append(result, index)
	}
	return result
}

// LodashTimesFunc mirrors _.times(n, iteratee), where iteratee receives the
// index.
func LodashTimesFunc(n int, iteratee func(index int) int) []int {
	result := make([]int, 0, n)
	for index := 0; index < n; index++ {
		result = append(result, iteratee(index))
	}
	return result
}

// NativeTimesFunc mirrors Array.from(Array(n), fn), where fn receives the
// element and then the index; the README snippet ignores the element.
func NativeTimesFunc(n int, iteratee func(index int) int) []int {
	result := []int{}
	for index := range make([]struct{}, n) {
		result = append(result, iteratee(index))
	}
	return result
}

// LodashLast mirrors _.last. The empty-slice row of the original suite is not
// translated because its result is `undefined`, so the second return value
// reports whether a value existed at all rather than modelling `undefined`.
func LodashLast[T any](array []T) (T, bool) {
	var zero T
	if len(array) == 0 {
		return zero, false
	}
	return array[len(array)-1], true
}

// LodashFirst mirrors _.first.
func LodashFirst[T any](array []T) (T, bool) {
	var zero T
	if len(array) == 0 {
		return zero, false
	}
	return array[0], true
}

// LodashHead mirrors _.head.
func LodashHead[T any](array []T) (T, bool) {
	if len(array) == 0 {
		var zero T
		return zero, false
	}
	return array[0], true
}

// NativeAt mirrors Array.prototype.at(index), including negative indices.
func NativeAt[T any](array []T, index int) (T, bool) {
	var zero T
	position := index
	if position < 0 {
		position += len(array)
	}
	if position < 0 || position >= len(array) {
		return zero, false
	}
	return array[position], true
}

// LodashUnionBy mirrors _.unionBy(...arrays, iteratee): the concatenation of
// every array, keeping the first element for each iteratee result.
func LodashUnionBy[T any, K comparable](arrays [][]T, iteratee func(value T) K) []T {
	seen := map[K]bool{}
	result := []T{}
	for _, array := range arrays {
		for _, value := range array {
			key := iteratee(value)
			if seen[key] {
				continue
			}
			seen[key] = true
			result = append(result, value)
		}
	}
	return result
}

// NativeUnionBy mirrors the README's flat-then-filter unionBy snippet.
func NativeUnionBy[T any, K comparable](arrays [][]T, iteratee func(value T) K) []T {
	flat := []T{}
	for _, array := range arrays {
		flat = append(flat, array...)
	}
	set := map[K]struct{}{}
	result := []T{}
	for _, value := range flat {
		if _, present := set[iteratee(value)]; present {
			continue
		}
		set[iteratee(value)] = struct{}{}
		result = append(result, value)
	}
	return result
}
