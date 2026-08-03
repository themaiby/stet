package generate

import (
	"strings"
	"testing"
)

func doc(rules string) []byte {
	return []byte(`<?xml version="1.0" encoding="utf-8"?><rules>` + rules + `</rules>`)
}

func tokens(result PatternResult) []string {
	var out []string
	for _, line := range strings.Split(result.Content, "\n") {
		if strings.HasPrefix(line, "  - ") {
			out = append(out, strings.TrimPrefix(line, "  - "))
		}
	}
	return out
}

func convert(t *testing.T, rules string, lemmas LemmaIndex) PatternResult {
	t.Helper()
	return Patterns([][]byte{doc(rules)}, lemmas, PatternOptions{
		Provenance: "test", Message: "m", Level: "warning",
	})
}

func TestPatternsConvertsLiteralTokens(t *testing.T) {
	got := convert(t, `<rule><pattern><token>перший</token><token>другий</token></pattern></rule>`, nil)
	if got.Converted != 1 {
		t.Fatalf("converted = %d, want 1:\n%s", got.Converted, got.Content)
	}
	if want := `'(?:перший)[ \t]+(?:другий)'`; tokens(got)[0] != want {
		t.Errorf("token = %s, want %s", tokens(got)[0], want)
	}
}

func TestPatternsSkipsWhatNeedsATagger(t *testing.T) {
	cases := map[string]string{
		"a part-of-speech attribute":  `<rule><pattern><token postag="verb"/><token>слово</token></pattern></rule>`,
		"an exception inside a token": `<rule><pattern><token>один<exception>два</exception></token><token>три</token></pattern></rule>`,
		"an antipattern":              `<rule><antipattern><token>ні</token></antipattern><pattern><token>перший</token><token>другий</token></pattern></rule>`,
		"a single mandatory token":    `<rule><pattern><token>самотній</token></pattern></rule>`,
		"a one-letter alternative":    `<rule><pattern><token regexp="yes">і|його</token><token>слово</token></pattern></rule>`,
		"a lone optional token": `<rule><pattern><token min="0">може</token>` +
			`<token>слово</token></pattern></rule>`,
	}
	for name, rule := range cases {
		t.Run(name, func(t *testing.T) {
			got := convert(t, rule, nil)
			if got.Converted != 0 {
				t.Errorf("converted %d rules, want 0:\n%s", got.Converted, got.Content)
			}
			if got.Skipped != 1 {
				t.Errorf("skipped = %d, want 1", got.Skipped)
			}
		})
	}
}

func TestPatternsRejectsOpenEndedWildcards(t *testing.T) {
	// The euphony rules write "any neighbouring word" as a wildcard, which
	// outside LanguageTool's tokeniser matches whole sentences.
	got := convert(t, `<rule><pattern><token regexp="yes">.*ння</token><token>слово</token></pattern></rule>`, nil)
	if got.Converted != 0 {
		t.Errorf("a sentence-swallowing wildcard survived:\n%s", got.Content)
	}
}

func TestPatternsExpandsInflectedTokens(t *testing.T) {
	lemmas := LemmaIndex{"робити": "робити робив робила"}
	got := convert(t, `<rule><pattern><token inflected="yes">робити</token><token>справу</token></pattern></rule>`, lemmas)
	if got.Converted != 1 {
		t.Fatalf("converted = %d, want 1. Skipped %d", got.Converted, got.Skipped)
	}
	// Forms are sorted so that the same corpus always yields the same rule.
	if want := `'(?:робив|робила|робити)[ \t]+(?:справу)'`; tokens(got)[0] != want {
		t.Errorf("token = %s, want %s", tokens(got)[0], want)
	}
}

func TestPatternsSkipsUnknownLemmas(t *testing.T) {
	got := convert(t, `<rule><pattern><token inflected="yes">невідоме</token><token>слово</token></pattern></rule>`, LemmaIndex{})
	if got.Converted != 0 {
		t.Errorf("a lemma with no forms produced a rule:\n%s", got.Content)
	}
}

func TestPatternsEscapesLikeTheReferenceImplementation(t *testing.T) {
	// The escape set is wider than Go's own, and the generated file carries
	// the difference. A hyphen and a hash are the ones Go would leave alone.
	if got, want := escapeRegex("a-b#c d"), `a\-b\#c\ d`; got != want {
		t.Errorf("escapeRegex = %q, want %q", got, want)
	}
	if got, want := escapeRegex("слово"), "слово"; got != want {
		t.Errorf("escapeRegex changed non-ASCII text: %q", got)
	}
}

func TestPatternsResolvesDeclaredEntities(t *testing.T) {
	// The Ukrainian rule files declare punctuation classes in the internal
	// subset and reference them inside tokens. A parser that cannot resolve
	// them reads no rules at all.
	document := []byte(`<?xml version="1.0"?>
<!DOCTYPE rules [ <!ENTITY dash "&quot;тире|риска&quot;"> ]>
<rules><rule><pattern><token regexp="yes">&dash;</token><token>слово</token></pattern></rule></rules>`)

	got := Patterns([][]byte{document}, nil, PatternOptions{Provenance: "test", Message: "m", Level: "warning"})
	if got.Converted != 1 {
		t.Fatalf("converted = %d, want 1. Skipped %d:\n%s", got.Converted, got.Skipped, got.Content)
	}
	if !strings.Contains(got.Content, "тире|риска") {
		t.Errorf("the entity was not resolved:\n%s", got.Content)
	}
}

func TestPatternsQuotesSuggestionsAsComments(t *testing.T) {
	rule := `<rule><pattern><token>перший</token><token>другий</token></pattern>` +
		`<message>Краще: <suggestion>інший \1</suggestion></message></rule>`
	got := convert(t, rule, nil)
	if !strings.Contains(got.Content, "  # інший …\n") {
		t.Errorf("the suggestion did not become a comment:\n%s", got.Content)
	}
}

func TestPatternsDropsDuplicateExpressions(t *testing.T) {
	rule := `<rule><pattern><token>перший</token><token>другий</token></pattern></rule>`
	got := convert(t, rule+rule, nil)
	if got.Converted != 1 {
		t.Errorf("converted = %d, want 1", got.Converted)
	}
}
