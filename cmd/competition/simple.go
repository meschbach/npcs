package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/npcs/competition"
	"github.com/meschbach/npcs/competition/simple"
	"github.com/meschbach/npcs/competition/wire"
	"github.com/meschbach/npcs/junk/proc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
)

type gameInstance struct {
	s competition.Component
}

func (g *gameInstance) Start(setup context.Context, run context.Context) error {
	// build service
	game := simple.NewGameService("")

	// export service
	net := competition.NetworkedGRPC
	if service, err := net.Listener(setup, "127.0.0.1:11235", func(ctx context.Context, server *grpc.Server) error {
		simple.RegisterSimpleGameServer(server, game)
		return nil
	}); err != nil {
		return err
	} else {
		g.s = service
	}
	if err := g.s.Start(run); err != nil {
		return err
	}
	// register with service
	slog.InfoContext(setup, "registering game")
	conn, err := net.Client(setup, "127.0.0.1:11234", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	r := wire.NewGameRegistryClient(conn)
	if _, err := r.RegisterGame(setup, &wire.RegisterGameIn{
		Name: "github.com/meschbach/npc/competition/simple/v0",
	}); err != nil {
		return err
	}
	i := wire.NewGameEngineOrchestrationClient(conn)
	if _, err := i.EngineAvailable(setup, &wire.EngineAvailableIn{
		ForGame:  "github.com/meschbach/npc/competition/simple/v0",
		StartURL: "127.0.0.1:11235",
	}); err != nil {
		return err
	}
	slog.InfoContext(setup, "simple game engine started")
	return nil
}

func (g *gameInstance) Stop(ctx context.Context) error {
	var problems []error
	if g.s != nil {
		problems = append(problems, g.s.Stop(ctx))
	}
	return errors.Join(problems...)
}

func runSimpleGameInstance(cmd *cobra.Command, args []string) {
	if err := proc.AsService(&gameInstance{}); err != nil {
		fmt.Printf("Failed to run because %s\n", err)
		return
	}
}

func simpleGamesCommands() *cobra.Command {
	list := &cobra.Command{
		Use:   "game-instance",
		Short: "Runs a single simple game instance once",
		Run:   runSimpleGameInstance,
	}

	root := &cobra.Command{
		Use:   "simple-game",
		Short: "Testable game service",
	}
	root.AddCommand(list)

	return root
}
