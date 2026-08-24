package unit_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/lodashdoc"
)

// This file is the translation of tests/unit/all.js. The original mocha suite
// does not exercise the ESLint plugin at all: it pairs a lodash call with the
// hand-written native JavaScript snippet documented in the README and asserts
// the two agree. Each ported it() below keeps the original title verbatim, the
// original inputs, and the original equivalence assertion, with both sides
// computed by their own helper in internal/lodashdoc.
//
// The cases that could not be ported are listed, with their reason, in
// testdata/untranslatable/manifest.json; the original JavaScript file is
// carried across verbatim next to it.

// assertDeepEqual ports assert.deepEqual / assert.deepStrictEqual.
func assertDeepEqual(t *testing.T, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

// assertEqual ports assert.equal / assert.strictEqual.
func assertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// assertNotEqual ports assert.notEqual.
func assertNotEqual[T comparable](t *testing.T, got, other T) {
	t.Helper()
	if got == other {
		t.Fatalf("got %v, expected it to differ from %v", got, other)
	}
}

// assertOK ports assert.ok.
func assertOK(t *testing.T, condition bool, description string) {
	t.Helper()
	if !condition {
		t.Fatalf("expected %s", description)
	}
}

func TestCodeSnippetExample(t *testing.T) {
	topLevelCases(t)
	fillCases(t)
	chunkCases(t)
	timesCases(t)
	isIntegerCases(t)
	getCases(t)
	splitCases(t)
	inRangeCases(t)
	randomCases(t)
	randomIntCases(t)
	clampCases(t)
	padStartCases(t)
	padEndCases(t)
	upperFirstCases(t)
	lowerFirstCases(t)
	forEachCases(t)
	startsWithCases(t)
	endsWithCases(t)
	throttleCases(t)
	isFunctionCases(t)
	unionByCases(t)
	capitalizeCases(t)
	defaultsCases(t)
	lastCases(t)
	firstCases(t)
	headCases(t)
}

func topLevelCases(t *testing.T) {
	t.Run("concat", func(t *testing.T) {
		lodashArray := []any{1}
		lodashResult := lodashdoc.LodashConcat(lodashArray, 2, []any{3}, []any{[]any{4}})

		nativeArray := []any{1}
		nativeResult := lodashdoc.NativeConcat(nativeArray, 2, []any{3}, []any{[]any{4}})

		assertDeepEqual(t, lodashResult, nativeResult)
	})

	t.Run("invert", func(t *testing.T) {
		// The JS object mixes a number and a string value; both become string
		// keys once inverted, so the fixture is a map[string]any.
		object := map[string]any{"a": 1, "b": "2", "c": 3}
		assertDeepEqual(t, lodashdoc.LodashInvert(object), lodashdoc.NativeInvert(object))
	})

	t.Run("mapKeys", func(t *testing.T) {
		object := map[string]int{"a": 1, "b": 2}
		// The JS callback is `key + value`, string concatenation with a number.
		iteratee := func(value int, key string) string {
			return fmt.Sprintf("%s%d", key, value)
		}
		assertDeepEqual(t,
			lodashdoc.LodashMapKeys(object, iteratee),
			lodashdoc.NativeMapKeys(object, iteratee),
		)
	})

	t.Run("pick", func(t *testing.T) {
		object := map[string]any{"a": 1, "b": "2", "c": 3}
		assertDeepEqual(t,
			lodashdoc.LodashPick(object, []string{"a", "c", "x"}),
			lodashdoc.NativePick(object, []string{"a", "c", "x"}),
		)
	})
}

func fillCases(t *testing.T) {
	t.Run("fill", func(t *testing.T) {
		t.Run("_.fill(array, 'a')", func(t *testing.T) {
			// As in the original, both calls fill the same array in place.
			array := []any{1, 2, 3}
			assertDeepEqual(t,
				lodashdoc.LodashFill(array, "a"),
				lodashdoc.NativeFill(array, "a"),
			)
		})
		t.Run("_.fill(Array(3), 2)", func(t *testing.T) {
			// Array(3) is a length-3 array of holes; every element is
			// overwritten by fill, so the holes are never observable.
			assertDeepEqual(t,
				lodashdoc.LodashFill(make([]any, 3), 2),
				lodashdoc.NativeFill(make([]any, 3), 2),
			)
		})

		t.Run("_.fill([4, 6, 8, 10], '*', 1, 3)", func(t *testing.T) {
			bounds := lodashdoc.Bounds{Start: 1, End: 3}
			assertDeepEqual(t,
				lodashdoc.LodashFillRange([]any{4, 6, 8, 10}, "*", bounds),
				lodashdoc.NativeFillRange([]any{4, 6, 8, 10}, "*", bounds),
			)
		})
	})
}

func chunkCases(t *testing.T) {
	t.Run("chunk", func(t *testing.T) {
		t.Run("_.chunk(['a', 'b', 'c', 'd'], 2);", func(t *testing.T) {
			input := []string{"a", "b", "c", "d"}
			assertDeepEqual(t, lodashdoc.LodashChunk(input, 2), lodashdoc.NativeChunk(input, 2))
		})
		t.Run("_.chunk(['a', 'b', 'c', 'd'], 3);", func(t *testing.T) {
			input := []string{"a", "b", "c", "d"}
			assertDeepEqual(t, lodashdoc.LodashChunk(input, 3), lodashdoc.NativeChunk(input, 3))
		})
	})
}

func timesCases(t *testing.T) {
	t.Run("times", func(t *testing.T) {
		t.Run("_.times(10);", func(t *testing.T) {
			assertDeepEqual(t, lodashdoc.LodashTimes(10), lodashdoc.NativeTimes(10))
		})
		t.Run("_.times(10, x => x + 1);", func(t *testing.T) {
			increment := func(x int) int { return x + 1 }
			assertDeepEqual(t,
				lodashdoc.LodashTimesFunc(10, increment),
				lodashdoc.NativeTimesFunc(10, increment),
			)
		})
	})
}
