package command

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/zieksef/degot/internal/gen"
	"github.com/zieksef/degot/internal/kitexgen"
)

type runResult struct {
	code         int
	stdout       string
	stderr       string
	genCalled    bool
	genOptions   gen.Options
	kitexCalled  bool
	kitexOptions kitexgen.Options
}

func runForTest(t *testing.T, args []string, genErr error, kitexErr error) runResult {
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
			return genErr
		},
		KitexGen: func(ctx context.Context, options kitexgen.Options) error {
			result.kitexCalled = true
			result.kitexOptions = options
			return kitexErr
		},
	}

	result.code = Run(context.Background(), args, &stdout, &stderr, deps)
	result.stdout = stdout.String()
	result.stderr = stderr.String()
	return result
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
			wantInStdout: "created orderservice",
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
		{
			name:     "help exits zero",
			args:     []string{"--help"},
			wantCode: 0,
		},
		{
			name:         "no arguments prints help",
			args:         []string{},
			wantCode:     0,
			wantInStdout: "gen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runForTest(t, tt.args, tt.generateErr, nil)

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

func TestRunKitexGen(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		generateErr  error
		wantCode     int
		wantCalled   bool
		wantOptions  kitexgen.Options
		wantInStdout string
		wantInStderr string
	}{
		{
			name:         "kitexgen with repo and idl",
			args:         []string{"kitexgen", "--repo", "https://github.com/acme/idl.git", "--idl", "order/order.proto"},
			wantCode:     0,
			wantCalled:   true,
			wantOptions:  kitexgen.Options{Repo: "https://github.com/acme/idl.git", IDL: "order/order.proto"},
			wantInStdout: "generated kitex code from order/order.proto",
		},
		{
			name:        "kitexgen with ssh repo",
			args:        []string{"kitexgen", "--repo", "git@github.com:acme/idl.git", "--idl", "order/order.proto"},
			wantCode:    0,
			wantCalled:  true,
			wantOptions: kitexgen.Options{Repo: "git@github.com:acme/idl.git", IDL: "order/order.proto"},
		},
		{
			name:         "kitexgen without repo is a usage error",
			args:         []string{"kitexgen", "--idl", "order/order.proto"},
			wantCode:     2,
			wantInStderr: "--repo",
		},
		{
			name:         "kitexgen without idl is a usage error",
			args:         []string{"kitexgen", "--repo", "https://github.com/acme/idl.git"},
			wantCode:     2,
			wantInStderr: "--idl",
		},
		{
			name:         "kitexgen with non-git repo prefix is a usage error",
			args:         []string{"kitexgen", "--repo", "ftp://acme/idl.git", "--idl", "order/order.proto"},
			wantCode:     2,
			wantInStderr: "git@",
		},
		{
			name:         "kitex failure",
			args:         []string{"kitexgen", "--repo", "https://github.com/acme/idl.git", "--idl", "order/order.proto"},
			generateErr:  errors.New("kitex exploded"),
			wantCode:     1,
			wantCalled:   true,
			wantOptions:  kitexgen.Options{Repo: "https://github.com/acme/idl.git", IDL: "order/order.proto"},
			wantInStderr: "kitex exploded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runForTest(t, tt.args, nil, tt.generateErr)

			if result.code != tt.wantCode {
				t.Errorf("exit code = %d, want %d\nstdout: %s\nstderr: %s", result.code, tt.wantCode, result.stdout, result.stderr)
			}
			if result.kitexCalled != tt.wantCalled {
				t.Errorf("generator called = %v, want %v", result.kitexCalled, tt.wantCalled)
			}
			if tt.wantCalled && result.kitexOptions != tt.wantOptions {
				t.Errorf("options = %+v, want %+v", result.kitexOptions, tt.wantOptions)
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
