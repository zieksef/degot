package command

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/zieksef/degot/internal/gen"
	"github.com/zieksef/degot/internal/sync"
)

type runResult struct {
	code                    int
	stdout                  string
	stderr                  string
	genCalled               bool
	genOptions              gen.Options
	syncLintersCalled       bool
	syncInstructionsCalled  bool
	syncInstructionsOptions sync.Options
	syncSkillsCalled        bool
	syncSkillsOptions       sync.Options
}

type runErrs struct {
	gen              error
	syncLinters      error
	syncInstructions error
	syncSkills       error
}

func runForTest(t *testing.T, args []string, errs runErrs) runResult {
	t.Helper()

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	result := runResult{}
	deps := Deps{
		Gen: func(ctx context.Context, options gen.Options) error {
			result.genCalled = true
			result.genOptions = options
			return errs.gen
		},
		SyncLinters: func(ctx context.Context) error {
			result.syncLintersCalled = true
			return errs.syncLinters
		},
		SyncInstructions: func(ctx context.Context, options sync.Options) error {
			result.syncInstructionsCalled = true
			result.syncInstructionsOptions = options
			return errs.syncInstructions
		},
		SyncSkills: func(ctx context.Context, options sync.Options) error {
			result.syncSkillsCalled = true
			result.syncSkillsOptions = options
			return errs.syncSkills
		},
	}

	result.code = Run(context.Background(), args, &stdout, &stderr, deps)
	result.stdout = stdout.String()
	result.stderr = stderr.String()
	return result
}

func TestRunRoot(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantCode     int
		wantInStdout string
	}{
		{
			name:     "help exits zero",
			args:     []string{"--help"},
			wantCode: 0,
		},
		{
			name:         "no arguments prints help",
			args:         []string{},
			wantCode:     0,
			wantInStdout: "sync",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runForTest(t, tt.args, runErrs{})

			if result.code != tt.wantCode {
				t.Errorf("exit code = %d, want %d\nstdout: %s\nstderr: %s", result.code, tt.wantCode, result.stdout, result.stderr)
			}
			if tt.wantInStdout != "" && !strings.Contains(result.stdout, tt.wantInStdout) {
				t.Errorf("stdout %q does not contain %q", result.stdout, tt.wantInStdout)
			}
		})
	}
}

func TestRunGen(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		generateErr  error
		wantCode     int
		wantCalled   bool
		wantOptions  gen.Options
		wantInStdout string
		wantInStderr string
	}{
		{
			name:         "gen with service",
			args:         []string{"gen", "--service", "orderservice"},
			wantCode:     0,
			wantCalled:   true,
			wantOptions:  gen.Options{Service: "orderservice"},
			wantInStdout: "[degot]: created orderservice",
		},
		{
			name:        "gen with service and mod",
			args:        []string{"gen", "--service", "orderservice", "--mod", "example.com/acme/orderservice"},
			wantCode:    0,
			wantCalled:  true,
			wantOptions: gen.Options{Service: "orderservice", Mod: "example.com/acme/orderservice"},
		},
		{
			name:         "gen without service is a usage error",
			args:         []string{"gen"},
			wantCode:     2,
			wantInStderr: "--service",
		},
		{
			name:         "gen with unknown flag is a usage error",
			args:         []string{"gen", "--bogus"},
			wantCode:     2,
			wantInStderr: "bogus",
		},
		{
			name:         "generator failure",
			args:         []string{"gen", "--service", "orderservice"},
			generateErr:  errors.New("disk full"),
			wantCode:     1,
			wantCalled:   true,
			wantOptions:  gen.Options{Service: "orderservice"},
			wantInStderr: "disk full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runForTest(t, tt.args, runErrs{gen: tt.generateErr})

			if result.code != tt.wantCode {
				t.Errorf("exit code = %d, want %d\nstdout: %s\nstderr: %s", result.code, tt.wantCode, result.stdout, result.stderr)
			}
			if result.genCalled != tt.wantCalled {
				t.Errorf("generator called = %v, want %v", result.genCalled, tt.wantCalled)
			}
			if tt.wantCalled && result.genOptions != tt.wantOptions {
				t.Errorf("options = %+v, want %+v", result.genOptions, tt.wantOptions)
			}
			if tt.wantInStdout != "" && !strings.Contains(result.stdout, tt.wantInStdout) {
				t.Errorf("stdout %q does not contain %q", result.stdout, tt.wantInStdout)
			}
			if tt.wantInStderr != "" && !strings.Contains(result.stderr, tt.wantInStderr) {
				t.Errorf("stderr %q does not contain %q", result.stderr, tt.wantInStderr)
			}
		})
	}
}

func TestRunSyncAll(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		errs             runErrs
		wantCode         int
		wantLinters      bool
		wantInstructions bool
		wantSkills       bool
		wantInStdout     string
		wantInStderr     string
	}{
		{
			name:             "sync all runs every subcommand",
			args:             []string{"sync", "all"},
			wantCode:         0,
			wantLinters:      true,
			wantInstructions: true,
			wantSkills:       true,
			wantInStdout:     "[degot]: synced skills to ~/.claude and ~/.codex",
		},
		{
			name:         "sync all rejects claude-only",
			args:         []string{"sync", "all", "--claude-only"},
			wantCode:     2,
			wantInStderr: "all cannot be combined",
		},
		{
			name:         "sync all rejects codex-only",
			args:         []string{"sync", "all", "--codex-only"},
			wantCode:     2,
			wantInStderr: "all cannot be combined",
		},
		{
			name:             "sync all stops at the first failure",
			args:             []string{"sync", "all"},
			errs:             runErrs{syncInstructions: errors.New("disk full")},
			wantCode:         1,
			wantLinters:      true,
			wantInstructions: true,
			wantSkills:       false,
			wantInStderr:     "disk full",
		},
		{
			name:     "bare sync still prints help",
			args:     []string{"sync"},
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runForTest(t, tt.args, tt.errs)

			if result.code != tt.wantCode {
				t.Errorf("exit code = %d, want %d\nstdout: %s\nstderr: %s", result.code, tt.wantCode, result.stdout, result.stderr)
			}
			if result.syncLintersCalled != tt.wantLinters {
				t.Errorf("sync linters called = %v, want %v", result.syncLintersCalled, tt.wantLinters)
			}
			if result.syncInstructionsCalled != tt.wantInstructions {
				t.Errorf("sync instructions called = %v, want %v", result.syncInstructionsCalled, tt.wantInstructions)
			}
			if result.syncSkillsCalled != tt.wantSkills {
				t.Errorf("sync skills called = %v, want %v", result.syncSkillsCalled, tt.wantSkills)
			}
			if tt.wantInStdout != "" && !strings.Contains(result.stdout, tt.wantInStdout) {
				t.Errorf("stdout %q does not contain %q", result.stdout, tt.wantInStdout)
			}
			if tt.wantInStderr != "" && !strings.Contains(result.stderr, tt.wantInStderr) {
				t.Errorf("stderr %q does not contain %q", result.stderr, tt.wantInStderr)
			}
		})
	}
}

func TestRunSyncLinters(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		syncErr      error
		wantCode     int
		wantCalled   bool
		wantInStdout string
		wantInStderr string
	}{
		{
			name:         "sync linters",
			args:         []string{"sync", "linters"},
			wantCode:     0,
			wantCalled:   true,
			wantInStdout: "[degot]: synced golangci-lint config to ~/.golangci.yml",
		},
		{
			name:         "sync linters failure",
			args:         []string{"sync", "linters"},
			syncErr:      errors.New("disk full"),
			wantCode:     1,
			wantCalled:   true,
			wantInStderr: "disk full",
		},
		{
			name:         "sync linters with unknown flag is a usage error",
			args:         []string{"sync", "linters", "--bogus"},
			wantCode:     2,
			wantInStderr: "bogus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runForTest(t, tt.args, runErrs{syncLinters: tt.syncErr})

			if result.code != tt.wantCode {
				t.Errorf("exit code = %d, want %d\nstdout: %s\nstderr: %s", result.code, tt.wantCode, result.stdout, result.stderr)
			}
			if result.syncLintersCalled != tt.wantCalled {
				t.Errorf("sync linters called = %v, want %v", result.syncLintersCalled, tt.wantCalled)
			}
			if tt.wantInStdout != "" && !strings.Contains(result.stdout, tt.wantInStdout) {
				t.Errorf("stdout %q does not contain %q", result.stdout, tt.wantInStdout)
			}
			if tt.wantInStderr != "" && !strings.Contains(result.stderr, tt.wantInStderr) {
				t.Errorf("stderr %q does not contain %q", result.stderr, tt.wantInStderr)
			}
		})
	}
}

func TestRunSyncInstructions(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		syncErr      error
		wantCode     int
		wantCalled   bool
		wantOptions  sync.Options
		wantInStdout string
		wantInStderr string
	}{
		{
			name:         "sync instructions with no flags targets both",
			args:         []string{"sync", "instructions"},
			wantCode:     0,
			wantCalled:   true,
			wantOptions:  sync.Options{},
			wantInStdout: "[degot]: synced instructions to ~/.claude and ~/.codex",
		},
		{
			name:         "sync instructions with --claude-only",
			args:         []string{"sync", "instructions", "--claude-only"},
			wantCode:     0,
			wantCalled:   true,
			wantOptions:  sync.Options{ClaudeOnly: true},
			wantInStdout: "synced instructions to ~/.claude\n",
		},
		{
			name:         "sync instructions with --codex-only",
			args:         []string{"sync", "instructions", "--codex-only"},
			wantCode:     0,
			wantCalled:   true,
			wantOptions:  sync.Options{CodexOnly: true},
			wantInStdout: "synced instructions to ~/.codex\n",
		},
		{
			name:         "sync instructions with both only flags is a usage error",
			args:         []string{"sync", "instructions", "--claude-only", "--codex-only"},
			wantCode:     2,
			wantInStderr: "mutually exclusive",
		},
		{
			name:         "sync instructions failure",
			args:         []string{"sync", "instructions"},
			syncErr:      errors.New("permission denied"),
			wantCode:     1,
			wantCalled:   true,
			wantOptions:  sync.Options{},
			wantInStderr: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runForTest(t, tt.args, runErrs{syncInstructions: tt.syncErr})

			if result.code != tt.wantCode {
				t.Errorf("exit code = %d, want %d\nstdout: %s\nstderr: %s", result.code, tt.wantCode, result.stdout, result.stderr)
			}
			if result.syncInstructionsCalled != tt.wantCalled {
				t.Errorf("generator called = %v, want %v", result.syncInstructionsCalled, tt.wantCalled)
			}
			if tt.wantCalled && result.syncInstructionsOptions != tt.wantOptions {
				t.Errorf("options = %+v, want %+v", result.syncInstructionsOptions, tt.wantOptions)
			}
			if tt.wantInStdout != "" && !strings.Contains(result.stdout, tt.wantInStdout) {
				t.Errorf("stdout %q does not contain %q", result.stdout, tt.wantInStdout)
			}
			if tt.wantInStderr != "" && !strings.Contains(result.stderr, tt.wantInStderr) {
				t.Errorf("stderr %q does not contain %q", result.stderr, tt.wantInStderr)
			}
		})
	}
}

func TestRunSyncSkills(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		syncErr      error
		wantCode     int
		wantCalled   bool
		wantOptions  sync.Options
		wantInStdout string
		wantInStderr string
	}{
		{
			name:         "sync skills with no flags targets both",
			args:         []string{"sync", "skills"},
			wantCode:     0,
			wantCalled:   true,
			wantOptions:  sync.Options{},
			wantInStdout: "[degot]: synced skills to ~/.claude and ~/.codex",
		},
		{
			name:         "sync skills with --claude-only",
			args:         []string{"sync", "skills", "--claude-only"},
			wantCode:     0,
			wantCalled:   true,
			wantOptions:  sync.Options{ClaudeOnly: true},
			wantInStdout: "synced skills to ~/.claude\n",
		},
		{
			name:         "sync skills with both only flags is a usage error",
			args:         []string{"sync", "skills", "--claude-only", "--codex-only"},
			wantCode:     2,
			wantInStderr: "mutually exclusive",
		},
		{
			name:         "sync skills failure",
			args:         []string{"sync", "skills"},
			syncErr:      errors.New("disk full"),
			wantCode:     1,
			wantCalled:   true,
			wantOptions:  sync.Options{},
			wantInStderr: "disk full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runForTest(t, tt.args, runErrs{syncSkills: tt.syncErr})

			if result.code != tt.wantCode {
				t.Errorf("exit code = %d, want %d\nstdout: %s\nstderr: %s", result.code, tt.wantCode, result.stdout, result.stderr)
			}
			if result.syncSkillsCalled != tt.wantCalled {
				t.Errorf("generator called = %v, want %v", result.syncSkillsCalled, tt.wantCalled)
			}
			if tt.wantCalled && result.syncSkillsOptions != tt.wantOptions {
				t.Errorf("options = %+v, want %+v", result.syncSkillsOptions, tt.wantOptions)
			}
			if tt.wantInStdout != "" && !strings.Contains(result.stdout, tt.wantInStdout) {
				t.Errorf("stdout %q does not contain %q", result.stdout, tt.wantInStdout)
			}
			if tt.wantInStderr != "" && !strings.Contains(result.stderr, tt.wantInStderr) {
				t.Errorf("stderr %q does not contain %q", result.stderr, tt.wantInStderr)
			}
		})
	}
}

func TestRunRejectsPositionalArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "gen", args: []string{"gen", "--service", "orderservice", "extra"}},
		{name: "sync parent", args: []string{"sync", "bogus"}},
		{name: "sync all", args: []string{"sync", "all", "extra"}},
		{name: "sync linters", args: []string{"sync", "linters", "extra"}},
		{name: "sync instructions", args: []string{"sync", "instructions", "extra"}},
		{name: "sync skills chained with instructions", args: []string{"sync", "skills", "instructions"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runForTest(t, tt.args, runErrs{})

			if result.code != 2 {
				t.Errorf("exit code = %d, want 2\nstdout: %s\nstderr: %s", result.code, result.stdout, result.stderr)
			}
			if result.genCalled || result.syncLintersCalled || result.syncInstructionsCalled || result.syncSkillsCalled {
				t.Errorf("no generator should run on a positional-arg usage error: %+v", result)
			}
		})
	}
}
