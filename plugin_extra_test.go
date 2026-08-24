package plugin_test

import (
	"encoding/json"
	"strings"
	"testing"

	plugin "github.com/you-dont-need/You-Dont-Need-Lodash-Underscore"
)

func TestConfigsKeys(t *testing.T) {
	want := []string{"all-warn", "all", "compatible-warn", "compatible"}
	got := plugin.ConfigsValue.Keys()
	if len(got) != len(want) {
		t.Fatalf("Keys() = %v, want %v", got, want)
	}
	for index := range want {
		t.Run(want[index], func(t *testing.T) {
			if got[index] != want[index] {
				t.Errorf("Keys()[%d] = %q, want %q", index, got[index], want[index])
			}
			if _, ok := plugin.ConfigsValue.Get(want[index]); !ok {
				t.Errorf("Get(%q) reports the configuration is absent", want[index])
			}
		})
	}
	t.Run("Keys returns a copy", func(t *testing.T) {
		first := plugin.ConfigsValue.Keys()
		first[0] = "mutated"
		if plugin.ConfigsValue.Keys()[0] != "all-warn" {
			t.Errorf("mutating the returned slice changed the configuration order")
		}
	})
	t.Run("unknown configuration", func(t *testing.T) {
		config, ok := plugin.ConfigsValue.Get("does-not-exist")
		if ok {
			t.Fatalf("Get returned ok = true for an unknown configuration")
		}
		if config.Rules != nil || config.Plugins != nil {
			t.Errorf("config = %+v, want the zero value", config)
		}
	})
}

func TestRuleLevelsAccessors(t *testing.T) {
	config, ok := plugin.ConfigsValue.Get("all")
	if !ok {
		t.Fatalf("configuration \"all\" is not exported")
	}
	levels := config.Rules

	t.Run("Len matches the exported rule count", func(t *testing.T) {
		if levels.Len() != len(plugin.Rules) {
			t.Errorf("Len() = %d, want %d", levels.Len(), len(plugin.Rules))
		}
	})

	t.Run("Keys length matches Len", func(t *testing.T) {
		if len(levels.Keys()) != levels.Len() {
			t.Errorf("len(Keys()) = %d, want %d", len(levels.Keys()), levels.Len())
		}
	})

	t.Run("Keys follow rule declaration order", func(t *testing.T) {
		keys := levels.Keys()
		for index, rule := range plugin.Rules {
			want := plugin.Name + "/" + rule.ID
			if keys[index] != want {
				t.Fatalf("Keys()[%d] = %q, want %q", index, keys[index], want)
			}
		}
	})

	t.Run("Keys returns a copy", func(t *testing.T) {
		first := levels.Keys()
		original := first[0]
		first[0] = "mutated"
		if levels.Keys()[0] != original {
			t.Errorf("mutating the returned slice changed the stored key order")
		}
	})

	t.Run("every key is set to Error", func(t *testing.T) {
		for _, key := range levels.Keys() {
			level, present := levels.Get(key)
			if !present {
				t.Fatalf("Get(%q) reports the rule is absent", key)
			}
			if level != plugin.Error {
				t.Fatalf("Get(%q) = %d, want %d", key, level, plugin.Error)
			}
		}
	})

	t.Run("unknown rule is absent", func(t *testing.T) {
		level, present := levels.Get("you-dont-need-lodash-underscore/not-a-rule")
		if present {
			t.Errorf("Get returned present = true for an unknown rule")
		}
		if level != 0 {
			t.Errorf("level = %d, want 0", level)
		}
	})
}

func TestRuleLevelsMarshalJSON(t *testing.T) {
	config, ok := plugin.ConfigsValue.Get("all-warn")
	if !ok {
		t.Fatalf("configuration \"all-warn\" is not exported")
	}
	encoded, err := json.Marshal(config.Rules)
	if err != nil {
		t.Fatalf("json.Marshal returned %v", err)
	}
	text := string(encoded)

	t.Run("is a JSON object", func(t *testing.T) {
		if !strings.HasPrefix(text, "{") || !strings.HasSuffix(text, "}") {
			t.Fatalf("encoded = %q, want a JSON object", text)
		}
	})

	t.Run("preserves key order", func(t *testing.T) {
		cursor := 0
		for _, key := range config.Rules.Keys() {
			needle := "\"" + key + "\":1"
			index := strings.Index(text[cursor:], needle)
			if index < 0 {
				t.Fatalf("%q does not appear after offset %d in %q", needle, cursor, text)
			}
			cursor += index + len(needle)
		}
	})

	t.Run("round trips to the same levels", func(t *testing.T) {
		var decoded map[string]int
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatalf("json.Unmarshal returned %v", err)
		}
		if len(decoded) != config.Rules.Len() {
			t.Fatalf("decoded %d entries, want %d", len(decoded), config.Rules.Len())
		}
		for key, level := range decoded {
			if level != plugin.Warn {
				t.Errorf("decoded[%q] = %d, want %d", key, level, plugin.Warn)
			}
		}
	})
}

func TestConfigsMarshalJSON(t *testing.T) {
	encoded, err := json.Marshal(plugin.ConfigsValue)
	if err != nil {
		t.Fatalf("json.Marshal returned %v", err)
	}
	text := string(encoded)

	t.Run("preserves configuration order", func(t *testing.T) {
		cursor := 0
		for _, key := range plugin.ConfigsValue.Keys() {
			needle := "\"" + key + "\":{"
			index := strings.Index(text[cursor:], needle)
			if index < 0 {
				t.Fatalf("%q does not appear after offset %d", needle, cursor)
			}
			cursor += index + len(needle)
		}
	})

	t.Run("each configuration carries the plugin name", func(t *testing.T) {
		var decoded map[string]struct {
			Plugins []string       `json:"plugins"`
			Rules   map[string]int `json:"rules"`
		}
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatalf("json.Unmarshal returned %v", err)
		}
		if len(decoded) != len(plugin.ConfigsValue.Keys()) {
			t.Fatalf("decoded %d configurations, want %d",
				len(decoded), len(plugin.ConfigsValue.Keys()))
		}
		for name, config := range decoded {
			if len(config.Plugins) != 1 || config.Plugins[0] != plugin.Name {
				t.Errorf("configs[%q].plugins = %v, want [%q]", name, config.Plugins, plugin.Name)
			}
			if len(config.Rules) == 0 {
				t.Errorf("configs[%q].rules is empty", name)
			}
		}
	})

	t.Run("compatible mixes both severities", func(t *testing.T) {
		config, ok := plugin.ConfigsValue.Get("compatible")
		if !ok {
			t.Fatalf("configuration \"compatible\" is not exported")
		}
		warns, errors := 0, 0
		for _, key := range config.Rules.Keys() {
			level, _ := config.Rules.Get(key)
			switch level {
			case plugin.Warn:
				warns++
			case plugin.Error:
				errors++
			default:
				t.Fatalf("rule %q has severity %d, want 1 or 2", key, level)
			}
		}
		if warns == 0 || errors == 0 {
			t.Errorf("compatible has %d warnings and %d errors, want both to be non-zero", warns, errors)
		}
		if warns+errors != config.Rules.Len() {
			t.Errorf("warns+errors = %d, want %d", warns+errors, config.Rules.Len())
		}
	})
}

func TestVerify(t *testing.T) {
	t.Run("reports a lodash member call", func(t *testing.T) {
		messages, err := plugin.Verify("input.js", "_.map(list, fn);\n")
		if err != nil {
			t.Fatalf("Verify returned %v", err)
		}
		if len(messages) != 1 {
			t.Fatalf("messages = %+v, want exactly one", messages)
		}
		if messages[0].RuleID != "map" {
			t.Errorf("RuleID = %q, want %q", messages[0].RuleID, "map")
		}
		if messages[0].Line != 1 || messages[0].Column != 1 {
			t.Errorf("position = %d:%d, want 1:1", messages[0].Line, messages[0].Column)
		}
		if !strings.Contains(messages[0].Message, "Consider using the native") {
			t.Errorf("Message = %q, want it to suggest a native alternative", messages[0].Message)
		}
	})

	t.Run("reports nothing for clean source", func(t *testing.T) {
		messages, err := plugin.Verify("input.js", "const doubled = list.map(fn);\n")
		if err != nil {
			t.Fatalf("Verify returned %v", err)
		}
		if len(messages) != 0 {
			t.Errorf("messages = %+v, want none", messages)
		}
	})

	t.Run("propagates a parse error", func(t *testing.T) {
		messages, err := plugin.Verify("broken.js", "const = ;")
		if err == nil {
			t.Fatalf("Verify returned %+v and no error, want a parse error", messages)
		}
		if !strings.Contains(err.Error(), "broken.js") {
			t.Errorf("error = %q, want it to mention broken.js", err.Error())
		}
	})
}
