package generate

import (
	"strings"
	"testing"
)

func TestParseExcessKeepsCorpusNotes(t *testing.T) {
	source := "# construction\tregister\tai\thuman\texcess\n" +
		"# machine-written corpus: 100 sentences\n" +
		"# control docs: 200 sentences\n" +
		"delve\tdocs\t0.312\t0.000\t999.00\n"

	excess, notes, err := ParseExcess(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	if got := excess["delve"]["docs"]; got != 999.00 {
		t.Errorf("excess = %v, want 999", got)
	}
	if len(notes) != 3 {
		t.Fatalf("notes = %v, want 3", notes)
	}
	if notes[1] != "machine-written corpus: 100 sentences" {
		t.Errorf("note = %q", notes[1])
	}
}

func TestParseExcessRejectsAShortRow(t *testing.T) {
	// A row this cannot read would otherwise become a zero, and a zero excess
	// drops a construction from every preset without saying so.
	if _, _, err := ParseExcess(strings.NewReader("delve\tdocs\t0.3\n")); err == nil {
		t.Error("a four-column row was accepted")
	}
}

// bands checks the thresholds the measurement defines. They are not chosen, so
// a change here is a change to what the tool claims.
func TestLevelForFollowsTheMeasuredBands(t *testing.T) {
	cases := []struct {
		Excess float64
		Want   string
	}{
		{999, "error"}, {8.0, "error"}, {7.9, "warning"},
		{4.0, "warning"}, {3.9, "suggestion"}, {2.0, "suggestion"},
		{1.9, ""}, {0, ""},
	}
	for _, c := range cases {
		if got := levelFor(c.Excess); got != c.Want {
			t.Errorf("levelFor(%v) = %q, want %q", c.Excess, got, c.Want)
		}
	}
}

func TestPresetsSendWhatIsHighEverywhereToTheBase(t *testing.T) {
	patterns := []Pattern{
		{"everywhere", `\beverywhere\b`},
		{"press only", `\bpress\b`},
	}
	excess := Excess{
		"everywhere": {"A": 9.0, "H": 5.0},
		"press only": {"A": 9.0, "H": 1.0},
	}

	set := Presets(patterns, excess, nil, "UK", "uk-excess.tsv")

	var base, press []string
	for _, f := range set.Files {
		switch f.Dir {
		case "UKBase":
			base = append(base, f.Content)
		case "UKPress":
			press = append(press, f.Content)
		}
	}
	if len(base) != 1 || !strings.Contains(base[0], "everywhere") {
		t.Errorf("the construction high in every register did not reach the base: %v", base)
	}
	if len(base) == 1 && strings.Contains(base[0], "press only") {
		t.Error("a construction high in one register reached the base")
	}
	if len(press) == 0 || !strings.Contains(press[0], "press only") {
		t.Errorf("the register-specific construction did not reach its preset: %v", press)
	}
	// The base takes the lower of the two, 5.0, which is a warning.
	if !strings.Contains(base[0], "level: warning") {
		t.Errorf("the base took the wrong level:\n%s", base[0])
	}
}

func TestPresetsSkipTheBaseWhenOneRegisterWasMeasured(t *testing.T) {
	// A base means "high in every register", which says nothing when only one
	// register exists.
	set := Presets(
		[]Pattern{{"delve", `\bdelve\b`}},
		Excess{"delve": {"docs": 999}},
		nil, "EN", "en-excess.tsv")

	for _, f := range set.Files {
		if f.Dir == "ENBase" {
			t.Errorf("a single measured register produced a base style:\n%s", f.Content)
		}
	}
	if len(set.Rows) != 1 || !strings.HasPrefix(set.Rows[0], "en|docs|ENDocs|") {
		t.Errorf("registry rows = %v", set.Rows)
	}
}

func TestPresetsMarkPreliminaryRegisters(t *testing.T) {
	// High in fiction alone, so it belongs to that preset rather than the base.
	set := Presets(
		[]Pattern{{"a", "a"}},
		Excess{"a": {"fiction": 9.0, "A": 1.0, "H": 1.0}},
		nil, "UK", "uk-excess.tsv")

	var row string
	for _, r := range set.Rows {
		if strings.Contains(r, "fiction") {
			row = r
		}
	}
	if !strings.HasSuffix(row, "|preliminary") {
		t.Errorf("the fiction row is not marked preliminary: %q", row)
	}
}

func TestPresetsSplitBySeverity(t *testing.T) {
	set := Presets(
		[]Pattern{{"loud", "loud"}, {"quiet", "quiet"}},
		Excess{"loud": {"docs": 9.0}, "quiet": {"docs": 2.5}},
		nil, "EN", "en-excess.tsv")

	names := map[string]bool{}
	for _, f := range set.Files {
		names[f.Name] = true
	}
	if !names["Slop-error.yml"] || !names["Slop-suggestion.yml"] {
		t.Errorf("two severities did not produce two files: %v", names)
	}
	if names["Slop.yml"] {
		t.Error("a split set also wrote the unsplit name")
	}
}

func TestMergeRegistryKeepsOtherLanguages(t *testing.T) {
	existing := []string{
		"# a header",
		"uk|docs|UKDocs|документація",
		"en|docs|ENDocs|superseded",
	}
	got := MergeRegistry(existing, "EN", []string{"en|docs|ENDocs|documentation"})

	if !strings.Contains(got, "uk|docs|UKDocs|документація") {
		t.Errorf("another language's row was dropped:\n%s", got)
	}
	if strings.Contains(got, "superseded") {
		t.Errorf("a stale row for this language survived:\n%s", got)
	}
	if strings.Count(got, "\n") != 3 {
		t.Errorf("unexpected shape:\n%s", got)
	}
}
