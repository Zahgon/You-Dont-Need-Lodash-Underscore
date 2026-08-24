package lodashdoc

import (
	"reflect"
	"testing"
)

func TestSplitEdges(t *testing.T) {
	t.Run("Lodash empty separator returns the whole string", func(t *testing.T) {
		got := LodashSplit("abc", "")
		want := []string{"abc"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("LodashSplit(\"abc\", \"\") = %v, want %v", got, want)
		}
	})
	t.Run("Lodash empty separator with limit zero yields nothing", func(t *testing.T) {
		got := LodashSplit("abc", "", 0)
		if len(got) != 0 {
			t.Errorf("LodashSplit(\"abc\", \"\", 0) = %v, want an empty slice", got)
		}
	})
	t.Run("Lodash limit zero yields nothing", func(t *testing.T) {
		got := LodashSplit("a-b-c", "-", 0)
		if len(got) != 0 {
			t.Errorf("LodashSplit with limit 0 = %v, want an empty slice", got)
		}
	})
	t.Run("Lodash limit stops the scan", func(t *testing.T) {
		got := LodashSplit("a-b-c", "-", 2)
		want := []string{"a", "b"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("LodashSplit with limit 2 = %v, want %v", got, want)
		}
	})
	t.Run("Lodash separator absent", func(t *testing.T) {
		got := LodashSplit("abc", "-")
		want := []string{"abc"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("LodashSplit(\"abc\", \"-\") = %v, want %v", got, want)
		}
	})
	t.Run("Lodash limit larger than the part count", func(t *testing.T) {
		got := LodashSplit("a-b", "-", 5)
		want := []string{"a", "b"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("LodashSplit(\"a-b\", \"-\", 5) = %v, want %v", got, want)
		}
	})
	t.Run("Native limit zero yields nothing", func(t *testing.T) {
		got := NativeSplit("a-b-c", "-", 0)
		if len(got) != 0 {
			t.Errorf("NativeSplit with limit 0 = %v, want an empty slice", got)
		}
	})
	t.Run("Native negative limit is ignored", func(t *testing.T) {
		got := NativeSplit("a-b", "-", -1)
		want := []string{"a", "b"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("NativeSplit with limit -1 = %v, want %v", got, want)
		}
	})
}

func TestPadEdges(t *testing.T) {
	cases := []struct {
		name   string
		text   string
		length int
		pad    string
		start  string
		end    string
	}{
		{"already long enough", "abc", 2, "-", "abc", "abc"},
		{"exactly the target length", "abc", 3, "-", "abc", "abc"},
		{"empty pad string", "abc", 6, "", "abc", "abc"},
		{"single character pad", "abc", 6, "-", "---abc", "abc---"},
		{"multi character pad truncated", "abc", 6, "12", "121abc", "abc121"},
		{"empty text", "", 3, "x", "xxx", "xxx"},
	}
	for _, testCase := range cases {
		t.Run("LodashPadStart "+testCase.name, func(t *testing.T) {
			if got := LodashPadStart(testCase.text, testCase.length, testCase.pad); got != testCase.start {
				t.Errorf("LodashPadStart(%q, %d, %q) = %q, want %q",
					testCase.text, testCase.length, testCase.pad, got, testCase.start)
			}
		})
		t.Run("NativePadStart "+testCase.name, func(t *testing.T) {
			if got := NativePadStart(testCase.text, testCase.length, testCase.pad); got != testCase.start {
				t.Errorf("NativePadStart(%q, %d, %q) = %q, want %q",
					testCase.text, testCase.length, testCase.pad, got, testCase.start)
			}
		})
		t.Run("LodashPadEnd "+testCase.name, func(t *testing.T) {
			if got := LodashPadEnd(testCase.text, testCase.length, testCase.pad); got != testCase.end {
				t.Errorf("LodashPadEnd(%q, %d, %q) = %q, want %q",
					testCase.text, testCase.length, testCase.pad, got, testCase.end)
			}
		})
		t.Run("NativePadEnd "+testCase.name, func(t *testing.T) {
			if got := NativePadEnd(testCase.text, testCase.length, testCase.pad); got != testCase.end {
				t.Errorf("NativePadEnd(%q, %d, %q) = %q, want %q",
					testCase.text, testCase.length, testCase.pad, got, testCase.end)
			}
		})
	}
}

func TestCaseHelperEdges(t *testing.T) {
	cases := []struct {
		name       string
		text       string
		capitalize string
		upperFirst string
		lowerFirst string
	}{
		{"empty", "", "", "", ""},
		{"single lowercase", "a", "A", "A", "a"},
		{"single uppercase", "A", "A", "A", "a"},
		{"all uppercase", "FRED", "Fred", "FRED", "fRED"},
		{"mixed", "fRED", "Fred", "FRED", "fRED"},
	}
	for _, testCase := range cases {
		t.Run("LodashCapitalize "+testCase.name, func(t *testing.T) {
			if got := LodashCapitalize(testCase.text); got != testCase.capitalize {
				t.Errorf("LodashCapitalize(%q) = %q, want %q", testCase.text, got, testCase.capitalize)
			}
		})
		t.Run("NativeCapitalize "+testCase.name, func(t *testing.T) {
			if got := NativeCapitalize(testCase.text); got != testCase.capitalize {
				t.Errorf("NativeCapitalize(%q) = %q, want %q", testCase.text, got, testCase.capitalize)
			}
		})
		t.Run("LodashUpperFirst "+testCase.name, func(t *testing.T) {
			if got := LodashUpperFirst(testCase.text); got != testCase.upperFirst {
				t.Errorf("LodashUpperFirst(%q) = %q, want %q", testCase.text, got, testCase.upperFirst)
			}
		})
		t.Run("NativeUpperFirst "+testCase.name, func(t *testing.T) {
			if got := NativeUpperFirst(testCase.text); got != testCase.upperFirst {
				t.Errorf("NativeUpperFirst(%q) = %q, want %q", testCase.text, got, testCase.upperFirst)
			}
		})
		t.Run("LodashLowerFirst "+testCase.name, func(t *testing.T) {
			if got := LodashLowerFirst(testCase.text); got != testCase.lowerFirst {
				t.Errorf("LodashLowerFirst(%q) = %q, want %q", testCase.text, got, testCase.lowerFirst)
			}
		})
		t.Run("NativeLowerFirst "+testCase.name, func(t *testing.T) {
			if got := NativeLowerFirst(testCase.text); got != testCase.lowerFirst {
				t.Errorf("NativeLowerFirst(%q) = %q, want %q", testCase.text, got, testCase.lowerFirst)
			}
		})
	}
}

func TestStartsWithEdges(t *testing.T) {
	cases := []struct {
		name     string
		text     string
		target   string
		position []int
		want     bool
	}{
		{"no position", "abc", "a", nil, true},
		{"no position mismatch", "abc", "b", nil, false},
		{"explicit position", "abc", "b", []int{1}, true},
		{"negative position is clamped to zero", "abc", "a", []int{-5}, true},
		{"position past the end", "abc", "a", []int{5}, false},
		{"target longer than the remainder", "abc", "abcd", nil, false},
		{"empty target", "abc", "", nil, true},
		{"empty text", "", "a", nil, false},
	}
	for _, testCase := range cases {
		t.Run("Lodash "+testCase.name, func(t *testing.T) {
			got := LodashStartsWith(testCase.text, testCase.target, testCase.position...)
			if got != testCase.want {
				t.Errorf("LodashStartsWith(%q, %q, %v) = %v, want %v",
					testCase.text, testCase.target, testCase.position, got, testCase.want)
			}
		})
		t.Run("Native "+testCase.name, func(t *testing.T) {
			got := NativeStartsWith(testCase.text, testCase.target, testCase.position...)
			if got != testCase.want {
				t.Errorf("NativeStartsWith(%q, %q, %v) = %v, want %v",
					testCase.text, testCase.target, testCase.position, got, testCase.want)
			}
		})
	}
}

func TestEndsWithEdges(t *testing.T) {
	cases := []struct {
		name     string
		text     string
		target   string
		position []int
		want     bool
	}{
		{"no position", "abc", "c", nil, true},
		{"no position mismatch", "abc", "b", nil, false},
		{"explicit position", "abc", "b", []int{2}, true},
		{"position past the end is clamped", "abc", "c", []int{10}, true},
		{"negative position", "abc", "c", []int{-1}, false},
		{"target longer than the text", "abc", "abcd", nil, false},
		{"empty target", "abc", "", nil, true},
		{"empty text", "", "a", nil, false},
	}
	for _, testCase := range cases {
		t.Run("Lodash "+testCase.name, func(t *testing.T) {
			got := LodashEndsWith(testCase.text, testCase.target, testCase.position...)
			if got != testCase.want {
				t.Errorf("LodashEndsWith(%q, %q, %v) = %v, want %v",
					testCase.text, testCase.target, testCase.position, got, testCase.want)
			}
		})
		t.Run("Native "+testCase.name, func(t *testing.T) {
			got := NativeEndsWith(testCase.text, testCase.target, testCase.position...)
			if got != testCase.want {
				t.Errorf("NativeEndsWith(%q, %q, %v) = %v, want %v",
					testCase.text, testCase.target, testCase.position, got, testCase.want)
			}
		})
	}
}
