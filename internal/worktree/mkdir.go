package worktree

import "os"

func mkdirAll(p string) error {
	return os.MkdirAll(p, 0o755)
}
