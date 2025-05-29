package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/npcs/competition/simple"
	"github.com/meschbach/npcs/competition/wire"
	"github.com/meschbach/npcs/junk"
	"github.com/meschbach/npcs/junk/proc"
	junk2 "github.com/meschbach/npcs/junk/realnet"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"time"
)

type gameInstance struct {
	serviceAddress     string
	advertiseAddress   string
	competitionService string
	initSleepTime      time.Duration
	grpcListener       junk.Component
}

func (g *gameInstance) Start(setup context.Context, run context.Context) error {
	// ensure advertise address is set
	if g.advertiseAddress == "" {
		g.advertiseAddress = g.serviceAddress
	}

	// build service
	game := simple.NewGameService()

	// export service
	net := junk2.NetworkedGRPC
	if service, err := net.Listener(setup, g.serviceAddress, func(ctx context.Context, server *grpc.Server) error {
		simple.RegisterSimpleGameServer(server, game)
		return nil
	}); err != nil {
		return fmt.Errorf("failed to setup listener: %w", err)
	} else {
		g.grpcListener = service
	}
	if err := g.grpcListener.Start(run); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	//
	slog.InfoContext(setup, "init sleep", "init", g.initSleepTime.String())
	//todo: for some reason initSleepTime is always zero by this time; need to fix that
	//if g.initSleepTime > 0 {
	time.Sleep(1 * time.Second)
	//}

	// register with service
	slog.InfoContext(setup, "registering game")
	conn, err := net.Client(setup, g.competitionService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	r := wire.NewGameRegistryClient(conn)
	if _, err := r.RegisterGame(setup, &wire.RegisterGameIn{
		Name: "github.com/meschbach/npc/competition/simple/v0",
	}); err != nil {
		return fmt.Errorf("failed to register game: %w", err)
	}
	id, _, err := game.RunGameInstance()
	if err != nil {
		return err
	}

	i := wire.NewGameEngineOrchestrationClient(conn)
	slog.InfoContext(setup, "marking engine as available", "instanceID", id)
	if _, err := i.EngineAvailable(setup, &wire.EngineAvailableIn{
		ForGame:    "github.com/meschbach/npc/competition/simple/v0",
		StartURL:   g.advertiseAddress,
		InstanceID: id,
	}); err != nil {
		return fmt.Errorf("failed to mark engine as available: %w", err)
	}
	slog.InfoContext(setup, "simple game engine started")
	return nil
}

func (g *gameInstance) Stop(ctx context.Context) error {
	var problems []error
	if g.grpcListener != nil {
		problems = append(problems, g.grpcListener.Stop(ctx))
	}
	return errors.Join(problems...)
}

func runSimpleGameInstance(cmd *cobra.Command, args []string, i *gameInstance) {
	if err := proc.AsService(i); err != nil {
		fmt.Printf("Failed to run because %grpcListener\n", err)
		return
	}
}

func singleGameInstance() *cobra.Command {
	i := &gameInstance{}
	gameInstance := &cobra.Command{
		Use:   "game-instance",
		Short: "Runs a single simple game instance once",
		Run: func(cmd *cobra.Command, args []string) {
			runSimpleGameInstance(cmd, args, i)
		},
	}
	gameInstanceFlags := gameInstance.PersistentFlags()
	gameInstanceFlags.StringVarP(&i.serviceAddress, "service-address", "s", "127.0.0.1:11235", "address of game to listen too")
	gameInstanceFlags.StringVarP(&i.advertiseAddress, "advertise-address", "a", "", "address to advertise to competition service (uses service-address if not set)")
	gameInstanceFlags.StringVarP(&i.competitionService, "competition-service", "c", "127.0.0.1:11234", "address of the competition service")
	//todo: really it would be better to retry but concentrated on delivery first
	gameInstanceFlags.DurationVarP(&i.initSleepTime, "init-sleep-time", "t", 2*time.Second, "time to sleep before registering the game")

	return gameInstance
}
