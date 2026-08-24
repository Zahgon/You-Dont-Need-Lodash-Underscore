// Package plugin is the translation of index.js, the entry point of the
// original repository. It exports the lint rules and the four shareable
// configurations built from them.
package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/kebabcase"
	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/linter"
	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/lib/rules"
)

// Name is the plugin name every rule is namespaced under.
const Name = "you-dont-need-lodash-underscore"

// Severity levels, matching the numeric levels ESLint configurations use.
const (
	Warn  = 1
	Error = 2
)

// Rules is every rule the plugin exports, in declaration order.
var Rules = rules.All

// RuleLevels maps a namespaced rule name to a severity. Key order is preserved
// because the JavaScript object it replaces was insertion ordered and that
// order is visible in the serialised configurations.
type RuleLevels struct {
	keys   []string
	levels map[string]int
}

func newRuleLevels() *RuleLevels {
	return &RuleLevels{levels: map[string]int{}}
}

func (r *RuleLevels) set(name string, level int) {
	if _, exists := r.levels[name]; !exists {
		r.keys = append(r.keys, name)
	}
	r.levels[name] = level
}

// Get returns the severity configured for a rule. The second result reports
// whether the rule is present at all, which is how a configuration that omits a
// rule differs from one that sets it to zero.
func (r *RuleLevels) Get(name string) (int, bool) {
	level, ok := r.levels[name]
	return level, ok
}

// Keys returns the rule names in order.
func (r *RuleLevels) Keys() []string {
	return append([]string(nil), r.keys...)
}

// Len returns the number of configured rules.
func (r *RuleLevels) Len() int { return len(r.keys) }

// MarshalJSON writes the rules in their original order.
func (r *RuleLevels) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteByte('{')
	for index, key := range r.keys {
		if index > 0 {
			buffer.WriteByte(',')
		}
		encoded, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buffer.Write(encoded)
		buffer.WriteByte(':')
		fmt.Fprintf(&buffer, "%d", r.levels[key])
	}
	buffer.WriteByte('}')
	return buffer.Bytes(), nil
}

// Config is a shareable ESLint configuration.
type Config struct {
	Plugins []string    `json:"plugins"`
	Rules   *RuleLevels `json:"rules"`
}

// Configs is the set of shareable configurations, in declaration order.
type Configs struct {
	keys    []string
	configs map[string]Config
}

func (c *Configs) set(name string, config Config) {
	if _, exists := c.configs[name]; !exists {
		c.keys = append(c.keys, name)
	}
	c.configs[name] = config
}

// Get returns a configuration by name.
func (c *Configs) Get(name string) (Config, bool) {
	config, ok := c.configs[name]
	return config, ok
}

// Keys returns the configuration names in order.
func (c *Configs) Keys() []string {
	return append([]string(nil), c.keys...)
}

// MarshalJSON writes the configurations in their original order.
func (c *Configs) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteByte('{')
	for index, key := range c.keys {
		if index > 0 {
			buffer.WriteByte(',')
		}
		encoded, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buffer.Write(encoded)
		buffer.WriteByte(':')
		config, err := json.Marshal(c.configs[key])
		if err != nil {
			return nil, err
		}
		buffer.Write(config)
	}
	buffer.WriteByte('}')
	return buffer.Bytes(), nil
}

// ruleNameFor returns the namespaced name a method is configured under.
func ruleNameFor(key string) string {
	definition, ok := rules.Lookup(key)
	if !ok {
		return Name + "/" + kebabcase.KebabCase(key)
	}
	if definition.RuleName != "" {
		return Name + "/" + definition.RuleName
	}
	return Name + "/" + kebabcase.KebabCase(key)
}

// configure builds a rule severity map that sets every listed method to level.
func configure(list []string, level int) *RuleLevels {
	levels := newRuleLevels()
	for _, key := range list {
		levels.set(ruleNameFor(key), level)
	}
	return levels
}

// merge applies the entries of extra on top of base, keeping the position of
// any key base already holds, which is what Object.assign does.
func merge(base, extra *RuleLevels) *RuleLevels {
	for _, key := range extra.keys {
		base.set(key, extra.levels[key])
	}
	return base
}

func methodNames(only func(rules.Definition) bool) []string {
	var list []string
	for _, entry := range rules.Definitions {
		if only == nil || only(entry.Definition) {
			list = append(list, entry.Key)
		}
	}
	return list
}

// ConfigsValue is the set of shareable configurations the plugin exports.
var ConfigsValue = buildConfigs()

func buildConfigs() *Configs {
	all := methodNames(nil)
	compatible := methodNames(func(d rules.Definition) bool { return d.Compatible })
	incompatible := methodNames(func(d rules.Definition) bool { return !d.Compatible })

	configs := &Configs{configs: map[string]Config{}}
	configs.set("all-warn", Config{Plugins: []string{Name}, Rules: configure(all, Warn)})
	configs.set("all", Config{Plugins: []string{Name}, Rules: configure(all, Error)})
	configs.set("compatible-warn", Config{Plugins: []string{Name}, Rules: configure(compatible, Warn)})
	configs.set("compatible", Config{
		Plugins: []string{Name},
		Rules:   merge(configure(compatible, Error), configure(incompatible, Warn)),
	})
	return configs
}

// Verify lints JavaScript source with every rule the plugin exports and returns
// the problems found.
func Verify(filename, source string) ([]linter.Message, error) {
	return linter.Verify(filename, source, Rules)
}
