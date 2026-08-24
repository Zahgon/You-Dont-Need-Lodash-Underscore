package plugin_test

import (
	"testing"

	plugin "github.com/you-dont-need/You-Dont-Need-Lodash-Underscore"
)

// This file is the translation of tests/index.js. The original suite asserted
// the severity a handful of rules receive in each shareable configuration, and
// that a rule missing from a configuration reads back as absent.

func level(t *testing.T, config, rule string) (int, bool) {
	t.Helper()
	configuration, ok := plugin.ConfigsValue.Get(config)
	if !ok {
		t.Fatalf("configuration %q is not exported", config)
	}
	return configuration.Rules.Get(rule)
}

func assertLevel(t *testing.T, config, rule string, want int) {
	t.Helper()
	got, present := level(t, config, rule)
	if !present {
		t.Fatalf("configs[%q].rules[%q] is absent, expected %d", config, rule, want)
	}
	if got != want {
		t.Fatalf("configs[%q].rules[%q] = %d, expected %d", config, rule, got, want)
	}
}

func assertAbsent(t *testing.T, config, rule string) {
	t.Helper()
	if got, present := level(t, config, rule); present {
		t.Fatalf("configs[%q].rules[%q] = %d, expected it to be absent", config, rule, got)
	}
}

func TestConfigs(t *testing.T) {
	t.Run("all-warn contains", func(t *testing.T) {
		assertLevel(t, "all-warn", "you-dont-need-lodash-underscore/contains", 1)
	})
	t.Run("all-warn trim", func(t *testing.T) {
		assertLevel(t, "all-warn", "you-dont-need-lodash-underscore/trim", 1)
	})
	t.Run("all every", func(t *testing.T) {
		assertLevel(t, "all", "you-dont-need-lodash-underscore/every", 2)
	})
	t.Run("all keys", func(t *testing.T) {
		assertLevel(t, "all", "you-dont-need-lodash-underscore/keys", 2)
	})
	t.Run("compatible-warn each is absent", func(t *testing.T) {
		assertAbsent(t, "compatible-warn", "you-dont-need-lodash-underscore/each")
	})
	t.Run("compatible-warn last-index-of", func(t *testing.T) {
		assertLevel(t, "compatible-warn", "you-dont-need-lodash-underscore/last-index-of", 1)
	})
	t.Run("compatible for-each", func(t *testing.T) {
		assertLevel(t, "compatible", "you-dont-need-lodash-underscore/for-each", 1)
	})
	t.Run("compatible is-nan", func(t *testing.T) {
		assertLevel(t, "compatible", "you-dont-need-lodash-underscore/is-nan", 2)
	})
}
