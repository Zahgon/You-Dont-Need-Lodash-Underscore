package lodashdoc

import (
	"reflect"
	"testing"
)

func TestLastFirstHeadOnEmptyInput(t *testing.T) {
	t.Run("LodashLast empty", func(t *testing.T) {
		value, ok := LodashLast([]int{})
		if ok {
			t.Errorf("ok = true, want false")
		}
		if value != 0 {
			t.Errorf("value = %d, want the zero value 0", value)
		}
	})
	t.Run("LodashLast populated", func(t *testing.T) {
		value, ok := LodashLast([]int{1, 2, 3})
		if !ok || value != 3 {
			t.Errorf("LodashLast = %d, %v, want 3, true", value, ok)
		}
	})
	t.Run("LodashFirst empty", func(t *testing.T) {
		value, ok := LodashFirst([]string{})
		if ok {
			t.Errorf("ok = true, want false")
		}
		if value != "" {
			t.Errorf("value = %q, want the empty string", value)
		}
	})
	t.Run("LodashFirst populated", func(t *testing.T) {
		value, ok := LodashFirst([]string{"a", "b"})
		if !ok || value != "a" {
			t.Errorf("LodashFirst = %q, %v, want \"a\", true", value, ok)
		}
	})
	t.Run("LodashHead empty", func(t *testing.T) {
		value, ok := LodashHead([]int{})
		if ok || value != 0 {
			t.Errorf("LodashHead = %d, %v, want 0, false", value, ok)
		}
	})
	t.Run("LodashHead populated", func(t *testing.T) {
		value, ok := LodashHead([]int{7, 8})
		if !ok || value != 7 {
			t.Errorf("LodashHead = %d, %v, want 7, true", value, ok)
		}
	})
}

func TestNativeAtIndices(t *testing.T) {
	array := []int{10, 20, 30}
	cases := []struct {
		name   string
		array  []int
		index  int
		want   int
		wantOK bool
	}{
		{name: "first", array: array, index: 0, want: 10, wantOK: true},
		{name: "last", array: array, index: 2, want: 30, wantOK: true},
		{name: "negative one", array: array, index: -1, want: 30, wantOK: true},
		{name: "negative three", array: array, index: -3, want: 10, wantOK: true},
		{name: "negative past the start", array: array, index: -4, want: 0, wantOK: false},
		{name: "past the end", array: array, index: 3, want: 0, wantOK: false},
		{name: "empty slice", array: []int{}, index: 0, want: 0, wantOK: false},
		{name: "empty slice negative", array: []int{}, index: -1, want: 0, wantOK: false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			value, ok := NativeAt(testCase.array, testCase.index)
			if value != testCase.want || ok != testCase.wantOK {
				t.Errorf("NativeAt(%v, %d) = %d, %v, want %d, %v",
					testCase.array, testCase.index, value, ok, testCase.want, testCase.wantOK)
			}
		})
	}
}

func TestFillRangeBounds(t *testing.T) {
	cases := []struct {
		name   string
		bounds Bounds
		want   []any
	}{
		{"negative start is clamped", Bounds{Start: -1, End: 2}, []any{0, 0, 3}},
		{"end past the length is clamped", Bounds{Start: 1, End: 10}, []any{1, 0, 0}},
		{"start past the length fills nothing", Bounds{Start: 5, End: 10}, []any{1, 2, 3}},
		{"empty window fills nothing", Bounds{Start: 2, End: 2}, []any{1, 2, 3}},
		{"whole array", Bounds{Start: 0, End: 3}, []any{0, 0, 0}},
	}
	for _, testCase := range cases {
		t.Run("Lodash "+testCase.name, func(t *testing.T) {
			got := LodashFillRange([]any{1, 2, 3}, 0, testCase.bounds)
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("LodashFillRange(%+v) = %v, want %v", testCase.bounds, got, testCase.want)
			}
		})
		t.Run("Native "+testCase.name, func(t *testing.T) {
			got := NativeFillRange([]any{1, 2, 3}, 0, testCase.bounds)
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("NativeFillRange(%+v) = %v, want %v", testCase.bounds, got, testCase.want)
			}
		})
	}
}

func TestChunkSizes(t *testing.T) {
	cases := []struct {
		name  string
		input []int
		size  int
		want  [][]int
	}{
		{"size zero", []int{1, 2, 3}, 0, [][]int{}},
		{"negative size", []int{1, 2, 3}, -2, [][]int{}},
		{"size one", []int{1, 2}, 1, [][]int{{1}, {2}}},
		{"size larger than input", []int{1, 2, 3}, 5, [][]int{{1, 2, 3}}},
		{"uneven final group", []int{1, 2, 3}, 2, [][]int{{1, 2}, {3}}},
		{"empty input", []int{}, 2, [][]int{}},
	}
	for _, testCase := range cases {
		t.Run("Lodash "+testCase.name, func(t *testing.T) {
			got := LodashChunk(testCase.input, testCase.size)
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("LodashChunk(%v, %d) = %v, want %v",
					testCase.input, testCase.size, got, testCase.want)
			}
		})
		t.Run("Native "+testCase.name, func(t *testing.T) {
			got := NativeChunk(testCase.input, testCase.size)
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("NativeChunk(%v, %d) = %v, want %v",
					testCase.input, testCase.size, got, testCase.want)
			}
		})
	}
}

func TestIsFunctionEdges(t *testing.T) {
	cases := []struct {
		name  string
		value any
		want  bool
	}{
		{"nil", nil, false},
		{"integer", 1, false},
		{"string", "abc", false},
		{"map", map[string]any{}, false},
		{"function", func() {}, true},
		{"function with arguments", func(int) string { return "" }, true},
	}
	for _, testCase := range cases {
		t.Run("Lodash "+testCase.name, func(t *testing.T) {
			if got := LodashIsFunction(testCase.value); got != testCase.want {
				t.Errorf("LodashIsFunction(%v) = %v, want %v", testCase.value, got, testCase.want)
			}
		})
		t.Run("Native "+testCase.name, func(t *testing.T) {
			if got := NativeIsFunction(testCase.value); got != testCase.want {
				t.Errorf("NativeIsFunction(%v) = %v, want %v", testCase.value, got, testCase.want)
			}
		})
	}
}
