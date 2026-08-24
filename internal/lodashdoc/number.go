package lodashdoc

import (
	"math"
	"math/rand"
)

// LodashIsInteger mirrors _.isInteger, which lodash defines as
// isNumber(value) && toInteger(value) === value. Numeric JavaScript literals
// are carried across as Go int when the original is written without a
// fractional part and as float64 otherwise, which is the representation that
// preserves the asserted behaviour (NaN only exists as a float64).
func LodashIsInteger(value any) bool {
	switch typed := value.(type) {
	case int:
		return true
	case float64:
		return toInteger(typed) == typed
	default:
		return false
	}
}

// NativeIsInteger mirrors Number.isInteger.
func NativeIsInteger(value any) bool {
	switch typed := value.(type) {
	case int:
		return true
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return false
		}
		return math.Trunc(typed) == typed
	default:
		return false
	}
}

// toInteger mirrors lodash's toInteger for a float64: NaN becomes 0 and the
// fractional part is discarded.
func toInteger(value float64) float64 {
	if math.IsNaN(value) {
		return 0
	}
	if math.IsInf(value, 0) {
		return value
	}
	return math.Trunc(value)
}

// LodashInRange mirrors _.inRange(number, start, end); when end is omitted the
// range is [0, start).
func LodashInRange(number float64, start float64, end ...float64) bool {
	lower, upper := 0.0, start
	if len(end) > 0 {
		lower, upper = start, end[0]
	}
	if lower > upper {
		lower, upper = upper, lower
	}
	return number >= lower && number < upper
}

// NativeInRange mirrors the README's inRange snippet, which swaps the bounds
// through Math.min/Math.max instead of comparing them.
func NativeInRange(num float64, init float64, final ...float64) bool {
	start, end := init, 0.0
	if len(final) > 0 {
		end = final[0]
	} else {
		start, end = 0, init
	}
	return num >= math.Min(start, end) && num < math.Max(start, end)
}

// NativeRandom mirrors the README's random snippet, whose parameters default to
// a = 1 and b = 0. The source of randomness is injected so the suite's
// thousand-iteration statistical assertions are deterministic; Math.random()
// and rand.Float64() both produce a float in [0, 1).
func NativeRandom(rng *rand.Rand, bounds ...float64) float64 {
	a, b := 1.0, 0.0
	if len(bounds) > 0 {
		a = bounds[0]
	}
	if len(bounds) > 1 {
		b = bounds[1]
	}
	lower := math.Min(a, b)
	upper := math.Max(a, b)
	return lower + rng.Float64()*(upper-lower)
}

// NativeRandomInt mirrors the README's randomInt snippet, whose parameters
// default to a = 1 and b = 0.
func NativeRandomInt(rng *rand.Rand, bounds ...float64) int {
	a, b := 1.0, 0.0
	if len(bounds) > 0 {
		a = bounds[0]
	}
	if len(bounds) > 1 {
		b = bounds[1]
	}
	lower := math.Ceil(math.Min(a, b))
	upper := math.Floor(math.Max(a, b))
	return int(math.Floor(lower + rng.Float64()*(upper-lower+1)))
}

// NativeClamp mirrors the README's clamp snippet. The JavaScript original tests
// `!boundTwo`, which is true both when the argument is omitted and when it is
// the number 0 or NaN; both conditions are reproduced here over the float64
// parameter, so no coercion across unrelated types is involved.
func NativeClamp(number float64, boundOne float64, boundTwo ...float64) float64 {
	if len(boundTwo) == 0 || boundTwo[0] == 0 || math.IsNaN(boundTwo[0]) {
		if math.Max(number, boundOne) == boundOne {
			return number
		}
		return boundOne
	}
	if math.Min(number, boundOne) == number {
		return boundOne
	}
	if math.Max(number, boundTwo[0]) == number {
		return boundTwo[0]
	}
	return number
}
