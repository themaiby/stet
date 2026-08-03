package generate

import (
	"strings"
	"testing"
)

func TestSubstitutionKeepsTheFirstSuggestionOnly(t *testing.T) {
	// Upstream writes alternatives, commentary and parentheses after the first
	// replacement. Vale shows one, so the rest is cut.
	source := strings.Join([]string{
		"# a comment",
		"",
		"wrong=right|alternative",
		"other=fine; some commentary",
		"third=good (a note)",
	}, "\n")

	got, pairs, err := Substitution(strings.NewReader(source), SubstitutionOptions{
		Provenance: "test", Message: "m", Level: "warning",
	})
	if err != nil {
		t.Fatal(err)
	}
	if pairs != 3 {
		t.Errorf("pairs = %d, want 3", pairs)
	}
	for _, want := range []string{`  "wrong": "right"`, `  "other": "fine"`, `  "third": "good"`} {
		if !strings.Contains(got, want+"\n") {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestSubstitutionDropsUnusableEntries(t *testing.T) {
	source := strings.Join([]string{
		"same=same",                  // no change to suggest
		"empty=",                     // nothing to suggest
		"tagged=value: with a colon", // a colon marks an annotation
		"commented=word\tcommentary", // upstream appended prose to the entry
	}, "\n")

	got, pairs, err := Substitution(strings.NewReader(source), SubstitutionOptions{
		Provenance: "test", Message: "m", Level: "warning",
	})
	if err != nil {
		t.Fatal(err)
	}
	if pairs != 0 {
		t.Errorf("pairs = %d, want 0, in:\n%s", pairs, got)
	}
}

func TestSubstitutionKeepsTheFirstEntryForAKey(t *testing.T) {
	// A key is settled by its first appearance even when that entry is
	// unusable. Letting a later line win would change which rule fires from
	// one upstream reordering to the next.
	source := "word=\nword=replacement\nother=first\nother=second"

	got, _, err := Substitution(strings.NewReader(source), SubstitutionOptions{
		Provenance: "test", Message: "m", Level: "warning",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "replacement") {
		t.Errorf("a later entry overrode an earlier one:\n%s", got)
	}
	if !strings.Contains(got, `  "other": "first"`) {
		t.Errorf("the first entry for a repeated key was lost:\n%s", got)
	}
	if strings.Contains(got, "second") {
		t.Errorf("a repeated key produced two entries:\n%s", got)
	}
}

func TestSubstitutionBareKeysOnly(t *testing.T) {
	source := "one word=fine\nsingle=fine"

	got, pairs, err := Substitution(strings.NewReader(source), SubstitutionOptions{
		Provenance: "test", Message: "m", Level: "warning", BareKeysOnly: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if pairs != 1 || strings.Contains(got, "one word") {
		t.Errorf("a multi-word key survived:\n%s", got)
	}
}

func TestSubstitutionDropsContextAnnotations(t *testing.T) {
	source := "word=ctx:something\nother=fine"

	_, pairs, err := Substitution(strings.NewReader(source), SubstitutionOptions{
		Provenance: "test", Message: "m", Level: "warning", DropContext: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if pairs != 1 {
		t.Errorf("pairs = %d, want 1", pairs)
	}
}

func TestSubstitutionHeader(t *testing.T) {
	got, _, err := Substitution(strings.NewReader(""), SubstitutionOptions{
		Note: "why this list exists", Provenance: "from somewhere", Message: "say '%s'", Level: "suggestion",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "# why this list exists\n" +
		"# from somewhere\n" +
		"extends: substitution\n" +
		"message: \"say '%s'\"\n" +
		"level: suggestion\n" +
		"ignorecase: true\n" +
		"swap:\n"
	if got != want {
		t.Errorf("header:\n%q\nwant:\n%q", got, want)
	}
}
