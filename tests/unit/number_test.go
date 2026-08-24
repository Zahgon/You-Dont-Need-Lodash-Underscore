package unit_test

import (
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/lodashdoc"
)

// iterations is the length of the `Array(1000).fill(0)` the random describes
// map over.
const iterations = 1000

// newRNG returns a deterministic source of randomness. Math.random() and
// rand.Float64() both yield a float in [0, 1); seeding keeps the suite's
// statistical assertions reproducible instead of flaky.
func newRNG() *rand.Rand {
	return rand.New(rand.NewSource(1))
}

func isIntegerCases(t *testing.T) {
	t.Run("isInteger", func(t *testing.T) {
		t.Run("_.isInteger(3)", func(t *testing.T) {
			assertEqual(t, lodashdoc.LodashIsInteger(3), lodashdoc.NativeIsInteger(3))
		})
		t.Run(`_.isInteger("3")`, func(t *testing.T) {
			assertEqual(t, lodashdoc.LodashIsInteger("3"), lodashdoc.NativeIsInteger("3"))
		})
		t.Run("_.isInteger(2.9)", func(t *testing.T) {
			assertEqual(t, lodashdoc.LodashIsInteger(2.9), lodashdoc.NativeIsInteger(2.9))
		})
		t.Run("_.isInteger(NaN)", func(t *testing.T) {
			// NaN only exists as a float64, so the JS number literal NaN is
			// carried across as math.NaN().
			assertEqual(t, lodashdoc.LodashIsInteger(math.NaN()), lodashdoc.NativeIsInteger(math.NaN()))
		})
	})
}

func inRangeCases(t *testing.T) {
	t.Run("inRange", func(t *testing.T) {
		t.Run("_.inRange(3, 2, 4)", func(t *testing.T) {
			assertEqual(t, lodashdoc.LodashInRange(3, 2, 4), lodashdoc.NativeInRange(3, 2, 4))
		})
		t.Run("_.inRange(4, 8)", func(t *testing.T) {
			assertEqual(t, lodashdoc.LodashInRange(4, 8), lodashdoc.NativeInRange(4, 8))
		})
		t.Run("_.inRange(4, 2)", func(t *testing.T) {
			assertEqual(t, lodashdoc.LodashInRange(4, 2), lodashdoc.NativeInRange(4, 2))
		})
		t.Run("_.inRange(2, 2)", func(t *testing.T) {
			assertEqual(t, lodashdoc.LodashInRange(2, 2), lodashdoc.NativeInRange(2, 2))
		})
		t.Run("_.inRange(1.2, 2)", func(t *testing.T) {
			assertEqual(t, lodashdoc.LodashInRange(1.2, 2), lodashdoc.NativeInRange(1.2, 2))
		})
		t.Run("_.inRange(5.2, 4)", func(t *testing.T) {
			assertEqual(t, lodashdoc.LodashInRange(5.2, 4), lodashdoc.NativeInRange(5.2, 4))
		})
		t.Run("_.inRange(-3, -2, -6)", func(t *testing.T) {
			assertEqual(t, lodashdoc.LodashInRange(-3, -2, -6), lodashdoc.NativeInRange(-3, -2, -6))
		})
		t.Run("_.inRange(1, 1, 5)", func(t *testing.T) {
			assertEqual(t, lodashdoc.LodashInRange(1, 1, 5), lodashdoc.NativeInRange(1, 1, 5))
		})
	})
}

// everyRandom mirrors `array.every(...)` over the thousand-iteration fixture.
func everyRandom(rng *rand.Rand, predicate func(value float64) bool) bool {
	for index := 0; index < iterations; index++ {
		if !predicate(lodashdoc.NativeRandom(rng)) {
			return false
		}
	}
	return true
}

func randomCases(t *testing.T) {
	t.Run("random", func(t *testing.T) {
		t.Run("random() in range [0, 1]", func(t *testing.T) {
			rng := newRNG()
			assertOK(t, everyRandom(rng, func(value float64) bool {
				return value >= 0 && value <= 1
			}), "every random() to fall in [0, 1]")
		})

		t.Run("random() is float", func(t *testing.T) {
			rng := newRNG()
			some := false
			for index := 0; index < iterations; index++ {
				if !lodashdoc.NativeIsInteger(lodashdoc.NativeRandom(rng)) {
					some = true
				}
			}
			assertOK(t, some, "some random() to be a float")
		})

		t.Run("random(5) in range [0, 5]", func(t *testing.T) {
			rng := newRNG()
			assertOK(t, everyBound(rng, func(value float64) bool {
				return value >= 0 && value <= 5
			}, 5), "every random(5) to fall in [0, 5]")
		})

		t.Run("random(5) is float", func(t *testing.T) {
			rng := newRNG()
			some := false
			for index := 0; index < iterations; index++ {
				if !lodashdoc.NativeIsInteger(lodashdoc.NativeRandom(rng, 5)) {
					some = true
				}
			}
			assertOK(t, some, "some random(5) to be a float")
		})

		t.Run("random(-10) supports negative", func(t *testing.T) {
			rng := newRNG()
			assertOK(t, everyBound(rng, func(value float64) bool {
				return value <= 0
			}, -10), "every random(-10) to be at most 0")
		})

		t.Run("random(10, 5) swap the bounds", func(t *testing.T) {
			rng := newRNG()
			assertOK(t, everyBound(rng, func(value float64) bool {
				return value >= 5 && value <= 10
			}, 10, 5), "every random(10, 5) to fall in [5, 10]")
		})

		t.Run("random(-10, 10) supports negative", func(t *testing.T) {
			rng := newRNG()
			assertOK(t, someBound(rng, func(value float64) bool {
				return value > 0
			}, -10, 10), "some random(-10, 10) to be positive")
			assertOK(t, someBound(rng, func(value float64) bool {
				return value < 0
			}, -10, 10), "some random(-10, 10) to be negative")
		})

		t.Run("random(-10, 10) in range [-10, 10]", func(t *testing.T) {
			rng := newRNG()
			assertOK(t, everyBound(rng, func(value float64) bool {
				return value >= -10 && value <= 10
			}, -10, 10), "every random(-10, 10) to fall in [-10, 10]")
		})

		t.Run("random(1.2, 5.2) supports floats", func(t *testing.T) {
			rng := newRNG()
			assertOK(t, everyBound(rng, func(value float64) bool {
				return value >= 1.2 && value <= 5.2
			}, 1.2, 5.2), "every random(1.2, 5.2) to fall in [1.2, 5.2]")
		})

		t.Run("random(100000, 100001) in range [100000, 100001]", func(t *testing.T) {
			rng := newRNG()
			assertOK(t, everyBound(rng, func(value float64) bool {
				return value >= 100000 && value <= 100001
			}, 100000, 100001), "every random(100000, 100001) to fall in [100000, 100001]")
		})
	})
}

func everyBound(rng *rand.Rand, predicate func(value float64) bool, bounds ...float64) bool {
	for index := 0; index < iterations; index++ {
		if !predicate(lodashdoc.NativeRandom(rng, bounds...)) {
			return false
		}
	}
	return true
}

func someBound(rng *rand.Rand, predicate func(value float64) bool, bounds ...float64) bool {
	for index := 0; index < iterations; index++ {
		if predicate(lodashdoc.NativeRandom(rng, bounds...)) {
			return true
		}
	}
	return false
}

func everyInt(rng *rand.Rand, predicate func(value int) bool, bounds ...float64) bool {
	for index := 0; index < iterations; index++ {
		if !predicate(lodashdoc.NativeRandomInt(rng, bounds...)) {
			return false
		}
	}
	return true
}

func someInt(rng *rand.Rand, predicate func(value int) bool, bounds ...float64) bool {
	for index := 0; index < iterations; index++ {
		if predicate(lodashdoc.NativeRandomInt(rng, bounds...)) {
			return true
		}
	}
	return false
}

// uniqSorted mirrors the describe's `uniq` helper followed by `.sort()`.
func uniqSorted(rng *rand.Rand, bounds ...float64) []int {
	seen := map[int]bool{}
	values := []int{}
	for index := 0; index < iterations; index++ {
		value := lodashdoc.NativeRandomInt(rng, bounds...)
		if seen[value] {
			continue
		}
		seen[value] = true
		values = append(values, value)
	}
	sort.Ints(values)
	return values
}

func randomIntCases(t *testing.T) {
	t.Run("randomInt", func(t *testing.T) {
		t.Run("randomInt() return `0` or `1`", func(t *testing.T) {
			assertDeepEqual(t, uniqSorted(newRNG()), []int{0, 1})
		})

		t.Run("randomInt(5) in range [0, 5]", func(t *testing.T) {
			assertOK(t, everyInt(newRNG(), func(value int) bool {
				return value >= 0 && value <= 5
			}, 5), "every randomInt(5) to fall in [0, 5]")
		})

		t.Run("randomInt(5) is integer", func(t *testing.T) {
			assertOK(t, someInt(newRNG(), func(value int) bool {
				return lodashdoc.NativeIsInteger(value)
			}, 5), "some randomInt(5) to be an integer")
		})

		t.Run("randomInt(-10) supports negative", func(t *testing.T) {
			assertOK(t, everyInt(newRNG(), func(value int) bool {
				return value <= 0
			}, -10), "every randomInt(-10) to be at most 0")
		})

		t.Run("randomInt(10, 5) swap the bounds", func(t *testing.T) {
			assertOK(t, everyInt(newRNG(), func(value int) bool {
				return value >= 5 && value <= 10
			}, 10, 5), "every randomInt(10, 5) to fall in [5, 10]")
		})

		t.Run("randomInt(-10, 10) supports negative", func(t *testing.T) {
			rng := newRNG()
			assertOK(t, someInt(rng, func(value int) bool {
				return value > 0
			}, -10, 10), "some randomInt(-10, 10) to be positive")
			assertOK(t, someInt(rng, func(value int) bool {
				return value < 0
			}, -10, 10), "some randomInt(-10, 10) to be negative")
		})

		t.Run("randomInt(-10, 10) in range [-10, 10]", func(t *testing.T) {
			assertOK(t, everyInt(newRNG(), func(value int) bool {
				return value >= -10 && value <= 10
			}, -10, 10), "every randomInt(-10, 10) to fall in [-10, 10]")
		})

		t.Run("randomInt(1.2, 5.2) supports floats", func(t *testing.T) {
			assertOK(t, everyInt(newRNG(), func(value int) bool {
				return value >= 2 && value <= 5
			}, 1.2, 5.2), "every randomInt(1.2, 5.2) to fall in [2, 5]")
		})

		t.Run("randomInt(100000, 100001) return `100000` or `100001`", func(t *testing.T) {
			assertDeepEqual(t, uniqSorted(newRNG(), 100000, 100001), []int{100000, 100001})
		})
	})
}

func clampCases(t *testing.T) {
	t.Run("clamp", func(t *testing.T) {
		t.Run("clamp(-10, -5, 5) returns lower bound if number is less than it", func(t *testing.T) {
			assertEqual(t, lodashdoc.NativeClamp(-10, -5, 5), -5.0)
		})
		t.Run("clamp(10, -5, 5) returns upper bound if number is greater than it", func(t *testing.T) {
			assertEqual(t, lodashdoc.NativeClamp(10, -5, 5), 5.0)
		})
		t.Run("clamp(10, -5) treats second parameter as upper bound", func(t *testing.T) {
			assertEqual(t, lodashdoc.NativeClamp(10, -5), -5.0)
		})
	})
}
