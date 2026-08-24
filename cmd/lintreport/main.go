// Command lintreport runs the plugin over a JavaScript file and writes the same
// three artefacts the original repository's entry point was exercised to
// produce: the shareable configurations, the exported rule names, and the
// problems reported for the file.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	plugin "github.com/you-dont-need/You-Dont-Need-Lodash-Underscore"
	"github.com/you-dont-need/You-Dont-Need-Lodash-Underscore/internal/linter"
)

func main() {
	fixture := flag.String("fixture", "", "path to the JavaScript file to lint")
	out := flag.String("out", "", "directory to write configs.json, rules.txt and lint_report.txt into")
	flag.Parse()

	if *fixture == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: lintreport -fixture <file.js> -out <directory>")
		os.Exit(2)
	}
	if err := run(*fixture, *out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(fixture, out string) error {
	source, err := os.ReadFile(fixture)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}

	configs, err := renderConfigs()
	if err != nil {
		return err
	}
	if err := write(filepath.Join(out, "configs.json"), configs); err != nil {
		return err
	}
	if err := write(filepath.Join(out, "rules.txt"), renderRuleNames()); err != nil {
		return err
	}

	messages, err := plugin.Verify(filepath.Base(fixture), string(source))
	if err != nil {
		return err
	}
	if err := write(filepath.Join(out, "lint_report.txt"), renderReport(messages)); err != nil {
		return err
	}

	fmt.Printf("configs: %d\n", len(plugin.ConfigsValue.Keys()))
	fmt.Printf("rules: %d\n", len(plugin.Rules))
	fmt.Printf("reports: %d\n", len(messages))
	return nil
}

// renderConfigs serialises the configurations exactly as JSON.stringify with an
// indent of two spaces did, so the result can be compared against the recorded
// baseline byte for byte.
func renderConfigs() (string, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(plugin.ConfigsValue); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func renderRuleNames() string {
	var buffer bytes.Buffer
	for _, rule := range plugin.Rules {
		buffer.WriteString(rule.ID)
		buffer.WriteByte('\n')
	}
	return buffer.String()
}

// renderReport formats the problems the way the baseline recorded them. The
// ordering is by position first and then by rule and message, so that two rules
// reporting the same node produce a stable listing.
func renderReport(messages []linter.Message) string {
	ordered := append([]linter.Message(nil), messages...)
	sort.SliceStable(ordered, func(i, j int) bool {
		a, b := ordered[i], ordered[j]
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		if a.Column != b.Column {
			return a.Column < b.Column
		}
		if a.RuleID != b.RuleID {
			return a.RuleID < b.RuleID
		}
		return a.Message < b.Message
	})

	var buffer bytes.Buffer
	for _, message := range ordered {
		fmt.Fprintf(&buffer, "%d:%d  %s/%s  %s\n",
			message.Line, message.Column, plugin.Name, message.RuleID, message.Message)
	}
	fmt.Fprintf(&buffer, "total: %d\n", len(ordered))
	return buffer.String()
}

func write(path, contents string) error {
	return os.WriteFile(path, []byte(contents), 0o644)
}
