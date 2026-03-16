package main

import (
	"log"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "simpled",
		Short: "simple game instance daemon",
	}
	root.AddCommand(singleGameInstance())

	if err := root.Execute(); err != nil {
		log.Fatalf("error: %e", err)
	}
}
