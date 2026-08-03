// Package grammar converts a grammar checker's report into the shape the rest
// of the output already uses.
//
// Vale matches text and cannot ask a tagger, so agreement, confused words and
// repetition are beyond it however many rules it carries. A checker that parses
// the sentence reaches them.
package grammar

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// report is the shape harper writes with --format json.
type report struct {
	File  string `json:"file"`
	Lints []struct {
		Rule        string `json:"rule"`
		Kind        string `json:"kind"`
		Line        int    `json:"line"`
		Column      int    `json:"column"`
		Message     string `json:"message"`
		MatchedText string `json:"matched_text"`
	} `json:"lints"`
}

// Finding is one result, named and placed the way Vale names and places its own.
type Finding struct {
	File     string
	Check    string
	Message  string
	Severity string
	Line     int
	Span     [2]int
	Match    string
}

// Parse reads a harper report.
func Parse(r io.Reader) ([]Finding, error) {
	var reports []report
	if err := json.NewDecoder(r).Decode(&reports); err != nil {
		return nil, err
	}
	var out []Finding
	for _, file := range reports {
		for _, lint := range file.Lints {
			out = append(out, Finding{
				File:     file.File,
				Check:    "harper." + lint.Rule,
				Message:  strings.TrimSpace(lint.Message),
				Severity: severity(lint.Kind),
				Line:     lint.Line,
				Span:     [2]int{lint.Column, lint.Column + len([]rune(lint.MatchedText)) - 1},
				Match:    lint.MatchedText,
			})
		}
	}
	return out, nil
}

// severity keeps the objective mistakes apart from the one measurement. A verb
// that disagrees with its subject is wrong; a long sentence is only long.
func severity(kind string) string {
	if kind == "Readability" {
		return "warning"
	}
	return "error"
}

// Lines renders the findings the way Vale's line output renders its own.
func Lines(findings []Finding) string {
	var b strings.Builder
	for _, f := range findings {
		fmt.Fprintf(&b, "%s:%d:%d:%s:%s\n", f.File, f.Line, f.Span[0], f.Check, f.Message)
	}
	return b.String()
}

// MergeJSON folds the findings into a Vale JSON report, so that a caller
// grouping by rule sees one document rather than two.
func MergeJSON(valeReport []byte, findings []Finding) ([]byte, error) {
	report := map[string][]map[string]any{}
	if len(valeReport) > 0 {
		if err := json.Unmarshal(valeReport, &report); err != nil {
			return nil, err
		}
	}
	for _, f := range findings {
		report[f.File] = append(report[f.File], map[string]any{
			"Check":    f.Check,
			"Message":  f.Message,
			"Severity": f.Severity,
			"Line":     f.Line,
			"Span":     []int{f.Span[0], f.Span[1]},
			"Match":    f.Match,
		})
	}
	return json.MarshalIndent(report, "", "  ")
}

// HasError reports whether anything found is worth failing a gate over.
func HasError(findings []Finding) bool {
	for _, f := range findings {
		if f.Severity == "error" {
			return true
		}
	}
	return false
}
