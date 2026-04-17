// Package config holds filesystem paths and user configuration for Armada.
//
// By default Armada keeps everything under ~/.armada:
//
//	~/.armada/bin/         — symlinks to installed binaries (add to $PATH)
//	~/.armada/pkgs/        — extracted package contents (one dir per tool)
//	~/.armada/cache/       — downloaded archives & repo clones
//	~/.armada/state.json   — installed package index
//	~/.armada/config.yaml  — user configuration (repositories, etc.)
//
// The environment variable ARMADA_HOME overrides the default location.
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// EnvHome is the environment variable that overrides the default Armada
// home directory.
const EnvHome = "ARMADA_HOME"

// DefaultDirName is the name of the home directory created under the user's
// home when EnvHome is unset.
const DefaultDirName = ".armada"

// Paths describes the on-disk layout used by Armada.
type Paths struct {
	Home       string // e.g. /home/user/.armada
	Bin        string // Home/bin
	Packages   string // Home/pkgs
	Cache      string // Home/cache
	Repos      string // Home/cache/repos
	State      string // Home/state.json
	ConfigFile string // Home/config.yaml
}

// ResolvePaths returns the Paths struct based on ARMADA_HOME or the default
// under the user's home directory.
func ResolvePaths() (Paths, error) {
	var home string
	if v := os.Getenv(EnvHome); v != "" {
		home = v
	} else {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return Paths{}, fmt.Errorf("cannot determine user home directory: %w", err)
		}
		home = filepath.Join(userHome, DefaultDirName)
	}
	return paths(home), nil
}

func paths(home string) Paths {
	return Paths{
		Home:       home,
		Bin:        filepath.Join(home, "bin"),
		Packages:   filepath.Join(home, "pkgs"),
		Cache:      filepath.Join(home, "cache"),
		Repos:      filepath.Join(home, "cache", "repos"),
		State:      filepath.Join(home, "state.json"),
		ConfigFile: filepath.Join(home, "config.yaml"),
	}
}

// EnsureDirs creates the home directory layout if it does not already exist.
func (p Paths) EnsureDirs() error {
	for _, d := range []string{p.Home, p.Bin, p.Packages, p.Cache, p.Repos} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", d, err)
		}
	}
	return nil
}

// Config is the user-level configuration, stored in config.yaml.
type Config struct {
	// Repositories is the ordered list of repositories Armada consults to
	// resolve manifests. Later entries override earlier entries.
	Repositories []Repository `yaml:"repositories"`
}

// Repository describes a source of manifests.
type Repository struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	Type     string `yaml:"type"`     // "git" or "http"
	Priority int    `yaml:"priority"` // higher wins
}

// Default returns the default configuration used when no config.yaml exists.
func Default() Config {
	return Config{
		Repositories: []Repository{
			{
				Name:     "core",
				URL:      "https://github.com/DiegoDev2/armada-registry",
				Type:     "git",
				Priority: 100,
			},
		},
	}
}

// Load reads the configuration from the given path. If the file does not
// exist, Load returns the Default configuration and nil. If the file
// exists but is empty, Default is also returned so first-time users get
// the seeded repository list. Once the file has any explicit content
// (even an empty repositories list) it is honoured verbatim so a user
// who intentionally removes every repository is not silently overridden.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return Default(), nil
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

// Save writes the configuration to the given path.
func (c Config) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// AddRepository appends r if no repository with the same name exists and
// returns a bool indicating whether the config was modified.
func (c *Config) AddRepository(r Repository) bool {
	for _, existing := range c.Repositories {
		if existing.Name == r.Name {
			return false
		}
	}
	c.Repositories = append(c.Repositories, r)
	return true
}

// RemoveRepository removes the repository with the given name and returns
// a bool indicating whether the config was modified.
func (c *Config) RemoveRepository(name string) bool {
	out := c.Repositories[:0]
	removed := false
	for _, r := range c.Repositories {
		if r.Name == name {
			removed = true
			continue
		}
		out = append(out, r)
	}
	c.Repositories = out
	return removed
}
