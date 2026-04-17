// Package installer orchestrates the install and uninstall workflows:
// download → verify checksum → extract → symlink binaries → record state.
package installer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/DiegoDev2/Fleet/internal/archive"
	"github.com/DiegoDev2/Fleet/internal/config"
	"github.com/DiegoDev2/Fleet/internal/download"
	"github.com/DiegoDev2/Fleet/internal/integrity"
	"github.com/DiegoDev2/Fleet/internal/platform"
	"github.com/DiegoDev2/Fleet/internal/state"
	"github.com/DiegoDev2/Fleet/pkg/manifest"
)

// Options tweak installer behaviour.
type Options struct {
	// Force reinstalls a tool even if the same version is already present.
	Force bool
	// Progress receives download progress updates. May be nil.
	Progress download.ProgressFunc
}

// Installer installs and removes tools described by manifest.Manifest
// documents.
type Installer struct {
	paths config.Paths
	store *state.Store
	info  platform.Info
}

// New returns an Installer wired to the given on-disk layout and state
// store.
func New(paths config.Paths, store *state.Store) *Installer {
	return &Installer{
		paths: paths,
		store: store,
		info:  platform.Detect(),
	}
}

// Platform returns the detected platform. Useful for tests.
func (i *Installer) Platform() platform.Info { return i.info }

// Install performs the install workflow for m. It returns the recorded
// state entry on success. Callers are expected to validate m with
// manifest.Validate before calling Install.
func (i *Installer) Install(ctx context.Context, m *manifest.Manifest, opts Options) (state.Installed, error) {
	if !m.UsesAssets() {
		return state.Installed{}, fmt.Errorf("%s: manifest has no assets; cannot install (legacy manifests are supported only by `armada simulate`)", m.Name)
	}

	key := i.info.Key()
	asset, ok := m.AssetFor(key)
	if !ok {
		return state.Installed{}, fmt.Errorf("%s: no asset for platform %s", m.Name, key)
	}

	if existing, already := i.store.Get(m.Name); already && !opts.Force {
		if existing.Version == m.Version {
			return existing, ErrAlreadyInstalled
		}
	}

	if err := i.paths.EnsureDirs(); err != nil {
		return state.Installed{}, err
	}

	// 1. Download to cache.
	cacheFile := filepath.Join(i.paths.Cache, cacheFileName(m, asset))
	if err := download.File(ctx, asset.URL, cacheFile, opts.Progress); err != nil {
		return state.Installed{}, err
	}

	// 2. Verify checksum.
	if err := integrity.VerifyFile(cacheFile, asset.Checksum); err != nil {
		// Remove the bad cache file so the next attempt re-downloads.
		_ = os.Remove(cacheFile)
		return state.Installed{}, err
	}

	// 3. Extract into ~/.armada/pkgs/<name>/<version>/.
	pkgDir := filepath.Join(i.paths.Packages, m.Name, m.Version)
	if err := os.RemoveAll(pkgDir); err != nil {
		return state.Installed{}, fmt.Errorf("clean package dir: %w", err)
	}
	if err := archive.ExtractAsset(cacheFile, pkgDir, asset); err != nil {
		return state.Installed{}, err
	}

	// 4. Link binaries into ~/.armada/bin.
	binaries := m.ResolvedBinaries()
	linked, err := i.linkBinaries(pkgDir, binaries)
	if err != nil {
		return state.Installed{}, err
	}

	// 5. Record state.
	entry := state.Installed{
		Name:        m.Name,
		Version:     m.Version,
		Platform:    key,
		InstalledAt: time.Now().UTC(),
		PackageDir:  pkgDir,
		Binaries:    linked,
	}
	if err := i.store.Put(entry); err != nil {
		return state.Installed{}, err
	}
	return entry, nil
}

// Uninstall removes a tool by name and cleans up its files.
func (i *Installer) Uninstall(name string) error {
	entry, ok := i.store.Get(name)
	if !ok {
		return fmt.Errorf("%s is not installed", name)
	}

	for _, bin := range entry.Binaries {
		link := filepath.Join(i.paths.Bin, bin)
		if err := os.Remove(link); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", link, err)
		}
	}

	if entry.PackageDir != "" {
		if err := os.RemoveAll(entry.PackageDir); err != nil {
			return fmt.Errorf("remove package dir: %w", err)
		}
		// Clean up the parent directory if empty.
		parent := filepath.Dir(entry.PackageDir)
		if entries, err := os.ReadDir(parent); err == nil && len(entries) == 0 {
			_ = os.Remove(parent)
		}
	}

	return i.store.Delete(name)
}

func (i *Installer) linkBinaries(pkgDir string, binaries []string) ([]string, error) {
	out := make([]string, 0, len(binaries))
	suffix := i.info.ExecSuffix()

	for _, bin := range binaries {
		target, err := findBinary(pkgDir, bin, suffix)
		if err != nil {
			return nil, err
		}
		if err := os.Chmod(target, 0o755); err != nil {
			return nil, fmt.Errorf("chmod %s: %w", target, err)
		}
		linkName := filepath.Base(bin)
		if suffix != "" && filepath.Ext(linkName) == "" {
			linkName += suffix
		}
		linkPath := filepath.Join(i.paths.Bin, linkName)
		_ = os.Remove(linkPath)

		if runtime.GOOS == "windows" {
			if err := copyFile(target, linkPath); err != nil {
				return nil, err
			}
		} else {
			if err := os.Symlink(target, linkPath); err != nil {
				return nil, fmt.Errorf("link %s: %w", linkPath, err)
			}
		}
		out = append(out, linkName)
	}
	return out, nil
}

// findBinary looks for bin (optionally with suffix) inside pkgDir. It walks
// the directory tree so manifests only need to declare the binary name and
// not its exact path within the archive.
func findBinary(pkgDir, bin, suffix string) (string, error) {
	candidates := []string{bin}
	if suffix != "" && filepath.Ext(bin) == "" {
		candidates = append(candidates, bin+suffix)
	}

	// Fast path: direct lookup.
	for _, c := range candidates {
		p := filepath.Join(pkgDir, c)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p, nil
		}
	}

	// Slow path: walk.
	var found string
	err := filepath.Walk(pkgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		name := info.Name()
		for _, c := range candidates {
			if name == c || name == filepath.Base(c) {
				found = path
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("search %s: %w", bin, err)
	}
	if found == "" {
		return "", fmt.Errorf("binary %q not found inside %s", bin, pkgDir)
	}
	return found, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := copyReader(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func cacheFileName(m *manifest.Manifest, asset manifest.Asset) string {
	name := fmt.Sprintf("%s-%s-%s", m.Name, m.Version, filepath.Base(asset.URL))
	// Keep it filesystem-safe.
	safe := make([]rune, 0, len(name))
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '.' || r == '-' || r == '_':
			safe = append(safe, r)
		default:
			safe = append(safe, '_')
		}
	}
	return string(safe)
}

// ErrAlreadyInstalled is returned by Install when the requested version is
// already on disk and Force is not set.
var ErrAlreadyInstalled = errors.New("tool already installed at this version")
