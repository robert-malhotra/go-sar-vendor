package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

// -----------------------------------------------------------------------------
// top-level command ------------------------------------------------------------
// -----------------------------------------------------------------------------
func main() {
	root := &cli.Command{
		Name:  "gosar",
		Usage: "Command-line helper for Capella Space Tasking & Access API",

		// global flags (still visible in every sub-command)

		// sub-commands (property renamed Subcommands → Commands)
		Commands: []*cli.Command{
			umbraCmd(),
			capellaCmd(),
			iceyeCmd(),
			airbusCmd(),
		},
	}

	if err := root.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
