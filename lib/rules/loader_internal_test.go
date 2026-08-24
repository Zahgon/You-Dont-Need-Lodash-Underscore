package rules

import (
	"strings"
	"testing"
)

func TestLoadDefinitionsValidDocument(t *testing.T) {
	document := []byte(`{"alpha":{"compatible":true,"alternative":"A","ruleName":"a-rule"},"beta":{"compatible":false,"alternative":"B"}}`)
	entries, err := loadDefinitions(document)
	if err != nil {
		t.Fatalf("loadDefinitions returned %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("loadDefinitions returned %d entries, want 2", len(entries))
	}
	if entries[0].Key != "alpha" || entries[1].Key != "beta" {
		t.Errorf("keys = %q, %q, want alpha, beta (document order)", entries[0].Key, entries[1].Key)
	}
	if !entries[0].Definition.Compatible {
		t.Errorf("alpha Compatible = false, want true")
	}
	if entries[0].Definition.RuleName != "a-rule" {
		t.Errorf("alpha RuleName = %q, want a-rule", entries[0].Definition.RuleName)
	}
	if entries[1].Definition.Compatible {
		t.Errorf("beta Compatible = true, want false")
	}
	if entries[1].Definition.RuleName != "" {
		t.Errorf("beta RuleName = %q, want the empty string", entries[1].Definition.RuleName)
	}
}

func TestLoadDefinitionsEmptyObject(t *testing.T) {
	entries, err := loadDefinitions([]byte(`{}`))
	if err != nil {
		t.Fatalf("loadDefinitions returned %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("entries = %v, want none", entries)
	}
}

func TestLoadDefinitionsErrors(t *testing.T) {
	cases := []struct {
		name     string
		document string
		wantText string
	}{
		{"empty document", ``, "EOF"},
		{"array instead of object", `[]`, "expected an object"},
		{"string instead of object", `"text"`, "expected an object"},
		{"value of the wrong type", `{"alpha":1}`, `key "alpha"`},
		{"value is an array", `{"alpha":[]}`, `key "alpha"`},
		{"unterminated object", `{"alpha":{"compatible":true}`, "unexpected end of JSON input"},
		{"truncated after the brace", `{`, "unexpected end of JSON input"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			entries, err := loadDefinitions([]byte(testCase.document))
			if err == nil {
				t.Fatalf("loadDefinitions(%q) returned %v and no error, want an error",
					testCase.document, entries)
			}
			if entries != nil {
				t.Errorf("entries = %v, want nil alongside the error", entries)
			}
			if !strings.Contains(err.Error(), testCase.wantText) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), testCase.wantText)
			}
		})
	}
}

func TestMustLoadDefinitionsMatchesTheEmbeddedDocument(t *testing.T) {
	entries := mustLoadDefinitions()
	if len(entries) != len(Definitions) {
		t.Fatalf("mustLoadDefinitions returned %d entries, want %d", len(entries), len(Definitions))
	}
	for index := range entries {
		if entries[index].Key != Definitions[index].Key {
			t.Fatalf("entry %d key = %q, want %q", index, entries[index].Key, Definitions[index].Key)
		}
	}
}
