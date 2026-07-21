package sync

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//go:embed agent/instructions
var instructionsFS embed.FS

//go:embed agent/skills
var skillsFS embed.FS

//go:embed golint/.golangci.yml
var golintFS embed.FS

const (
	instructionsRoot = "agent/instructions"
	skillsRoot       = "agent/skills"
	golintConfigSrc  = "golint/.golangci.yml"
	golintConfigDest = ".golangci.yml"
)

type Options struct {
	ClaudeOnly bool
	CodexOnly  bool
}

func targetDirs(options Options) ([]string, error) {
	if options.ClaudeOnly && options.CodexOnly {
		return nil, errors.New("--claude-only and --codex-only are mutually exclusive")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}

	switch {
	case options.ClaudeOnly:
		return []string{filepath.Join(home, ".claude")}, nil
	case options.CodexOnly:
		return []string{filepath.Join(home, ".codex")}, nil
	default:
		return []string{filepath.Join(home, ".claude"), filepath.Join(home, ".codex")}, nil
	}
}

// Instructions copies the top-level files under agent/instructions into each
// selected target's root directory (~/.claude and/or ~/.codex), creating the
// directory if it does not exist yet and overwriting files already there.
// AGENTS.md is renamed to CLAUDE.md for the ~/.claude target, since Claude
// Code only reads global memory from CLAUDE.md; ~/.codex already reads
// AGENTS.md, so its filenames are left unchanged.
func Instructions(ctx context.Context, options Options) error {
	targets, err := targetDirs(options)
	if err != nil {
		return err
	}

	entries, err := fs.ReadDir(instructionsFS, instructionsRoot)
	if err != nil {
		return fmt.Errorf("read embedded instructions: %w", err)
	}

	for _, target := range targets {
		if err := os.MkdirAll(target, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", target, err)
		}

		renameAgentsToClaude := filepath.Base(target) == ".claude"
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if renameAgentsToClaude && name == "AGENTS.md" {
				name = "CLAUDE.md"
			}

			src := path.Join(instructionsRoot, entry.Name())
			if err := copyEmbeddedFile(instructionsFS, src, filepath.Join(target, name)); err != nil {
				return err
			}
		}
	}

	return nil
}

// Skills recursively copies every file under each agent/skills/<name>
// directory into <target>/skills/<name>/... for each selected target,
// creating directories as needed and overwriting existing files. Files
// already present at the destination that don't exist in the source are
// left untouched.
func Skills(ctx context.Context, options Options) error {
	targets, err := targetDirs(options)
	if err != nil {
		return err
	}

	for _, target := range targets {
		skillsDir := filepath.Join(target, "skills")
		if err := os.MkdirAll(skillsDir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", skillsDir, err)
		}

		walkErr := fs.WalkDir(skillsFS, skillsRoot, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			rel := strings.TrimPrefix(p, skillsRoot+"/")
			dest := filepath.Join(skillsDir, filepath.FromSlash(rel))
			return copyEmbeddedFile(skillsFS, p, dest)
		})
		if walkErr != nil {
			return fmt.Errorf("sync skills to %s: %w", skillsDir, walkErr)
		}
	}

	return nil
}

// Linters copies the bundled team golangci-lint baseline into the home
// directory as ~/.golangci.yml, overwriting an existing file. golangci-lint
// falls back to the home-directory config when the analyzed project has no
// .golangci.yml of its own, regardless of where the project lives; a
// project-local config always wins over this fallback.
func Linters(ctx context.Context) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}
	return copyEmbeddedFile(golintFS, golintConfigSrc, filepath.Join(home, golintConfigDest))
}

func copyEmbeddedFile(fsys fs.FS, srcPath, destPath string) error {
	content, err := fs.ReadFile(fsys, srcPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", srcPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(destPath), err)
	}

	if err := os.WriteFile(destPath, content, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", destPath, err)
	}

	return nil
}
