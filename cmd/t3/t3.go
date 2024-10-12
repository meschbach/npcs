package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {

	root := &cobra.Command{
		Use:   "t3",
		Short: "t3 game interactions",
	}
	root.AddCommand(hciCommands())
	root.AddCommand(serviceCommands())

	if err := root.Execute(); err != nil {
		if _, err := fmt.Fprintf(os.Stderr, "error: %v\n", err); err != nil {
			panic(err)
		}
	}
}
