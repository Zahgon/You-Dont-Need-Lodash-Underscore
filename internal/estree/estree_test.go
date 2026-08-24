package estree

import "testing"

func sameNodes(got, want []Node) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

func TestNodeTypeAndChildren(t *testing.T) {
	a := &Identifier{Name: "a"}
	b := &Identifier{Name: "b"}
	lit := &Literal{Value: "lodash", IsString: true}

	cases := []struct {
		name     string
		node     Node
		wantType string
		want     []Node
	}{
		{"Program", &Program{Body: []Node{a, b}}, "Program", []Node{a, b}},
		{"Program empty", &Program{}, "Program", nil},
		{"Identifier", a, "Identifier", nil},
		{"Literal", lit, "Literal", nil},
		{"CallExpression full", &CallExpression{Callee: a, Arguments: []Node{b, lit}}, "CallExpression", []Node{a, b, lit}},
		{"CallExpression nil callee", &CallExpression{Arguments: []Node{b}}, "CallExpression", []Node{b}},
		{"CallExpression bare", &CallExpression{}, "CallExpression", nil},
		{"MemberExpression full", &MemberExpression{Object: a, Property: b}, "MemberExpression", []Node{a, b}},
		{"MemberExpression object only", &MemberExpression{Object: a}, "MemberExpression", []Node{a}},
		{"MemberExpression property only", &MemberExpression{Property: b}, "MemberExpression", []Node{b}},
		{"MemberExpression empty", &MemberExpression{}, "MemberExpression", nil},
		{"VariableDeclaration", &VariableDeclaration{Kind: "const", Declarations: []Node{a}}, "VariableDeclaration", []Node{a}},
		{"VariableDeclaration empty", &VariableDeclaration{}, "VariableDeclaration", nil},
		{"VariableDeclarator full", &VariableDeclarator{ID: a, Init: b}, "VariableDeclarator", []Node{a, b}},
		{"VariableDeclarator id only", &VariableDeclarator{ID: a}, "VariableDeclarator", []Node{a}},
		{"VariableDeclarator init only", &VariableDeclarator{Init: b}, "VariableDeclarator", []Node{b}},
		{"VariableDeclarator empty", &VariableDeclarator{}, "VariableDeclarator", nil},
		{"AssignmentExpression full", &AssignmentExpression{Operator: "=", Left: a, Right: b}, "AssignmentExpression", []Node{a, b}},
		{"AssignmentExpression left only", &AssignmentExpression{Left: a}, "AssignmentExpression", []Node{a}},
		{"AssignmentExpression right only", &AssignmentExpression{Right: b}, "AssignmentExpression", []Node{b}},
		{"AssignmentExpression empty", &AssignmentExpression{}, "AssignmentExpression", nil},
		{"ObjectPattern", &ObjectPattern{Properties: []Node{a, b}}, "ObjectPattern", []Node{a, b}},
		{"ObjectPattern empty", &ObjectPattern{}, "ObjectPattern", nil},
		{"ObjectExpression", &ObjectExpression{Properties: []Node{b}}, "ObjectExpression", []Node{b}},
		{"ObjectExpression empty", &ObjectExpression{}, "ObjectExpression", nil},
		{"Property full", &Property{Key: a, Value: b}, "Property", []Node{a, b}},
		{"Property key only", &Property{Key: a}, "Property", []Node{a}},
		{"Property value only", &Property{Value: b}, "Property", []Node{b}},
		{"Property empty", &Property{}, "Property", nil},
		{"RestElement", &RestElement{Argument: a}, "RestElement", []Node{a}},
		{"RestElement nil argument", &RestElement{}, "RestElement", nil},
		{"ImportDeclaration full", &ImportDeclaration{Specifiers: []Node{a}, Source: lit}, "ImportDeclaration", []Node{a, lit}},
		{"ImportDeclaration no source", &ImportDeclaration{Specifiers: []Node{a}}, "ImportDeclaration", []Node{a}},
		{"ImportSpecifier full", &ImportSpecifier{Imported: a, Local: b}, "ImportSpecifier", []Node{a, b}},
		{"ImportSpecifier imported only", &ImportSpecifier{Imported: a}, "ImportSpecifier", []Node{a}},
		{"ImportSpecifier local only", &ImportSpecifier{Local: b}, "ImportSpecifier", []Node{b}},
		{"ImportSpecifier empty", &ImportSpecifier{}, "ImportSpecifier", nil},
		{"ImportDefaultSpecifier", &ImportDefaultSpecifier{Local: a}, "ImportDefaultSpecifier", []Node{a}},
		{"ImportDefaultSpecifier nil local", &ImportDefaultSpecifier{}, "ImportDefaultSpecifier", nil},
		{"ImportNamespaceSpecifier", &ImportNamespaceSpecifier{Local: a}, "ImportNamespaceSpecifier", []Node{a}},
		{"ImportNamespaceSpecifier nil local", &ImportNamespaceSpecifier{}, "ImportNamespaceSpecifier", nil},
		{"ExpressionStatement", &ExpressionStatement{Expression: a}, "ExpressionStatement", []Node{a}},
		{"ExpressionStatement nil expression", &ExpressionStatement{}, "ExpressionStatement", nil},
		{"ArrayExpression", &ArrayExpression{Elements: []Node{a, b}}, "ArrayExpression", []Node{a, b}},
		{"ArrayExpression empty", &ArrayExpression{}, "ArrayExpression", nil},
		{"Unknown", &Unknown{Kind: "SIf", Nodes: []Node{a}}, "Unknown", []Node{a}},
		{"Unknown empty", &Unknown{}, "Unknown", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := testCase.node.Type(); got != testCase.wantType {
				t.Errorf("Type() = %q, want %q", got, testCase.wantType)
			}
			if got := testCase.node.Children(); !sameNodes(got, testCase.want) {
				t.Errorf("Children() = %v (len %d), want %v (len %d)",
					got, len(got), testCase.want, len(testCase.want))
			}
		})
	}
}

func TestBaseAccessors(t *testing.T) {
	t.Run("defaults to zero", func(t *testing.T) {
		node := &Identifier{Name: "x"}
		if node.Line() != 0 || node.Column() != 0 {
			t.Errorf("Line/Column = %d/%d, want 0/0", node.Line(), node.Column())
		}
		if node.Parent() != nil {
			t.Errorf("Parent() = %v, want nil", node.Parent())
		}
	})

	t.Run("setLoc and setParent", func(t *testing.T) {
		parent := &Program{}
		node := &Identifier{Name: "x"}
		node.setLoc(7, 13)
		node.setParent(parent)
		if node.Line() != 7 {
			t.Errorf("Line() = %d, want 7", node.Line())
		}
		if node.Column() != 13 {
			t.Errorf("Column() = %d, want 13", node.Column())
		}
		if node.Parent() != Node(parent) {
			t.Errorf("Parent() = %v, want %v", node.Parent(), parent)
		}
	})
}

func TestWalk(t *testing.T) {
	t.Run("nil node returns immediately", func(t *testing.T) {
		visits := 0
		Walk(nil, func(Node) { visits++ })
		if visits != 0 {
			t.Errorf("visits = %d, want 0", visits)
		}
	})

	t.Run("single node", func(t *testing.T) {
		var seen []string
		Walk(&Identifier{Name: "only"}, func(n Node) { seen = append(seen, n.Type()) })
		if len(seen) != 1 || seen[0] != "Identifier" {
			t.Errorf("seen = %v, want [Identifier]", seen)
		}
	})

	t.Run("document order over a nested tree", func(t *testing.T) {
		program := &Program{Body: []Node{
			&ExpressionStatement{Expression: &CallExpression{
				Callee:    &MemberExpression{Object: &Identifier{Name: "_"}, Property: &Identifier{Name: "map"}},
				Arguments: []Node{&ArrayExpression{Elements: []Node{&Literal{Value: "1"}}}},
			}},
		}}
		var seen []string
		Walk(program, func(n Node) { seen = append(seen, n.Type()) })
		want := []string{
			"Program", "ExpressionStatement", "CallExpression", "MemberExpression",
			"Identifier", "Identifier", "ArrayExpression", "Literal",
		}
		if len(seen) != len(want) {
			t.Fatalf("seen = %v (len %d), want %v (len %d)", seen, len(seen), want, len(want))
		}
		for index := range want {
			if seen[index] != want[index] {
				t.Errorf("seen[%d] = %q, want %q", index, seen[index], want[index])
			}
		}
	})

	t.Run("nil child in slice is skipped", func(t *testing.T) {
		program := &Program{Body: []Node{nil, &Identifier{Name: "after"}}}
		var seen []string
		Walk(program, func(n Node) { seen = append(seen, n.Type()) })
		want := []string{"Program", "Identifier"}
		if len(seen) != len(want) || seen[0] != want[0] || seen[1] != want[1] {
			t.Errorf("seen = %v, want %v", seen, want)
		}
	})
}
