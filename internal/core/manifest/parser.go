// Package manifest provides manifest parsing, validation and
// management utilities for Armada. It builds on the public types from
// github.com/DiegoDev2/Fleet/pkg/manifest.
package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DiegoDev2/Fleet/pkg/manifest"
	"gopkg.in/yaml.v3"
)

// ParseFile reads and parses a manifest YAML file.
func ParseFile(path string) (*manifest.Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest %s: %w", path, err)
	}
	return Parse(data)
}

// Parse decodes YAML bytes into a Manifest.
func Parse(data []byte) (*manifest.Manifest, error) {
	var m manifest.Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return &m, nil
}

// ParseDirectory recursively parses every .yaml/.yml file under dirPath.
// Unparseable files are returned as errors instead of being silently
// skipped.
func ParseDirectory(dirPath string) (map[string]*manifest.Manifest, error) {
	manifests := make(map[string]*manifest.Manifest)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}
		m, err := ParseFile(path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		if m.Name == "" {
			return fmt.Errorf("%s: manifest is missing the 'name' field", path)
		}
		manifests[m.Name] = m
		return nil
	})
	if err != nil {
		return nil, err
	}
	return manifests, nil
}
