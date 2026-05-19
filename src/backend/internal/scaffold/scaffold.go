// Package scaffold initialises a new AI Engine workspace in a given directory.
// It creates the .ai-engine/ folder structure with the minimum required files.
package scaffold

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

//go:embed templates/engine_context.md templates/swarmito_system_prompt.md templates/swarmito_agent.json
var templates embed.FS

const configJSON = `{
  "provider": "",
  "default_model": "",
  "root_agent": "swarmito",
  "port": 8080,
  "max_tool_retries": 3,
  "max_tool_calls": 50
}
`

const modelPricingJSON = `{
  "claude-sonnet-4-5": {
    "input_per_million": 3.00,
    "output_per_million": 15.00,
    "currency": "USD"
  },
  "claude-sonnet-4-6": {
    "input_per_million": 3.00,
    "output_per_million": 15.00,
    "currency": "USD"
  },
  "claude-opus-4": {
    "input_per_million": 15.00,
    "output_per_million": 75.00,
    "currency": "USD"
  },
  "claude-haiku-3-5": {
    "input_per_million": 0.80,
    "output_per_million": 4.00,
    "currency": "USD"
  }
}
`

const dotEnv = `ANTHROPIC_API_KEY=
`

// Init creates the .ai-engine/ workspace structure inside dir.
// It returns an error if any file already exists (safe — never overwrites).
func Init(dir string) error {
	base := filepath.Join(dir, ".ai-engine")

	// Directories to create.
	dirs := []string{
		base,
		filepath.Join(base, "agents", "swarmito"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("scaffold: failed to create directory %q: %w", d, err)
		}
	}

	// Files to write from inline content.
	inlineFiles := map[string]string{
		filepath.Join(base, "config.json"):        configJSON,
		filepath.Join(base, ".env"):               dotEnv,
		filepath.Join(base, "model-pricing.json"): modelPricingJSON,
	}
	for path, content := range inlineFiles {
		if err := writeNew(path, []byte(content)); err != nil {
			return err
		}
	}

	// Files to write from embedded templates.
	embeddedFiles := map[string]string{
		filepath.Join(base, "engine_context.md"):                          "templates/engine_context.md",
		filepath.Join(base, "agents", "swarmito", "system_prompt.md"):     "templates/swarmito_system_prompt.md",
		filepath.Join(base, "agents", "swarmito", "agent.json"):           "templates/swarmito_agent.json",
	}
	for dst, src := range embeddedFiles {
		data, err := templates.ReadFile(src)
		if err != nil {
			return fmt.Errorf("scaffold: failed to read embedded template %q: %w", src, err)
		}
		if err := writeNew(dst, data); err != nil {
			return err
		}
	}

	return nil
}

// writeNew writes data to path only if the file does not already exist.
// If the file already exists, it logs a message and returns nil (safe to re-run init).
func writeNew(path string, data []byte) error {
	if _, err := os.Stat(path); err == nil {
		// File already exists — skip silently (safe to re-run init).
		log.Printf("scaffold: skipping existing file: %s", path)
		return nil
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("scaffold: failed to write %q: %w", path, err)
	}
	return nil
}
