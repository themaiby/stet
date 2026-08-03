// Package generate turns upstream data into Vale rule files.
//
// Every function here is a transform: text in, text out. The download, the
// temporary directory and the destination path belong to the caller. That is
// what lets a test pin the exact bytes of a rule file without a network.
package generate

import (
	"bufio"
	"io"
	"strings"
)

// posixSpace is the [[:space:]] class the shell version matched on. Go's
// unicode.IsSpace is wider, and widening the class would drop keys the old
// pipeline kept.
const posixSpace = " \t\n\v\f\r"

// SubstitutionOptions describes one rule file built from a LanguageTool
// replace list.
type SubstitutionOptions struct {
	// Note is an optional first comment line saying what this list is for.
	Note string
	// Provenance is the comment naming the source and its licence.
	Provenance string
	// Message is the Vale message template.
	Message string
	// Level is the Vale severity.
	Level string
	// BareKeysOnly drops entries whose key contains whitespace. The Ukrainian
	// lists carry multi-word keys that belong to the pattern rules instead.
	BareKeysOnly bool
	// DropContext drops entries whose replacement is a "ctx:" annotation rather
	// than a word.
	DropContext bool
}

// Substitution converts a LanguageTool replace list into a Vale substitution
// rule. It returns the file text and the number of pairs kept.
func Substitution(r io.Reader, o SubstitutionOptions) (string, int, error) {
	var b strings.Builder
	if o.Note != "" {
		b.WriteString("# " + o.Note + "\n")
	}
	b.WriteString("# " + o.Provenance + "\n")
	b.WriteString("extends: substitution\n")
	b.WriteString("message: \"" + o.Message + "\"\n")
	b.WriteString("level: " + o.Level + "\n")
	b.WriteString("ignorecase: true\n")
	b.WriteString("swap:\n")

	pairs := 0
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.Trim(line, posixSpace) == "" {
			continue
		}
		key, value, ok := splitOnce(line, '=')
		if !ok {
			continue
		}
		if o.BareKeysOnly && strings.ContainsAny(key, posixSpace) {
			continue
		}
		value = strings.TrimRight(firstSegment(value), posixSpace)
		if o.DropContext && strings.HasPrefix(value, "ctx:") {
			continue
		}
		// The key is marked seen before the value is judged, which is what the
		// shell pipeline did: a key whose first entry is unusable stays out even
		// if a later line would have supplied a good one.
		if seen[key] {
			continue
		}
		seen[key] = true
		// A tab means upstream appended commentary to the replacement rather
		// than ending the entry. Vale takes one replacement, and the commentary
		// would arrive inside it.
		if strings.ContainsAny(key, "\t") || strings.ContainsAny(value, "\t") {
			continue
		}
		if value == "" || key == value || strings.Contains(value, ":") {
			continue
		}
		b.WriteString("  \"" + key + "\": \"" + value + "\"\n")
		pairs++
	}
	if err := scanner.Err(); err != nil {
		return "", 0, err
	}
	return b.String(), pairs, nil
}

// splitOnce splits on a single separator and reports whether there was exactly
// one. A line carrying two separators is ambiguous, and guessing which one
// divides the pair would invent entries nobody wrote.
func splitOnce(line string, sep byte) (string, string, bool) {
	if strings.Count(line, string(sep)) != 1 {
		return "", "", false
	}
	i := strings.IndexByte(line, sep)
	return line[:i], line[i+1:], true
}

// firstSegment keeps the text up to the first alternative, commentary or
// parenthesis. Vale shows one replacement, and the rest belong in a dictionary.
func firstSegment(value string) string {
	if i := strings.IndexAny(value, "|;("); i >= 0 {
		return value[:i]
	}
	return value
}
