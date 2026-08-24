package unit_test

import (
	"math"
	"testing"
	"time"

	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/lodashdoc"
)

func forEachCases(t *testing.T) {
	t.Run("forEach", func(t *testing.T) {
		t.Run("_.forEach(array)", func(t *testing.T) {
			testArray := []int{1, 2, 3, 4}

			lodashOutput := []int{}
			nativeOutput := []int{}

			lodashdoc.LodashForEach(testArray, func(element int) {
				lodashOutput = append(lodashOutput, element)
			})
			lodashdoc.NativeForEach(testArray, func(element int) {
				nativeOutput = append(nativeOutput, element)
			})

			assertDeepEqual(t, lodashOutput, nativeOutput)
		})

		t.Run("_.forEach(object)", func(t *testing.T) {
			testObject := map[string]int{
				"one":   1,
				"two":   2,
				"three": 3,
				"four":  4,
			}

			lodashOutput := []int{}
			nativeOutput := []int{}

			lodashdoc.LodashForEachObject(testObject, func(value int) {
				lodashOutput = append(lodashOutput, value)
			})

			lodashdoc.NativeObjectEntriesForEach(testObject, func(key string, value int) {
				nativeOutput = append(nativeOutput, value)
			})

			assertDeepEqual(t, lodashOutput, nativeOutput)
		})
	})
}

func throttleCases(t *testing.T) {
	t.Run("throttle", func(t *testing.T) {
		t.Run("throttle is not called more than once within timeframe", func(t *testing.T) {
			// The JS original reads `new Date()` three times in a row, so all
			// three calls land inside the same time frame; a fixed clock makes
			// that deterministic instead of dependent on machine speed.
			clock := func() time.Time { return time.Unix(1700000000, 0) }
			callCount := 0
			fn := lodashdoc.Throttle(func() { callCount++ }, 100*time.Millisecond, clock)

			fn()
			fn()
			fn()

			assertEqual(t, callCount, 1)
		})
	})
}

func isFunctionCases(t *testing.T) {
	t.Run("isFunction", func(t *testing.T) {
		// time.AfterFunc stands in for the host-provided setTimeout.
		t.Run("_.isFunction(setTimeout)", func(t *testing.T) {
			assertDeepEqual(t,
				lodashdoc.LodashIsFunction(time.AfterFunc),
				lodashdoc.NativeIsFunction(time.AfterFunc),
			)
		})

		t.Run("_.isFunction(1)", func(t *testing.T) {
			assertDeepEqual(t, lodashdoc.LodashIsFunction(1), lodashdoc.NativeIsFunction(1))
		})

		t.Run("_.isFunction(abc)", func(t *testing.T) {
			assertDeepEqual(t, lodashdoc.LodashIsFunction("abc"), lodashdoc.NativeIsFunction("abc"))
		})
	})
}

// point carries the { x, y } object literals of the second unionBy case.
type point struct {
	X int
	Y int
}

func unionByCases(t *testing.T) {
	t.Run("unionBy", func(t *testing.T) {
		t.Run("should take an iteratee function", func(t *testing.T) {
			arrays := [][]float64{{2.1}, {1.2, 2.3}}
			floor := func(value float64) float64 { return math.Floor(value) }
			assertDeepEqual(t,
				lodashdoc.LodashUnionBy(arrays, floor),
				lodashdoc.NativeUnionBy(arrays, floor),
			)
		})

		t.Run("should output values from the first possible array", func(t *testing.T) {
			arrays := [][]point{{{X: 1, Y: 1}}, {{X: 1, Y: 2}}}
			byX := func(value point) int { return value.X }
			assertDeepEqual(t,
				lodashdoc.LodashUnionBy(arrays, byX),
				lodashdoc.NativeUnionBy(arrays, byX),
			)
		})
	})
}

func defaultsCases(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		t.Run("sets up default values the same way", func(t *testing.T) {
			defaultValues := map[string]any{"a": 1, "b": 2, "c": 3}
			givenValues := map[string]any{"b": 4}

			lodashObject := lodashdoc.LodashDefaults(givenValues, defaultValues)
			vanillaObject := lodashdoc.NativeAssign(map[string]any{}, defaultValues, givenValues)

			assertDeepEqual(t, vanillaObject, map[string]any{"a": 1, "b": 4, "c": 3})
			assertDeepEqual(t, vanillaObject, lodashObject)
		})

		t.Run("should handle nested values equally", func(t *testing.T) {
			defaultValues := map[string]any{"a": 1, "b": 2, "c": map[string]any{"x": 3, "y": 4}}
			givenValues := map[string]any{"c": map[string]any{"x": 5}}

			lodashObject := lodashdoc.LodashDefaults(givenValues, defaultValues)
			vanillaObject := lodashdoc.NativeAssign(map[string]any{}, defaultValues, givenValues)

			assertDeepEqual(t, vanillaObject, map[string]any{"a": 1, "b": 2, "c": map[string]any{"x": 5}})
			assertDeepEqual(t, vanillaObject, lodashObject)
		})
	})
}

func lastCases(t *testing.T) {
	t.Run("last", func(t *testing.T) {
		t.Run("_.last([1,2,3,4,5])", func(t *testing.T) {
			value, present := lodashdoc.LodashLast([]int{1, 2, 3, 4, 5})
			nativeValue, nativePresent := lodashdoc.NativeAt([]int{1, 2, 3, 4, 5}, -1)
			assertEqual(t, present, nativePresent)
			assertDeepEqual(t, value, nativeValue)
		})
	})
}

func firstCases(t *testing.T) {
	t.Run("first", func(t *testing.T) {
		t.Run("_.first([1,2,3,4,5])", func(t *testing.T) {
			value, present := lodashdoc.LodashFirst([]int{1, 2, 3, 4, 5})
			nativeValue, nativePresent := lodashdoc.NativeAt([]int{1, 2, 3, 4, 5}, 0)
			assertEqual(t, present, nativePresent)
			assertDeepEqual(t, value, nativeValue)
		})
	})
}

func headCases(t *testing.T) {
	t.Run("head", func(t *testing.T) {
		t.Run("_.head([1,2,3,4,5])", func(t *testing.T) {
			value, present := lodashdoc.LodashHead([]int{1, 2, 3, 4, 5})
			nativeValue, nativePresent := lodashdoc.NativeAt([]int{1, 2, 3, 4, 5}, 0)
			assertEqual(t, present, nativePresent)
			assertDeepEqual(t, value, nativeValue)
		})
	})
}
