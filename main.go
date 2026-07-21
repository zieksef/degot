package main

import (
	"context"
	"os"

	"github.com/zieksef/degot/internal/command"
	"github.com/zieksef/degot/internal/gen"
	"github.com/zieksef/degot/internal/sync"
)

func main() {
	os.Exit(command.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr, command.Deps{
		Gen:              gen.Generate,
		SyncLinters:      sync.Linters,
		SyncInstructions: sync.Instructions,
		SyncSkills:       sync.Skills,
	}))
}
