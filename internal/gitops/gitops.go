// Package gitops wraps git CLI operations used by mergesiding.
package gitops

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Error is returned when a git command fails.
type Error struct {
	Args   []string
	Stderr string
	Stdout string
}

func (e *Error) Error() string {
	msg := strings.TrimSpace(e.Stderr)
	if msg == "" {
		msg = strings.TrimSpace(e.Stdout)
	}
	if msg == "" {
		msg = fmt.Sprintf("git %s failed", strings.Join(e.Args, " "))
	}
	return msg
}

// Run executes git in cwd with args.
func Run(cwd string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := stdout.String()
	if err != nil {
		return out, &Error{Args: args, Stderr: stderr.String(), Stdout: out}
	}
	return out, nil
}

// RunAllow runs git and returns stdout/stderr/exit without wrapping failure as Error.
func RunAllow(cwd string, args ...string) (stdout, stderr string, code int) {
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	code = 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			code = 1
		}
	}
	return outBuf.String(), errBuf.String(), code
}

// CurrentBranch returns the current HEAD branch name.
func CurrentBranch(cwd string) (string, error) {
	out, err := Run(cwd, "rev-parse", "--abbrev-ref", "HEAD")
	return strings.TrimSpace(out), err
}

// IsClean reports whether the working tree has no uncommitted changes.
func IsClean(cwd string) (bool, error) {
	out, err := Run(cwd, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "", nil
}

// CreateTaskWorktree creates branch at startPoint and checks it out in worktreePath.
func CreateTaskWorktree(repo, worktreePath, branch, startPoint string) error {
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0o755); err != nil {
		return err
	}
	_, err := Run(repo, "worktree", "add", "-b", branch, worktreePath, startPoint)
	return err
}

// RebaseResult is the outcome of a rebase attempt.
type RebaseResult struct {
	OK        bool
	Conflicts []string
	Message   string
}

// ConflictedFiles returns paths with unresolved conflicts.
func ConflictedFiles(cwd string) ([]string, error) {
	out, _, _ := RunAllow(cwd, "diff", "--name-only", "--diff-filter=U")
	var files []string
	for _, ln := range strings.Split(out, "\n") {
		ln = strings.TrimSpace(ln)
		if ln != "" {
			files = append(files, ln)
		}
	}
	return files, nil
}

// RebaseOnto rebases the current branch onto upstream.
func RebaseOnto(cwd, upstream string) RebaseResult {
	_, stderr, code := RunAllow(cwd, "rebase", upstream)
	if code == 0 {
		return RebaseResult{OK: true}
	}
	files, _ := ConflictedFiles(cwd)
	if len(files) > 0 {
		return RebaseResult{OK: false, Conflicts: files, Message: stderr}
	}
	return RebaseResult{OK: false, Message: stderr}
}

// AbortRebase aborts an in-progress rebase if any.
func AbortRebase(cwd string) {
	_, _, _ = RunAllow(cwd, "rebase", "--abort")
}

// CommitsAhead counts commits on taskBranch not in integrationBranch.
func CommitsAhead(repo, integrationBranch, taskBranch string) (int, error) {
	out, err := Run(repo, "rev-list", "--count", integrationBranch+".."+taskBranch)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(out))
}

// MergeFFOrNoFF merges taskBranch into the current branch of integrationRepo.
func MergeFFOrNoFF(integrationRepo, taskBranch string) error {
	_, _, code := RunAllow(integrationRepo, "merge", "--ff-only", taskBranch)
	if code == 0 {
		return nil
	}
	_, err := Run(integrationRepo, "merge", "--no-ff", taskBranch, "-m", "merge "+taskBranch)
	return err
}

// RemoveWorktree removes a linked worktree path.
func RemoveWorktree(repo, worktreePath string) error {
	_, err := Run(repo, "worktree", "remove", "--force", worktreePath)
	return err
}

// DeleteBranch deletes a local branch.
func DeleteBranch(repo, branch string) error {
	_, err := Run(repo, "branch", "-D", branch)
	return err
}
