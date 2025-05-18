package main

import (
	"fmt"
	"github.com/meschbach/npcs/competition/wire"
	"github.com/meschbach/npcs/junk/realnet"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func enginesAvailable(cmd *cobra.Command, args []string) {
	game, service := args[0], args[1]

	ctx := cmd.Context()
	net := realnet.NetworkedGRPC
	clientWire, err := net.Client(ctx, "127.0.0.1:11234", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Failed to connect because %s\n", err)
		return
	}
	o := wire.NewGameEngineOrchestrationClient(clientWire)
	result, err := o.EngineAvailable(ctx, &wire.EngineAvailableIn{
		ForGame:  game,
		StartURL: service,
	})
	if err != nil {
		fmt.Printf("Failed to register as available because %s\n", err)
		return
	}
	fmt.Printf("Game engine ID %s\n", result.GameID)
}

func enginesCommand() *cobra.Command {
	available := &cobra.Command{
		Use:   "available <game> <service-url>",
		Short: "Registered a game engine as available for play",
		Args:  cobra.ExactArgs(2),
		Run:   enginesAvailable,
	}

	root := &cobra.Command{
		Use:   "engines",
		Short: "operations against the engines registry",
	}
	root.AddCommand(available)

	return root
}
