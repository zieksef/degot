package command

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zieksef/degot/internal/gen"
)

func newGenCommand(generator Generator) *cobra.Command {
	var options gen.Options

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate service directories and base files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(options.Service) == "" {
				return &usageError{err: errors.New("--service is required")}
			}
			cmd.SilenceUsage = true

			if err := generator(cmd.Context(), options); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", options.Service)
			return nil
		},
	}
	cmd.Flags().StringVar(&options.Service, "service", "", "service directory to generate")
	cmd.Flags().StringVar(&options.Mod, "mod", "", "go module")
	return cmd
}
