package main

import (
	"github.com/spf13/cobra"
	"log"
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
