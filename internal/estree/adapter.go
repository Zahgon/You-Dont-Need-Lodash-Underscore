package estree

import (
	"fmt"
	"reflect"
	"strings"
	"unicode/utf16"

	"github.com/ije/esbuild-internal/ast"
	"github.com/ije/esbuild-internal/config"
	"github.com/ije/esbuild-internal/js_ast"
	"github.com/ije/esbuild-internal/js_parser"
	"github.com/ije/esbuild-internal/logger"
)

// SyntaxError reports a source text that could not be parsed.
type SyntaxError struct {
	File     string
	Messages []string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("%s: %s", e.File, strings.Join(e.Messages, "; "))
}

// Parse parses JavaScript module source text and returns it as an ESTree tree
// with parent pointers and 1-based positions already assigned.
func Parse(filename, source string) (*Program, error) {
	log := logger.NewDeferLog(logger.DeferLogAll, nil)
	src := logger.Source{
		Index:    0,
		KeyPath:  logger.Path{Text: filename},
		Contents: source,
	}
	tree, ok := js_parser.Parse(log, src, js_parser.OptionsFromConfig(&config.Options{}))
	if !ok {
		var messages []string
		for _, msg := range log.Done() {
			if msg.Kind == logger.Error {
				messages = append(messages, msg.Data.Text)
			}
		}
		if len(messages) == 0 {
			messages = []string{"parse failed"}
		}
		return nil, &SyntaxError{File: filename, Messages: messages}
	}

	conv := &converter{tree: &tree, positions: newPositions(source), source: source}
	program := &Program{SourceType: "module"}
	program.setLoc(1, 1)
	for _, part := range tree.Parts {
		for _, stmt := range part.Stmts {
			if stmt.Data == nil {
				continue
			}
			program.Body = append(program.Body, conv.stmt(stmt))
		}
	}
	wireParents(program)
	return program, nil
}

func wireParents(node Node) {
	for _, child := range node.Children() {
		if child == nil {
			continue
		}
		child.setParent(node)
		wireParents(child)
	}
}

// positions converts byte offsets into 1-based line and column pairs. Columns
// are counted in UTF-16 code units so that they match the columns ESLint
// reports for the same source.
type positions struct {
	source     string
	lineStarts []int32
}

func newPositions(source string) *positions {
	starts := []int32{0}
	for i := 0; i < len(source); i++ {
		if source[i] == '\n' {
			starts = append(starts, int32(i+1))
		}
	}
	return &positions{source: source, lineStarts: starts}
}

func (p *positions) at(offset int32) (line, column int) {
	if offset < 0 {
		offset = 0
	}
	if int(offset) > len(p.source) {
		offset = int32(len(p.source))
	}
	low, high := 0, len(p.lineStarts)-1
	for low < high {
		mid := (low + high + 1) / 2
		if p.lineStarts[mid] <= offset {
			low = mid
		} else {
			high = mid - 1
		}
	}
	units := 0
	for _, r := range p.source[p.lineStarts[low]:offset] {
		units += utf16.RuneLen(r)
	}
	return low + 1, units + 1
}

type converter struct {
	tree      *js_ast.AST
	positions *positions
	source    string
	seen      map[uintptr]bool
}

func (c *converter) place(node Node, loc logger.Loc) Node {
	line, column := c.positions.at(loc.Start)
	node.setLoc(line, column)
	return node
}

func (c *converter) symbolName(ref ast.Ref) string {
	if int(ref.InnerIndex) >= len(c.tree.Symbols) {
		return ""
	}
	return c.tree.Symbols[ref.InnerIndex].OriginalName
}

// stmt converts a statement. Only the kinds the rules inspect are modelled
// precisely; everything else becomes an Unknown node that still carries its
// descendants so traversal reaches every call expression.
func (c *converter) stmt(stmt js_ast.Stmt) Node {
	switch data := stmt.Data.(type) {
	case *js_ast.SImport:
		return c.importDeclaration(data, stmt.Loc)
	case *js_ast.SLocal:
		return c.variableDeclaration(data, stmt.Loc)
	case *js_ast.SExpr:
		node := &ExpressionStatement{}
		if data.Value.Data != nil {
			node.Expression = c.expr(data.Value)
		}
		return c.place(node, stmt.Loc)
	default:
		return c.unknown(kindName(stmt.Data), stmt.Data, stmt.Loc)
	}
}

func (c *converter) importDeclaration(data *js_ast.SImport, loc logger.Loc) Node {
	node := &ImportDeclaration{}
	path := ""
	if int(data.ImportRecordIndex) < len(c.tree.ImportRecords) {
		path = c.tree.ImportRecords[data.ImportRecordIndex].Path.Text
	}
	source := &Literal{Value: path, IsString: true}
	node.Source = c.place(source, loc)

	if data.DefaultName != nil {
		local := &Identifier{Name: c.symbolName(data.DefaultName.Ref)}
		specifier := &ImportDefaultSpecifier{Local: c.place(local, data.DefaultName.Loc)}
		node.Specifiers = append(node.Specifiers, c.place(specifier, data.DefaultName.Loc))
	}
	if data.StarNameLoc != nil {
		local := &Identifier{Name: c.symbolName(data.NamespaceRef)}
		specifier := &ImportNamespaceSpecifier{Local: c.place(local, *data.StarNameLoc)}
		node.Specifiers = append(node.Specifiers, c.place(specifier, *data.StarNameLoc))
	}
	if data.Items != nil {
		for _, item := range *data.Items {
			imported := &Identifier{Name: item.Alias}
			local := &Identifier{Name: item.OriginalName}
			specifier := &ImportSpecifier{
				Imported: c.place(imported, item.AliasLoc),
				Local:    c.place(local, item.Name.Loc),
			}
			node.Specifiers = append(node.Specifiers, c.place(specifier, item.AliasLoc))
		}
	}
	return c.place(node, loc)
}

func (c *converter) variableDeclaration(data *js_ast.SLocal, loc logger.Loc) Node {
	node := &VariableDeclaration{Kind: localKindName(data.Kind)}
	// One statement can bind several declarators. Each becomes its own
	// VariableDeclarator, because the rules resolve a call expression's parent
	// to decide whether it is the initialiser of a destructuring binding.
	for _, decl := range data.Decls {
		declarator := &VariableDeclarator{}
		if decl.Binding.Data != nil {
			declarator.ID = c.binding(decl.Binding)
		}
		if decl.ValueOrNil.Data != nil {
			declarator.Init = c.expr(decl.ValueOrNil)
		}
		node.Declarations = append(node.Declarations, c.place(declarator, decl.Binding.Loc))
	}
	return c.place(node, loc)
}

func localKindName(kind js_ast.LocalKind) string {
	switch kind {
	case js_ast.LocalVar:
		return "var"
	case js_ast.LocalLet:
		return "let"
	case js_ast.LocalConst:
		return "const"
	default:
		return "var"
	}
}

func (c *converter) binding(binding js_ast.Binding) Node {
	switch data := binding.Data.(type) {
	case *js_ast.BIdentifier:
		return c.place(&Identifier{Name: c.symbolName(data.Ref)}, binding.Loc)
	case *js_ast.BObject:
		node := &ObjectPattern{}
		for _, property := range data.Properties {
			node.Properties = append(node.Properties, c.propertyBinding(property))
		}
		return c.place(node, binding.Loc)
	default:
		return c.unknown(kindName(binding.Data), binding.Data, binding.Loc)
	}
}

func (c *converter) propertyBinding(property js_ast.PropertyBinding) Node {
	loc := property.Value.Loc
	if property.Key.Data != nil {
		loc = property.Key.Loc
	}
	// A rest binding carries no key at all. The rules must therefore not assume
	// every pattern property exposes one.
	if property.IsSpread || property.Key.Data == nil {
		node := &RestElement{}
		if property.Value.Data != nil {
			node.Argument = c.binding(property.Value)
		}
		return c.place(node, loc)
	}
	node := &Property{Computed: property.IsComputed, Key: c.propertyKey(property.Key)}
	if property.Value.Data != nil {
		node.Value = c.binding(property.Value)
	}
	if identifier, ok := node.Key.(*Identifier); ok {
		if local, ok := node.Value.(*Identifier); ok && local.Name == identifier.Name {
			node.Shorthand = true
		}
	}
	return c.place(node, loc)
}

// propertyKey rebuilds the distinction the parser discards. A bare key is an
// Identifier and a quoted key is a Literal; the rules read key.name, which is
// absent on a Literal, so collapsing the two would change behaviour.
func (c *converter) propertyKey(key js_ast.Expr) Node {
	if key.Data == nil {
		return nil
	}
	if str, ok := key.Data.(*js_ast.EString); ok {
		value := decodeUTF16(str.Value)
		if c.wasQuoted(key.Loc) {
			return c.place(&Literal{Value: value, IsString: true}, key.Loc)
		}
		return c.place(&Identifier{Name: value}, key.Loc)
	}
	return c.expr(key)
}

func (c *converter) wasQuoted(loc logger.Loc) bool {
	if int(loc.Start) >= len(c.source) {
		return false
	}
	switch c.source[loc.Start] {
	case '\'', '"', '`':
		return true
	default:
		return false
	}
}

func (c *converter) expr(expr js_ast.Expr) Node {
	switch data := expr.Data.(type) {
	case *js_ast.ECall:
		node := &CallExpression{}
		if data.Target.Data != nil {
			node.Callee = c.expr(data.Target)
		}
		for _, arg := range data.Args {
			if arg.Data == nil {
				continue
			}
			node.Arguments = append(node.Arguments, c.expr(arg))
		}
		return c.place(node, expr.Loc)

	case *js_ast.EDot:
		node := &MemberExpression{Computed: false}
		if data.Target.Data != nil {
			node.Object = c.expr(data.Target)
		}
		node.Property = c.place(&Identifier{Name: data.Name}, data.NameLoc)
		return c.place(node, expr.Loc)

	case *js_ast.EIndex:
		node := &MemberExpression{Computed: true}
		if data.Target.Data != nil {
			node.Object = c.expr(data.Target)
		}
		if data.Index.Data != nil {
			node.Property = c.expr(data.Index)
		}
		return c.place(node, expr.Loc)

	case *js_ast.EIdentifier:
		return c.place(&Identifier{Name: c.symbolName(data.Ref)}, expr.Loc)

	case *js_ast.EImportIdentifier:
		return c.place(&Identifier{Name: c.symbolName(data.Ref)}, expr.Loc)

	case *js_ast.EString:
		return c.place(&Literal{Value: decodeUTF16(data.Value), IsString: true}, expr.Loc)

	case *js_ast.EObject:
		node := &ObjectExpression{}
		for _, property := range data.Properties {
			node.Properties = append(node.Properties, c.objectProperty(property))
		}
		return c.place(node, expr.Loc)

	case *js_ast.EArray:
		node := &ArrayExpression{}
		for _, item := range data.Items {
			if item.Data == nil {
				continue
			}
			node.Elements = append(node.Elements, c.expr(item))
		}
		return c.place(node, expr.Loc)

	case *js_ast.EBinary:
		if data.Op == js_ast.BinOpAssign {
			node := &AssignmentExpression{Operator: "="}
			if data.Left.Data != nil {
				node.Left = c.assignmentTarget(data.Left)
			}
			if data.Right.Data != nil {
				node.Right = c.expr(data.Right)
			}
			return c.place(node, expr.Loc)
		}
		return c.unknown("BinaryExpression", data, expr.Loc)

	default:
		return c.unknown(kindName(expr.Data), expr.Data, expr.Loc)
	}
}

// assignmentTarget reinterprets an object literal used on the left of an
// assignment as a destructuring pattern, which is what ESTree produces for
// ({ a } = value).
func (c *converter) assignmentTarget(expr js_ast.Expr) Node {
	object, ok := expr.Data.(*js_ast.EObject)
	if !ok {
		return c.expr(expr)
	}
	node := &ObjectPattern{}
	for _, property := range object.Properties {
		node.Properties = append(node.Properties, c.objectProperty(property))
	}
	return c.place(node, expr.Loc)
}

func (c *converter) objectProperty(property js_ast.Property) Node {
	loc := property.Loc
	if property.Key.Data != nil {
		loc = property.Key.Loc
	} else if property.ValueOrNil.Data != nil {
		loc = property.ValueOrNil.Loc
	}
	if property.Key.Data == nil {
		node := &RestElement{}
		if property.ValueOrNil.Data != nil {
			node.Argument = c.expr(property.ValueOrNil)
		}
		return c.place(node, loc)
	}
	node := &Property{Key: c.propertyKey(property.Key)}
	if property.ValueOrNil.Data != nil {
		node.Value = c.expr(property.ValueOrNil)
	}
	if identifier, ok := node.Key.(*Identifier); ok {
		if local, ok := node.Value.(*Identifier); ok && local.Name == identifier.Name {
			node.Shorthand = true
		}
	}
	return c.place(node, loc)
}

func decodeUTF16(units []uint16) string {
	return string(utf16.Decode(units))
}

func kindName(data any) string {
	if data == nil {
		return "Unknown"
	}
	name := reflect.TypeOf(data).String()
	if index := strings.LastIndex(name, "."); index >= 0 {
		name = name[index+1:]
	}
	return name
}

// unknown wraps a construct the rules never inspect. Its descendants are still
// discovered so that traversal remains complete over the whole program.
func (c *converter) unknown(kind string, data any, loc logger.Loc) Node {
	node := &Unknown{Kind: kind}
	if data != nil {
		c.seen = map[uintptr]bool{}
		c.descend(reflect.ValueOf(data), &node.Nodes, 0)
		c.seen = nil
	}
	return c.place(node, loc)
}

var (
	stmtType    = reflect.TypeOf(js_ast.Stmt{})
	exprType    = reflect.TypeOf(js_ast.Expr{})
	bindingType = reflect.TypeOf(js_ast.Binding{})
)

// descend finds every statement, expression and binding reachable from an
// unmodelled construct, without enumerating the entire grammar by hand.
func (c *converter) descend(value reflect.Value, out *[]Node, depth int) {
	if depth > 128 || !value.IsValid() {
		return
	}
	switch value.Kind() {
	case reflect.Pointer:
		if value.IsNil() {
			return
		}
		address := value.Pointer()
		if c.seen[address] {
			return
		}
		c.seen[address] = true
		c.descend(value.Elem(), out, depth+1)

	case reflect.Interface:
		if value.IsNil() {
			return
		}
		c.descend(value.Elem(), out, depth+1)

	case reflect.Slice, reflect.Array:
		for index := 0; index < value.Len(); index++ {
			c.descend(value.Index(index), out, depth+1)
		}

	case reflect.Struct:
		switch value.Type() {
		case stmtType:
			if stmt, ok := value.Interface().(js_ast.Stmt); ok && stmt.Data != nil {
				*out = append(*out, c.stmt(stmt))
			}
			return
		case exprType:
			if expr, ok := value.Interface().(js_ast.Expr); ok && expr.Data != nil {
				*out = append(*out, c.expr(expr))
			}
			return
		case bindingType:
			if binding, ok := value.Interface().(js_ast.Binding); ok && binding.Data != nil {
				*out = append(*out, c.binding(binding))
			}
			return
		}
		if isOpaque(value.Type()) {
			return
		}
		for index := 0; index < value.NumField(); index++ {
			field := value.Field(index)
			if !field.CanInterface() {
				continue
			}
			c.descend(field, out, depth+1)
		}
	}
}

// isOpaque reports whether a struct type holds compiler bookkeeping rather than
// syntax. Scopes link back to their parents, so descending into them would both
// cycle and reach nothing the rules care about.
func isOpaque(t reflect.Type) bool {
	switch t.Name() {
	case "Scope", "Symbol", "Ref", "LocRef", "Loc", "Range", "Path", "ImportRecord", "CharFreq":
		return true
	default:
		return false
	}
}
