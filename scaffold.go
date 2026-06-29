package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	serviceDirs = []string{
		"api",
		"cmd",
		"configs",
		"deployments",
		"internal",
		"internal/cfg",
		"internal/infra",
		"internal/repo",
		"internal/svc",
		"migrations",
		"scripts",
	}
)

type Options struct {
	Service string
	Mod     string
}

func Generate(ctx context.Context, options Options) error {
	servicePath := strings.TrimSpace(options.Service)
	if servicePath == "" {
		return errors.New("--service is required")
	}

	servicePath = filepath.Clean(servicePath)
	// Clean keeps leading "..", so reject relative paths that escape the
	// working directory; absolute paths stay allowed (caller is explicit).
	if servicePath == ".." || strings.HasPrefix(servicePath, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("service path %q escapes the working directory", options.Service)
	}
	serviceName := filepath.Base(servicePath)

	serviceExisted := true
	if _, err := os.Stat(servicePath); errors.Is(err, os.ErrNotExist) {
		serviceExisted = false
	}

	if err := os.MkdirAll(servicePath, 0o755); err != nil {
		return fmt.Errorf("create service directory: %w", err)
	}

	for _, name := range serviceDirs {
		path := filepath.Join(servicePath, name)
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("create %s directory: %w", name, err)
		}
	}

	files := []struct {
		relPath string
		content string
	}{
		{"README.md", readmeContent(serviceName)},
		{"AGENTS.md", ""},
		{".gitignore", ""},
		{filepath.Join("cmd", "main.go"), cmdMainContent()},
	}
	for _, f := range files {
		if err := writeFileIfMissing(filepath.Join(servicePath, f.relPath), f.content); err != nil {
			return err
		}
	}

	modPath := strings.TrimSpace(options.Mod)
	if modPath == "" {
		return nil
	}

	cmd := exec.CommandContext(ctx, "go", "mod", "init", modPath)
	cmd.Dir = servicePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Roll back only the tree we created this run; never delete a service
		// directory that already existed before Generate was called.
		if !serviceExisted {
			_ = os.RemoveAll(servicePath)
		}
		return fmt.Errorf("go mod init %s: %w: %s", modPath, err, strings.TrimSpace(string(output)))
	}

	return nil
}

func writeFileIfMissing(path string, content string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	if _, err := file.WriteString(content); err != nil {
		_ = file.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

func readmeContent(serviceName string) string {
	return fmt.Sprintf("# %s\n", serviceName)
}

func cmdMainContent() string {
	return `package main

func main() {}
`
}
