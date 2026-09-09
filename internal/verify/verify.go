// Package verify runs repo verify commands against a worktree checkout.
package verify

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Result is the outcome of running verify commands.
type Result struct {
	OK  bool
	Log string
}

// Run executes each command in worktreeDir via the system shell.
func Run(worktreeDir string, commands []string) Result {
	if len(commands) == 0 {
		return Result{OK: true, Log: ""}
	}
	var log bytes.Buffer
	for _, c := range commands {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", c)
		} else {
			cmd = exec.Command("sh", "-c", c)
		}
		cmd.Dir = worktreeDir
		out, err := cmd.CombinedOutput()
		log.Write(out)
		if err != nil {
			fmt.Fprintf(&log, "\nverify failed: %v\n", err)
			return Result{OK: false, Log: log.String()}
		}
	}
	return Result{OK: true, Log: log.String()}
}
