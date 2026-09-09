// Package worktree resolves final worktree directories (sibling default, abs/rel roots).
package worktree

import (
	"os"
	"path/filepath"
	"strings"
)

// EnvWorktreeRoot overrides worktree root when set (shared-parent semantics for multi-repo).
const EnvWorktreeRoot = "MERGESIDING_WORKTREE_ROOT"

// EnvWorktreeRootLegacy is accepted for migration from agent-git.
const EnvWorktreeRootLegacy = "AGENT_GIT_WORKTREE_ROOT"

// ResolveInput carries optional overrides for one repo+slug resolution.
type ResolveInput struct {
	RepoRoot       string  // absolute or cleanable repo path
	Slug           string  // task slug
	SharedRoot     *string // CLI/MCP flag; relative → repo root; multi-repo uses shared/<repo>/<slug>
	RepoConfigRoot *string // from .mergesiding.json / .agent-git.json worktreeRoot
	MultiRepo      bool    // when SharedRoot set and true, append repo name before slug
}

// ResolveWorktreePath returns the final worktree directory for one repo+slug.
func ResolveWorktreePath(in ResolveInput) (string, error) {
	repoRoot, err := filepath.Abs(in.RepoRoot)
	if err != nil {
		return "", err
	}
	repoName := filepath.Base(repoRoot)

	if in.SharedRoot != nil && strings.TrimSpace(*in.SharedRoot) != "" {
		root, err := resolveAgainst(repoRoot, *in.SharedRoot)
		if err != nil {
			return "", err
		}
		if in.MultiRepo {
			return filepath.Join(root, repoName, in.Slug), nil
		}
		return filepath.Join(root, in.Slug), nil
	}

	if env := firstNonEmpty(os.Getenv(EnvWorktreeRoot), os.Getenv(EnvWorktreeRootLegacy)); env != "" {
		root, err := resolveAgainst(repoRoot, env)
		if err != nil {
			return "", err
		}
		if in.MultiRepo {
			return filepath.Join(root, repoName, in.Slug), nil
		}
		return filepath.Join(root, in.Slug), nil
	}

	if in.RepoConfigRoot != nil && strings.TrimSpace(*in.RepoConfigRoot) != "" {
		root, err := resolveAgainst(repoRoot, *in.RepoConfigRoot)
		if err != nil {
			return "", err
		}
		return filepath.Join(root, in.Slug), nil
	}

	parent := filepath.Dir(repoRoot)
	return filepath.Join(parent, ".agent-git-worktrees", repoName, in.Slug), nil
}

func resolveAgainst(repoRoot, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if filepath.IsAbs(raw) {
		return filepath.Clean(raw), nil
	}
	return filepath.Abs(filepath.Join(repoRoot, raw))
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
