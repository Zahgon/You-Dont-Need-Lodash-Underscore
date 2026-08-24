package estree

import (
	"strings"
	"testing"
)

func mustParse(t *testing.T, source string) *Program {
	t.Helper()
	program, err := Parse("fixture.js", source)
	if err != nil {
		t.Fatalf("Parse(%q) returned error %v", source, err)
	}
	return program
}

// collect returns every node in the tree whose Type() equals want.
func collect(program *Program, want string) []Node {
	var found []Node
	Walk(program, func(n Node) {
		if n.Type() == want {
			found = append(found, n)
		}
	})
	return found
}

func typeSequence(program *Program) []string {
	var seen []string
	Walk(program, func(n Node) { seen = append(seen, n.Type()) })
	return seen
}

func TestSyntaxError(t *testing.T) {
	t.Run("Error formats file and messages", func(t *testing.T) {
		err := &SyntaxError{File: "broken.js", Messages: []string{"first", "second"}}
		if got := err.Error(); got != "broken.js: first; second" {
			t.Fatalf("Error() = %q, want %q", got, "broken.js: first; second")
		}
	})

	t.Run("Parse reports unparsable source", func(t *testing.T) {
		program, err := Parse("broken.js", "const = ;")
		if err == nil {
			t.Fatalf("Parse returned program %v and no error, want a syntax error", program)
		}
		if program != nil {
			t.Errorf("Parse returned program %v, want nil", program)
		}
		syntaxErr, ok := err.(*SyntaxError)
		if !ok {
			t.Fatalf("Parse returned %T, want *SyntaxError", err)
		}
		if syntaxErr.File != "broken.js" {
			t.Errorf("File = %q, want %q", syntaxErr.File, "broken.js")
		}
		if len(syntaxErr.Messages) == 0 {
			t.Errorf("Messages is empty, want at least one diagnostic")
		}
		if !strings.Contains(syntaxErr.Error(), "broken.js") {
			t.Errorf("Error() = %q, want it to mention broken.js", syntaxErr.Error())
		}
	})
}

func TestParseProgram(t *testing.T) {
	t.Run("empty source has no body", func(t *testing.T) {
		program := mustParse(t, "")
		if len(program.Body) != 0 {
			t.Errorf("Body has %d entries, want 0", len(program.Body))
		}
		if program.SourceType != "module" {
			t.Errorf("SourceType = %q, want %q", program.SourceType, "module")
		}
		if program.Line() != 1 || program.Column() != 1 {
			t.Errorf("position = %d:%d, want 1:1", program.Line(), program.Column())
		}
		if program.Parent() != nil {
			t.Errorf("Parent() = %v, want nil", program.Parent())
		}
	})

	t.Run("statement is an ExpressionStatement", func(t *testing.T) {
		program := mustParse(t, "f();\n")
		if len(program.Body) != 1 {
			t.Fatalf("Body has %d entries, want 1", len(program.Body))
		}
		if got := program.Body[0].Type(); got != "ExpressionStatement" {
			t.Fatalf("Body[0].Type() = %q, want ExpressionStatement", got)
		}
	})
}

func TestVariableDeclarationKinds(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   string
	}{
		{"var", "var a = 1;", "var"},
		{"let", "let a = 1;", "let"},
		{"const", "const a = 1;", "const"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			program := mustParse(t, testCase.source)
			declarations := collect(program, "VariableDeclaration")
			if len(declarations) != 1 {
				t.Fatalf("found %d VariableDeclaration nodes, want 1", len(declarations))
			}
			declaration := declarations[0].(*VariableDeclaration)
			if declaration.Kind != testCase.want {
				t.Errorf("Kind = %q, want %q", declaration.Kind, testCase.want)
			}
			if len(declaration.Declarations) != 1 {
				t.Fatalf("Declarations has %d entries, want 1", len(declaration.Declarations))
			}
		})
	}

	t.Run("multiple declarators", func(t *testing.T) {
		program := mustParse(t, "var a = 1, b = 2, c;")
		declaration := collect(program, "VariableDeclaration")[0].(*VariableDeclaration)
		if len(declaration.Declarations) != 3 {
			t.Fatalf("Declarations has %d entries, want 3", len(declaration.Declarations))
		}
		last := declaration.Declarations[2].(*VariableDeclarator)
		if last.Init != nil {
			t.Errorf("third declarator Init = %v, want nil", last.Init)
		}
		if id, ok := last.ID.(*Identifier); !ok || id.Name != "c" {
			t.Errorf("third declarator ID = %v, want Identifier c", last.ID)
		}
	})
}

func TestParseImports(t *testing.T) {
	t.Run("default import", func(t *testing.T) {
		program := mustParse(t, "import lodash from 'lodash';")
		declarations := collect(program, "ImportDeclaration")
		if len(declarations) != 1 {
			t.Fatalf("found %d ImportDeclaration nodes, want 1", len(declarations))
		}
		declaration := declarations[0].(*ImportDeclaration)
		source, ok := declaration.Source.(*Literal)
		if !ok || source.Value != "lodash" || !source.IsString {
			t.Fatalf("Source = %v, want string Literal lodash", declaration.Source)
		}
		if len(declaration.Specifiers) != 1 {
			t.Fatalf("Specifiers has %d entries, want 1", len(declaration.Specifiers))
		}
		specifier, ok := declaration.Specifiers[0].(*ImportDefaultSpecifier)
		if !ok {
			t.Fatalf("Specifiers[0] is %T, want *ImportDefaultSpecifier", declaration.Specifiers[0])
		}
		if local, ok := specifier.Local.(*Identifier); !ok || local.Name != "lodash" {
			t.Errorf("Local = %v, want Identifier lodash", specifier.Local)
		}
	})

	t.Run("namespace import", func(t *testing.T) {
		program := mustParse(t, "import * as _ from 'lodash';")
		declaration := collect(program, "ImportDeclaration")[0].(*ImportDeclaration)
		if len(declaration.Specifiers) != 1 {
			t.Fatalf("Specifiers has %d entries, want 1", len(declaration.Specifiers))
		}
		specifier, ok := declaration.Specifiers[0].(*ImportNamespaceSpecifier)
		if !ok {
			t.Fatalf("Specifiers[0] is %T, want *ImportNamespaceSpecifier", declaration.Specifiers[0])
		}
		if local, ok := specifier.Local.(*Identifier); !ok || local.Name != "_" {
			t.Errorf("Local = %v, want Identifier _", specifier.Local)
		}
	})

	t.Run("named imports", func(t *testing.T) {
		program := mustParse(t, "import { map, filter as keep } from 'lodash';")
		declaration := collect(program, "ImportDeclaration")[0].(*ImportDeclaration)
		if len(declaration.Specifiers) != 2 {
			t.Fatalf("Specifiers has %d entries, want 2", len(declaration.Specifiers))
		}
		first := declaration.Specifiers[0].(*ImportSpecifier)
		if imported := first.Imported.(*Identifier); imported.Name != "map" {
			t.Errorf("first Imported = %q, want map", imported.Name)
		}
		second := declaration.Specifiers[1].(*ImportSpecifier)
		if imported := second.Imported.(*Identifier); imported.Name != "filter" {
			t.Errorf("second Imported = %q, want filter", imported.Name)
		}
		if local := second.Local.(*Identifier); local.Name != "keep" {
			t.Errorf("second Local = %q, want keep", local.Name)
		}
	})

	t.Run("default and named together", func(t *testing.T) {
		program := mustParse(t, "import lodash, { map } from 'lodash';")
		declaration := collect(program, "ImportDeclaration")[0].(*ImportDeclaration)
		if len(declaration.Specifiers) != 2 {
			t.Fatalf("Specifiers has %d entries, want 2", len(declaration.Specifiers))
		}
		if _, ok := declaration.Specifiers[0].(*ImportDefaultSpecifier); !ok {
			t.Errorf("Specifiers[0] is %T, want *ImportDefaultSpecifier", declaration.Specifiers[0])
		}
		if _, ok := declaration.Specifiers[1].(*ImportSpecifier); !ok {
			t.Errorf("Specifiers[1] is %T, want *ImportSpecifier", declaration.Specifiers[1])
		}
	})

	t.Run("bare import has no specifiers", func(t *testing.T) {
		program := mustParse(t, "import 'lodash';")
		declaration := collect(program, "ImportDeclaration")[0].(*ImportDeclaration)
		if len(declaration.Specifiers) != 0 {
			t.Fatalf("Specifiers has %d entries, want 0", len(declaration.Specifiers))
		}
		children := declaration.Children()
		if len(children) != 1 || children[0] != declaration.Source {
			t.Errorf("Children() = %v, want just the source literal", children)
		}
	})
}
