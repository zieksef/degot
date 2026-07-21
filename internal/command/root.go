package command

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/zieksef/degot/internal/gen"
	"github.com/zieksef/degot/internal/output"
	"github.com/zieksef/degot/internal/sync"
)

// usageTemplate is cobra's default usage template with the
// `(eq .Name "help")` exemptions removed, so the hidden help command does not
// show up in command listings.
const (
	usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) .IsAvailableCommand)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") .IsAvailableCommand)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
)

type Generator func(context.Context, gen.Options) error

type SyncLinters func(context.Context) error

type SyncInstructions func(context.Context, sync.Options) error

type SyncSkills func(context.Context, sync.Options) error

type Deps struct {
	Gen              Generator
	SyncLinters      SyncLinters
	SyncInstructions SyncInstructions
	SyncSkills       SyncSkills
}

// usageError marks errors caused by invalid invocation (bad flags, missing
// required flags) so Run can map them to exit code 2 instead of 1.
type usageError struct {
	err error
}

func (e *usageError) Error() string {
	return e.err.Error()
}

func (e *usageError) Unwrap() error {
	return e.err
}

// noArgs rejects positional arguments as usage errors (exit code 2), the
// same convention as flag errors. Without it cobra silently ignores extras,
// so e.g. `degot sync skills instructions` would sync only skills while
// looking successful.
func noArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.NoArgs(cmd, args); err != nil {
		return &usageError{err: err}
	}
	return nil
}

func Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer, deps Deps) int {
	root := newRootCommand(deps)
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	// Print errors ourselves so they carry the [degot] prefix instead of
	// cobra's default "Error:" line. Usage text on flag/arg errors is left
	// on, so those still print the command usage above our message.
	root.SilenceErrors = true

	if err := root.ExecuteContext(ctx); err != nil {
		_, _ = fmt.Fprint(stderr, output.Sprintf("%s\n", err))
		if _, ok := errors.AsType[*usageError](err); ok {
			return 2
		}
		return 1
	}
	return 0
}

func newRootCommand(deps Deps) *cobra.Command {
	root := &cobra.Command{
		Use:   "degot",
		Short: "Go service scaffolds-depot.",
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}
	root.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return &usageError{err: err}
	})
	root.SetUsageTemplate(usageTemplate)
	root.AddCommand(newGenCommand(deps.Gen))
	root.AddCommand(newSyncCommand(deps.SyncLinters, deps.SyncInstructions, deps.SyncSkills))

	// The default help command and help flags only exist after these init
	// calls; hiding them keeps -h/--help and `degot help` functional.
	root.InitDefaultHelpCmd()
	for _, cmd := range append([]*cobra.Command{root}, root.Commands()...) {
		if cmd.Name() == "help" {
			cmd.Hidden = true
		}
		cmd.InitDefaultHelpFlag()
		if flag := cmd.Flags().Lookup("help"); flag != nil {
			flag.Hidden = true
		}
	}
	return root
}
