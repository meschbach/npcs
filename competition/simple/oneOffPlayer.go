package simple

import (
	"context"
	"github.com/meschbach/npcs/competition/wire"
	"github.com/meschbach/npcs/junk/realnet"
	"google.golang.org/grpc"
	"log/slog"
)

type RunOncePlayerOpt func(*RunOncePlayer)

type RunOncePlayer struct {
	net realnet.GRPCNetwork

	matcherAddress string
	matcherOptions []grpc.DialOption
	gameOptions    []grpc.DialOption
}

func WithPlayerMatcherOption(opts ...grpc.DialOption) RunOncePlayerOpt {
	return func(r *RunOncePlayer) {
		r.matcherOptions = append(r.matcherOptions, opts...)
		r.gameOptions = append(r.gameOptions, opts...)
	}
}

func WithPlayerMatcherAddress(address string) RunOncePlayerOpt {
	return func(player *RunOncePlayer) {
		player.matcherAddress = address
	}
}

func WithPlayerNetwork(net realnet.GRPCNetwork) RunOncePlayerOpt {
	return func(once *RunOncePlayer) {
		once.net = net
	}
}

func NewRunOnce(opts ...RunOncePlayerOpt) *RunOncePlayer {
	r := &RunOncePlayer{}
	for _, o := range opts {
		o(r)
	}
	if r.net == nil {
		r.net = realnet.NetworkedGRPC
	}
	return r
}

func (r *RunOncePlayer) Run(ctx context.Context) error {
	competitionClientWire, err := r.net.Client(ctx, r.matcherAddress, r.matcherOptions...)
	if err != nil {
		return err
	}
	defer competitionClientWire.Close()
	competitionClient := wire.NewCompetitionV1Client(competitionClientWire)
	matchOut, err := competitionClient.QuickMatch(ctx, &wire.QuickMatchIn{
		PlayerName: "test-1234",
		Game:       "github.com/meschbach/npc/competition/simple/v0",
	})
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "Matched URL", "matchURL", matchOut.MatchURL)

	gameClient, err := r.net.Client(ctx, matchOut.MatchURL, r.matcherOptions...)
	if err != nil {
		return err
	}
	defer gameClient.Close()
	simpleGameClient := NewSimpleGameClient(gameClient)
	slog.InfoContext(ctx, "Connecting to game", "game.id", matchOut.UUID)
	result, err := simpleGameClient.Joined(ctx, &JoinedIn{})
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "Connected to game", "result", result)
	return nil
}
