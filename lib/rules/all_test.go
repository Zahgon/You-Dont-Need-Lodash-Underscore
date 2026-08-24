package rules_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/linter"
	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/lib/rules"
)

// This file is the translation of tests/lib/rules/all.js. The original suite
// drove ESLint's RuleTester, which lints a snippet with a single rule and
// compares the reported messages against an expected list. The same contract is
// reproduced below so the snippets and expected messages can be carried across
// unchanged.

type invalidCase struct {
	code   string
	errors []string
}

func lookup(t *testing.T, name string) linter.Rule {
	t.Helper()
	rule, ok := rules.Get(name)
	if !ok {
		t.Fatalf("rule %q is not exported by the plugin", name)
	}
	return rule
}

func messageTexts(messages []linter.Message) []string {
	texts := make([]string, 0, len(messages))
	for _, message := range messages {
		texts = append(texts, message.Message)
	}
	return texts
}

func run(t *testing.T, title string, rule linter.Rule, valid []string, invalid []invalidCase) {
	t.Helper()
	t.Run(title, func(t *testing.T) {
		for index, code := range valid {
			t.Run(fmt.Sprintf("valid[%d]", index), func(t *testing.T) {
				messages, err := linter.Verify("input.js", code, []linter.NamedRule{{ID: title, Rule: rule}})
				if err != nil {
					t.Fatalf("code %q: %v", code, err)
				}
				if len(messages) != 0 {
					t.Fatalf("code %q: expected no problems, got %q", code, messageTexts(messages))
				}
			})
		}
		for index, testCase := range invalid {
			t.Run(fmt.Sprintf("invalid[%d]", index), func(t *testing.T) {
				messages, err := linter.Verify("input.js", testCase.code, []linter.NamedRule{{ID: title, Rule: rule}})
				if err != nil {
					t.Fatalf("code %q: %v", testCase.code, err)
				}
				got := messageTexts(messages)
				if !reflect.DeepEqual(got, testCase.errors) {
					t.Fatalf("code %q:\n  expected %q\n  got      %q", testCase.code, testCase.errors, got)
				}
			})
		}
	})
}

// Only a couple of smoke tests because otherwise it would get very redundant.
func TestRules(t *testing.T) {
	run(t, "_.concat", lookup(t, "concat"),
		[]string{`array.concat(2, [3], [[4]])`},
		[]invalidCase{{
			code:   `_.concat(array, 2, [3], [[4]])`,
			errors: []string{`Consider using the native Array.prototype.concat()`},
		}})

	run(t, "lodash.keys", lookup(t, "keys"),
		[]string{`Object.keys({one: 1, two: 2, three: 3})`},
		[]invalidCase{{
			code:   `lodash.keys({one: 1, two: 2, three: 3})`,
			errors: []string{`Consider using the native Object.keys()`},
		}})

	run(t, "Import lodash.isnan", lookup(t, "is-nan"),
		[]string{`{ x: require('lodash') }`},
		[]invalidCase{
			{
				code:   `import isNaN from 'lodash/isNaN';`,
				errors: []string{`Import from 'lodash/isNaN' detected. Consider using the native Number.isNaN()`},
			},
			{
				code:   `import isNaN from 'lodash.isnan';`,
				errors: []string{`Import from 'lodash.isnan' detected. Consider using the native Number.isNaN()`},
			},
			{
				code:   `import { isNaN as x } from 'lodash';`,
				errors: []string{`Import { isNaN } from 'lodash' detected. Consider using the native Number.isNaN()`},
			},
			{
				code:   `const { isNaN: x } = require('lodash');`,
				errors: []string{`{ isNaN } = require('lodash') detected. Consider using the native Number.isNaN()`},
			},
			{
				code:   `({ isNaN: x } = require('lodash'));`,
				errors: []string{`{ isNaN } = require('lodash') detected. Consider using the native Number.isNaN()`},
			},
			{
				code:   `require('lodash/isNaN');`,
				errors: []string{`require('lodash/isNaN') detected. Consider using the native Number.isNaN()`},
			},
			{
				code:   `require('lodash.isnan');`,
				errors: []string{`require('lodash.isnan') detected. Consider using the native Number.isNaN()`},
			},
		})

	run(t, "Import { isNaN } from lodash-es", lookup(t, "is-nan"),
		[]string{`{ x: require('lodash-es') }`},
		[]invalidCase{
			{
				code:   `import { isNaN } from 'lodash-es';`,
				errors: []string{`Import { isNaN } from 'lodash-es' detected. Consider using the native Number.isNaN()`},
			},
			{
				code:   `import isNaN from 'lodash-es/isNaN';`,
				errors: []string{`Import from 'lodash-es/isNaN' detected. Consider using the native Number.isNaN()`},
			},
			{
				code:   `import { isNaN as x } from 'lodash-es';`,
				errors: []string{`Import { isNaN } from 'lodash-es' detected. Consider using the native Number.isNaN()`},
			},
			{
				code:   `const { isNaN: x } = require('lodash-es');`,
				errors: []string{`{ isNaN } = require('lodash-es') detected. Consider using the native Number.isNaN()`},
			},
			{
				code:   `({ isNaN: x } = require('lodash-es'));`,
				errors: []string{`{ isNaN } = require('lodash-es') detected. Consider using the native Number.isNaN()`},
			},
			{
				code:   `require('lodash-es/isNaN');`,
				errors: []string{`require('lodash-es/isNaN') detected. Consider using the native Number.isNaN()`},
			},
		})

	run(t, "underscore.forEach", lookup(t, "for-each"),
		[]string{
			`[0, 1].forEach()`,
			`Object.entries({'one':1,'two':2}).forEach()`,
		},
		[]invalidCase{{
			code:   `underscore.forEach()`,
			errors: []string{`Consider using the native Array.prototype.forEach() or Object.entries().forEach()`},
		}})

	run(t, "underscore.isNaN", lookup(t, "is-nan"),
		[]string{`Number.isNaN(NaN);`},
		[]invalidCase{{
			code:   `underscore.isNaN(NaN)`,
			errors: []string{`Consider using the native Number.isNaN()`},
		}})

	run(t, "_.first", lookup(t, "first"),
		[]string{
			`[0, 1, 3][0]`,
			`[0, 1, 3].at(0)`,
			`[0, 1, 3].slice(0, 2)`,
		},
		[]invalidCase{
			{
				code:   `_.first([0, 1, 3])`,
				errors: []string{`Consider using the native Array.prototype.at(0) or Array.prototype.slice()`},
			},
			{
				code:   `_.first([0, 1, 3], 2)`,
				errors: []string{`Consider using the native Array.prototype.at(0) or Array.prototype.slice()`},
			},
		})

	run(t, "_.last", lookup(t, "last"),
		[]string{
			`var numbers = [0, 1, 3]; numbers[numbers.length - 1]`,
			`[0, 1, 3].at(-1)`,
			`[0, 1, 3].slice(-2)`,
		},
		[]invalidCase{
			{
				code:   `_.last([0, 1, 3])`,
				errors: []string{`Consider using the native Array.prototype.at(-1) or Array.prototype.slice()`},
			},
			{
				code:   `_.last([0, 1, 3], 2)`,
				errors: []string{`Consider using the native Array.prototype.at(-1) or Array.prototype.slice()`},
			},
		})

	run(t, "_", lookup(t, "concat"),
		[]string{`_(2, [3], [[4]])`},
		nil)

	run(t, "_.isUndefined", lookup(t, "is-undefined"),
		[]string{`2 === undefined`},
		[]invalidCase{
			{
				code:   `_.isUndefined(2)`,
				errors: []string{`Consider using the native value === undefined`},
			},
			{
				code:   `_(2).isUndefined()`,
				errors: []string{`Consider using the native value === undefined`},
			},
		})

	// This is to make sure that You-Dont-Need-Lodash can handle the evaluation
	// of nested functions that had caused an error noted in the comments of
	// Pull Request #219.
	run(t, "Nested functions", lookup(t, "is-undefined"),
		[]string{`function myNestedFunction(firstInput) {
      return (secondInput) => {
        return firstInput + secondInput
      }
    }
    myNestedFunction(2)(2)`},
		nil)

	// Test for new flatten rule.
	run(t, "_.flatten", lookup(t, "flatten"),
		[]string{
			`[1,2,[3,4]].reduce((a,b) => a.concat(b), [])`,
			`[1,2,[3,4]].flat()`,
			`[1,2,[3,4]].flatMap(a => a)`,
		},
		[]invalidCase{
			{
				code:   `_.flatten([1,2,[3,4]])`,
				errors: []string{`Consider using the native Array.prototype.reduce((a,b) => a.concat(b), [])`},
			},
			{
				code:   `_([1,2,[3,4]]).flatten()`,
				errors: []string{`Consider using the native Array.prototype.reduce((a,b) => a.concat(b), [])`},
			},
		})

	// The original suite declares this block twice, verbatim.
	run(t, "_.isUndefined", lookup(t, "is-undefined"),
		[]string{`2 === undefined`},
		[]invalidCase{
			{
				code:   `_.isUndefined(2)`,
				errors: []string{`Consider using the native value === undefined`},
			},
			{
				code:   `_(2).isUndefined()`,
				errors: []string{`Consider using the native value === undefined`},
			},
		})

	run(t, "_.startsWith", lookup(t, "starts-with"),
		[]string{
			`"abc".startsWith("a")`,
			`"abc".startsWith("b", 1)`,
		},
		[]invalidCase{
			{
				code:   `_.startsWith("abc", "a")`,
				errors: []string{`Consider using the native String.prototype.startsWith()`},
			},
			{
				code:   `_.startsWith("abc", "b", 1)`,
				errors: []string{`Consider using the native String.prototype.startsWith()`},
			},
		})

	run(t, "_.endsWith", lookup(t, "ends-with"),
		[]string{
			`"abc".endsWith("c")`,
			`"abc".endsWith("b", 1)`,
		},
		[]invalidCase{
			{
				code:   `_.endsWith("abc", "c")`,
				errors: []string{`Consider using the native String.prototype.endsWith()`},
			},
			{
				code:   `_.endsWith("abc", "b", 1)`,
				errors: []string{`Consider using the native String.prototype.endsWith()`},
			},
		})

	run(t, "_.head", lookup(t, "head"),
		[]string{`[0, 1, 3].at(0)`},
		[]invalidCase{{
			code:   `_.head([0, 1, 3])`,
			errors: []string{`Consider using the native Array.prototype.at(0)`},
		}})
}
