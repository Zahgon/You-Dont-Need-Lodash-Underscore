// Package estree defines the subset of the ESTree specification that the
// translated lint rules operate on, together with parent pointers and 1-based
// source positions.
//
// The original JavaScript plugin received this tree from ESLint, which builds
// it with espree and assigns a parent pointer to every node while traversing.
// Go has no equivalent, so the tree is reconstructed here from the AST produced
// by the JavaScript parser this repository depends on. Keeping the node names
// and field names identical to ESTree is deliberate: the rule implementation in
// rules/all.go is a line-for-line translation of lib/rules/all.js and reads
// exactly the same properties.
//
// Node kinds the rules never inspect are represented by Unknown so that
// traversal stays complete without modelling the whole grammar.
package estree

// Node is an ESTree node.
type Node interface {
	// Type is the ESTree "type" discriminator.
	Type() string
	// Line is the 1-based line of the node start.
	Line() int
	// Column is the 1-based column of the node start, measured in UTF-16 code
	// units, matching the columns ESLint reports.
	Column() int
	// Parent is the enclosing node, or nil for the Program node. It mirrors the
	// node.parent property ESLint attaches during traversal.
	Parent() Node
	// Children are the child nodes in document order.
	Children() []Node

	setParent(Node)
	setLoc(line, column int)
}

type base struct {
	line   int
	column int
	parent Node
}

func (b *base) Line() int            { return b.line }
func (b *base) Column() int          { return b.column }
func (b *base) Parent() Node         { return b.parent }
func (b *base) setParent(n Node)     { b.parent = n }
func (b *base) setLoc(line, col int) { b.line, b.column = line, col }

// Program is the root node.
type Program struct {
	base
	Body       []Node
	SourceType string
}

func (*Program) Type() string       { return "Program" }
func (n *Program) Children() []Node { return n.Body }

// Identifier is a bare name.
type Identifier struct {
	base
	Name string
}

func (*Identifier) Type() string       { return "Identifier" }
func (n *Identifier) Children() []Node { return nil }

// Literal is a primitive literal. IsString distinguishes a string literal from
// every other literal kind, because the rules compare the value of a literal
// against module specifiers and a non-string literal must never match.
type Literal struct {
	base
	Value    string
	IsString bool
}

func (*Literal) Type() string       { return "Literal" }
func (n *Literal) Children() []Node { return nil }

// CallExpression is a function call.
type CallExpression struct {
	base
	Callee    Node
	Arguments []Node
}

func (*CallExpression) Type() string { return "CallExpression" }
func (n *CallExpression) Children() []Node {
	children := make([]Node, 0, 1+len(n.Arguments))
	if n.Callee != nil {
		children = append(children, n.Callee)
	}
	children = append(children, n.Arguments...)
	return children
}

// MemberExpression is a property access.
type MemberExpression struct {
	base
	Object   Node
	Property Node
	Computed bool
}

func (*MemberExpression) Type() string { return "MemberExpression" }
func (n *MemberExpression) Children() []Node {
	var children []Node
	if n.Object != nil {
		children = append(children, n.Object)
	}
	if n.Property != nil {
		children = append(children, n.Property)
	}
	return children
}

// VariableDeclaration is a var, let or const statement.
type VariableDeclaration struct {
	base
	Kind         string
	Declarations []Node
}

func (*VariableDeclaration) Type() string       { return "VariableDeclaration" }
func (n *VariableDeclaration) Children() []Node { return n.Declarations }

// VariableDeclarator is a single binding within a VariableDeclaration.
type VariableDeclarator struct {
	base
	ID   Node
	Init Node
}

func (*VariableDeclarator) Type() string { return "VariableDeclarator" }
func (n *VariableDeclarator) Children() []Node {
	var children []Node
	if n.ID != nil {
		children = append(children, n.ID)
	}
	if n.Init != nil {
		children = append(children, n.Init)
	}
	return children
}

// AssignmentExpression is an assignment.
type AssignmentExpression struct {
	base
	Operator string
	Left     Node
	Right    Node
}

func (*AssignmentExpression) Type() string { return "AssignmentExpression" }
func (n *AssignmentExpression) Children() []Node {
	var children []Node
	if n.Left != nil {
		children = append(children, n.Left)
	}
	if n.Right != nil {
		children = append(children, n.Right)
	}
	return children
}

// ObjectPattern is a destructuring pattern.
type ObjectPattern struct {
	base
	Properties []Node
}

func (*ObjectPattern) Type() string       { return "ObjectPattern" }
func (n *ObjectPattern) Children() []Node { return n.Properties }

// ObjectExpression is an object literal.
type ObjectExpression struct {
	base
	Properties []Node
}

func (*ObjectExpression) Type() string       { return "ObjectExpression" }
func (n *ObjectExpression) Children() []Node { return n.Properties }

// Property is a member of an ObjectPattern or ObjectExpression.
//
// Key is an Identifier when the source used a bare name and a Literal when the
// source used a quoted key. The rules read property.key.name, which in
// JavaScript is undefined for a Literal, so that distinction changes behaviour
// and is preserved here rather than flattened.
type Property struct {
	base
	Key       Node
	Value     Node
	Shorthand bool
	Computed  bool
}

func (*Property) Type() string { return "Property" }
func (n *Property) Children() []Node {
	var children []Node
	if n.Key != nil {
		children = append(children, n.Key)
	}
	if n.Value != nil {
		children = append(children, n.Value)
	}
	return children
}

// RestElement is a rest binding inside a pattern. It has no Key, which is why
// the rules must not assume every pattern property exposes one.
type RestElement struct {
	base
	Argument Node
}

func (*RestElement) Type() string { return "RestElement" }
func (n *RestElement) Children() []Node {
	if n.Argument == nil {
		return nil
	}
	return []Node{n.Argument}
}

// ImportDeclaration is an ES module import statement.
type ImportDeclaration struct {
	base
	Specifiers []Node
	Source     Node
}

func (*ImportDeclaration) Type() string { return "ImportDeclaration" }
func (n *ImportDeclaration) Children() []Node {
	children := make([]Node, 0, len(n.Specifiers)+1)
	children = append(children, n.Specifiers...)
	if n.Source != nil {
		children = append(children, n.Source)
	}
	return children
}

// ImportSpecifier is a named import binding.
type ImportSpecifier struct {
	base
	Imported Node
	Local    Node
}

func (*ImportSpecifier) Type() string { return "ImportSpecifier" }
func (n *ImportSpecifier) Children() []Node {
	var children []Node
	if n.Imported != nil {
		children = append(children, n.Imported)
	}
	if n.Local != nil {
		children = append(children, n.Local)
	}
	return children
}

// ImportDefaultSpecifier is a default import binding. It is a distinct type so
// that the rules' check for the ImportSpecifier type excludes it.
type ImportDefaultSpecifier struct {
	base
	Local Node
}

func (*ImportDefaultSpecifier) Type() string { return "ImportDefaultSpecifier" }
func (n *ImportDefaultSpecifier) Children() []Node {
	if n.Local == nil {
		return nil
	}
	return []Node{n.Local}
}

// ImportNamespaceSpecifier is a namespace import binding. It is likewise a
// distinct type so that the ImportSpecifier check excludes it.
type ImportNamespaceSpecifier struct {
	base
	Local Node
}

func (*ImportNamespaceSpecifier) Type() string { return "ImportNamespaceSpecifier" }
func (n *ImportNamespaceSpecifier) Children() []Node {
	if n.Local == nil {
		return nil
	}
	return []Node{n.Local}
}

// ExpressionStatement is an expression used as a statement.
type ExpressionStatement struct {
	base
	Expression Node
}

func (*ExpressionStatement) Type() string { return "ExpressionStatement" }
func (n *ExpressionStatement) Children() []Node {
	if n.Expression == nil {
		return nil
	}
	return []Node{n.Expression}
}

// ArrayExpression is an array literal.
type ArrayExpression struct {
	base
	Elements []Node
}

func (*ArrayExpression) Type() string       { return "ArrayExpression" }
func (n *ArrayExpression) Children() []Node { return n.Elements }

// Unknown stands for any node kind the rules never inspect. Keeping it in the
// tree preserves traversal completeness and the parent chain without modelling
// the entire grammar. Kind carries the underlying construct for diagnostics.
type Unknown struct {
	base
	Kind  string
	Nodes []Node
}

func (*Unknown) Type() string       { return "Unknown" }
func (n *Unknown) Children() []Node { return n.Nodes }

// Walk visits node and every descendant in document order.
func Walk(node Node, visit func(Node)) {
	if node == nil {
		return
	}
	visit(node)
	for _, child := range node.Children() {
		Walk(child, visit)
	}
}
