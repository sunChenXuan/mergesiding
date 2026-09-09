// Package config loads per-repo mergesiding settings from JSON.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ConfigFilePrimary is the preferred repo config filename.
const ConfigFilePrimary = ".mergesiding.json"

// ConfigFileLegacy is accepted for migration from agent-git.
const ConfigFileLegacy = ".agent-git.json"

// Config holds per-repo settings.
type Config struct {
	IntegrationBranch string   `json:"integrationBranch"`
	Verify            []string `json:"verify"`
	WorktreeRoot      string   `json:"worktreeRoot"`
}

type rawConfig struct {
	IntegrationBranch string   `json:"integrationBranch"`
	Verify            []string `json:"verify"`
	WorktreeRoot      string   `json:"worktreeRoot"`
}

// Load reads .mergesiding.json or falls back to .agent-git.json; missing file yields empty config.
func Load(repoRoot string) (Config, error) {
	for _, name := range []string{ConfigFilePrimary, ConfigFileLegacy} {
		path := filepath.Join(repoRoot, name)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return Config{}, err
		}
		var raw rawConfig
		if err := json.Unmarshal(data, &raw); err != nil {
			return Config{}, err
		}
		cfg := Config{
			IntegrationBranch: raw.IntegrationBranch,
			Verify:            raw.Verify,
			WorktreeRoot:      raw.WorktreeRoot,
		}
		if cfg.Verify == nil {
			cfg.Verify = []string{}
		}
		return cfg, nil
	}
	return Config{Verify: []string{}}, nil
}
