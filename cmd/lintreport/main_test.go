package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	plugin "github.com/you-dont-need/You-Dont-Need-Lodash-Underscore"
	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/linter"
)

const (
	fixturePath  = "../../baseline_outputs/fixture.js"
	baselineRoot = "../../baseline_outputs"
)

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}
	return contents
}

func TestRunMatchesTheRecordedBaseline(t *testing.T) {
	out := t.TempDir()
	if err := run(fixturePath, out); err != nil {
		t.Fatalf("run returned %v", err)
	}

	for _, name := range []string{"configs.json", "rules.txt", "lint_report.txt"} {
		t.Run(name, func(t *testing.T) {
			want := readFile(t, filepath.Join(baselineRoot, name))
			got := readFile(t, filepath.Join(out, name))
			if !bytes.Equal(got, want) {
				t.Errorf("%s differs from the baseline: got %d bytes, want %d bytes",
					name, len(got), len(want))
			}
		})
	}
}

func TestRunCreatesTheOutputDirectory(t *testing.T) {
	out := filepath.Join(t.TempDir(), "nested", "deeper")
	if err := run(fixturePath, out); err != nil {
		t.Fatalf("run returned %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "configs.json")); err != nil {
		t.Fatalf("configs.json was not created: %v", err)
	}
}

func TestRunErrors(t *testing.T) {
	t.Run("fixture does not exist", func(t *testing.T) {
		err := run(filepath.Join(t.TempDir(), "missing.js"), t.TempDir())
		if err == nil {
			t.Fatalf("run returned no error for a missing fixture")
		}
		if !os.IsNotExist(err) {
			t.Errorf("error = %v, want a not-exist error", err)
		}
	})

	t.Run("output path is an existing file", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "not-a-directory")
		if err := os.WriteFile(out, []byte("x"), 0o644); err != nil {
			t.Fatalf("cannot create the blocking file: %v", err)
		}
		if err := run(fixturePath, out); err == nil {
			t.Fatalf("run returned no error when the output path is a file")
		}
	})

	for _, blocked := range []string{"configs.json", "rules.txt", "lint_report.txt"} {
		t.Run("cannot write "+blocked, func(t *testing.T) {
			out := t.TempDir()
			if err := os.Mkdir(filepath.Join(out, blocked), 0o755); err != nil {
				t.Fatalf("cannot create the blocking directory: %v", err)
			}
			if err := run(fixturePath, out); err == nil {
				t.Fatalf("run returned no error when %s cannot be written", blocked)
			}
		})
	}

	t.Run("fixture does not parse", func(t *testing.T) {
		fixture := filepath.Join(t.TempDir(), "broken.js")
		if err := os.WriteFile(fixture, []byte("const = ;"), 0o644); err != nil {
			t.Fatalf("cannot write the broken fixture: %v", err)
		}
		err := run(fixture, t.TempDir())
		if err == nil {
			t.Fatalf("run returned no error for an unparsable fixture")
		}
		if !strings.Contains(err.Error(), "broken.js") {
			t.Errorf("error = %q, want it to mention broken.js", err.Error())
		}
	})
}

func TestWrite(t *testing.T) {
	t.Run("writes the contents", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "file.txt")
		if err := write(path, "hello"); err != nil {
			t.Fatalf("write returned %v", err)
		}
		if got := string(readFile(t, path)); got != "hello" {
			t.Errorf("contents = %q, want %q", got, "hello")
		}
	})

	t.Run("parent directory does not exist", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing", "file.txt")
		if err := write(path, "hello"); err == nil {
			t.Fatalf("write returned no error for a missing parent directory")
		}
	})

	t.Run("path is a directory", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "directory")
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatalf("cannot create the directory: %v", err)
		}
		if err := write(path, "hello"); err == nil {
			t.Fatalf("write returned no error when the path is a directory")
		}
	})
}

func TestRenderConfigs(t *testing.T) {
	got, err := renderConfigs()
	if err != nil {
		t.Fatalf("renderConfigs returned %v", err)
	}
	want := string(readFile(t, filepath.Join(baselineRoot, "configs.json")))

	t.Run("matches the baseline", func(t *testing.T) {
		if got != want {
			t.Errorf("renderConfigs differs from the baseline: got %d bytes, want %d bytes",
				len(got), len(want))
		}
	})
	t.Run("ends with a newline", func(t *testing.T) {
		if !strings.HasSuffix(got, "\n") {
			t.Errorf("renderConfigs does not end with a newline")
		}
	})
	t.Run("indents with two spaces", func(t *testing.T) {
		if !strings.Contains(got, "\n  \"all-warn\": {") {
			t.Errorf("renderConfigs is not indented with two spaces:\n%s", got[:120])
		}
	})
}

func TestRenderRuleNames(t *testing.T) {
	got := renderRuleNames()

	t.Run("matches the baseline", func(t *testing.T) {
		want := string(readFile(t, filepath.Join(baselineRoot, "rules.txt")))
		if got != want {
			t.Errorf("renderRuleNames differs from the baseline")
		}
	})
	t.Run("one line per rule", func(t *testing.T) {
		lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
		if len(lines) != len(plugin.Rules) {
			t.Fatalf("renderRuleNames produced %d lines, want %d", len(lines), len(plugin.Rules))
		}
		for index, rule := range plugin.Rules {
			if lines[index] != rule.ID {
				t.Fatalf("line %d = %q, want %q", index, lines[index], rule.ID)
			}
		}
	})
}

func TestRenderReport(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		if got := renderReport(nil); got != "total: 0\n" {
			t.Errorf("renderReport(nil) = %q, want %q", got, "total: 0\n")
		}
	})

	t.Run("single message", func(t *testing.T) {
		messages := []linter.Message{{RuleID: "map", Line: 2, Column: 5, Message: "text"}}
		want := "2:5  " + plugin.Name + "/map  text\ntotal: 1\n"
		if got := renderReport(messages); got != want {
			t.Errorf("renderReport = %q, want %q", got, want)
		}
	})

	t.Run("orders by line", func(t *testing.T) {
		messages := []linter.Message{
			{RuleID: "b", Line: 3, Column: 1, Message: "later"},
			{RuleID: "a", Line: 1, Column: 1, Message: "earlier"},
		}
		got := renderReport(messages)
		if !strings.HasPrefix(got, "1:1  ") {
			t.Errorf("renderReport = %q, want the line 1 message first", got)
		}
	})

	t.Run("orders by column within a line", func(t *testing.T) {
		messages := []linter.Message{
			{RuleID: "a", Line: 1, Column: 9, Message: "right"},
			{RuleID: "a", Line: 1, Column: 2, Message: "left"},
		}
		got := renderReport(messages)
		if !strings.HasPrefix(got, "1:2  ") {
			t.Errorf("renderReport = %q, want the column 2 message first", got)
		}
	})

	t.Run("orders by rule at the same position", func(t *testing.T) {
		messages := []linter.Message{
			{RuleID: "zeta", Line: 1, Column: 1, Message: "z"},
			{RuleID: "alpha", Line: 1, Column: 1, Message: "a"},
		}
		got := renderReport(messages)
		want := "1:1  " + plugin.Name + "/alpha  a\n1:1  " + plugin.Name + "/zeta  z\ntotal: 2\n"
		if got != want {
			t.Errorf("renderReport = %q, want %q", got, want)
		}
	})

	t.Run("orders by message for the same rule and position", func(t *testing.T) {
		messages := []linter.Message{
			{RuleID: "same", Line: 1, Column: 1, Message: "second"},
			{RuleID: "same", Line: 1, Column: 1, Message: "first"},
		}
		got := renderReport(messages)
		want := "1:1  " + plugin.Name + "/same  first\n1:1  " + plugin.Name + "/same  second\ntotal: 2\n"
		if got != want {
			t.Errorf("renderReport = %q, want %q", got, want)
		}
	})

	t.Run("does not mutate its argument", func(t *testing.T) {
		messages := []linter.Message{
			{RuleID: "b", Line: 3, Column: 1, Message: "later"},
			{RuleID: "a", Line: 1, Column: 1, Message: "earlier"},
		}
		renderReport(messages)
		if messages[0].Line != 3 {
			t.Errorf("the input slice was reordered: %+v", messages)
		}
	})
}

func TestMain_HappyPath(t *testing.T) {
	out := t.TempDir()

	savedArgs := os.Args
	savedFlags := flag.CommandLine
	t.Cleanup(func() {
		os.Args = savedArgs
		flag.CommandLine = savedFlags
	})

	os.Args = []string{"lintreport", "-fixture", fixturePath, "-out", out}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	main()

	for _, name := range []string{"configs.json", "rules.txt", "lint_report.txt"} {
		want := readFile(t, filepath.Join(baselineRoot, name))
		got := readFile(t, filepath.Join(out, name))
		if !bytes.Equal(got, want) {
			t.Errorf("%s written by main differs from the baseline", name)
		}
	}
}
