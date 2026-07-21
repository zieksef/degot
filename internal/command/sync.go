package command

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/zieksef/degot/internal/output"
	"github.com/zieksef/degot/internal/sync"
)

func newSyncCommand(syncLinters SyncLinters, syncInstructions SyncInstructions, syncSkills SyncSkills) *cobra.Command {
	var (
		options sync.Options
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync bundled agent instructions, skills and lint config.",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// The bare parent prints the subcommand help and exits zero.
			return cmd.Help()
		},
	}
	cmd.PersistentFlags().BoolVar(&options.ClaudeOnly, "claude-only", false, "only sync for claude")
	cmd.PersistentFlags().BoolVar(&options.CodexOnly, "codex-only", false, "only sync for codex")

	cmd.AddCommand(newSyncAllCommand(syncLinters, syncInstructions, syncSkills, &options))
	cmd.AddCommand(newSyncLintersCommand(syncLinters))
	cmd.AddCommand(newSyncInstructionsCommand(syncInstructions, &options))
	cmd.AddCommand(newSyncSkillsCommand(syncSkills, &options))
	return cmd
}

func newSyncAllCommand(syncLinters SyncLinters, syncInstructions SyncInstructions, syncSkills SyncSkills, options *sync.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "Run every sync subcommand (linters, instructions, skills).",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// all always writes both targets; combining it with a target
			// filter would silently surprise either way.
			if options.ClaudeOnly || options.CodexOnly {
				return &usageError{err: errors.New("all cannot be combined with --claude-only or --codex-only")}
			}
			cmd.SilenceUsage = true

			if err := syncLinters(cmd.Context()); err != nil {
				return err
			}
			printLintersSynced(cmd.OutOrStdout())

			if err := syncInstructions(cmd.Context(), *options); err != nil {
				return err
			}
			printInstructionsSynced(cmd.OutOrStdout(), *options)

			if err := syncSkills(cmd.Context(), *options); err != nil {
				return err
			}
			printSkillsSynced(cmd.OutOrStdout(), *options)
			return nil
		},
	}
}

func newSyncLintersCommand(generator SyncLinters) *cobra.Command {
	return &cobra.Command{
		Use:   "linters",
		Short: "Sync the team golangci-lint baseline to ~/.golangci.yml.",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if err := generator(cmd.Context()); err != nil {
				return err
			}

			printLintersSynced(cmd.OutOrStdout())
			return nil
		},
	}
}

func newSyncInstructionsCommand(generator SyncInstructions, options *sync.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "instructions",
		Short: "Sync agent instructions files.",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateSyncOptions(*options); err != nil {
				return &usageError{err: err}
			}
			cmd.SilenceUsage = true

			if err := generator(cmd.Context(), *options); err != nil {
				return err
			}

			printInstructionsSynced(cmd.OutOrStdout(), *options)
			return nil
		},
	}
}

func newSyncSkillsCommand(generator SyncSkills, options *sync.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "skills",
		Short: "Sync agent skills files.",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateSyncOptions(*options); err != nil {
				return &usageError{err: err}
			}
			cmd.SilenceUsage = true

			if err := generator(cmd.Context(), *options); err != nil {
				return err
			}

			printSkillsSynced(cmd.OutOrStdout(), *options)
			return nil
		},
	}
}

func validateSyncOptions(options sync.Options) error {
	if options.ClaudeOnly && options.CodexOnly {
		return errors.New("--claude-only and --codex-only are mutually exclusive")
	}
	return nil
}

func syncTargetsDescription(options sync.Options) string {
	switch {
	case options.ClaudeOnly:
		return "~/.claude"
	case options.CodexOnly:
		return "~/.codex"
	default:
		return "~/.claude and ~/.codex"
	}
}

func printLintersSynced(w io.Writer) {
	_, _ = fmt.Fprint(w, output.Sprintf("synced golangci-lint config to ~/.golangci.yml\n"))
}

func printInstructionsSynced(w io.Writer, options sync.Options) {
	_, _ = fmt.Fprint(w, output.Sprintf("synced instructions to %s\n", syncTargetsDescription(options)))
}

func printSkillsSynced(w io.Writer, options sync.Options) {
	_, _ = fmt.Fprint(w, output.Sprintf("synced skills to %s\n", syncTargetsDescription(options)))
}
