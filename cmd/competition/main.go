package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "competition",
		Short: "Competition service",
	}
	root.AddCommand(daemonCommand())
	root.AddCommand(gamesCommand())
	root.AddCommand(enginesCommand())

	if err := root.Execute(); err != nil {
		if _, err := fmt.Fprintf(os.Stderr, "error: %v\n", err); err != nil {
			panic(err)
		}
	}
}
