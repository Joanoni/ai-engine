package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DynamicContextConfig controls which dynamic context providers are active.
// If Providers is empty, all registered providers are enabled by default.
type DynamicContextConfig struct {
	Providers []string `json:"providers"`
}

// Config holds the global engine settings parsed from .ai-engine/config.json.
type Config struct {
	Provider       string               `json:"provider"`
	DefaultModel   string               `json:"default_model"`
	Port           int                  `json:"port"`
	MaxToolRetries int                  `json:"max_tool_retries"`
	MaxToolCalls   int                  `json:"max_tool_calls"`
	DynamicContext DynamicContextConfig `json:"dynamic_context"`
}

// Load reads and parses .ai-engine/config.json from the given workspace path.
// It is intended to be called on every session start to support hot-reload.
func Load(workspacePath string) (*Config, error) {
	configPath := filepath.Join(workspacePath, ".ai-engine", "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config: failed to read %s: %w", configPath, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: failed to parse %s: %w", configPath, err)
	}

	if cfg.Port == 0 {
		cfg.Port = 8080
	}

	if cfg.MaxToolRetries == 0 {
		cfg.MaxToolRetries = 3
	}

	if cfg.MaxToolCalls == 0 {
		cfg.MaxToolCalls = 50
	}

	if cfg.Provider == "" {
		return nil, fmt.Errorf("config: 'provider' is required in config.json (e.g., \"anthropic\")")
	}
	if cfg.DefaultModel == "" {
		return nil, fmt.Errorf("config: 'default_model' is required in config.json (e.g., \"claude-sonnet-4-6\")")
	}

	return &cfg, nil
}

// loadEnvFile reads the file at path and sets any variables not already
// present in the environment. It is optional — if the file does not exist,
// loadEnvFile returns nil without error.
func loadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // optional file — not an error
		}
		return fmt.Errorf("config: failed to read %s: %w", path, err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			continue // malformed line — skip silently
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		// Strip surrounding quotes (single or double).
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if key == "" {
			continue
		}
		// Only set if not already defined in the environment.
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("config: failed to set env var %s: %w", key, err)
			}
		}
	}
	return nil
}

// LoadEnv reads .ai-engine/.env from the given workspace path and sets any
// variables not already present in the environment. It is optional — if the
// file does not exist, LoadEnv returns nil without error.
func LoadEnv(workspacePath string) error {
	return loadEnvFile(filepath.Join(workspacePath, ".ai-engine", ".env"))
}

