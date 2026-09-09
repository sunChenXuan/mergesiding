package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMergesidingJSON(t *testing.T) {
	dir := t.TempDir()
	body := `{"integrationBranch":"main","verify":["go test ./..."],"worktreeRoot":"../.agent-git-worktrees/app"}`
	if err := os.WriteFile(filepath.Join(dir, ConfigFilePrimary), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.IntegrationBranch != "main" || len(cfg.Verify) != 1 || cfg.WorktreeRoot == "" {
		t.Fatalf("%+v", cfg)
	}
}

func TestLoadLegacyFallback(t *testing.T) {
	dir := t.TempDir()
	body := `{"integrationBranch":"develop","verify":[]}`
	if err := os.WriteFile(filepath.Join(dir, ConfigFileLegacy), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.IntegrationBranch != "develop" {
		t.Fatalf("%+v", cfg)
	}
}
