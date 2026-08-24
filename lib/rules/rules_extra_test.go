package rules_test

import (
	"reflect"
	"testing"

	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/kebabcase"
	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/linter"
	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/lib/rules"
)

func TestGetUnknownRule(t *testing.T) {
	t.Run("missing name", func(t *testing.T) {
		rule, ok := rules.Get("you-dont-need-lodash-underscore/not-a-rule")
		if ok {
			t.Fatalf("Get returned ok = true for a name that is not exported")
		}
		if rule.Create != nil {
			t.Errorf("Create is non-nil, want nil for the zero rule")
		}
	})
	t.Run("empty name", func(t *testing.T) {
		if _, ok := rules.Get(""); ok {
			t.Fatalf("Get(\"\") returned ok = true, want false")
		}
	})
	t.Run("known name", func(t *testing.T) {
		if _, ok := rules.Get("trim"); !ok {
			t.Fatalf("Get(\"trim\") returned ok = false, want true")
		}
	})
}

func TestLookup(t *testing.T) {
	t.Run("missing key", func(t *testing.T) {
		definition, ok := rules.Lookup("thisMethodDoesNotExist")
		if ok {
			t.Fatalf("Lookup returned ok = true for a key that is not defined")
		}
		if definition != (rules.Definition{}) {
			t.Errorf("definition = %+v, want the zero value", definition)
		}
	})
	t.Run("known key", func(t *testing.T) {
		definition, ok := rules.Lookup("trim")
		if !ok {
			t.Fatalf("Lookup(\"trim\") returned ok = false, want true")
		}
		if definition.Alternative == "" {
			t.Errorf("Alternative is empty, want the native replacement text")
		}
	})
}

func TestDefinitionsTable(t *testing.T) {
	if len(rules.Definitions) != 73 {
		t.Fatalf("rules.json declares %d entries, want 73", len(rules.Definitions))
	}
	if len(rules.All) != len(rules.Definitions) {
		t.Fatalf("rules.All has %d entries, want %d", len(rules.All), len(rules.Definitions))
	}

	seen := map[string]bool{}
	for index, entry := range rules.Definitions {
		t.Run(entry.Key, func(t *testing.T) {
			name := rules.RuleName(entry)
			if name == "" {
				t.Fatalf("resolved rule name is empty")
			}
			want := entry.Definition.RuleName
			if want == "" {
				want = kebabcase.KebabCase(entry.Key)
			}
			if name != want {
				t.Errorf("RuleName = %q, want %q", name, want)
			}
			if entry.Definition.Alternative == "" {
				t.Errorf("Alternative is empty for %q", entry.Key)
			}
			if rules.All[index].ID != name {
				t.Errorf("rules.All[%d].ID = %q, want %q", index, rules.All[index].ID, name)
			}
			if rules.All[index].Rule.Create == nil {
				t.Errorf("rules.All[%d].Rule.Create is nil", index)
			}
			if _, ok := rules.Get(name); !ok {
				t.Errorf("rules.Get(%q) reports the rule is not exported", name)
			}
			if seen[name] {
				t.Errorf("rule name %q is used more than once", name)
			}
			seen[name] = true
		})
	}
}

func lintWith(t *testing.T, name, code string) []string {
	t.Helper()
	rule, ok := rules.Get(name)
	if !ok {
		t.Fatalf("rule %q is not exported", name)
	}
	messages, err := linter.Verify("input.js", code, []linter.NamedRule{{ID: name, Rule: rule}})
	if err != nil {
		t.Fatalf("code %q: %v", code, err)
	}
	texts := make([]string, 0, len(messages))
	for _, message := range messages {
		texts = append(texts, message.Message)
	}
	return texts
}

func TestNonStringRequireArguments(t *testing.T) {
	cases := []struct {
		name string
		code string
	}{
		{"identifier argument", "var name = 'lodash'; var map = require(name);"},
		{"numeric argument", "var map = require(123);"},
		{"template argument", "var map = require(`lodash`);"},
		{"no arguments", "var map = require();"},
		{"two arguments", "var map = require('lodash', 'extra');"},
		{"member argument", "var map = require(config.lodash);"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := lintWith(t, "map", testCase.code)
			if len(got) != 0 {
				t.Errorf("code %q reported %q, want no problems", testCase.code, got)
			}
		})
	}
}

func TestDestructuredRequireShapes(t *testing.T) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "shorthand binding is reported",
			code: "var { map } = require('lodash');",
			want: []string{"{ map } = require('lodash') detected. Consider using the native Array.prototype.map()"},
		},
		{
			name: "quoted key has no identifier name",
			code: "var { 'map': m } = require('lodash');",
			want: nil,
		},
		{
			name: "rest element carries no key",
			code: "var { ...everything } = require('lodash');",
			want: nil,
		},
		{
			name: "a different method is not reported",
			code: "var { filter } = require('lodash');",
			want: nil,
		},
		{
			name: "identifier binding is not a pattern",
			code: "var lodash = require('lodash');",
			want: nil,
		},
		{
			name: "require result used directly",
			code: "require('lodash');",
			want: nil,
		},
		{
			name: "destructuring assignment is reported",
			code: "var map; ({ map } = require('lodash'));",
			want: []string{"{ map } = require('lodash') detected. Consider using the native Array.prototype.map()"},
		},
		{
			name: "submodule require is reported",
			code: "var map = require('lodash/map');",
			want: []string{"require('lodash/map') detected. Consider using the native Array.prototype.map()"},
		},
		{
			name: "dotted package require is reported",
			code: "var map = require('lodash.map');",
			want: []string{"require('lodash.map') detected. Consider using the native Array.prototype.map()"},
		},
		{
			name: "unrelated module is ignored",
			code: "var map = require('not-lodash');",
			want: nil,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := lintWith(t, "map", testCase.code)
			if len(got) == 0 && len(testCase.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("code %q reported\n  got  %q\n  want %q", testCase.code, got, testCase.want)
			}
		})
	}
}

func TestImportShapes(t *testing.T) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "named import is reported",
			code: "import { map } from 'lodash';",
			want: []string{"Import { map } from 'lodash' detected. Consider using the native Array.prototype.map()"},
		},
		{
			name: "default import binds the module itself",
			code: "import lodash from 'lodash';",
			want: nil,
		},
		{
			name: "namespace import binds the module itself",
			code: "import * as lodash from 'lodash';",
			want: nil,
		},
		{
			name: "bare import has no specifiers",
			code: "import 'lodash';",
			want: nil,
		},
		{
			name: "submodule import is reported",
			code: "import map from 'lodash/map';",
			want: []string{"Import from 'lodash/map' detected. Consider using the native Array.prototype.map()"},
		},
		{
			name: "lodash-es named import is reported",
			code: "import { map } from 'lodash-es';",
			want: []string{"Import { map } from 'lodash-es' detected. Consider using the native Array.prototype.map()"},
		},
		{
			name: "unrelated module import is ignored",
			code: "import { map } from 'not-lodash';",
			want: nil,
		},
		{
			name: "renamed named import still reports the exported name",
			code: "import { map as mapper } from 'lodash';",
			want: []string{"Import { map } from 'lodash' detected. Consider using the native Array.prototype.map()"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := lintWith(t, "map", testCase.code)
			if len(got) == 0 && len(testCase.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("code %q reported\n  got  %q\n  want %q", testCase.code, got, testCase.want)
			}
		})
	}
}

func TestMemberCallShapes(t *testing.T) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{"underscore receiver", "_.map(list, fn);", []string{"Consider using the native Array.prototype.map()"}},
		{"lodash receiver", "lodash.map(list, fn);", []string{"Consider using the native Array.prototype.map()"}},
		{"underscore name receiver", "underscore.map(list, fn);", []string{"Consider using the native Array.prototype.map()"}},
		{"chained receiver", "_(list).map(fn);", []string{"Consider using the native Array.prototype.map()"}},
		{"unrelated receiver", "other.map(list, fn);", nil},
		{"bare call", "map(list, fn);", nil},
		{"computed property", "_['map'](list, fn);", nil},
		{"different method on lodash", "_.filter(list, fn);", nil},
		{"call on a call result of an unrelated name", "other(list).map(fn);", nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := lintWith(t, "map", testCase.code)
			if len(got) == 0 && len(testCase.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("code %q reported\n  got  %q\n  want %q", testCase.code, got, testCase.want)
			}
		})
	}
}
