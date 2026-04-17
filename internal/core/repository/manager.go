// Package repository aggregates manifest sources (git clones, HTTP
// indexes) and caches them on disk.
package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/DiegoDev2/Fleet/internal/integrity"
	"github.com/DiegoDev2/Fleet/pkg/manifest"
)

// Manager owns a collection of Repository instances and an on-disk cache.
type Manager struct {
	repositories map[string]Repository
	cacheDir     string
	cache        *Cache
	mutex        sync.RWMutex
}

// NewManager creates a Manager backed by the cache directory cacheDir.
// The directory is created if it does not already exist.
func NewManager(cacheDir string) (*Manager, error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}
	cache, err := NewCache(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("init cache: %w", err)
	}
	return &Manager{
		repositories: make(map[string]Repository),
		cacheDir:     cacheDir,
		cache:        cache,
	}, nil
}

// AddRepository registers a new repository with the Manager.
func (m *Manager) AddRepository(name, url, repoType string, priority int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.repositories[name]; exists {
		return fmt.Errorf("repository %q already exists", name)
	}

	repoCacheDir := filepath.Join(m.cacheDir, name)
	if err := os.MkdirAll(repoCacheDir, 0o755); err != nil {
		return fmt.Errorf("create repo dir: %w", err)
	}

	var (
		repo Repository
		err  error
	)
	switch repoType {
	case "git":
		repo, err = NewGitRepository(name, url, repoCacheDir, priority)
	case "http":
		repo, err = NewHTTPRepository(name, url, repoCacheDir, priority)
	default:
		return fmt.Errorf("unsupported repository type %q", repoType)
	}
	if err != nil {
		return err
	}

	m.repositories[name] = repo
	m.cache.AddRepo(name, url, repoType)
	if err := m.cache.SaveMetadata(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: save cache metadata: %v\n", err)
	}
	return nil
}

// RemoveRepository deletes a repository and its on-disk cache directory.
func (m *Manager) RemoveRepository(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.repositories[name]; !exists {
		return fmt.Errorf("repository %q not found", name)
	}
	delete(m.repositories, name)

	repoCacheDir := filepath.Join(m.cacheDir, name)
	if err := os.RemoveAll(repoCacheDir); err != nil {
		return fmt.Errorf("remove repo dir: %w", err)
	}

	m.cache.RemoveRepo(name)
	if err := m.cache.SaveMetadata(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: save cache metadata: %v\n", err)
	}
	return nil
}

// SyncRepository refreshes a repository from its remote source and caches
// the resulting manifests.
func (m *Manager) SyncRepository(name string) error {
	m.mutex.RLock()
	repo, exists := m.repositories[name]
	m.mutex.RUnlock()
	if !exists {
		return fmt.Errorf("repository %q not found", name)
	}

	if err := repo.Sync(); err != nil {
		return err
	}

	manifests, err := repo.ListManifests()
	if err != nil {
		return fmt.Errorf("list manifests: %w", err)
	}

	for _, mf := range manifests {
		data, err := manifest.ToYAML(mf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: marshal %s: %v\n", mf.Name, err)
			continue
		}
		sum := integrity.HashBytes("sha256", data)
		if err := m.cache.AddManifest(mf.Name, name, sum, data); err != nil {
			fmt.Fprintf(os.Stderr, "warning: cache %s: %v\n", mf.Name, err)
		}
	}

	if err := m.cache.SaveMetadata(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: save cache metadata: %v\n", err)
	}
	return nil
}

// SyncAllRepositories syncs every registered repository concurrently and
// returns the first error (if any). Remaining errors are attached via
// error wrapping.
func (m *Manager) SyncAllRepositories() error {
	m.mutex.RLock()
	names := make([]string, 0, len(m.repositories))
	for name := range m.repositories {
		names = append(names, name)
	}
	m.mutex.RUnlock()

	var wg sync.WaitGroup
	errCh := make(chan error, len(names))

	for _, name := range names {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			if err := m.SyncRepository(n); err != nil {
				errCh <- fmt.Errorf("sync %s: %w", n, err)
			}
		}(name)
	}

	wg.Wait()
	close(errCh)

	var errs []string
	for err := range errCh {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}

// GetManifest resolves a manifest by tool name, trying the on-disk cache
// first and falling back to the highest-priority repository that knows
// about it.
func (m *Manager) GetManifest(toolName string) (*manifest.Manifest, error) {
	if cached, err := m.cache.GetManifest(toolName); err == nil {
		return cached, nil
	}

	m.mutex.RLock()
	repos := make([]Repository, 0, len(m.repositories))
	for _, repo := range m.repositories {
		repos = append(repos, repo)
	}
	m.mutex.RUnlock()

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].GetPriority() > repos[j].GetPriority()
	})

	for _, repo := range repos {
		mf, err := repo.GetManifest(toolName)
		if err == nil && mf != nil {
			if data, err := manifest.ToYAML(mf); err == nil {
				_ = m.cache.AddManifest(toolName, repo.GetName(), integrity.HashBytes("sha256", data), data)
				_ = m.cache.SaveMetadata()
			}
			return mf, nil
		}
	}
	return nil, fmt.Errorf("manifest not found for %q", toolName)
}

// ListAllManifests returns every manifest known across all registered
// repositories. Duplicate names from lower-priority repositories are
// dropped in favour of the highest priority entry.
func (m *Manager) ListAllManifests() ([]*manifest.Manifest, error) {
	m.mutex.RLock()
	repos := make([]Repository, 0, len(m.repositories))
	for _, repo := range m.repositories {
		repos = append(repos, repo)
	}
	m.mutex.RUnlock()

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].GetPriority() > repos[j].GetPriority()
	})

	seen := make(map[string]bool)
	var out []*manifest.Manifest
	for _, repo := range repos {
		list, err := repo.ListManifests()
		if err != nil {
			return nil, fmt.Errorf("list %s: %w", repo.GetName(), err)
		}
		for _, mf := range list {
			if seen[mf.Name] {
				continue
			}
			seen[mf.Name] = true
			out = append(out, mf)
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// SearchManifests returns every manifest whose name, description or
// category matches pattern (case-insensitive substring match).
func (m *Manager) SearchManifests(pattern string) ([]*manifest.Manifest, error) {
	all, err := m.ListAllManifests()
	if err != nil {
		return nil, err
	}
	needle := strings.ToLower(pattern)
	var results []*manifest.Manifest
	for _, mf := range all {
		if strings.Contains(strings.ToLower(mf.Name), needle) ||
			strings.Contains(strings.ToLower(mf.Description), needle) {
			results = append(results, mf)
			continue
		}
		for _, cat := range mf.Categories {
			if strings.Contains(strings.ToLower(cat), needle) {
				results = append(results, mf)
				break
			}
		}
	}
	return results, nil
}

// GetRepositoryByName returns a repository by name.
func (m *Manager) GetRepositoryByName(name string) (Repository, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	repo, exists := m.repositories[name]
	return repo, exists
}

// ClearCache wipes the on-disk cache.
func (m *Manager) ClearCache() error {
	return m.cache.ClearCache()
}
