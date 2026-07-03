package kitexgen

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type Options struct {
	Repo string
	IDL  string
}

// Generate shells out to the kitex tool, passing the remote IDL repository as
// a git include path (kitex clones it into ~/.kitex/cache and resolves the
// IDL path inside the checkout). Output lands in the current directory; the
// go module name is inferred by kitex from go.mod.
func Generate(ctx context.Context, options Options) error {
	if strings.TrimSpace(options.Repo) == "" {
		return errors.New("--repo is required")
	}

	if strings.TrimSpace(options.IDL) == "" {
		return errors.New("--idl is required")
	}

	kitexPath, err := exec.LookPath("kitex")
	if err != nil {
		return errors.New("kitex tool not found, install it with: go install github.com/cloudwego/kitex/tool/cmd/kitex@latest")
	}

	args := []string{"-I", options.Repo, options.IDL}
	cmd := exec.CommandContext(ctx, kitexPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kitex %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}

	return nil
}
