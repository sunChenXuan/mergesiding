// Package escalate writes conflict/verify escalation markdown and optional resume hooks.
package escalate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sunChenXuan/mergesiding/internal/models"
	"github.com/sunChenXuan/mergesiding/internal/paths"
	"github.com/sunChenXuan/mergesiding/internal/slug"
)

// EnvResumeCmd is an optional shell command with {writer_id} placeholder.
const EnvResumeCmd = "MERGESIDING_RESUME_CMD"

// EnvResumeCmdLegacy is accepted for migration.
const EnvResumeCmdLegacy = "AGENT_GIT_RESUME_CMD"

func escalatePath(p paths.Paths, taskSlug string) (string, error) {
	if err := slug.Validate(taskSlug); err != nil {
		return "", err
	}
	return filepath.Join(p.EscalateDir(), taskSlug+".md"), nil
}

// WriteConflictEscalation writes rebase conflict instructions for the writer.
func WriteConflictEscalation(p paths.Paths, task *models.TaskRecord, binding models.RepoBinding, conflicts []string) (string, error) {
	if err := p.Ensure(); err != nil {
		return "", err
	}
	path, err := escalatePath(p, task.Slug)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Conflict: %s\n\n", task.Slug)
	fmt.Fprintf(&b, "writerId: %s\n", deref(task.WriterID))
	fmt.Fprintf(&b, "brief: %s\n", deref(task.BriefPath))
	fmt.Fprintf(&b, "repo: %s\n", binding.Path)
	fmt.Fprintf(&b, "worktree: %s\n", binding.WorktreePath)
	fmt.Fprintf(&b, "branch: %s\n", binding.Branch)
	fmt.Fprintf(&b, "integration: %s\n\n", binding.IntegrationBranch)
	b.WriteString("## Conflicted files\n")
	for _, c := range conflicts {
		fmt.Fprintf(&b, "- %s\n", c)
	}
	b.WriteString("\n## Instructions for writer\n")
	b.WriteString("1. Open the worktree path above.\n")
	b.WriteString("2. Finish the rebase by resolving ONLY conflict markers.\n")
	b.WriteString("3. `git add` resolved files and `git rebase --continue`.\n")
	fmt.Fprintf(&b, "4. Ensure clean worktree, then `mergesiding ready --slug %s`.\n", task.Slug)
	b.WriteString("5. Do not use --ours/--theirs unless brief says so.\n\n")
	b.WriteString("## Next steps (who runs what)\n")
	fmt.Fprintf(&b, "- **Writer**: after fix → `mergesiding ready --slug %s`\n", task.Slug)
	fmt.Fprintf(&b, "- **Human/scheduler**: then → `mergesiding integrate` (or `mergesiding integrate --slug %s`)\n", task.Slug)
	b.WriteString("- Writer must **not** merge into the integration branch.\n")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// WriteVerifyEscalation writes verify-failure instructions for the writer.
func WriteVerifyEscalation(p paths.Paths, task *models.TaskRecord, binding models.RepoBinding, verifyLog string) (string, error) {
	if err := p.Ensure(); err != nil {
		return "", err
	}
	path, err := escalatePath(p, task.Slug)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Verify failed: %s\n\n", task.Slug)
	fmt.Fprintf(&b, "writerId: %s\n", deref(task.WriterID))
	fmt.Fprintf(&b, "brief: %s\n", deref(task.BriefPath))
	fmt.Fprintf(&b, "repo: %s\n", binding.Path)
	fmt.Fprintf(&b, "worktree: %s\n", binding.WorktreePath)
	fmt.Fprintf(&b, "branch: %s\n", binding.Branch)
	fmt.Fprintf(&b, "integration: %s\n\n", binding.IntegrationBranch)
	b.WriteString("## Verify output\n```\n")
	b.WriteString(verifyLog)
	b.WriteString("\n```\n\n## Instructions for writer\n")
	b.WriteString("1. Open the worktree path above.\n")
	b.WriteString("2. Fix the verify failures shown above.\n")
	fmt.Fprintf(&b, "3. Ensure clean worktree, then `mergesiding ready --slug %s`.\n\n", task.Slug)
	b.WriteString("## Next steps (who runs what)\n")
	fmt.Fprintf(&b, "- **Writer**: after fix → `mergesiding ready --slug %s`\n", task.Slug)
	b.WriteString("- **Human/scheduler**: then → `mergesiding integrate`\n")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// TryResumeWriter runs MERGESIDING_RESUME_CMD with {writer_id} substituted when set.
// writer_id must pass ValidateWriterID; otherwise the hook is skipped to avoid shell injection.
func TryResumeWriter(writerID *string) {
	cmdTpl := os.Getenv(EnvResumeCmd)
	if cmdTpl == "" {
		cmdTpl = os.Getenv(EnvResumeCmdLegacy)
	}
	if cmdTpl == "" {
		return
	}
	id := ""
	if writerID != nil {
		id = *writerID
	}
	if id == "" || slug.ValidateWriterID(id) != nil {
		return
	}
	cmdTpl = strings.ReplaceAll(cmdTpl, "{writer_id}", id)
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdTpl)
	} else {
		cmd = exec.Command("sh", "-c", cmdTpl)
	}
	_ = cmd.Run()
}

// PathFor returns escalate markdown path if it exists.
func PathFor(p paths.Paths, taskSlug string) string {
	path, err := escalatePath(p, taskSlug)
	if err != nil {
		return ""
	}
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

func deref(s *string) string {
	if s == nil || *s == "" {
		return "(none)"
	}
	return *s
}
