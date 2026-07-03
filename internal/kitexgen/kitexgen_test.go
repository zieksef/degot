package kitexgen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateWithoutKitexInPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := Generate(context.Background(), Options{Repo: "https://github.com/acme/idl.git", IDL: "order/order.proto"})
	if err == nil {
		t.Fatal("expected error when kitex is not installed")
	}
	if !strings.Contains(err.Error(), "go install github.com/cloudwego/kitex/tool/cmd/kitex@latest") {
		t.Errorf("error %q does not carry the install hint", err)
	}
}

func installFakeKitex(t *testing.T, argsFile string, exitCode int) {
	t.Helper()

	dir := t.TempDir()
	script := fmt.Sprintf("#!/bin/sh\necho \"$@\" > %s\necho boom >&2\nexit %d\n", argsFile, exitCode)
	if err := os.WriteFile(filepath.Join(dir, "kitex"), []byte(script), 0o755); err != nil {
		t.Fatalf("write fake kitex: %v", err)
	}
	t.Setenv("PATH", dir)
}

func TestGenerateInvokesKitex(t *testing.T) {
	argsFile := filepath.Join(t.TempDir(), "args.txt")
	installFakeKitex(t, argsFile, 0)

	options := Options{Repo: "https://github.com/acme/idl.git", IDL: "order/order.proto"}
	if err := Generate(context.Background(), options); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	got, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("read recorded args: %v", err)
	}
	want := "-I https://github.com/acme/idl.git order/order.proto"
	if strings.TrimSpace(string(got)) != want {
		t.Errorf("kitex argv = %q, want %q", strings.TrimSpace(string(got)), want)
	}
}

func TestGenerateReportsKitexFailure(t *testing.T) {
	argsFile := filepath.Join(t.TempDir(), "args.txt")
	installFakeKitex(t, argsFile, 3)

	err := Generate(context.Background(), Options{Repo: "https://github.com/acme/idl.git", IDL: "order/order.proto"})
	if err == nil {
		t.Fatal("expected error when kitex exits non-zero")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("error %q does not include kitex output", err)
	}
}
