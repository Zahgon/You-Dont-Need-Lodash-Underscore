package unit_test

import (
	"testing"

	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/lodashdoc"
)

func splitCases(t *testing.T) {
	t.Run("split", func(t *testing.T) {
		const source = "a-b-c"
		const separator = "-"
		const limit = 2
		t.Run(`_.split("a-b-c", "-")`, func(t *testing.T) {
			assertDeepEqual(t,
				lodashdoc.LodashSplit(source, separator),
				lodashdoc.NativeSplit(source, separator),
			)
		})
		t.Run(`_.split("a-b-c", "-", 2)`, func(t *testing.T) {
			assertDeepEqual(t,
				lodashdoc.LodashSplit(source, separator, limit),
				lodashdoc.NativeSplit(source, separator, limit),
			)
		})
	})
}

func padStartCases(t *testing.T) {
	t.Run("padStart", func(t *testing.T) {
		t.Run(`_.padStart("123", 5, "0")`, func(t *testing.T) {
			assertEqual(t,
				lodashdoc.LodashPadStart("123", 5, "0"),
				lodashdoc.NativePadStart("123", 5, "0"),
			)
		})

		t.Run(`_.padStart("123", 6, "_-")`, func(t *testing.T) {
			assertEqual(t,
				lodashdoc.LodashPadStart("123", 6, "_-"),
				lodashdoc.NativePadStart("123", 6, "_-"),
			)
		})
	})
}

func padEndCases(t *testing.T) {
	t.Run("padEnd", func(t *testing.T) {
		t.Run(`_.padEnd("123", 5, "0")`, func(t *testing.T) {
			assertEqual(t,
				lodashdoc.LodashPadEnd("123", 5, "0"),
				lodashdoc.NativePadEnd("123", 5, "0"),
			)
		})

		t.Run(`_.padEnd("123", 6, "_-")`, func(t *testing.T) {
			assertEqual(t,
				lodashdoc.LodashPadEnd("123", 6, "_-"),
				lodashdoc.NativePadEnd("123", 6, "_-"),
			)
		})
	})
}

func upperFirstCases(t *testing.T) {
	t.Run("upperFirst", func(t *testing.T) {
		t.Run(`_.upperFirst("george")`, func(t *testing.T) {
			assertEqual(t,
				lodashdoc.LodashUpperFirst("george"),
				lodashdoc.NativeUpperFirst("george"),
			)
		})

		t.Run(`_.upperFirst("")`, func(t *testing.T) {
			assertEqual(t,
				lodashdoc.LodashUpperFirst(""),
				lodashdoc.NativeUpperFirst(""),
			)
		})
	})
}

func lowerFirstCases(t *testing.T) {
	t.Run("lowerFirst", func(t *testing.T) {
		// The original title reads "Fred" while the call passes 'fred'; the
		// title is kept verbatim and the input is kept as written.
		t.Run(`_.lowerFirst("Fred")`, func(t *testing.T) {
			assertEqual(t,
				lodashdoc.LodashLowerFirst("fred"),
				lodashdoc.NativeLowerFirst("fred"),
			)
		})

		t.Run(`_.lowerFirst("")`, func(t *testing.T) {
			assertEqual(t,
				lodashdoc.LodashLowerFirst(""),
				lodashdoc.NativeLowerFirst(""),
			)
		})
	})
}

func startsWithCases(t *testing.T) {
	t.Run("startsWith", func(t *testing.T) {
		t.Run(`_.startsWith('abc', 'a')`, func(t *testing.T) {
			assertDeepEqual(t,
				lodashdoc.LodashStartsWith("abc", "a"),
				lodashdoc.NativeStartsWith("abc", "a"),
			)
		})
		t.Run(`_.startsWith('abc', 'b')`, func(t *testing.T) {
			assertDeepEqual(t,
				lodashdoc.LodashStartsWith("abc", "b"),
				lodashdoc.NativeStartsWith("abc", "b"),
			)
		})
		t.Run(`_.startsWith('abc', 'b', 1)`, func(t *testing.T) {
			assertDeepEqual(t,
				lodashdoc.LodashStartsWith("abc", "b", 1),
				lodashdoc.NativeStartsWith("abc", "b", 1),
			)
		})
	})
}

func endsWithCases(t *testing.T) {
	t.Run("endsWith", func(t *testing.T) {
		t.Run(`_.endsWith('abc', 'c')`, func(t *testing.T) {
			assertDeepEqual(t,
				lodashdoc.LodashEndsWith("abc", "c"),
				lodashdoc.NativeEndsWith("abc", "c"),
			)
		})
		t.Run(`_.endsWith('abc', 'b')`, func(t *testing.T) {
			assertDeepEqual(t,
				lodashdoc.LodashEndsWith("abc", "b"),
				lodashdoc.NativeEndsWith("abc", "b"),
			)
		})
		t.Run(`_.endsWith('abc', 'b', 2)`, func(t *testing.T) {
			assertDeepEqual(t,
				lodashdoc.LodashEndsWith("abc", "b", 2),
				lodashdoc.NativeEndsWith("abc", "b", 2),
			)
		})
	})
}

func capitalizeCases(t *testing.T) {
	t.Run("capitalize", func(t *testing.T) {
		t.Run(`_.capitalize("FRED")`, func(t *testing.T) {
			assertDeepEqual(t, lodashdoc.LodashCapitalize("FRED"), lodashdoc.NativeCapitalize("FRED"))
		})

		t.Run(`_.capitalize("fred")`, func(t *testing.T) {
			assertDeepEqual(t, lodashdoc.LodashCapitalize("fred"), lodashdoc.NativeCapitalize("fred"))
		})

		t.Run(`_.capitalize("HELLO WORLD")`, func(t *testing.T) {
			assertDeepEqual(t,
				lodashdoc.LodashCapitalize("HELLO WORLD"),
				lodashdoc.NativeCapitalize("HELLO WORLD"),
			)
		})

		t.Run(`_.capitalize("hello world")`, func(t *testing.T) {
			assertDeepEqual(t,
				lodashdoc.LodashCapitalize("hello world"),
				lodashdoc.NativeCapitalize("hello world"),
			)
		})
	})
}
