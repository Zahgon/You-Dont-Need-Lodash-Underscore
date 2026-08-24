// Package rules is the translation of lib/rules in the original repository. It
// carries rules.json across unchanged and builds one lint rule per entry, the
// way lib/rules/all.js did.
package rules

import (
	"fmt"
	"strings"

	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/estree"
	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/kebabcase"
	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/linter"
)

// forbiddenLibs are the packages whose named exports the rules object to.
var forbiddenLibs = []string{"lodash", "lodash/fp", "lodash-es"}

func isForbiddenLib(name string) bool {
	for _, lib := range forbiddenLibs {
		if lib == name {
			return true
		}
	}
	return false
}

// getAssignmentLeftHandSide returns the binding target of a declarator or an
// assignment, and nothing for any other construct.
func getAssignmentLeftHandSide(node estree.Node) estree.Node {
	switch typed := node.(type) {
	case *estree.VariableDeclarator:
		return typed.ID
	case *estree.AssignmentExpression:
		return typed.Left
	default:
		return nil
	}
}

// identifierName returns the name of an identifier node. A key that was written
// as a quoted string is a Literal rather than an Identifier and therefore has no
// name, which is why the second result matters.
func identifierName(node estree.Node) (string, bool) {
	identifier, ok := node.(*estree.Identifier)
	if !ok {
		return "", false
	}
	return identifier.Name, true
}

// calleeObjectName resolves the receiver a call is made on. It covers a bare
// call, a call on an identifier, and a call on the result of another call, which
// is what lodash method chaining produces.
func calleeObjectName(callee estree.Node) string {
	if name, ok := identifierName(callee); ok {
		return name
	}
	member, ok := callee.(*estree.MemberExpression)
	if !ok {
		return ""
	}
	if name, ok := identifierName(member.Object); ok {
		return name
	}
	if call, ok := member.Object.(*estree.CallExpression); ok {
		if name, ok := identifierName(call.Callee); ok {
			return name
		}
	}
	return ""
}

// stringLiteralValue returns the value of a string literal. Any other node, and
// any non-string literal, has no value to compare against a module specifier.
func stringLiteralValue(node estree.Node) (string, bool) {
	literal, ok := node.(*estree.Literal)
	if !ok || !literal.IsString {
		return "", false
	}
	return literal.Value, true
}

// build creates the rule for a single rules.json entry.
func build(rule string, definition Definition) linter.Rule {
	alternative := definition.Alternative
	forbiddenImports := map[string]bool{
		"lodash/" + rule:                  true,
		"lodash/fp/" + rule:               true,
		"lodash-es/" + rule:               true,
		"lodash." + strings.ToLower(rule): true,
	}

	return linter.Rule{
		Create: func(context linter.Context) linter.Visitor {
			return linter.Visitor{
				CallExpression: func(node *estree.CallExpression) {
					callee := node.Callee
					objectName := calleeObjectName(callee)

					if objectName == "require" && len(node.Arguments) == 1 {
						requiredModuleName, _ := stringLiteralValue(node.Arguments[0])
						parent := node.Parent()
						if isForbiddenLib(requiredModuleName) {
							leftHandSide := getAssignmentLeftHandSide(parent)
							pattern, ok := leftHandSide.(*estree.ObjectPattern)
							if !ok {
								return
							}
							for _, element := range pattern.Properties {
								property, ok := element.(*estree.Property)
								if !ok {
									// A rest element has no key at all.
									continue
								}
								name, ok := identifierName(property.Key)
								if !ok || name != rule {
									continue
								}
								context.Report(node, fmt.Sprintf(
									"{ %s } = require('%s') detected. Consider using the native %s",
									rule, requiredModuleName, alternative))
							}
						} else if forbiddenImports[requiredModuleName] {
							context.Report(node, fmt.Sprintf(
								"require('%s') detected. Consider using the native %s",
								requiredModuleName, alternative))
						}
						return
					}

					if objectName != "_" && objectName != "lodash" && objectName != "underscore" {
						return
					}
					member, ok := callee.(*estree.MemberExpression)
					if !ok {
						return
					}
					if name, ok := identifierName(member.Property); ok && name == rule {
						context.Report(node, "Consider using the native "+alternative)
					}
				},

				ImportDeclaration: func(node *estree.ImportDeclaration) {
					sourceValue, _ := stringLiteralValue(node.Source)
					if isForbiddenLib(sourceValue) {
						for _, element := range node.Specifiers {
							specifier, ok := element.(*estree.ImportSpecifier)
							if !ok {
								// A default or namespace import binds the module
								// itself, not one of its named exports.
								continue
							}
							name, ok := identifierName(specifier.Imported)
							if !ok || name != rule {
								continue
							}
							context.Report(node, fmt.Sprintf(
								"Import { %s } from '%s' detected. Consider using the native %s",
								rule, sourceValue, alternative))
						}
					} else if forbiddenImports[sourceValue] {
						context.Report(node, fmt.Sprintf(
							"Import from '%s' detected. Consider using the native %s",
							sourceValue, alternative))
					}
				},
			}
		},
	}
}

// RuleName returns the name an entry is published under: the explicit override
// when rules.json carries one, and the kebab-cased method name otherwise.
func RuleName(entry Entry) string {
	if entry.Definition.RuleName != "" {
		return entry.Definition.RuleName
	}
	return kebabcase.KebabCase(entry.Key)
}

// All is every rule the plugin exports, in the order rules.json declares them.
var All = buildAll()

func buildAll() []linter.NamedRule {
	all := make([]linter.NamedRule, 0, len(Definitions))
	for _, entry := range Definitions {
		all = append(all, linter.NamedRule{
			ID:   RuleName(entry),
			Rule: build(entry.Key, entry.Definition),
		})
	}
	return all
}

// Get returns the rule published under name.
func Get(name string) (linter.Rule, bool) {
	for _, rule := range All {
		if rule.ID == name {
			return rule.Rule, true
		}
	}
	return linter.Rule{}, false
}
