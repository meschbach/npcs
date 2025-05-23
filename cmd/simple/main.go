package main

import (
	"context"
	"fmt"
	"github.com/meschbach/npcs/competition/simple"
	"github.com/meschbach/npcs/competition/wire"
	"github.com/meschbach/npcs/junk/proc/tproc"
	"github.com/meschbach/npcs/junk/realnet"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"os"
)

type playerInstance struct {
	competitionAddress string
	playerID           string
}

func (p *playerInstance) run(ctx context.Context) error {
	net := realnet.NetworkedGRPC
	competitionClientWire, err := net.Client(ctx, p.competitionAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer competitionClientWire.Close()
	competitionClient := wire.NewCompetitionV1Client(competitionClientWire)
	matchOut, err := competitionClient.QuickMatch(ctx, &wire.QuickMatchIn{
		PlayerName: p.playerID,
		Game:       "github.com/meschbach/npc/competition/simple/v0",
	})
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "Matched URL", "matchURL", matchOut.MatchURL)

	gameClient, err := net.Client(ctx, matchOut.MatchURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer gameClient.Close()
	simpleGameClient := simple.NewSimpleGameClient(gameClient)
	slog.InfoContext(ctx, "Connecting to game", "game.id", matchOut.UUID)
	result, err := simpleGameClient.Joined(ctx, &simple.JoinedIn{})
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "Connected to game", "result", result)
	return nil
}

func main() {
	i := &playerInstance{}
	root := &cobra.Command{
		Use:   "simple",
		Short: "Simple game client",
		Run: func(cmd *cobra.Command, args []string) {
			if err := tproc.RunOnce("simple-game-client", i.run); err != nil {
				slog.Error("tproc.AsService failed", "error", err)
			}
		},
	}
	flags := root.PersistentFlags()
	flags.StringVarP(&i.competitionAddress, "competition-address", "c", "127.0.0.1:11234", "address of the competition service")
	flags.StringVarP(&i.playerID, "player-id", "p", "test-1234", "player ID to use")

	if err := root.Execute(); err != nil {
		if _, err := fmt.Fprintf(os.Stderr, "error: %v\n", err); err != nil {
			panic(err)
		}
	}
}
