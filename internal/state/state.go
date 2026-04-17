// Package state persists information about the tools Armada has installed.
//
// State is stored as a single JSON file (typically ~/.armada/state.json).
// The format is intentionally simple; older states are upgraded by adding
// new optional fields with sensible zero-value defaults.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// State is the top-level document stored on disk.
type State struct {
	Version   int                  `json:"version"`
	Installed map[string]Installed `json:"installed"`
}

// Installed records a single installed tool.
type Installed struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Platform    string    `json:"platform"`
	InstalledAt time.Time `json:"installed_at"`
	Source      string    `json:"source,omitempty"`
	PackageDir  string    `json:"package_dir"`
	Binaries    []string  `json:"binaries"`
}

// Store is a concurrency-safe handle to a state file.
type Store struct {
	path string
	mu   sync.Mutex
	data State
}

// Open loads the state file at path. If the file does not exist an empty
// store backed by path is returned.
func Open(path string) (*Store, error) {
	s := &Store{path: path, data: State{Version: 1, Installed: map[string]Installed{}}}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("read state: %w", err)
	}
	if len(raw) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(raw, &s.data); err != nil {
		return nil, fmt.Errorf("parse state: %w", err)
	}
	if s.data.Installed == nil {
		s.data.Installed = map[string]Installed{}
	}
	if s.data.Version == 0 {
		s.data.Version = 1
	}
	return s, nil
}

// Save persists the store to disk atomically.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.save()
}

func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("rename state: %w", err)
	}
	return nil
}

// Get returns the installation record for name, if any.
func (s *Store) Get(name string) (Installed, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i, ok := s.data.Installed[name]
	return i, ok
}

// Put records an installation. It replaces any previous record with the
// same name. The store is persisted before returning.
func (s *Store) Put(i Installed) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Installed[i.Name] = i
	return s.save()
}

// Delete removes an installation record.
func (s *Store) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data.Installed, name)
	return s.save()
}

// List returns a deterministic snapshot of all installation records sorted
// by name.
func (s *Store) List() []Installed {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Installed, 0, len(s.data.Installed))
	for _, v := range s.data.Installed {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
