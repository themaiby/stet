package warmup

import "testing"

func TestParseRoundTrip(t *testing.T) {
	want := State{Phase: Building, Step: 2, Total: 5, Message: "uk (word forms)"}
	if got := Parse(want.String()); got != want {
		t.Errorf("Parse(%q) = %+v, want %+v", want.String(), got, want)
	}
}

func TestParseRejectsWhatItCannotRead(t *testing.T) {
	// A state file this tool did not write says nothing about what was built,
	// and reading it as ready would let a lint run report on absent rules.
	for _, line := range []string{
		"", "ready", "ready|1", "ready|a|2|message", "elsewhere|1|2|message",
	} {
		if got := Parse(line); got.Phase != Cold {
			t.Errorf("Parse(%q) = %+v, want cold", line, got)
		}
	}
}

func TestOnlyAFailedBuildIsPartial(t *testing.T) {
	if !(State{Phase: Failed}).Partial() {
		t.Error("a failed build did not report itself as partial")
	}
	for _, phase := range []Phase{Cold, Building, Ready} {
		if (State{Phase: phase}).Partial() {
			t.Errorf("%s reported itself as partial", phase)
		}
	}
}
