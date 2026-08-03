//go:build !windows

package cli

import (
	"os"
	"os/exec"
	"syscall"
)

// detachProcess puts the child in its own process group so that closing the
// session does not take the warm-up down with it.
func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

// processAlive reports whether a recorded warm-up is still running. Signal zero
// asks the question without sending anything.
func processAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}
