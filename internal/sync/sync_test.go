package sync

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func withHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	return home
}

func TestInstructionsCreatesAndRenamesForClaude(t *testing.T) {
	home := withHome(t)

	if err := Instructions(context.Background(), Options{}); err != nil {
		t.Fatalf("Instructions: %v", err)
	}

	want, err := os.ReadFile("agent/instructions/AGENTS.md")
	if err != nil {
		t.Fatalf("read source AGENTS.md: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(home, ".claude", "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read ~/.claude/CLAUDE.md: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("~/.claude/CLAUDE.md content = %q, want %q", got, want)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "AGENTS.md")); !os.IsNotExist(err) {
		t.Errorf("~/.claude/AGENTS.md should not exist (expected rename to CLAUDE.md), stat err = %v", err)
	}

	got, err = os.ReadFile(filepath.Join(home, ".codex", "AGENTS.md"))
	if err != nil {
		t.Fatalf("read ~/.codex/AGENTS.md: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("~/.codex/AGENTS.md content = %q, want %q", got, want)
	}
}

func TestInstructionsOverwritesExistingFile(t *testing.T) {
	home := withHome(t)

	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "CLAUDE.md"), []byte("stale content"), 0o644); err != nil {
		t.Fatalf("seed stale CLAUDE.md: %v", err)
	}

	if err := Instructions(context.Background(), Options{}); err != nil {
		t.Fatalf("Instructions: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(claudeDir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	if string(got) == "stale content" {
		t.Error("CLAUDE.md was not overwritten")
	}
}

func TestInstructionsClaudeOnly(t *testing.T) {
	home := withHome(t)

	if err := Instructions(context.Background(), Options{ClaudeOnly: true}); err != nil {
		t.Fatalf("Instructions: %v", err)
	}

	if _, err := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); err != nil {
		t.Errorf("~/.claude/CLAUDE.md: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".codex")); !os.IsNotExist(err) {
		t.Errorf("~/.codex should not have been created, stat err = %v", err)
	}
}

func TestInstructionsCodexOnly(t *testing.T) {
	home := withHome(t)

	if err := Instructions(context.Background(), Options{CodexOnly: true}); err != nil {
		t.Fatalf("Instructions: %v", err)
	}

	if _, err := os.Stat(filepath.Join(home, ".codex", "AGENTS.md")); err != nil {
		t.Errorf("~/.codex/AGENTS.md: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude")); !os.IsNotExist(err) {
		t.Errorf("~/.claude should not have been created, stat err = %v", err)
	}
}

func TestInstructionsMutuallyExclusiveFlags(t *testing.T) {
	withHome(t)

	err := Instructions(context.Background(), Options{ClaudeOnly: true, CodexOnly: true})
	if err == nil {
		t.Fatal("expected error for mutually exclusive options")
	}
}

func TestSkillsRecursiveCopyPreservesExtraFiles(t *testing.T) {
	home := withHome(t)

	for _, root := range []string{".claude", ".codex"} {
		dir := filepath.Join(home, root, "skills", "grilling")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "extra.md"), []byte("keep me"), 0o644); err != nil {
			t.Fatalf("seed extra file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("stale skill content"), 0o644); err != nil {
			t.Fatalf("seed stale SKILL.md: %v", err)
		}
	}

	if err := Skills(context.Background(), Options{}); err != nil {
		t.Fatalf("Skills: %v", err)
	}

	want, err := os.ReadFile("agent/skills/grilling/SKILL.md")
	if err != nil {
		t.Fatalf("read source SKILL.md: %v", err)
	}

	for _, root := range []string{".claude", ".codex"} {
		dir := filepath.Join(home, root, "skills", "grilling")

		got, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
		if err != nil {
			t.Fatalf("read %s/SKILL.md: %v", root, err)
		}
		if string(got) != string(want) {
			t.Errorf("%s/skills/grilling/SKILL.md not overwritten, got %q", root, got)
		}

		extra, err := os.ReadFile(filepath.Join(dir, "extra.md"))
		if err != nil {
			t.Fatalf("%s/skills/grilling/extra.md missing: %v", root, err)
		}
		if string(extra) != "keep me" {
			t.Errorf("%s/skills/grilling/extra.md was modified, got %q", root, extra)
		}
	}
}

func TestSkillsMutuallyExclusiveFlags(t *testing.T) {
	withHome(t)

	err := Skills(context.Background(), Options{ClaudeOnly: true, CodexOnly: true})
	if err == nil {
		t.Fatal("expected error for mutually exclusive options")
	}
}

func TestLintersWritesHomeConfig(t *testing.T) {
	home := withHome(t)

	if err := Linters(context.Background()); err != nil {
		t.Fatalf("Linters: %v", err)
	}

	want, err := os.ReadFile("golint/.golangci.yml")
	if err != nil {
		t.Fatalf("read source golangci.yml: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(home, ".golangci.yml"))
	if err != nil {
		t.Fatalf("read ~/.golangci.yml: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("~/.golangci.yml content = %q, want %q", got, want)
	}
}

func TestLintersOverwritesExistingConfig(t *testing.T) {
	home := withHome(t)

	if err := os.WriteFile(filepath.Join(home, ".golangci.yml"), []byte("stale content"), 0o644); err != nil {
		t.Fatalf("seed stale .golangci.yml: %v", err)
	}

	if err := Linters(context.Background()); err != nil {
		t.Fatalf("Linters: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(home, ".golangci.yml"))
	if err != nil {
		t.Fatalf("read ~/.golangci.yml: %v", err)
	}
	if string(got) == "stale content" {
		t.Error("~/.golangci.yml was not overwritten")
	}
}
