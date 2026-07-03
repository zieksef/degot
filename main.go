package main

import (
	"context"
	"os"

	"github.com/zieksef/degot/internal/command"
	"github.com/zieksef/degot/internal/gen"
	"github.com/zieksef/degot/internal/kitexgen"
)

func main() {
	os.Exit(command.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr, command.Deps{
		Gen:      gen.Generate,
		KitexGen: kitexgen.Generate,
	}))
}
