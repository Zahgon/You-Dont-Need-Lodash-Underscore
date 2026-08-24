package estree

import (
	"reflect"
	"testing"

	"github.com/ije/esbuild-internal/js_ast"
	"github.com/ije/esbuild-internal/logger"
)

func newTestConverter(source string) *converter {
	return &converter{
		tree:      &js_ast.AST{},
		positions: newPositions(source),
		source:    source,
		seen:      map[uintptr]bool{},
	}
}

func TestExprDefensiveBranches(t *testing.T) {
	c := newTestConverter("")

	t.Run("call with no target and no arguments", func(t *testing.T) {
		node := c.expr(js_ast.Expr{Data: &js_ast.ECall{}}).(*CallExpression)
		if node.Callee != nil {
			t.Errorf("Callee = %v, want nil", node.Callee)
		}
		if len(node.Arguments) != 0 {
			t.Errorf("Arguments has %d entries, want 0", len(node.Arguments))
		}
	})

	t.Run("call argument without data is skipped", func(t *testing.T) {
		data := &js_ast.ECall{Args: []js_ast.Expr{{}, {Data: &js_ast.EIdentifier{}}}}
		node := c.expr(js_ast.Expr{Data: data}).(*CallExpression)
		if len(node.Arguments) != 1 {
			t.Errorf("Arguments has %d entries, want 1", len(node.Arguments))
		}
	})

	t.Run("dot access with no target", func(t *testing.T) {
		node := c.expr(js_ast.Expr{Data: &js_ast.EDot{Name: "prop"}}).(*MemberExpression)
		if node.Object != nil {
			t.Errorf("Object = %v, want nil", node.Object)
		}
		if property := node.Property.(*Identifier); property.Name != "prop" {
			t.Errorf("Property.Name = %q, want prop", property.Name)
		}
		if node.Computed {
			t.Errorf("Computed = true, want false")
		}
	})

	t.Run("index access with no target and no index", func(t *testing.T) {
		node := c.expr(js_ast.Expr{Data: &js_ast.EIndex{}}).(*MemberExpression)
		if node.Object != nil || node.Property != nil {
			t.Errorf("Object/Property = %v/%v, want nil/nil", node.Object, node.Property)
		}
		if !node.Computed {
			t.Errorf("Computed = false, want true")
		}
	})

	t.Run("array item without data is skipped", func(t *testing.T) {
		data := &js_ast.EArray{Items: []js_ast.Expr{{}, {Data: &js_ast.EIdentifier{}}, {}}}
		node := c.expr(js_ast.Expr{Data: data}).(*ArrayExpression)
		if len(node.Elements) != 1 {
			t.Errorf("Elements has %d entries, want 1", len(node.Elements))
		}
	})

	t.Run("assignment with no operands", func(t *testing.T) {
		data := &js_ast.EBinary{Op: js_ast.BinOpAssign}
		node := c.expr(js_ast.Expr{Data: data}).(*AssignmentExpression)
		if node.Operator != "=" {
			t.Errorf("Operator = %q, want =", node.Operator)
		}
		if node.Left != nil || node.Right != nil {
			t.Errorf("Left/Right = %v/%v, want nil/nil", node.Left, node.Right)
		}
	})

	t.Run("import identifier resolves through the symbol table", func(t *testing.T) {
		node := c.expr(js_ast.Expr{Data: &js_ast.EImportIdentifier{}}).(*Identifier)
		if node.Type() != "Identifier" {
			t.Errorf("Type() = %q, want Identifier", node.Type())
		}
	})

	t.Run("unmodelled expression becomes Unknown", func(t *testing.T) {
		node := c.expr(js_ast.Expr{Data: &js_ast.ENumber{Value: 1}}).(*Unknown)
		if node.Kind != "ENumber" {
			t.Errorf("Kind = %q, want ENumber", node.Kind)
		}
	})

	t.Run("expression statement with no value", func(t *testing.T) {
		node := c.stmt(js_ast.Stmt{Data: &js_ast.SExpr{}}).(*ExpressionStatement)
		if node.Expression != nil {
			t.Errorf("Expression = %v, want nil", node.Expression)
		}
	})
}

func TestPropertyKeyComputed(t *testing.T) {
	t.Run("computed key falls through to expr", func(t *testing.T) {
		program := mustParse(t, "var key = 'a'; var o = { [key]: 1 };")
		properties := collect(program, "Property")
		if len(properties) != 1 {
			t.Fatalf("found %d Property nodes, want 1", len(properties))
		}
		property := properties[0].(*Property)
		key, ok := property.Key.(*Identifier)
		if !ok {
			t.Fatalf("Key is %T, want *Identifier from the computed expression", property.Key)
		}
		if key.Name != "key" {
			t.Errorf("Key.Name = %q, want key", key.Name)
		}
	})

	t.Run("non string computed key", func(t *testing.T) {
		c := newTestConverter("")
		node := c.propertyKey(js_ast.Expr{Data: &js_ast.ENumber{Value: 2}})
		unknown, ok := node.(*Unknown)
		if !ok {
			t.Fatalf("propertyKey returned %T, want *Unknown", node)
		}
		if unknown.Kind != "ENumber" {
			t.Errorf("Kind = %q, want ENumber", unknown.Kind)
		}
	})
}

func TestImportedBindingUse(t *testing.T) {
	t.Run("named import used as a callee", func(t *testing.T) {
		program := mustParse(t, "import { map } from 'lodash';\nmap(list, fn);\n")
		call := collect(program, "CallExpression")[0].(*CallExpression)
		callee, ok := call.Callee.(*Identifier)
		if !ok {
			t.Fatalf("Callee is %T, want *Identifier", call.Callee)
		}
		if callee.Name != "map" {
			t.Errorf("Callee.Name = %q, want map", callee.Name)
		}
	})
}

type unexportedHolder struct {
	hidden js_ast.Expr
	Shown  js_ast.Expr
}

func TestDescendSkipsUnexportedFields(t *testing.T) {
	c := newTestConverter("")
	holder := unexportedHolder{
		hidden: js_ast.Expr{Data: &js_ast.EIdentifier{}},
		Shown:  js_ast.Expr{Data: &js_ast.EIdentifier{}},
	}
	var out []Node
	c.descend(reflect.ValueOf(holder), &out, 0)
	if len(out) != 1 {
		t.Fatalf("out has %d entries, want 1 (the unexported field must be skipped)", len(out))
	}
	if out[0].Type() != "Identifier" {
		t.Errorf("out[0].Type() = %q, want Identifier", out[0].Type())
	}
}

func TestUnknownWithoutData(t *testing.T) {
	c := newTestConverter("")
	node := c.unknown("Empty", nil, logger.Loc{}).(*Unknown)
	if node.Kind != "Empty" {
		t.Errorf("Kind = %q, want Empty", node.Kind)
	}
	if len(node.Nodes) != 0 {
		t.Errorf("Nodes has %d entries, want 0", len(node.Nodes))
	}
}
