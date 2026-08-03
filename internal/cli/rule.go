package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// runRule answers which lever reaches a rule. A rule matching below the word
// level never reads accept.txt, and reaching for that file instead of .vale.ini
// looks like it worked and changes nothing.
func runRule(e *env, args []string) int {
	if len(args) != 1 || !strings.Contains(args[0], ".") {
		fmt.Fprintln(e.Err, "usage: stet rule <Style>.<Name>")
		return 2
	}
	style, name, _ := strings.Cut(args[0], ".")
	path := e.Layout.Style(style, name+".yml")

	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(e.Err, "stet: no rule %q. Looked in %s\n", args[0], path)
		return 1
	}
	defer file.Close()

	level, nonword := "", false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		switch line := strings.TrimSpace(scanner.Text()); {
		case strings.HasPrefix(line, "level:"):
			level = strings.TrimSpace(strings.TrimPrefix(line, "level:"))
		case line == "nonword: true":
			nonword = true
		}
	}
	if level == "" {
		level = "suggestion"
	}

	matching, vocabulary := "whole words", "accept.txt exempts a term from this rule"
	if nonword {
		matching = "below the word level"
		vocabulary = "accept.txt does not reach this rule; demote or switch it off in .vale.ini"
	}

	fmt.Fprintf(e.Out, "%s\n", args[0])
	fmt.Fprintf(e.Out, "  %-12s %s\n", "level", level)
	fmt.Fprintf(e.Out, "  %-12s %s\n", "matches", matching)
	fmt.Fprintf(e.Out, "  %-12s %s\n", "vocabulary", vocabulary)
	rel, err := filepath.Rel(e.Layout.Root, path)
	if err != nil {
		rel = path
	}
	fmt.Fprintf(e.Out, "  %-12s %s\n", "file", rel)
	return 0
}
