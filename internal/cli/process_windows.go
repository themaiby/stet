//go:build windows

package cli

import (
	"os"
	"os/exec"
)

// detachProcess is a no-op on Windows: a child started this way already
// outlives its parent, and the console flags that would hide it are not worth
// the platform-specific handle work.
func detachProcess(cmd *exec.Cmd) {}

// processAlive reports whether a recorded warm-up is still running. FindProcess
// on Windows fails for a process that has exited, which is the answer.
func processAlive(pid int) bool {
	_, err := os.FindProcess(pid)
	return err == nil
}
