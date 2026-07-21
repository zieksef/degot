package gen

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGenerate(t *testing.T) {
	servicePath := filepath.Join(t.TempDir(), "orderservice")

	if err := Generate(context.Background(), Options{Service: servicePath}); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	for _, name := range serviceDirs {
		info, err := os.Stat(filepath.Join(servicePath, name))
		if err != nil {
			t.Errorf("directory %s: %v", name, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", name)
		}
	}

	files := map[string]string{
		".gitignore":                    gitignoreTemplate,
		"README.md":                     "# orderservice\n",
		"AGENTS.md":                     "",
		"CLAUDE.md":                     "@AGENTS.md\n",
		filepath.Join("cmd", "main.go"): "package main\n\nfunc main() {}\n",
	}
	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(servicePath, name))
		if err != nil {
			t.Errorf("read %s: %v", name, err)
			continue
		}
		if string(got) != want {
			t.Errorf("%s content = %q, want %q", name, got, want)
		}
	}
}

func TestGenerateGitignoreMatchesTemplateFile(t *testing.T) {
	want, err := os.ReadFile("templates/gitignore")
	if err != nil {
		t.Fatalf("read template: %v", err)
	}
	if gitignoreTemplate != string(want) {
		t.Fatal("embedded gitignore template differs from templates/gitignore")
	}
	if len(gitignoreTemplate) == 0 {
		t.Fatal("gitignore template is empty")
	}
}

func TestGenerateKeepsExistingFiles(t *testing.T) {
	servicePath := filepath.Join(t.TempDir(), "orderservice")

	if err := Generate(context.Background(), Options{Service: servicePath}); err != nil {
		t.Fatalf("first Generate: %v", err)
	}
	custom := []byte("custom content\n")
	if err := os.WriteFile(filepath.Join(servicePath, ".gitignore"), custom, 0o644); err != nil {
		t.Fatalf("write custom .gitignore: %v", err)
	}

	if err := Generate(context.Background(), Options{Service: servicePath}); err != nil {
		t.Fatalf("second Generate: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(servicePath, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if string(got) != string(custom) {
		t.Errorf(".gitignore was overwritten: %q", got)
	}
}

func TestGenerateRejectsEscapingPath(t *testing.T) {
	err := Generate(context.Background(), Options{Service: filepath.Join("..", "escape")})
	if err == nil {
		t.Fatal("expected error for escaping relative path")
	}
}

func TestGenerateRollsBackOnModInitFailure(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go binary not available")
	}

	servicePath := filepath.Join(t.TempDir(), "orderservice")
	err := Generate(context.Background(), Options{Service: servicePath, Mod: "invalid!module!path"})
	if err == nil {
		t.Fatal("expected go mod init to fail for invalid module path")
	}
	if _, statErr := os.Stat(servicePath); !os.IsNotExist(statErr) {
		t.Errorf("service directory was not rolled back: %v", statErr)
	}
}
