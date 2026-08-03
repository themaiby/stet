// Package warmup models the state of the background rule build.
//
// The state is a single line on disk so that any process can read it without a
// lock. A failed build has to stay visible: silence there would let a rule
// check nothing at all, and the report would read as clean because it was blind.
package warmup

import (
	"fmt"
	"strconv"
	"strings"
)

// Phase is where the build has got to.
type Phase string

const (
	// Cold means nothing has been built yet.
	Cold Phase = "cold"
	// Building means a build is running now.
	Building Phase = "building"
	// Ready means every builder finished.
	Ready Phase = "ready"
	// Failed means a builder stopped, and the rules it feeds are inactive.
	Failed Phase = "failed"
)

// State is one line of the state file: phase, progress, and something to show a
// person.
type State struct {
	Phase   Phase
	Step    int
	Total   int
	Message string
}

// Missing is the state before any build has run.
func Missing() State {
	return State{Phase: Cold, Message: "nothing built yet"}
}

// Parse reads a state line. Anything it cannot read is cold, because a state
// file this tool did not write says nothing about what was built.
func Parse(line string) State {
	parts := strings.SplitN(strings.TrimSpace(line), "|", 4)
	if len(parts) < 4 {
		return Missing()
	}
	step, err := strconv.Atoi(parts[1])
	if err != nil {
		return Missing()
	}
	total, err := strconv.Atoi(parts[2])
	if err != nil {
		return Missing()
	}
	switch Phase(parts[0]) {
	case Cold, Building, Ready, Failed:
	default:
		return Missing()
	}
	return State{Phase: Phase(parts[0]), Step: step, Total: total, Message: parts[3]}
}

// String renders the state line.
func (s State) String() string {
	return fmt.Sprintf("%s|%d|%d|%s", s.Phase, s.Step, s.Total, s.Message)
}

// Partial reports whether a report built on this state is missing rules, which
// the reader has to be told.
func (s State) Partial() bool { return s.Phase == Failed }
