package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const dataDir = "data"

// Load reads a JSON array from data/<filename> into a slice of T.
// If the file does not exist, an empty slice is returned with no error.
func Load[T any](filename string) ([]T, error) {
	path := filepath.Join(dataDir, filename)
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []T{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var items []T
	if err := json.NewDecoder(f).Decode(&items); err != nil {
		// Empty or malformed file — return empty slice.
		return []T{}, nil
	}
	return items, nil
}

// Save writes a slice of T as a pretty-printed JSON array to data/<filename>.
// It creates the data/ directory if it does not already exist.
func Save[T any](filename string, items []T) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	out, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dataDir, filename)
	return os.WriteFile(path, out, 0644)
}
