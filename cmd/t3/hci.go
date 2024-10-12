package main

import (
	"fmt"
	"github.com/meschbach/npcs/t3/tui"
	"github.com/spf13/cobra"
	"os"
)

func hciCommands() *cobra.Command {
	play := &cobra.Command{
		Use:   "play <grpc-host>",
		Short: "play a game against a remote unit",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			if err := tui.RunGame(ctx, args[0]); err != nil {
				if _, err := fmt.Fprintf(os.Stderr, "failed because %s\n", err.Error()); err != nil {
					panic(err)
				}
			}
		},
	}

	fillIn := &cobra.Command{
		Use:   "fill-in",
		Short: "Play a game against the fill-in AI",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			if err := tui.RunFillIn(ctx); err != nil {
				if _, err := fmt.Fprintf(os.Stderr, "failed because %s\n", err.Error()); err != nil {
					panic(err)
				}
			}
		},
	}

	hci := &cobra.Command{
		Use:   "hci",
		Short: "human computer interactions with services",
	}
	hci.AddCommand(play)
	hci.AddCommand(fillIn)
	return hci
}
