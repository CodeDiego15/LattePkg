package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	internalValidate "github.com/DiegoDev2/Fleet/internal/core/manifest"
	"github.com/DiegoDev2/Fleet/pkg/manifest"
	"gopkg.in/yaml.v3"
)

// defaultHTTPTimeout is applied to each individual request the
// HTTPRepository makes when syncing. The timeout is per-request (not for
// the whole Sync call) so larger repositories can still complete.
const defaultHTTPTimeout = 30 * time.Second

// HTTPRepository implements the Repository interface for repositories
// served over plain HTTP(S).
type HTTPRepository struct {
	name      string
	url       string
	cacheDir  string
	priority  int
	lastSync  time.Time
	manifests map[string]*manifest.Manifest
	client    *http.Client
}

// NewHTTPRepository creates a new HTTP-backed repository.
func NewHTTPRepository(name, url, cacheDir string, priority int) (*HTTPRepository, error) {
	return &HTTPRepository{
		name:      name,
		url:       url,
		cacheDir:  cacheDir,
		priority:  priority,
		manifests: make(map[string]*manifest.Manifest),
		client:    &http.Client{Timeout: defaultHTTPTimeout},
	}, nil
}

// GetName returns the repository name.
func (r *HTTPRepository) GetName() string {
	return r.name
}

// GetURL returns the repository URL.
func (r *HTTPRepository) GetURL() string {
	return r.url
}

// GetType returns the repository type.
func (r *HTTPRepository) GetType() string {
	return "http"
}

// GetPriority returns the configured repository priority.
func (r *HTTPRepository) GetPriority() int {
	return r.priority
}

// GetLastSync returns the timestamp of the last successful sync.
func (r *HTTPRepository) GetLastSync() time.Time {
	return r.lastSync
}

// Sync fetches the index and every referenced manifest from the remote
// repository.
func (r *HTTPRepository) Sync() error {
	manifestsDir := filepath.Join(r.cacheDir, "manifests")
	if err := os.MkdirAll(manifestsDir, 0o755); err != nil {
		return fmt.Errorf("create manifests dir: %w", err)
	}

	indexURL := fmt.Sprintf("%s/index.json", r.url)
	resp, err := r.client.Get(indexURL)
	if err != nil {
		return fmt.Errorf("fetch manifest index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch manifest index: status %d", resp.StatusCode)
	}

	var index struct {
		Manifests []string `json:"manifests"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		return fmt.Errorf("decode manifest index: %w", err)
	}

	next := make(map[string]*manifest.Manifest, len(index.Manifests))
	for _, manifestName := range index.Manifests {
		m, err := r.fetchManifest(manifestsDir, manifestName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "armada: %s: %v\n", manifestName, err)
			continue
		}
		if m == nil {
			continue
		}
		next[m.Name] = m
	}

	r.manifests = next
	r.lastSync = time.Now()
	return nil
}

// fetchManifest downloads a single manifest into the cache and returns the
// parsed+validated value, or nil if the manifest should be skipped.
func (r *HTTPRepository) fetchManifest(manifestsDir, manifestName string) (*manifest.Manifest, error) {
	manifestURL := fmt.Sprintf("%s/manifests/%s.yaml", r.url, manifestName)
	resp, err := r.client.Get(manifestURL)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	manifestPath := filepath.Join(manifestsDir, manifestName+".yaml")
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		return nil, fmt.Errorf("cache manifest: %w", err)
	}

	var m manifest.Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	if errs := internalValidate.Validate(&m); len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "armada: skipping invalid manifest %s in repository %s\n", manifestName, r.name)
		return nil, nil
	}

	return &m, nil
}

// GetManifest looks up a manifest by tool name.
func (r *HTTPRepository) GetManifest(toolName string) (*manifest.Manifest, error) {
	if m, exists := r.manifests[toolName]; exists {
		return m, nil
	}
	return nil, fmt.Errorf("manifest not found: %s", toolName)
}

// ListManifests returns all manifests currently cached in the repository.
func (r *HTTPRepository) ListManifests() ([]*manifest.Manifest, error) {
	manifests := make([]*manifest.Manifest, 0, len(r.manifests))
	for _, m := range r.manifests {
		manifests = append(manifests, m)
	}
	return manifests, nil
}
