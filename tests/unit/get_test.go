package unit_test

import (
	"testing"

	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/lodashdoc"
)

// getCases ports the `get` describe. The fixture keeps the original keys
// exactly, including the dotted keys "gg.h", "kk.ll", "mm.n" and "oo.p"; the
// numeric key 1 of the nested object becomes the string key "1" because
// JavaScript property keys are strings.
func getCases(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		obj := map[string]any{
			"aa":    []any{map[string]any{"b": map[string]any{"c": 0}, "1": 0}},
			"dd":    map[string]any{"ee": map[string]any{"ff": 2}},
			"gg":    map[string]any{"h": 2},
			"gg.h":  1,
			"kk.ll": map[string]any{"mm.n": []any{3, 4, map[string]any{"oo.p": 5}}},
		}

		t.Run("should handle falsy values", func(t *testing.T) {
			val := lodashdoc.LodashGet(obj, "aa[0].b.c", 1)
			assertEqual(t, val, lodashdoc.NativeGet(obj, "aa[0].b.c", 1))
			assertNotEqual(t, val, any(1))
		})
		t.Run("should handle just bracket notation", func(t *testing.T) {
			val := lodashdoc.LodashGet(obj, "aa[0][1]", 1)
			assertEqual(t, val, lodashdoc.NativeGet(obj, "aa[0][1]", 1))
			assertNotEqual(t, val, any(1))
		})
		t.Run("should handle just period notation", func(t *testing.T) {
			val := lodashdoc.LodashGet(obj, "dd.ee.ff", 1)
			assertEqual(t, val, lodashdoc.NativeGet(obj, "dd.ee.ff", 1))
			assertNotEqual(t, val, any(1))
		})
		t.Run("should handle neither notation", func(t *testing.T) {
			val := lodashdoc.LodashGet(obj, "aa", 1)
			assertDeepEqual(t, val, lodashdoc.NativeGet(obj, "aa", 1))
			assertNotEqual(t, val, any(1))
		})
		t.Run("should handle both notation", func(t *testing.T) {
			val := lodashdoc.LodashGet(obj, "aa[0].b.c", 1)
			assertEqual(t, val, lodashdoc.NativeGet(obj, "aa[0].b.c", 1))
			assertNotEqual(t, val, any(1))
		})
		t.Run("should handle array path", func(t *testing.T) {
			path := []any{"aa", []any{0}, "b", "c"}
			val := lodashdoc.LodashGetPath(obj, path, 1)
			assertEqual(t, val, lodashdoc.NativeGetPath(obj, path, 1))
			assertNotEqual(t, val, any(1))
		})
		t.Run("should handle undefined with default", func(t *testing.T) {
			val := lodashdoc.LodashGet(obj, "dd.b", 1)
			assertEqual(t, val, lodashdoc.NativeGet(obj, "dd.b", 1))
		})
		t.Run("should handle deep undefined with default", func(t *testing.T) {
			val := lodashdoc.LodashGet(obj, "dd.b.c", 1)
			assertEqual(t, val, lodashdoc.NativeGet(obj, "dd.b.c", 1))
			assertEqual(t, val, any(1))
		})
		t.Run("should handle null default", func(t *testing.T) {
			val := lodashdoc.LodashGet(obj, "dd.b", nil)
			assertEqual(t, val, lodashdoc.NativeGet(obj, "dd.b", nil))
			assertEqual(t, val, any(nil))
		})
		t.Run("should handle empty path", func(t *testing.T) {
			val := lodashdoc.LodashGet(obj, "", 1)
			assertEqual(t, val, lodashdoc.NativeGet(obj, "", 1))
			assertOK(t, !lodashdoc.IsSameObject(val, obj), "the result not to be the fixture itself")
		})
		t.Run("should handle path contains a key with dots", func(t *testing.T) {
			val := lodashdoc.LodashGet(obj, "gg.h", nil)
			assertEqual(t, val, lodashdoc.NativeGet(obj, "gg.h", nil))
			assertEqual(t, val, any(1))
		})
	})
}
