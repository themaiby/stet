//go:build windows

package cli

import (
	"os"
	"os/exec"
)

// detachProcess is a no-op on Windows, where a child already outlives its
// parent.
func detachProcess(cmd *exec.Cmd) {}

// processAlive reports whether a recorded warm-up is still running. FindProcess
// on Windows fails for a process that has exited, which is the answer.
func processAlive(pid int) bool {
	_, err := os.FindProcess(pid)
	return err == nil
}
