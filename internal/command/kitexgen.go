package command

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zieksef/degot/internal/kitexgen"
)

func newKitexGenCommand(generator KitexGenerator) *cobra.Command {
	var options kitexgen.Options

	cmd := &cobra.Command{
		Use:   "kitexgen",
		Short: "Generate kitex code from a proto IDL in a remote git repository.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(options.Repo) == "" {
				return &usageError{err: errors.New("--repo is required")}
			}
			if strings.TrimSpace(options.IDL) == "" {
				return &usageError{err: errors.New("--idl is required")}
			}
			// kitex only treats includes with these prefixes as remote git
			// repositories; anything else would be read as a local path.
			if !strings.HasPrefix(options.Repo, "git@") && !strings.HasPrefix(options.Repo, "http://") && !strings.HasPrefix(options.Repo, "https://") {
				return &usageError{err: fmt.Errorf("--repo must start with git@, http:// or https://, got %q", options.Repo)}
			}
			cmd.SilenceUsage = true

			if err := generator(cmd.Context(), options); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "generated kitex code from %s\n", options.IDL)
			return nil
		},
	}
	cmd.Flags().StringVar(&options.Repo, "repo", "", "IDL git repository URL (git@... or https://...)")
	cmd.Flags().StringVar(&options.IDL, "idl", "", "proto file path inside the IDL repository")
	return cmd
}
