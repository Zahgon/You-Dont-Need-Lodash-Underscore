package rules

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
)

// definitionsJSON is lib/rules/rules.json, carried across from the original
// repository byte for byte. In JavaScript it was loaded with require, which
// preserves the order the keys appear in the document; that order is load
// bearing, because it determines rule registration order and the order of the
// keys in every generated configuration.
//
//go:embed rules.json
var definitionsJSON []byte

// Definition describes one lodash or underscore method that has a native
// equivalent.
type Definition struct {
	// Compatible reports whether the native alternative behaves identically to
	// the lodash or underscore version.
	Compatible bool `json:"compatible"`
	// Alternative is the native construct to use instead.
	Alternative string `json:"alternative"`
	// RuleName overrides the name derived from the method name.
	RuleName string `json:"ruleName"`
}

// Entry is a definition together with the method name that keys it.
type Entry struct {
	Key        string
	Definition Definition
}

// Definitions holds every entry of rules.json in document order.
var Definitions = mustLoadDefinitions()

// Lookup returns the definition for a method name.
func Lookup(key string) (Definition, bool) {
	for _, entry := range Definitions {
		if entry.Key == key {
			return entry.Definition, true
		}
	}
	return Definition{}, false
}

// mustLoadDefinitions decodes rules.json while preserving key order. The
// standard map decoding would lose it, so the document is streamed token by
// token instead.
func mustLoadDefinitions() []Entry {
	entries, err := loadDefinitions(definitionsJSON)
	if err != nil {
		panic(fmt.Sprintf("lib/rules: cannot load rules.json: %v", err))
	}
	return entries
}

func loadDefinitions(document []byte) ([]Entry, error) {
	decoder := json.NewDecoder(bytes.NewReader(document))
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		return nil, fmt.Errorf("expected an object, got %v", token)
	}

	var entries []Entry
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyToken.(string)
		if !ok {
			return nil, fmt.Errorf("expected a string key, got %v", keyToken)
		}
		var definition Definition
		if err := decoder.Decode(&definition); err != nil {
			return nil, fmt.Errorf("key %q: %w", key, err)
		}
		entries = append(entries, Entry{Key: key, Definition: definition})
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	return entries, nil
}
