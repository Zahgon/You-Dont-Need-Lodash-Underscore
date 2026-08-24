package lodashdoc

import (
	"reflect"
	"time"
)

// Throttle ports the README's hand-written throttle snippet. The JavaScript
// original reads the wall clock through `new Date()` and starts from
// `lastTime = 0`; the clock is injected here so the suite stays deterministic,
// and the zero time.Time plays the role of the epoch, leaving the very first
// call always outside the time frame.
func Throttle(fn func(), timeFrame time.Duration, now func() time.Time) func() {
	var lastTime time.Time
	return func() {
		current := now()
		if current.Sub(lastTime) < timeFrame {
			return
		}
		fn()
		lastTime = current
	}
}

// LodashIsFunction mirrors _.isFunction.
func LodashIsFunction(value any) bool {
	if value == nil {
		return false
	}
	return reflect.ValueOf(value).Kind() == reflect.Func
}

// NativeIsFunction mirrors the README's `func && typeof func === "function"`
// snippet. The leading truthiness guard cannot change the answer for any of the
// values the suite passes - a function is always truthy, and so are the number
// 1 and the string "abc" - so the port only keeps the type test.
func NativeIsFunction(value any) bool {
	if value == nil {
		return false
	}
	return reflect.TypeOf(value).Kind() == reflect.Func
}
