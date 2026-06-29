package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
)

type Generator func(context.Context, Options) error

func Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer, generator Generator) int {
	flags := flag.NewFlagSet("degot", flag.ContinueOnError)
	flags.SetOutput(stderr)

	service := flags.String("service", "", "service directory to generate")
	modPath := flags.String("mod", "", "go module")

	if err := flags.Parse(args); err != nil {
		// flag.ErrHelp means the user asked for usage (-h/-help); that is a
		// successful request, not a usage error, so exit 0.
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if *service == "" {
		writeln(stderr, "--service is required")
		flags.Usage()
		return 2
	}

	options := Options{
		Service: *service,
		Mod:     *modPath,
	}
	if err := generator(ctx, options); err != nil {
		writef(stderr, "%v\n", err)
		return 1
	}

	writef(stdout, "created %s\n", *service)
	return 0
}

func writef(writer io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(writer, format, args...)
}

func writeln(writer io.Writer, value string) {
	_, _ = fmt.Fprintln(writer, value)
}
