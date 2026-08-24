package estree

import (
	"reflect"
	"testing"

	"github.com/ije/esbuild-internal/ast"
	"github.com/ije/esbuild-internal/js_ast"
	"github.com/ije/esbuild-internal/logger"
)

func TestPositionsAt(t *testing.T) {
	source := "abc\ndefg\n\nhi"
	p := newPositions(source)

	cases := []struct {
		name       string
		offset     int32
		wantLine   int
		wantColumn int
	}{
		{"negative clamps to start", -5, 1, 1},
		{"start of file", 0, 1, 1},
		{"middle of first line", 2, 1, 3},
		{"start of second line", 4, 2, 1},
		{"middle of second line", 6, 2, 3},
		{"empty third line", 9, 3, 1},
		{"fourth line", 10, 4, 1},
		{"end of file", int32(len(source)), 4, 3},
		{"beyond end clamps", 9999, 4, 3},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			line, column := p.at(testCase.offset)
			if line != testCase.wantLine || column != testCase.wantColumn {
				t.Errorf("at(%d) = %d:%d, want %d:%d",
					testCase.offset, line, column, testCase.wantLine, testCase.wantColumn)
			}
		})
	}

	t.Run("columns count UTF-16 units", func(t *testing.T) {
		// The emoji is a surrogate pair, so it occupies two UTF-16 units.
		astral := newPositions("\U0001F600x")
		line, column := astral.at(4)
		if line != 1 || column != 3 {
			t.Errorf("at(4) = %d:%d, want 1:3", line, column)
		}
	})
}

func TestSymbolNameOutOfRange(t *testing.T) {
	c := &converter{tree: &js_ast.AST{}}
	t.Run("index past the symbol table", func(t *testing.T) {
		if got := c.symbolName(ast.Ref{InnerIndex: 5}); got != "" {
			t.Errorf("symbolName = %q, want the empty string", got)
		}
	})
	t.Run("index inside the symbol table", func(t *testing.T) {
		withSymbols := &converter{tree: &js_ast.AST{Symbols: []ast.Symbol{{OriginalName: "named"}}}}
		if got := withSymbols.symbolName(ast.Ref{InnerIndex: 0}); got != "named" {
			t.Errorf("symbolName = %q, want %q", got, "named")
		}
	})
}

func TestLocalKindName(t *testing.T) {
	cases := []struct {
		name string
		kind js_ast.LocalKind
		want string
	}{
		{"var", js_ast.LocalVar, "var"},
		{"let", js_ast.LocalLet, "let"},
		{"const", js_ast.LocalConst, "const"},
		{"unrecognised falls back to var", js_ast.LocalKind(99), "var"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := localKindName(testCase.kind); got != testCase.want {
				t.Errorf("localKindName(%v) = %q, want %q", testCase.kind, got, testCase.want)
			}
		})
	}
}

func TestWasQuoted(t *testing.T) {
	c := &converter{source: "'a\"b`c d"}
	cases := []struct {
		name  string
		start int32
		want  bool
	}{
		{"single quote", 0, true},
		{"double quote", 2, true},
		{"backtick", 4, true},
		{"letter", 1, false},
		{"space", 7, false},
		{"offset past the source", 100, false},
		{"offset exactly at the end", int32(len(c.source)), false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := c.wasQuoted(logger.Loc{Start: testCase.start}); got != testCase.want {
				t.Errorf("wasQuoted(%d) = %v, want %v", testCase.start, got, testCase.want)
			}
		})
	}
}

func TestKindName(t *testing.T) {
	cases := []struct {
		name string
		data any
		want string
	}{
		{"nil", nil, "Unknown"},
		{"pointer to package type", &js_ast.SBlock{}, "SBlock"},
		{"builtin without a package qualifier", 42, "int"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := kindName(testCase.data); got != testCase.want {
				t.Errorf("kindName(%v) = %q, want %q", testCase.data, got, testCase.want)
			}
		})
	}
}

func TestPropertyKeyNil(t *testing.T) {
	c := &converter{tree: &js_ast.AST{}, positions: newPositions(""), source: ""}
	if got := c.propertyKey(js_ast.Expr{}); got != nil {
		t.Fatalf("propertyKey with no data = %v, want nil", got)
	}
}

func TestWireParentsSkipsNilChildren(t *testing.T) {
	child := &Identifier{Name: "kept"}
	program := &Program{Body: []Node{nil, child}}
	wireParents(program)
	if child.Parent() != Node(program) {
		t.Fatalf("child Parent() = %v, want the program", child.Parent())
	}
}

func TestParentPointersAndPositions(t *testing.T) {
	source := "import lodash from 'lodash';\n\nlodash.map(list, fn);\n"
	program := mustParse(t, source)

	t.Run("call expression position", func(t *testing.T) {
		call := collect(program, "CallExpression")[0]
		if call.Line() != 3 || call.Column() != 1 {
			t.Errorf("call position = %d:%d, want 3:1", call.Line(), call.Column())
		}
	})

	t.Run("member property position", func(t *testing.T) {
		member := collect(program, "MemberExpression")[0].(*MemberExpression)
		if member.Property.Line() != 3 || member.Property.Column() != 8 {
			t.Errorf("property position = %d:%d, want 3:8",
				member.Property.Line(), member.Property.Column())
		}
	})

	t.Run("every node except the program has a parent", func(t *testing.T) {
		Walk(program, func(n Node) {
			if n == Node(program) {
				return
			}
			if n.Parent() == nil {
				t.Errorf("node %s at %d:%d has no parent", n.Type(), n.Line(), n.Column())
			}
		})
	})

	t.Run("parent of a member expression is the call", func(t *testing.T) {
		member := collect(program, "MemberExpression")[0]
		parent := member.Parent()
		if parent == nil || parent.Type() != "CallExpression" {
			t.Errorf("parent = %v, want a CallExpression", parent)
		}
	})
}

func TestDescendBranches(t *testing.T) {
	c := &converter{tree: &js_ast.AST{}, positions: newPositions(""), source: "", seen: map[uintptr]bool{}}

	t.Run("invalid value is ignored", func(t *testing.T) {
		var out []Node
		c.descend(reflect.Value{}, &out, 0)
		if len(out) != 0 {
			t.Errorf("out has %d entries, want 0", len(out))
		}
	})

	t.Run("depth limit stops recursion", func(t *testing.T) {
		var out []Node
		stmt := js_ast.Stmt{Loc: logger.Loc{}, Data: &js_ast.SBlock{}}
		c.descend(reflect.ValueOf(stmt), &out, 129)
		if len(out) != 0 {
			t.Errorf("out has %d entries, want 0 at depth 129", len(out))
		}
	})

	t.Run("nil pointer is ignored", func(t *testing.T) {
		var out []Node
		var nilPointer *js_ast.SBlock
		c.descend(reflect.ValueOf(nilPointer), &out, 0)
		if len(out) != 0 {
			t.Errorf("out has %d entries, want 0", len(out))
		}
	})

	t.Run("nil interface is ignored", func(t *testing.T) {
		var out []Node
		holder := struct{ Data js_ast.S }{}
		c.descend(reflect.ValueOf(holder).Field(0), &out, 0)
		if len(out) != 0 {
			t.Errorf("out has %d entries, want 0", len(out))
		}
	})

	t.Run("an unseen pointer is descended", func(t *testing.T) {
		block := &js_ast.SBlock{Stmts: []js_ast.Stmt{{Data: &js_ast.SEmpty{}}}}
		var out []Node
		c.seen = map[uintptr]bool{}
		c.descend(reflect.ValueOf(block), &out, 0)
		if len(out) != 1 {
			t.Fatalf("out has %d entries, want 1", len(out))
		}
		if out[0].Type() != "Unknown" {
			t.Errorf("out[0].Type() = %q, want Unknown", out[0].Type())
		}
	})

	t.Run("a pointer already seen is not revisited", func(t *testing.T) {
		block := &js_ast.SBlock{Stmts: []js_ast.Stmt{{Data: &js_ast.SEmpty{}}}}
		var out []Node
		c.seen = map[uintptr]bool{reflect.ValueOf(block).Pointer(): true}
		c.descend(reflect.ValueOf(block), &out, 0)
		if len(out) != 0 {
			t.Errorf("out has %d entries, want 0 for an already seen pointer", len(out))
		}
	})

	t.Run("opaque struct is not descended", func(t *testing.T) {
		var out []Node
		c.seen = map[uintptr]bool{}
		c.descend(reflect.ValueOf(logger.Path{Text: "x"}), &out, 0)
		if len(out) != 0 {
			t.Errorf("out has %d entries, want 0", len(out))
		}
	})

	t.Run("array of expressions is descended", func(t *testing.T) {
		var out []Node
		c.seen = map[uintptr]bool{}
		values := [2]js_ast.Expr{
			{Data: &js_ast.EIdentifier{}},
			{Data: &js_ast.EIdentifier{}},
		}
		c.descend(reflect.ValueOf(values), &out, 0)
		if len(out) != 2 {
			t.Errorf("out has %d entries, want 2", len(out))
		}
	})

	t.Run("binding value is descended", func(t *testing.T) {
		var out []Node
		c.seen = map[uintptr]bool{}
		binding := js_ast.Binding{Data: &js_ast.BIdentifier{}}
		c.descend(reflect.ValueOf(binding), &out, 0)
		if len(out) != 1 || out[0].Type() != "Identifier" {
			t.Errorf("out = %v, want a single Identifier", out)
		}
	})

	t.Run("empty statement data is ignored", func(t *testing.T) {
		var out []Node
		c.seen = map[uintptr]bool{}
		c.descend(reflect.ValueOf(js_ast.Stmt{}), &out, 0)
		c.descend(reflect.ValueOf(js_ast.Expr{}), &out, 0)
		c.descend(reflect.ValueOf(js_ast.Binding{}), &out, 0)
		if len(out) != 0 {
			t.Errorf("out has %d entries, want 0", len(out))
		}
	})
}
