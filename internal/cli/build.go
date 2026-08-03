package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/themaiby/stet/internal/build"
)

func runBuild(e *env, args []string) int {
	mode := "run"
	codes := "all"
	for _, arg := range args {
		switch arg {
		case "--detach":
			mode = "detach"
		case "--status":
			mode = "status"
		default:
			codes = arg
		}
	}

	runner := build.New(e.Layout, e.Err)

	if mode == "status" {
		fmt.Fprintln(e.Out, runner.State())
		return 0
	}

	requested := strings.Split(codes, ",")
	if codes == "all" {
		languages, err := loadLanguages(e)
		if err != nil {
			fmt.Fprintln(e.Err, err)
			return 1
		}
		requested = languages.Codes()
	}

	if mode == "detach" {
		if err := detach(e, requested); err != nil {
			fmt.Fprintln(e.Err, err)
			return 1
		}
		return 0
	}

	if err := runner.Run(requested); err != nil {
		fmt.Fprintln(e.Err, err)
		return 1
	}
	return 0
}

// detach starts the build in the background and returns at once, so that a
// session hook can warm the rules up without holding the session.
//
// A lock left by a killed run must not wedge the warm-up forever, so a recorded
// process is believed only while it is still alive.
func detach(e *env, codes []string) error {
	if running(e.Layout.PID()) {
		return nil
	}
	self, err := os.Executable()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(e.Layout.Data, 0o755); err != nil {
		return err
	}
	log, err := os.OpenFile(e.Layout.Log(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer log.Close()

	cmd := exec.Command(self, "build", strings.Join(codes, ","))
	cmd.Stdout, cmd.Stderr = log, log
	detachProcess(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := os.WriteFile(e.Layout.PID(), []byte(strconv.Itoa(cmd.Process.Pid)), 0o644); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func running(pidFile string) bool {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false
	}
	return processAlive(pid)
}
