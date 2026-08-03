package grammar

import (
	"encoding/json"
	"strings"
	"testing"
)

const sample = `[
  {"file": "doc.md", "lint_count": 2, "lints": [
    {"rule": "PronounVerbAgreement", "kind": "Agreement", "line": 3, "column": 5,
     "message": "The form of the verb must agree.", "matched_text": "maintains"},
    {"rule": "LongSentences", "kind": "Readability", "line": 7, "column": 1,
     "message": "This sentence is 45 words long.", "matched_text": "The"}
  ]}
]`

func TestParse(t *testing.T) {
	got, err := Parse(strings.NewReader(sample))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("findings = %d, want 2", len(got))
	}
	if got[0].Check != "harper.PronounVerbAgreement" {
		t.Errorf("Check = %q", got[0].Check)
	}
	// The span covers the matched text, so a caller can highlight it.
	if got[0].Span != [2]int{5, 13} {
		t.Errorf("Span = %v, want [5 13]", got[0].Span)
	}
}

func TestSeverityKeepsMeasurementApartFromMistakes(t *testing.T) {
	// A verb that disagrees with its subject is wrong, where a long sentence is
	// only long, and the register decides whether that matters.
	got, _ := Parse(strings.NewReader(sample))
	if got[0].Severity != "error" {
		t.Errorf("agreement severity = %q, want error", got[0].Severity)
	}
	if got[1].Severity != "warning" {
		t.Errorf("readability severity = %q, want warning", got[1].Severity)
	}
}

func TestHasErrorIgnoresWarnings(t *testing.T) {
	got, _ := Parse(strings.NewReader(sample))
	if !HasError(got) {
		t.Error("an agreement finding did not count as an error")
	}
	if HasError(got[1:]) {
		t.Error("a readability finding counted as an error")
	}
}

func TestLines(t *testing.T) {
	got, _ := Parse(strings.NewReader(sample))
	want := "doc.md:3:5:harper.PronounVerbAgreement:The form of the verb must agree.\n"
	if !strings.HasPrefix(Lines(got), want) {
		t.Errorf("Lines = %q, want it to start with %q", Lines(got), want)
	}
}

func TestMergeJSONJoinsBothReports(t *testing.T) {
	// A caller grouping by rule has to see one document, not two.
	vale := []byte(`{"doc.md":[{"Check":"ai-tells.Delve","Message":"m","Severity":"error","Line":1}]}`)
	findings, _ := Parse(strings.NewReader(sample))

	merged, err := MergeJSON(vale, findings)
	if err != nil {
		t.Fatal(err)
	}
	var report map[string][]map[string]any
	if err := json.Unmarshal(merged, &report); err != nil {
		t.Fatal(err)
	}
	if len(report["doc.md"]) != 3 {
		t.Errorf("entries = %d, want 3", len(report["doc.md"]))
	}
}

func TestMergeJSONWithoutAValeReport(t *testing.T) {
	findings, _ := Parse(strings.NewReader(sample))
	merged, err := MergeJSON(nil, findings)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(merged), "harper.PronounVerbAgreement") {
		t.Errorf("merged = %s", merged)
	}
}
