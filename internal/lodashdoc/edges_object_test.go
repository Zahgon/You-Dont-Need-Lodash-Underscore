package lodashdoc

import (
	"math"
	"reflect"
	"testing"
)

func TestToInteger(t *testing.T) {
	cases := []struct {
		name  string
		value float64
		want  float64
	}{
		{"zero", 0, 0},
		{"positive whole", 3, 3},
		{"positive fractional", 3.9, 3},
		{"negative fractional", -3.9, -3},
		{"negative whole", -7, -7},
		{"positive infinity", math.Inf(1), math.Inf(1)},
		{"negative infinity", math.Inf(-1), math.Inf(-1)},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := toInteger(testCase.value); got != testCase.want {
				t.Errorf("toInteger(%v) = %v, want %v", testCase.value, got, testCase.want)
			}
		})
	}
	t.Run("NaN becomes zero", func(t *testing.T) {
		if got := toInteger(math.NaN()); got != 0 {
			t.Errorf("toInteger(NaN) = %v, want 0", got)
		}
	})
}

func TestNativeClampBranches(t *testing.T) {
	cases := []struct {
		name     string
		number   float64
		boundOne float64
		boundTwo []float64
		want     float64
	}{
		{"upper bound only, below it", 5, 10, nil, 5},
		{"upper bound only, above it", 15, 10, nil, 10},
		{"upper bound only, equal", 10, 10, nil, 10},
		{"zero second bound behaves as omitted", 5, 10, []float64{0}, 5},
		{"zero second bound clamps", 15, 10, []float64{0}, 10},
		{"below the lower bound", -5, 0, []float64{10}, 0},
		{"above the upper bound", 15, 0, []float64{10}, 10},
		{"inside the range", 5, 0, []float64{10}, 5},
		{"at the lower bound", 0, 0, []float64{10}, 0},
		{"at the upper bound", 10, 0, []float64{10}, 10},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := NativeClamp(testCase.number, testCase.boundOne, testCase.boundTwo...)
			if got != testCase.want {
				t.Errorf("NativeClamp(%v, %v, %v) = %v, want %v",
					testCase.number, testCase.boundOne, testCase.boundTwo, got, testCase.want)
			}
		})
	}

	t.Run("NaN second bound behaves as omitted", func(t *testing.T) {
		if got := NativeClamp(15, 10, math.NaN()); got != 10 {
			t.Errorf("NativeClamp(15, 10, NaN) = %v, want 10", got)
		}
	})
}

func TestNativePickEdges(t *testing.T) {
	object := map[string]any{"a": 1, "b": 2}
	cases := []struct {
		name   string
		object map[string]any
		keys   []string
		want   map[string]any
	}{
		{"all keys present", object, []string{"a", "b"}, map[string]any{"a": 1, "b": 2}},
		{"one key missing", object, []string{"a", "zz"}, map[string]any{"a": 1}},
		{"every key missing", object, []string{"zz", "yy"}, map[string]any{}},
		{"no keys", object, nil, map[string]any{}},
		{"nil object", nil, []string{"a"}, map[string]any{}},
	}
	for _, testCase := range cases {
		t.Run("Native "+testCase.name, func(t *testing.T) {
			got := NativePick(testCase.object, testCase.keys)
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("NativePick(%v, %v) = %v, want %v",
					testCase.object, testCase.keys, got, testCase.want)
			}
		})
		t.Run("Lodash "+testCase.name, func(t *testing.T) {
			got := LodashPick(testCase.object, testCase.keys)
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("LodashPick(%v, %v) = %v, want %v",
					testCase.object, testCase.keys, got, testCase.want)
			}
		})
	}
}

func TestProperty(t *testing.T) {
	object := map[string]any{"a": 1}
	list := []any{"x", "y"}
	cases := []struct {
		name      string
		container any
		key       string
		want      any
		wantOK    bool
	}{
		{"map hit", object, "a", 1, true},
		{"map miss", object, "b", nil, false},
		{"slice index", list, "1", "y", true},
		{"slice index zero", list, "0", "x", true},
		{"slice non numeric key", list, "a", nil, false},
		{"slice negative index", list, "-1", nil, false},
		{"slice index past the end", list, "2", nil, false},
		{"integer container", 42, "a", nil, false},
		{"string container", "abc", "0", nil, false},
		{"nil container", nil, "a", nil, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			value, ok := property(testCase.container, testCase.key)
			if ok != testCase.wantOK {
				t.Fatalf("property(%v, %q) ok = %v, want %v",
					testCase.container, testCase.key, ok, testCase.wantOK)
			}
			if !reflect.DeepEqual(value, testCase.want) {
				t.Errorf("property(%v, %q) = %v, want %v",
					testCase.container, testCase.key, value, testCase.want)
			}
		})
	}
}

func TestIsKey(t *testing.T) {
	object := map[string]any{"a": 1, "a.b": 2}
	cases := []struct {
		name string
		path string
		want bool
	}{
		{"plain word", "abc", true},
		{"empty path", "", true},
		{"digits", "123", true},
		{"underscore", "a_b", true},
		{"space is neither plain nor deep", "a b", true},
		{"dotted path present on the object", "a.b", true},
		{"dotted path absent from the object", "x.y", false},
		{"bracket path absent from the object", "a[0]", false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := isKey(object, testCase.path); got != testCase.want {
				t.Errorf("isKey(object, %q) = %v, want %v", testCase.path, got, testCase.want)
			}
		})
	}
}

func TestStringToPath(t *testing.T) {
	cases := []struct {
		name string
		path string
		want []string
	}{
		{"empty", "", []string{}},
		{"single key", "a", []string{"a"}},
		{"dotted", "a.b.c", []string{"a", "b", "c"}},
		{"bracketed", "a[0].b", []string{"a", "0", "b"}},
		{"trailing separator", "a.", []string{"a"}},
		{"leading separator", ".a", []string{"a"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := stringToPath(testCase.path)
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("stringToPath(%q) = %v, want %v", testCase.path, got, testCase.want)
			}
		})
	}
}

func TestJSStringCoercion(t *testing.T) {
	cases := []struct {
		name  string
		value any
		want  string
	}{
		{"string", "abc", "abc"},
		{"integer", 12, "12"},
		{"nil", nil, "<nil>"},
		{"flat array", []any{"a", 1}, "a,1"},
		{"nested array", []any{"a", []any{"b", "c"}}, "a,b,c"},
		{"empty array", []any{}, ""},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := jsString(testCase.value); got != testCase.want {
				t.Errorf("jsString(%v) = %q, want %q", testCase.value, got, testCase.want)
			}
		})
	}
}

func TestIsSameObject(t *testing.T) {
	object := map[string]any{"a": 1}
	t.Run("same map", func(t *testing.T) {
		if !IsSameObject(object, object) {
			t.Errorf("IsSameObject(object, object) = false, want true")
		}
	})
	t.Run("equal but distinct map", func(t *testing.T) {
		if IsSameObject(map[string]any{"a": 1}, object) {
			t.Errorf("IsSameObject(copy, object) = true, want false")
		}
	})
	t.Run("non map value", func(t *testing.T) {
		if IsSameObject(42, object) {
			t.Errorf("IsSameObject(42, object) = true, want false")
		}
	})
}
