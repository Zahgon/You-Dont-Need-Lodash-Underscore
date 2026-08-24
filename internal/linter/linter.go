// Package linter provides the rule-running machinery that ESLint supplied to
// the original JavaScript plugin.
//
// The plugin itself never contained a linter: it exported rule objects and
// ESLint parsed the source, walked the tree and collected the reports. A Go
// translation has no such host, so the small part of ESLint the plugin actually
// relies on is reproduced here: parse the source into an ESTree tree, walk it
// once in document order, dispatch each node to every registered rule, and
// return the collected messages ordered by position.
package linter

import (
	"sort"

	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/estree"
)

// Context is the object a rule receives from ESLint. Only the report method is
// used by this plugin.
type Context interface {
	Report(node estree.Node, message string)
}

// Visitor is the object a rule's create method returns. Its fields are the node
// types the plugin listens for.
type Visitor struct {
	CallExpression    func(node *estree.CallExpression)
	ImportDeclaration func(node *estree.ImportDeclaration)
}

// Rule is an ESLint rule.
type Rule struct {
	Create func(context Context) Visitor
}

// NamedRule pairs a rule with the identifier it is reported under.
type NamedRule struct {
	ID   string
	Rule Rule
}

// Message is a single reported problem.
type Message struct {
	RuleID  string
	Line    int
	Column  int
	Message string
}

type reporter struct {
	ruleID   string
	messages *[]Message
}

func (r *reporter) Report(node estree.Node, message string) {
	*r.messages = append(*r.messages, Message{
		RuleID:  r.ruleID,
		Line:    node.Line(),
		Column:  node.Column(),
		Message: message,
	})
}

// Verify parses source and runs every rule over it.
//
// Rules are dispatched in registration order at each node, and the resulting
// messages are then ordered by line and column with ties left in the order they
// were produced, which is how ESLint orders the messages it returns.
func Verify(filename, source string, rules []NamedRule) ([]Message, error) {
	program, err := estree.Parse(filename, source)
	if err != nil {
		return nil, err
	}

	var messages []Message
	visitors := make([]Visitor, 0, len(rules))
	for _, rule := range rules {
		context := &reporter{ruleID: rule.ID, messages: &messages}
		visitors = append(visitors, rule.Rule.Create(context))
	}

	estree.Walk(program, func(node estree.Node) {
		switch typed := node.(type) {
		case *estree.CallExpression:
			for _, visitor := range visitors {
				if visitor.CallExpression != nil {
					visitor.CallExpression(typed)
				}
			}
		case *estree.ImportDeclaration:
			for _, visitor := range visitors {
				if visitor.ImportDeclaration != nil {
					visitor.ImportDeclaration(typed)
				}
			}
		}
	})

	sort.SliceStable(messages, func(i, j int) bool {
		if messages[i].Line != messages[j].Line {
			return messages[i].Line < messages[j].Line
		}
		return messages[i].Column < messages[j].Column
	})
	return messages, nil
}
