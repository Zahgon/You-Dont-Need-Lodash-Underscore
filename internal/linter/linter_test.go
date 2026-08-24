package linter

import (
	"testing"

	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/estree"
)

// reportEveryCall builds a rule that reports message on every call expression.
func reportEveryCall(message string) Rule {
	return Rule{Create: func(context Context) Visitor {
		return Visitor{CallExpression: func(node *estree.CallExpression) {
			context.Report(node, message)
		}}
	}}
}

// reportEveryImport builds a rule that reports message on every import.
func reportEveryImport(message string) Rule {
	return Rule{Create: func(context Context) Visitor {
		return Visitor{ImportDeclaration: func(node *estree.ImportDeclaration) {
			context.Report(node, message)
		}}
	}}
}

func positions(messages []Message) [][2]int {
	out := make([][2]int, 0, len(messages))
	for _, message := range messages {
		out = append(out, [2]int{message.Line, message.Column})
	}
	return out
}

func TestVerifyParseError(t *testing.T) {
	messages, err := Verify("broken.js", "const = ;", []NamedRule{{ID: "any", Rule: reportEveryCall("x")}})
	if err == nil {
		t.Fatalf("Verify returned %v and no error, want a parse error", messages)
	}
	if messages != nil {
		t.Errorf("messages = %v, want nil", messages)
	}
}

func TestVerifyNoRules(t *testing.T) {
	messages, err := Verify("input.js", "f();", nil)
	if err != nil {
		t.Fatalf("Verify returned %v", err)
	}
	if len(messages) != 0 {
		t.Errorf("messages = %v, want none", messages)
	}
}

func TestVerifyNilHandlers(t *testing.T) {
	t.Run("rule with an entirely empty visitor", func(t *testing.T) {
		empty := Rule{Create: func(Context) Visitor { return Visitor{} }}
		messages, err := Verify("input.js", "import 'lodash';\nf();\n", []NamedRule{{ID: "empty", Rule: empty}})
		if err != nil {
			t.Fatalf("Verify returned %v", err)
		}
		if len(messages) != 0 {
			t.Errorf("messages = %v, want none", messages)
		}
	})

	t.Run("call handler present, import handler nil", func(t *testing.T) {
		rules := []NamedRule{{ID: "calls", Rule: reportEveryCall("call")}}
		messages, err := Verify("input.js", "import 'lodash';\nf();\n", rules)
		if err != nil {
			t.Fatalf("Verify returned %v", err)
		}
		if len(messages) != 1 {
			t.Fatalf("messages = %v, want exactly one", messages)
		}
		if messages[0].Message != "call" {
			t.Errorf("Message = %q, want %q", messages[0].Message, "call")
		}
	})

	t.Run("import handler present, call handler nil", func(t *testing.T) {
		rules := []NamedRule{{ID: "imports", Rule: reportEveryImport("import")}}
		messages, err := Verify("input.js", "import 'lodash';\nf();\n", rules)
		if err != nil {
			t.Fatalf("Verify returned %v", err)
		}
		if len(messages) != 1 {
			t.Fatalf("messages = %v, want exactly one", messages)
		}
		if messages[0].Line != 1 || messages[0].Column != 1 {
			t.Errorf("position = %d:%d, want 1:1", messages[0].Line, messages[0].Column)
		}
	})
}

func TestVerifyMultipleRulesOnOneNode(t *testing.T) {
	rules := []NamedRule{
		{ID: "first", Rule: reportEveryCall("alpha")},
		{ID: "second", Rule: reportEveryCall("beta")},
		{ID: "third", Rule: reportEveryCall("gamma")},
	}
	messages, err := Verify("input.js", "f();\n", rules)
	if err != nil {
		t.Fatalf("Verify returned %v", err)
	}
	if len(messages) != 3 {
		t.Fatalf("messages = %v, want three", messages)
	}
	wantIDs := []string{"first", "second", "third"}
	wantTexts := []string{"alpha", "beta", "gamma"}
	for index := range wantIDs {
		if messages[index].RuleID != wantIDs[index] {
			t.Errorf("messages[%d].RuleID = %q, want %q", index, messages[index].RuleID, wantIDs[index])
		}
		if messages[index].Message != wantTexts[index] {
			t.Errorf("messages[%d].Message = %q, want %q", index, messages[index].Message, wantTexts[index])
		}
	}
}

func TestVerifyOrdersByLineThenColumn(t *testing.T) {
	// The reporter walks the tree in document order, so a rule that reports the
	// outer call first produces messages that are already ordered. Reporting the
	// arguments of a call instead produces them out of order, which the sort
	// must repair.
	reportArgumentsFirst := Rule{Create: func(context Context) Visitor {
		return Visitor{CallExpression: func(node *estree.CallExpression) {
			for index := len(node.Arguments) - 1; index >= 0; index-- {
				context.Report(node.Arguments[index], "argument")
			}
			context.Report(node, "call")
		}}
	}}

	source := "outer(\n  inner(),\n  other()\n);\n"
	messages, err := Verify("input.js", source, []NamedRule{{ID: "r", Rule: reportArgumentsFirst}})
	if err != nil {
		t.Fatalf("Verify returned %v", err)
	}

	got := positions(messages)
	for index := 1; index < len(got); index++ {
		previous, current := got[index-1], got[index]
		if previous[0] > current[0] || (previous[0] == current[0] && previous[1] > current[1]) {
			t.Fatalf("messages are not ordered by position: %v", got)
		}
	}
	if len(got) == 0 {
		t.Fatalf("no messages were produced")
	}
	if got[0][0] != 1 || got[0][1] != 1 {
		t.Errorf("first message position = %v, want [1 1]", got[0])
	}
}

func TestVerifyTiesKeepRegistrationOrder(t *testing.T) {
	rules := []NamedRule{
		{ID: "a", Rule: reportEveryCall("one")},
		{ID: "b", Rule: reportEveryCall("two")},
	}
	messages, err := Verify("input.js", "f();\ng();\n", rules)
	if err != nil {
		t.Fatalf("Verify returned %v", err)
	}
	if len(messages) != 4 {
		t.Fatalf("messages = %v, want four", messages)
	}
	want := []struct {
		line int
		id   string
	}{{1, "a"}, {1, "b"}, {2, "a"}, {2, "b"}}
	for index, expected := range want {
		if messages[index].Line != expected.line || messages[index].RuleID != expected.id {
			t.Errorf("messages[%d] = line %d rule %q, want line %d rule %q",
				index, messages[index].Line, messages[index].RuleID, expected.line, expected.id)
		}
	}
}

func TestReporterCapturesNodePosition(t *testing.T) {
	program, err := estree.Parse("input.js", "\n\n\nvar v = target();\n")
	if err != nil {
		t.Fatalf("Parse returned %v", err)
	}
	var node estree.Node
	estree.Walk(program, func(candidate estree.Node) {
		if node == nil && candidate.Type() == "CallExpression" {
			node = candidate
		}
	})
	if node == nil {
		t.Fatalf("no CallExpression was parsed out of the source")
	}

	var captured []Message
	r := &reporter{ruleID: "rule", messages: &captured}
	r.Report(node, "text")

	if len(captured) != 1 {
		t.Fatalf("captured %d messages, want 1", len(captured))
	}
	want := Message{RuleID: "rule", Line: 4, Column: 9, Message: "text"}
	if captured[0] != want {
		t.Errorf("captured[0] = %+v, want %+v", captured[0], want)
	}
}
