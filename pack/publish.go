// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pack

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// PublishConfig configures pack publishing preparation.
type PublishConfig struct {
	// SkillsDir is the directory containing skills to publish.
	SkillsDir string

	// PackName is the name of the skill pack.
	PackName string

	// Version is the pack version (defaults to git commit).
	Version string

	// OutputDir is where to write the publish bundle.
	OutputDir string

	// Strict treats validation warnings as errors.
	Strict bool
}

// PublishBundle represents a prepared pack ready for publishing.
type PublishBundle struct {
	// Manifest contains pack metadata.
	Manifest *PackManifest

	// BundlePath is the path to the tarball.
	BundlePath string

	// Checksum is the SHA256 of the bundle.
	Checksum string

	// Size is the bundle size in bytes.
	Size int64

	// Validation is the validation result.
	Validation *ValidationResult
}

// PackManifest contains metadata for a published pack.
type PackManifest struct {
	// Name is the pack identifier.
	Name string `json:"name"`

	// Version is the pack version.
	Version string `json:"version"`

	// Description is a human-readable description.
	Description string `json:"description,omitempty"`

	// Skills lists the skill names in this pack.
	Skills []string `json:"skills"`

	// Author is the pack author.
	Author string `json:"author,omitempty"`

	// License is the pack license.
	License string `json:"license,omitempty"`

	// Repository is the source repository URL.
	Repository string `json:"repository,omitempty"`

	// Homepage is the pack homepage URL.
	Homepage string `json:"homepage,omitempty"`

	// Keywords are searchable keywords.
	Keywords []string `json:"keywords,omitempty"`

	// CreatedAt is when the pack was created.
	CreatedAt time.Time `json:"created_at"`
}

// PrepareForPublish validates and bundles a pack for publishing.
//
// This creates a tarball containing all skills and a manifest file,
// ready for upload to ClawHub or another registry.
func PrepareForPublish(cfg PublishConfig) (*PublishBundle, error) {
	if cfg.SkillsDir == "" {
		cfg.SkillsDir = "skills"
	}
	if cfg.PackName == "" {
		cfg.PackName = filepath.Base(cfg.SkillsDir)
	}
	if cfg.Version == "" {
		cfg.Version = getGitCommitHash()
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "dist"
	}

	// Validate the pack
	validation, err := ValidatePack(ValidateConfig{
		SkillsDir: cfg.SkillsDir,
		Strict:    cfg.Strict,
	})
	if err != nil {
		return nil, fmt.Errorf("validate pack: %w", err)
	}

	if !validation.Valid {
		return &PublishBundle{
			Validation: validation,
		}, fmt.Errorf("validation failed: %d errors", len(validation.Errors))
	}

	// Create manifest
	manifest := &PackManifest{
		Name:      cfg.PackName,
		Version:   cfg.Version,
		Skills:    validation.Skills,
		CreatedAt: time.Now().UTC(),
	}

	// Ensure output directory exists
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	// Create bundle filename
	bundleName := fmt.Sprintf("%s-%s.tar.gz", cfg.PackName, cfg.Version[:min(8, len(cfg.Version))])
	bundlePath := filepath.Join(cfg.OutputDir, bundleName)

	// Create tarball
	size, err := createBundle(bundlePath, cfg.SkillsDir, manifest)
	if err != nil {
		return nil, fmt.Errorf("create bundle: %w", err)
	}

	return &PublishBundle{
		Manifest:   manifest,
		BundlePath: bundlePath,
		Size:       size,
		Validation: validation,
	}, nil
}

// createBundle creates a tarball containing skills and manifest.
func createBundle(bundlePath, skillsDir string, manifest *PackManifest) (int64, error) {
	f, err := os.Create(bundlePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Add manifest
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("marshal manifest: %w", err)
	}

	if err := addToTar(tw, "manifest.json", manifestData); err != nil {
		return 0, fmt.Errorf("add manifest to tar: %w", err)
	}

	// Add skills
	err = filepath.WalkDir(skillsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(skillsDir, path)
		if err != nil {
			return err
		}

		// Skip root
		if relPath == "." {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.Join("skills", relPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Write file content
		if !d.IsDir() {
			//nolint:gosec // G122: Path from WalkDir is safe within skillsDir
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if _, err := tw.Write(content); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("walk skills: %w", err)
	}

	// Get final size
	info, err := f.Stat()
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

// addToTar adds a file with content to a tar archive.
func addToTar(tw *tar.Writer, name string, content []byte) error {
	header := &tar.Header{
		Name:    name,
		Mode:    0644,
		Size:    int64(len(content)),
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err := tw.Write(content)
	return err
}

// ExtractBundle extracts a pack bundle to a directory.
func ExtractBundle(bundlePath, targetDir string) (*PackManifest, error) {
	f, err := os.Open(bundlePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	var manifest *PackManifest

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		//nolint:gosec // G305: Tar entries are from trusted pack bundle
		targetPath := filepath.Join(targetDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return nil, err
			}

		case tar.TypeReg:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return nil, err
			}

			// Read content
			content, err := io.ReadAll(tr)
			if err != nil {
				return nil, err
			}

			// Parse manifest if this is it
			if header.Name == "manifest.json" {
				var m PackManifest
				if err := json.Unmarshal(content, &m); err != nil {
					return nil, fmt.Errorf("parse manifest: %w", err)
				}
				manifest = &m
			}

			// Write file
			//nolint:gosec // G115: Mode from trusted tar header
			if err := os.WriteFile(targetPath, content, os.FileMode(header.Mode)); err != nil {
				return nil, err
			}
		}
	}

	if manifest == nil {
		return nil, fmt.Errorf("manifest.json not found in bundle")
	}

	return manifest, nil
}
