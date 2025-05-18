package main

import (
	"fmt"
	"github.com/meschbach/npcs/competition/wire"
	"github.com/meschbach/npcs/junk/realnet"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func listGames(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	net := realnet.NetworkedGRPC
	clientWire, err := net.Client(ctx, "127.0.0.1:11234", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Failed to connect because %s\n", err)
		return
	}
	r := wire.NewGameRegistryClient(clientWire)
	out, err := r.ListRegisteredGames(ctx, &wire.ListRegisteredGamesIn{})
	if err != nil {
		fmt.Printf("Failed to list games because %s\n", err)
		return
	}
	if len(out.Games) == 0 {
		fmt.Println("No games registered")
	}
	for _, g := range out.Games {
		fmt.Printf("%s\t%t\t%s\n", g.Name, g.Active, g.Id)
	}
}

func registerGame(cmd *cobra.Command, args []string) {
	name := args[0]
	ctx := cmd.Context()
	net := realnet.NetworkedGRPC
	clientWire, err := net.Client(ctx, "127.0.0.1:11234", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Failed to connect because %s\n", err)
		return
	}
	r := wire.NewGameRegistryClient(clientWire)
	out, err := r.RegisterGame(ctx, &wire.RegisterGameIn{
		Name: name,
	})
	if err != nil {
		fmt.Printf("Failed to register: %s\n", err)
		return
	}
	fmt.Printf("Registered game %s with id %s\n", name, out.Id)
}

func gamesCommand() *cobra.Command {
	list := &cobra.Command{
		Use:   "list",
		Short: "Lists games in the registry",
		Run:   listGames,
	}

	register := &cobra.Command{
		Use:   "register <name>",
		Short: "Registers a game",
		Args:  cobra.ExactArgs(1),
		Run:   registerGame,
	}

	root := &cobra.Command{
		Use:   "games",
		Short: "Lists games in the registry",
	}
	root.AddCommand(list)
	root.AddCommand(register)

	return root
}
